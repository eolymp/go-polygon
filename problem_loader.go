package polygon

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	atlaspb "github.com/eolymp/go-sdk/eolymp/atlas"
	ecmpb "github.com/eolymp/go-sdk/eolymp/ecm"
	executorpb "github.com/eolymp/go-sdk/eolymp/executor"
	keeperpb "github.com/eolymp/go-sdk/eolymp/keeper"
	"github.com/eolymp/go-sdk/eolymp/typewriter"
	"github.com/google/uuid"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const objectChunkSize = 5242880

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

// FetchProblem downloads, parses and normalizes problem for it to be imported into Eolymp database.
//
// The link must be a valid url with following parameters:
//   - schema=polygon
//   - username=api-key
//   - password=api-secret
//   - host, path and port can be omitted
//
// An example of a link: polygon://api-key:api-secret@/?problemId=123
func (p *ProblemLoader) FetchProblem(ctx context.Context, link string) (*atlaspb.Snapshot, error) {
	origin, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid problem origin: %w", err)
	}

	if origin.Scheme != "polygon" {
		return nil, fmt.Errorf("invalid problem origin: scheme %#v is not supported", origin.Scheme)
	}

	pid, err := strconv.ParseInt(origin.Query().Get("problemId"), 10, 32)
	if err != nil {
		return nil, errors.New("invalid problem origin: query parameter problemId must be a valid integer")
	}

	secret, _ := origin.User.Password()
	poly := New(origin.User.Username(), secret)

	// create workspace
	path := filepath.Join(os.TempDir(), uuid.New().String())
	if err := os.Mkdir(path, 0777); err != nil {
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	defer p.cleanup(path)

	// download and unpack
	if err := p.download(ctx, path, poly, int(pid)); err != nil {
		return nil, fmt.Errorf("unable to download problem archive: %w", err)
	}

	if err := p.unpack(ctx, path); err != nil {
		return nil, fmt.Errorf("unable to unpack problem archive: %w", err)
	}

	// read problem.xml
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

	return &atlaspb.Snapshot{
		Problem:     &atlaspb.Problem{},
		Checker:     checker,
		Interactor:  interactor,
		Statements:  statements,
		Templates:   templates,
		Attachments: attachments,
		Testsets:    testsets,
		Tests:       tests,
	}, nil
}

// download problem archive and save it locally for parsing
func (p *ProblemLoader) download(ctx context.Context, path string, poly *Client, id int) error {
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

func (p *ProblemLoader) checker(ctx context.Context, path string, spec *Specification) (*executorpb.Verifier, error) {
	switch spec.Checker.Name {
	case "std::rcmp4.cpp", // Single or more double, max any error 1E-4
		"std::ncmp.cpp": // Single or more int64, ignores whitespaces
		return &executorpb.Verifier{Type: executorpb.Verifier_TOKENS, Precision: 4, CaseSensitive: true}, nil
	case "std::rcmp6.cpp": // Single or more double, max any error 1E-6
		return &executorpb.Verifier{Type: executorpb.Verifier_TOKENS, Precision: 6, CaseSensitive: true}, nil
	case "std::rcmp9.cpp": // Single or more double, max any error 1E-9
		return &executorpb.Verifier{Type: executorpb.Verifier_TOKENS, Precision: 9, CaseSensitive: true}, nil
	case "std::wcmp.cpp": // Sequence of tokens
		return &executorpb.Verifier{Type: executorpb.Verifier_TOKENS, Precision: 5, CaseSensitive: true}, nil
	case "std::nyesno.cpp", // Zero or more yes/no, case-insensitive
		"std::yesno.cpp": // Single yes or no, case-insensitive
		return &executorpb.Verifier{Type: executorpb.Verifier_TOKENS, Precision: 5, CaseSensitive: false}, nil
	case "std::fcmp.cpp", // Lines, doesn't ignore whitespaces
		"std::hcmp.cpp", // Single huge integer
		"std::lcmp.cpp": // Lines, ignores whitespaces
		return &executorpb.Verifier{Type: executorpb.Verifier_LINES}, nil
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

			return &executorpb.Verifier{Type: executorpb.Verifier_PROGRAM, Lang: lang, Source: string(data)}, nil
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

		locale, err := mapLanguageToLocale(statement.Language)
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

		statements = append(statements, &atlaspb.Statement{
			Locale:  locale,
			Title:   props.Name,
			Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: strings.Join(parts, "\n\n")}},
			Author:  props.AuthorName,
		})
	}

	return statements, nil
}

// todo: add grader to the templates
func (p *ProblemLoader) templates(ctx context.Context, path string, spec *Specification) (templates []*atlaspb.Template, err error) {
	languages := map[string][]string{
		"files/template_cpp.cpp":   {"gpp", "cpp:17-gnu10"},
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

		asset, err := p.assets.UploadAsset(ctx, &typewriter.UploadAssetInput{Filename: name, Data: data})
		if err != nil {
			return nil, fmt.Errorf("unable to upload attachment (material): %w", err)
		}

		attachments = append(attachments, &atlaspb.Attachment{Name: name, Link: asset.GetLink()})
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

	// no idea... ask Anton
	tags := p.tags(spec)
	blockMin := tags["block_min"] || tags["min_block"]

	groupByName := map[string]SpecificationGroup{}
	for _, group := range polyset.Groups {
		groupByName[group.Name] = group
	}

	testsetIndexByGroup := p.mapGroupToIndex(polyset)
	testsetIDByGroup := map[string]string{}

	// read testsets
	for name, index := range testsetIndexByGroup {
		testset := &atlaspb.Testset{
			Id:             uuid.New().String(),
			Index:          index,
			CpuLimit:       uint32(polyset.TimeLimit),
			MemoryLimit:    uint64(polyset.MemoryLimit),
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

		testsetIDByGroup[name] = testset.Id
		testsets = append(testsets, testset)
	}

	// read tests
	for index, polytest := range polyset.Tests {
		input, err := p.uploadObject(ctx, filepath.Join(path, fmt.Sprintf(polyset.InputPathPattern, index+1)))
		if err != nil {
			return nil, nil, err
		}

		answer, err := p.uploadObject(ctx, filepath.Join(path, fmt.Sprintf(polyset.AnswerPathPattern, index+1)))
		if err != nil {
			return nil, nil, err
		}

		tests = append(tests, &atlaspb.Test{
			TestsetId:      testsetIDByGroup[polytest.Group],
			Index:          int32(index),
			Example:        polytest.Sample,
			Score:          0,
			InputObjectId:  input,
			AnswerObjectId: answer,
		})
	}

	// todo: set points

	return
}

func (p *ProblemLoader) tags(spec *Specification) map[string]bool {
	tags := map[string]bool{}
	for _, tag := range spec.Tags {
		tags[tag.Value] = true
	}

	return tags
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
	sort.Strings(names)

	// assign numbers starting from 1, except if group is called "sample"
	index := uint32(1)
	mapping := map[string]uint32{}

	for _, name := range names {
		if _, ok := mapping[name]; ok {
			continue
		}

		if strings.Contains(strings.ToLower(name), "sample") {
			mapping[name] = 0
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

	upload, err := p.blobs.StartMultipartUpload(ctx, &keeperpb.StartMultipartUploadInput{})
	if err != nil {
		return "", err
	}

	var parts []*keeperpb.CompleteMultipartUploadInput_Part

	chunk := make([]byte, objectChunkSize)

	for index := 1; ; index++ {
		size, err := reader.Read(chunk)
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		part, err := p.blobs.UploadPart(ctx, &keeperpb.UploadPartInput{
			ObjectId:   upload.GetObjectId(),
			UploadId:   upload.GetUploadId(),
			PartNumber: uint32(index),
			Data:       chunk[0:size],
		})

		if err != nil {
			return "", err
		}

		parts = append(parts, &keeperpb.CompleteMultipartUploadInput_Part{
			Number: uint32(index),
			Etag:   part.GetEtag(),
		})
	}

	_, err = p.blobs.CompleteMultipartUpload(ctx, &keeperpb.CompleteMultipartUploadInput{
		ObjectId: upload.GetObjectId(),
		UploadId: upload.GetUploadId(),
		Parts:    parts,
	})

	if err != nil {
		return "", err
	}

	return upload.GetObjectId(), nil
}
