package polygon

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	assetpb "github.com/eolymp/go-sdk/eolymp/asset"
	atlaspb "github.com/eolymp/go-sdk/eolymp/atlas"
	ecmpb "github.com/eolymp/go-sdk/eolymp/ecm"
	executorpb "github.com/eolymp/go-sdk/eolymp/executor"
	"github.com/google/uuid"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const objectChunkSize = 5242880

var imageFinder = regexp.MustCompile("(\\\\includegraphics.*?{)(.+?)(})")

type ProblemLoader struct {
	assets assetUploader
	blobs  blobUploader
	log    logger
}

func NewProblemLoader(assets assetUploader, blobs blobUploader, log logger) *ProblemLoader {
	return &ProblemLoader{
		assets: assets,
		blobs:  blobs,
		log:    log,
	}
}

// Fetch downloads, parses and normalizes problem for it to be imported into Eolymp database.
//
// The link must be a valid url with following parameters:
//   - schema=polygon
//   - username=api-key
//   - password=api-secret
//   - host, path and port can be omitted
//
// An example of a link: polygon://api-key:api-secret@/?problemId=123
func (p *ProblemLoader) Fetch(ctx context.Context, link string) (*atlaspb.Snapshot, error) {
	// create workspace
	path := filepath.Join(os.TempDir(), uuid.New().String())
	if err := os.Mkdir(path, 0777); err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	defer p.cleanup(path)

	// download and unpack
	if err := p.download(ctx, path, link); err != nil {
		return nil, fmt.Errorf("unable to download problem archive: %w", err)
	}

	if err := p.unpack(ctx, path); err != nil {
		return nil, fmt.Errorf("unable to unpack problem archive: %w", err)
	}

	return p.Snapshot(ctx, path)
}

func (p *ProblemLoader) Snapshot(ctx context.Context, path string) (*atlaspb.Snapshot, error) {
	file, err := os.Open(filepath.Join(path, "problem.xml"))
	if err != nil {
		return nil, fmt.Errorf("unable to open problem.xml: %w", err)
	}

	defer file.Close()

	spec := &Specification{}

	if err := xml.NewDecoder(file).Decode(spec); err != nil {
		return nil, fmt.Errorf("unable to decode problem.xml: %w", err)
	}

	// import...
	checker, err := p.checker(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read checker configuration: %w", err)
	}

	interactor, err := p.interactor(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read interactor configuration: %w", err)
	}

	statements, err := p.statements(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read statements: %w", err)
	}

	templates, err := p.templates(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read templates: %w", err)
	}

	attachments, err := p.attachments(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read attachments (materials): %w", err)
	}

	testsets, tests, err := p.testing(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read tests: %w", err)
	}

	editorials, err := p.editorials(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read tutorials: %w", err)
	}

	solutions, err := p.solutions(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read solutions: %w", err)
	}

	scripts, err := p.scripts(ctx, path, spec)
	if err != nil {
		return nil, fmt.Errorf("unable to read solutions: %w", err)
	}

	runs := uint32(spec.Judging.RunCount)
	if runs <= 0 {
		runs = 1
	}

	return &atlaspb.Snapshot{
		Problem:     &atlaspb.Problem{Topics: TopicsFromTags(spec.Tags)},
		Testing:     &atlaspb.TestingConfig{RunCount: runs},
		Checker:     checker,
		Interactor:  interactor,
		Statements:  statements,
		Templates:   templates,
		Attachments: attachments,
		Testsets:    testsets,
		Tests:       tests,
		Editorials:  editorials,
		Solutions:   solutions,
		Scripts:     scripts,
	}, nil
}

// download problem archive and save it locally for parsing
func (p *ProblemLoader) download(ctx context.Context, path string, link string) error {
	origin, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("invalid problem origin: %w", err)
	}

	switch {
	case origin.Scheme == "polygon":
		pid, err := strconv.ParseInt(origin.Query().Get("problemId"), 10, 32)
		if err != nil {
			return errors.New("invalid problem origin: query parameter problemId must be a valid integer")
		}

		secret, _ := origin.User.Password()
		poly := New(origin.User.Username(), secret)

		return p.downloadByID(ctx, path, poly, int(pid))
	case origin.Scheme == "https" && origin.Hostname() == "polygon.codeforces.com" &&
		origin.Port() == "":

		return p.downloadByLink(ctx, path, origin)
	default:
		return fmt.Errorf("invalid problem origin: schema %#v is not supported", origin.Scheme)
	}
}

func (p *ProblemLoader) downloadByLink(ctx context.Context, path string, link *url.URL) error {
	username := link.User.Username()
	password, _ := link.User.Password()

	link.User = nil

	var pkgType []string
	if !link.Query().Has("type") {
		pkgType = []string{"windows"}
	} else if v := link.Query().Get("type"); v != "" {
		pkgType = []string{v}
	}

	query := url.Values{"login": {username}, "password": {password}, "type": pkgType}

	req, err := http.NewRequest(http.MethodPost, link.String(), strings.NewReader(query.Encode()))
	if err != nil {
		return fmt.Errorf("unable to compose HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("HTTP request has failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("problem link %#v leads to a file which does not exist", link.String())
	}

	kind, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("unable to read response content-type: %w", err)
	}

	if kind != "application/zip" {
		return fmt.Errorf("problem link %#v does not seem to lead to problem archive (check link and credentials)", link.String())
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("problem link %#v requires valid credentials", link.String())
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("problem link %#v is invalid: server response code is %v", link.String(), resp.StatusCode)
	}

	file, err := os.Create(filepath.Join(path, "problem.zip"))
	if err != nil {
		return fmt.Errorf("unable to create problem archieve: %w", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("unable to write problem archieve: %w", err)
	}

	return nil
}

func (p *ProblemLoader) downloadByID(ctx context.Context, path string, poly *Client, id int) error {
	pack, err := p.pickPackage(ctx, poly, id)
	if err != nil {
		return fmt.Errorf("unable to find package: %w", err)
	}

	src, err := poly.DownloadPackage(ctx, DownloadPackageInput{
		ProblemID: id,
		PackageID: pack.ID,
		Type:      pack.Type,
	})

	if err != nil {
		return fmt.Errorf("unable to download package: %w", err)
	}

	defer src.Close()

	dst, err := os.Create(filepath.Join(path, "problem.zip"))
	if err != nil {
		return fmt.Errorf("unable to create problem archieve: %w", err)
	}

	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("unable to save problem archive locally: %w", err)
	}

	return nil
}

// pickPackage to download, it has to be in the right status, and it has to be windows, so we can use generated tests
func (p *ProblemLoader) pickPackage(ctx context.Context, poly *Client, problem int) (*Package, error) {
	packages, err := poly.ListPackages(ctx, ListPackagesInput{ProblemID: problem})
	if err != nil {
		return nil, err
	}

	for _, pack := range packages {
		if pack.Type == "windows" && pack.State == "READY" {
			return &pack, nil
		}
	}

	return nil, errors.New("no suitable packages")
}

// unpack problem archive
func (p *ProblemLoader) unpack(ctx context.Context, path string) error {
	reader, err := zip.OpenReader(filepath.Join(path, "problem.zip"))
	if err != nil {
		return err
	}

	defer reader.Close()

	for _, file := range reader.File {
		file := file

		err := func() error {
			// sanitize file path
			name := strings.TrimPrefix(filepath.Clean(filepath.Join("/", file.Name)), string([]rune{filepath.Separator}))
			fpath := filepath.Join(path, name)

			if file.FileInfo().IsDir() {
				if err := os.MkdirAll(fpath, 0777); err != nil && !os.IsExist(err) {
					return fmt.Errorf("unable to create folder %#v: %w", name, err)
				}

				return nil
			}

			if err := os.MkdirAll(filepath.Dir(fpath), 0777); err != nil && !os.IsExist(err) {
				return fmt.Errorf("unable to create folder %#v: %w", filepath.Dir(name), err)
			}

			sf, err := file.Open()
			if err != nil {
				return fmt.Errorf("unable to open %#v for reading: %w", name, err)
			}

			defer sf.Close()

			df, err := os.Create(fpath)
			if err != nil {
				return fmt.Errorf("unable to open %#v for writing: %w", name, err)
			}

			defer df.Close()

			if _, err = io.Copy(df, sf); err != nil {
				return fmt.Errorf("unable to write %#v: %w", name, err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

// cleanup after import
func (p *ProblemLoader) cleanup(path string) {
	if err := os.RemoveAll(path); err != nil {
		p.log.Warning("Unable to cleanup workspace path", map[string]any{"error": err, "path": path})
	}
}

func (p *ProblemLoader) checker(ctx context.Context, path string, spec *Specification) (*executorpb.Checker, error) {
	switch spec.Checker.Name {
	case "std::ncmp.cpp": // Single or more int64, ignores whitespaces
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 0, CaseSensitive: true}, nil
	case "std::rcmp4.cpp": // Single or more double, max any error 1E-4
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 4, CaseSensitive: true}, nil
	case "std::rcmp6.cpp": // Single or more double, max any error 1E-6
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 6, CaseSensitive: true}, nil
	case "std::rcmp9.cpp": // Single or more double, max any error 1E-9
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 9, CaseSensitive: true}, nil
	case "std::wcmp.cpp": // Sequence of tokens
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 0, CaseSensitive: true}, nil
	case "std::nyesno.cpp", // Zero or more yes/no, case-insensitive
		"std::yesno.cpp": // Single yes or no, case-insensitive
		return &executorpb.Checker{Type: executorpb.Checker_TOKENS, Precision: 0, CaseSensitive: false}, nil
	case "std::fcmp.cpp", // Lines, doesn't ignore whitespaces
		"std::hcmp.cpp", // Single huge integer
		"std::lcmp.cpp": // Lines, ignores whitespaces
		return &executorpb.Checker{Type: executorpb.Checker_LINES}, nil
	default:
		for lang, types := range runtimeMapping {
			source, ok := SourceByType(spec.Checker.Sources, types...)
			if !ok {
				continue
			}

			data, err := os.ReadFile(filepath.Join(path, source.Path))
			if err != nil {
				return nil, err
			}

			return &executorpb.Checker{Type: executorpb.Checker_PROGRAM, Lang: lang, Source: string(data)}, nil
		}
	}

	return nil, fmt.Errorf("checker \"%s\" not supported", spec.Checker.Name)
}

func (p *ProblemLoader) interactor(ctx context.Context, path string, spec *Specification) (*executorpb.Interactor, error) {
	if len(spec.Interactor.Sources) == 0 {
		return nil, nil
	}

	for lang, types := range runtimeMapping {
		source, ok := SourceByType(spec.Interactor.Sources, types...)
		if !ok {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, source.Path))
		if err != nil {
			return nil, err
		}

		return &executorpb.Interactor{Type: executorpb.Interactor_PROGRAM, Lang: lang, Source: string(data)}, nil
	}

	return nil, errors.New("interactor is not supported")
}

func (p *ProblemLoader) statements(ctx context.Context, path string, spec *Specification) (statements []*atlaspb.Statement, err error) {
	for _, statement := range spec.Statements {
		if statement.Type != "application/x-tex" {
			continue
		}

		locale, err := LocaleFromLanguage(statement.Language)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, filepath.Dir(statement.Path), "problem-properties.json"))
		if err != nil {
			return nil, fmt.Errorf("unable to read problem-properties.json: %w", err)
		}

		props := ProblemProperties{}

		if err := json.Unmarshal(data, &props); err != nil {
			return nil, fmt.Errorf("unable to unmrashal problem-properties.json: %w", err)
		}

		parts := []string{props.Legend}
		if props.Input != "" {
			parts = append(parts, fmt.Sprintf("\\InputFile\n\n%v", props.Input))
		}

		if props.Interaction != "" {
			parts = append(parts, fmt.Sprintf("\\Interaction\n\n%v", props.Interaction))
		}

		if props.Output != "" {
			parts = append(parts, fmt.Sprintf("\\OutputFile\n\n%v", props.Output))
		}

		if props.Notes != "" {
			parts = append(parts, fmt.Sprintf("\\Note\n\n%v", props.Notes))
		}

		if props.Scoring != "" {
			parts = append(parts, fmt.Sprintf("\\Scoring\n\n%v", props.Scoring))
		}

		if len(spec.Interactor.Sources) > 0 {
			tests, _ := findTestFiles(filepath.Join(path, filepath.Dir(statement.Path)))

			examples := "\\Examples\n\n"
			for _, test := range tests {
				input, err := os.ReadFile(test.input)
				if err != nil {
					return nil, err
				}

				output, err := os.ReadFile(test.output)
				if err != nil {
					return nil, err
				}

				examples += "\\exmp{" + string(input) + "}{" + string(output) + "\n}\n"
			}

			parts = append(parts, examples)
		}

		latex := strings.Join(parts, "\n\n")
		latex = p.uploadImagesFromLatex(ctx, filepath.Join(path, filepath.Dir(statement.Path)), latex)

		statements = append(statements, &atlaspb.Statement{
			Locale:  locale,
			Title:   props.Name,
			Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: latex}},
			Author:  props.AuthorName,
		})
	}

	return statements, nil
}

func (p *ProblemLoader) editorials(ctx context.Context, path string, spec *Specification) (editorials []*atlaspb.Editorial, err error) {
	for _, tutorial := range spec.Tutorials {
		if tutorial.Type != "application/x-tex" {
			continue
		}

		locale, err := LocaleFromLanguage(tutorial.Language)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, tutorial.Path))
		if err != nil {
			return nil, fmt.Errorf("unable to read problem-properties.json: %w", err)
		}

		latex := p.uploadImagesFromLatex(ctx, filepath.Join(path, filepath.Dir(tutorial.Path)), string(data))

		editorials = append(editorials, &atlaspb.Editorial{
			Locale:  locale,
			Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: latex}},
		})
	}

	return editorials, nil
}

func (p *ProblemLoader) solutions(ctx context.Context, path string, spec *Specification) (solutions []*atlaspb.Solution, err error) {
	for _, solution := range spec.Solutions {
		runtime, ok := SourceTypeToRuntime(solution.Source.Type)
		if !ok {
			continue
		}

		kind := atlaspb.Solution_UNSET
		switch solution.Tag {
		case "main", "accepted":
			kind = atlaspb.Solution_CORRECT
		case "rejected":
			kind = atlaspb.Solution_INCORRECT
		case "wrong-answer":
			kind = atlaspb.Solution_WRONG_ANSWER
		case "time-limit-exceeded":
			kind = atlaspb.Solution_TIMEOUT
		case "time-limit-exceeded-or-accepted":
			kind = atlaspb.Solution_TIMEOUT_OR_ACCEPTED
		case "memory-limit-exceeded":
			kind = atlaspb.Solution_OVERFLOW
		case "failed":
			kind = atlaspb.Solution_FAILURE
		case "time-limit-exceeded-or-memory-limit-exceeded", "presentation-error":
			kind = atlaspb.Solution_DONT_RUN
		default:
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, solution.Source.Path))
		if err != nil {
			return nil, fmt.Errorf("unable to read solution source %#v: %w", solution.Source.Path, err)
		}

		solutions = append(solutions, &atlaspb.Solution{
			Name:    filepath.Base(solution.Source.Path),
			Runtime: runtime,
			Source:  string(data),
			Type:    kind,
		})
	}

	return solutions, nil
}

func (p *ProblemLoader) scripts(ctx context.Context, path string, spec *Specification) (scripts []*atlaspb.Script, err error) {
	for _, script := range spec.Templates {
		runtime, ok := SourceTypeToRuntime(script.Source.Type)
		if !ok {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, script.Source.Path))
		if err != nil {
			return nil, fmt.Errorf("unable to read script source %#v: %w", script.Source.Path, err)
		}

		scripts = append(scripts, &atlaspb.Script{
			Name:    strings.TrimSuffix(filepath.Base(script.Source.Path), filepath.Ext(script.Source.Path)),
			Runtime: runtime,
			Source:  string(data),
		})
	}

	for _, solution := range spec.Solutions {
		if solution.Tag != "main" {
			continue
		}

		runtime, ok := SourceTypeToRuntime(solution.Source.Type)
		if !ok {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, solution.Source.Path))
		if err != nil {
			return nil, fmt.Errorf("unable to read solution script source %#v: %w", solution.Source.Path, err)
		}

		scripts = append(scripts, &atlaspb.Script{
			Name:    "solution",
			Runtime: runtime,
			Source:  string(data),
		})
	}

	return scripts, nil
}

// todo: add grader to the templates
func (p *ProblemLoader) templates(ctx context.Context, path string, spec *Specification) (templates []*atlaspb.Template, err error) {
	languages := map[string][]string{
		"files/template_cpp.cpp":   {"gpp", "cpp:17-gnu10", "cpp:20-gnu10"},
		"files/template_java.java": {"java"},
		"files/template_pas.pas":   {"fpc"},
		"files/template_py.py":     {"pypy", "python"},
	}

	for _, file := range spec.Templates {
		name := file.Source.Path

		list, ok := languages[name]
		if !ok {
			continue
		}

		for _, lang := range list {
			source, err := os.ReadFile(filepath.Join(path, name))
			if err != nil {
				return nil, err
			}

			templates = append(templates, &atlaspb.Template{
				Runtime: lang,
				Source:  string(source),
			})
		}
	}

	return
}

func (p *ProblemLoader) attachments(ctx context.Context, path string, spec *Specification) (attachments []*atlaspb.Attachment, err error) {
	for _, material := range spec.Materials {
		if material.Publish != "with-statement" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, material.Path))
		if err != nil {
			return nil, fmt.Errorf("unable to read attachment (material): %w", err)
		}

		name := filepath.Base(material.Path)

		asset, err := p.assets.UploadFile(ctx, &assetpb.UploadFileInput{Name: name, Data: data})
		if err != nil {
			return nil, fmt.Errorf("unable to upload attachment (material): %w", err)
		}

		attachments = append(attachments, &atlaspb.Attachment{Name: name, Link: asset.GetFileUrl()})
	}

	return
}

func (p *ProblemLoader) testing(ctx context.Context, path string, spec *Specification) (testsets []*atlaspb.Testset, tests []*atlaspb.Test, err error) {
	// don't bother if there are no tests
	if len(spec.Judging.Testsets) < 0 {
		return
	}

	// pick testset called "tests" or first one
	polyset := p.pickTestset(spec)

	// eolymp specific overrides
	blockMin := false
	timeLimit := polyset.TimeLimit
	memLimit := polyset.MemoryLimit

	for _, tag := range spec.Tags {
		switch {
		case tag.Value == "block_min" || tag.Value == "min_block":
			blockMin = true
		case strings.HasPrefix(tag.Value, "eolymp_tl="):
			if val, err := strconv.Atoi(tag.Value[10:]); err == nil {
				timeLimit = val
			}

		case strings.HasPrefix(tag.Value, "eolymp_ml="):
			if val, err := strconv.Atoi(tag.Value[10:]); err == nil {
				memLimit = val
			}
		}
	}

	groupByName := map[string]SpecificationGroup{}
	for _, group := range polyset.Groups {
		groupByName[group.Name] = group
	}

	testsetIndexByGroup := p.mapGroupToIndex(polyset)
	testsetByGroup := map[string]*atlaspb.Testset{}

	// read testsets
	for name, index := range testsetIndexByGroup {
		testset := &atlaspb.Testset{
			Id:             uuid.New().String(),
			Index:          index,
			CpuLimit:       uint32(timeLimit),
			MemoryLimit:    uint64(memLimit),
			FileSizeLimit:  536870912,
			ScoringMode:    atlaspb.ScoringMode_ALL, // assume the problem is ICPC and uses typical ICPC feedback
			FeedbackPolicy: atlaspb.FeedbackPolicy_ICPC_EXPANDED,
		}

		// normally group with index 0 is samples
		if index == 0 {
			testset.ScoringMode = atlaspb.ScoringMode_EACH
			testset.FeedbackPolicy = atlaspb.FeedbackPolicy_COMPLETE
		}

		// check if group is defined and inherit any parameters
		if group, ok := groupByName[name]; ok {
			testset.ScoringMode = atlaspb.ScoringMode_EACH

			if group.PointsPolicy == "complete-group" {
				testset.ScoringMode = atlaspb.ScoringMode_ALL
			}

			if blockMin && index != 0 {
				testset.ScoringMode = atlaspb.ScoringMode_WORST
			}

			testset.FeedbackPolicy = atlaspb.FeedbackPolicy_COMPLETE
			if group.FeedbackPolicy == "icpc" || group.FeedbackPolicy == "points" || group.FeedbackPolicy == "none" {
				testset.FeedbackPolicy = atlaspb.FeedbackPolicy_ICPC
			} else if group.FeedbackPolicy == "icpc-expanded" {
				testset.FeedbackPolicy = atlaspb.FeedbackPolicy_ICPC_EXPANDED
			}

			for _, dep := range group.Dependencies {
				testset.Dependencies = append(testset.Dependencies, testsetIndexByGroup[dep.Group])
			}
		}

		testsetByGroup[name] = testset
		testsets = append(testsets, testset)
	}

	// read tests
	var total float32
	for index, polytest := range polyset.Tests {
		testset, ok := testsetByGroup[polytest.Group]
		if !ok {
			continue
		}

		test := &atlaspb.Test{
			TestsetId: testset.GetId(),
			Index:     int32(index),
			Example:   polytest.Sample,
			Score:     polytest.Points,
		}

		// make input
		input := filepath.Join(path, fmt.Sprintf(polyset.InputPathPattern, index+1))
		if polytest.Method == "generated" && !fileExists(input) {
			command := strings.Split(polytest.Command, " ")
			test.Input = &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: command[0], Arguments: command[1:]}}
		} else {
			link, err := p.uploadObject(ctx, input)
			if err != nil {
				return nil, nil, err
			}

			test.Input = &atlaspb.Test_InputUrl{InputUrl: link}
		}

		// make answer
		answer := filepath.Join(path, fmt.Sprintf(polyset.AnswerPathPattern, index+1))
		if polytest.Method == "generated" && !fileExists(answer) {
			test.Answer = &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}
		} else {
			link, err := p.uploadObject(ctx, answer)
			if err != nil {
				return nil, nil, err
			}

			test.Answer = &atlaspb.Test_AnswerUrl{AnswerUrl: link}
		}

		// add test to the list
		tests = append(tests, test)
		total += test.GetScore()
	}

	// set points evenly if total is 0
	if total == 0 {
		var credit float64 = 100
		for i, test := range tests {
			test.Score = float32(math.Min(math.Floor(credit/float64(len(tests)-i)), credit))
			credit -= float64(test.Score)
		}
	}

	return
}

// uploadImagesFromLatex finds images in text, uploads them and replaces original names with links.
// e.g. \includegraphics[width=12cm]{myimage.png} -> \includegraphics[width=12cm]{https://...}
func (p *ProblemLoader) uploadImagesFromLatex(ctx context.Context, path, text string) string {
	images := imageFinder.FindAllStringSubmatch(text, -1)

	replaced := map[string]bool{}
	for _, image := range images {
		if want, got := 4, len(image); want != got {
			p.log.Warning("Captured unexpected number of groups", map[string]any{"want": want, "got": got})
		}

		full := image[0]
		prefix := image[1]
		name := image[2]
		suffix := image[3]

		if _, ok := replaced[full]; ok {
			continue
		}

		data, err := os.ReadFile(filepath.Join(path, name))
		if err != nil {
			p.log.Warning("unable to read image", map[string]any{"error": err.Error()})
			continue
		}

		asset, err := p.assets.UploadImage(ctx, &assetpb.UploadImageInput{Name: name, Data: data})
		if err != nil {
			p.log.Warning("unable to upload image", map[string]any{"error": err.Error()})
			continue
		}

		text = strings.Replace(text, full, prefix+asset.GetImageUrl()+suffix, -1)
	}
	return text
}

// pickTestset find "main" testset for a problem
func (p *ProblemLoader) pickTestset(spec *Specification) SpecificationTestset {
	for _, set := range spec.Judging.Testsets {
		if strings.ToLower(set.Name) == "tests" {
			return set
		}
	}

	if len(spec.Judging.Testsets) > 0 {
		return spec.Judging.Testsets[0]
	}

	return SpecificationTestset{}
}

// mapGroupToIndex creates a map which allows to translate polygon's group names into eolymp's testset indexes.
// Eolymp uses testsets identified by a number, while polygon uses string names. This function creates name to index
// mapping to translate string names to numbers.
func (p *ProblemLoader) mapGroupToIndex(testset SpecificationTestset) map[string]uint32 {
	var names []string

	// collect group names defined as groups, or their dependencies
	for _, group := range testset.Groups {
		names = append(names, group.Name)

		for _, dep := range group.Dependencies {
			names = append(names, dep.Group)
		}
	}

	// collect groups defined in tests
	for _, test := range testset.Tests {
		names = append(names, test.Group)
	}

	// sort everything
	sort.Slice(names, func(i, j int) bool {
		firstValue, err1 := strconv.Atoi(names[i])
		secondValue, err2 := strconv.Atoi(names[j])
		if err1 == nil && err2 == nil {
			return firstValue < secondValue
		} else {
			return names[i] < names[j]
		}
	})

	// assign numbers starting from 1, except if group is called "sample"
	index := uint32(1)
	mapping := map[string]uint32{}

	for _, name := range names {
		if _, ok := mapping[name]; ok {
			continue
		}

		if strings.Contains(strings.ToLower(name), "sample") || name == "0" {
			mapping[name] = 0
			continue
		}

		mapping[name] = index
		index++
	}

	return mapping
}

// uploadObject to eolymp's blob storage, used to upload test data
func (p *ProblemLoader) uploadObject(ctx context.Context, path string) (string, error) {
	reader, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer reader.Close()

	upload, err := p.blobs.StartMultipartUpload(ctx, &assetpb.StartMultipartUploadInput{
		Name: filepath.Base(path),
		Type: "text/plain",
	})

	if err != nil {
		return "", err
	}

	var parts []*assetpb.CompleteMultipartUploadInput_Part

	chunk := make([]byte, objectChunkSize)

	for index := 1; ; index++ {
		size, err := reader.Read(chunk)
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		part, err := p.blobs.UploadPart(ctx, &assetpb.UploadPartInput{
			UploadId:   upload.GetUploadId(),
			PartNumber: uint32(index),
			Data:       chunk[0:size],
		})

		if err != nil {
			return "", err
		}

		parts = append(parts, &assetpb.CompleteMultipartUploadInput_Part{
			Number: uint32(index),
			Token:  part.GetToken(),
		})
	}

	if len(parts) == 0 {
		return "", nil
	}

	out, err := p.blobs.CompleteMultipartUpload(ctx, &assetpb.CompleteMultipartUploadInput{
		UploadId: upload.GetUploadId(),
		Parts:    parts,
	})

	if err != nil {
		return "", err
	}

	return out.GetAssetUrl(), nil
}
