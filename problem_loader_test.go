package polygon

import (
	"context"
	atlaspb "github.com/eolymp/go-sdk/eolymp/atlas"
	ecmpb "github.com/eolymp/go-sdk/eolymp/ecm"
	"net/url"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestProblemLoader_FetchViaID(t *testing.T) {
	if os.Getenv("POLYGON_API_KEY") == "" {
		t.Skip("This test requires polygon password in env variable POLYGON_API_KEY and POLYGON_API_SECRET")
	}

	ctx := context.Background()

	loader := NewProblemLoader(
		&assetMock{},
		&blobMock{},
		&loggerMock{t: t},
	)

	link := url.URL{
		Scheme:   "polygon",
		User:     url.UserPassword(os.Getenv("POLYGON_API_KEY"), os.Getenv("POLYGON_API_SECRET")),
		Path:     "/",
		RawQuery: "problemId=270574",
	}

	_, err := loader.Fetch(ctx, link.String())
	if err != nil {
		t.Fatal(err)
	}

	// todo: make some assertions
}

func TestProblemLoader_FetchViaLink(t *testing.T) {
	if os.Getenv("POLYGON_USERNAME") == "" {
		t.Skip("This test requires polygon password in env variable POLYGON_USERNAME and POLYGON_PASSWORD")
	}

	ctx := context.Background()

	loader := NewProblemLoader(
		&assetMock{},
		&blobMock{},
		&loggerMock{t: t},
	)

	link := url.URL{
		Scheme: "https",
		Host:   "polygon.codeforces.com",
		User:   url.UserPassword(os.Getenv("POLYGON_USERNAME"), os.Getenv("POLYGON_PASSWORD")),
		Path:   "/p8bWTsG/eolymp/example-a-plus-b-testdata",
	}

	_, err := loader.Fetch(ctx, link.String())
	if err != nil {
		t.Fatal(err)
	}

	// todo: make some assertions
}

func TestProblemLoader_Snapshot(t *testing.T) {
	ctx := context.Background()
	loader := NewProblemLoader(&assetMock{}, &blobMock{}, &loggerMock{t: t})

	t.Run("import topics", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/01-topics")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		want := []string{"mougogmuf10i3b5gpp7ur935l0", "pjjft5joql5j95u7radbchs51g"}
		got := snap.GetProblem().GetTopics()
		sort.Strings(got)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("Problem topics do not match:\n want %v\n  got %v", want, got)
		}
	})

	t.Run("import statements", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/02-statements")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetStatements()
		want := []*atlaspb.Statement{{
			Locale:  "uk",
			Title:   "Сума масиву",
			Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "Дано $n$ цілих чисел $a_1, a_2, \\ldots, a_n$. Знайдіть їхню суму.\n\n\\InputFile\n\nПерший рядок містить ціле число $n$ ($1 \\leq n \\leq 2 \\cdot 10^6$)~--- кількість чисел.\r\n\r\nДругий рядок містить $n$ цілих чисел $a_1, a_2, \\ldots, a_n$ ($0 \\leq a_i \\leq 10^9$)~--- числа масиву.\n\n\\OutputFile\n\nВиведіть одне число~--- суму масиву.\n\n\\Scoring\n\n\\begin{enumerate}\r\n\\item ($10$ балів): $n \\leq 1\\,000$, $a_i \\leq 1\\,000$;\r\n\\item ($10$ балів): $n \\leq 10\\,000$;\r\n\\item ($8$ балів): $n \\leq 200\\,000$;\r\n\\item ($8$ балів): $n \\leq 400\\,000$;\r\n\\item ($8$ балів): $n \\leq 600\\,000$;\r\n\\item ($8$ балів): $n \\leq 800\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,000\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,200\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,400\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,600\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,800\\,000$;\r\n\\item ($8$ балів): повні обмеження.\r\n\\end{enumerate}\r\n"}},
			Author:  "Anton Tsypko",
		}}

		if len(got) != len(want) {
			t.Fatalf("Number of solutions does not match: want %v, got %v", len(want), len(got))
		}

		for i := range want {
			// erase content if it matches to simplify error output
			if want[i].GetContent().GetLatex() == got[i].GetContent().GetLatex() {
				want[i].Content = nil
				got[i].Content = nil
			}

			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("Problem statements[%v] do not match:\n want %v\n  got %v", i, want[i], got[i])
			}
		}
	})

	t.Run("import test points from problem.xml", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/03-test-scoring-with-points")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		// total score should be 100
		var score float32
		for _, test := range snap.GetTests() {
			score += test.GetScore()
		}

		if score != 100 {
			t.Errorf("Problem score does not add up to 100, got %v instead", score)
		}

		want := []float32{0, 0, 0, 0, 10, 10, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4}
		var got []float32
		for _, test := range snap.GetTests() {
			got = append(got, test.GetScore())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("Scores in group 2 do not match:\n want: %v\n  got: %v\n", want, got)
		}
	})

	t.Run("set 100 points evenly if there are none in problem.xml", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/04-test-scoring-without-points")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		// total score should be 100
		var score float32
		for _, test := range snap.GetTests() {
			score += test.GetScore()
		}

		if score != 100 {
			t.Errorf("Problem score does not add up to 100, got %v instead", score)
		}

		want := []float32{33, 33, 34}
		var got []float32
		for _, test := range snap.GetTests() {
			got = append(got, test.GetScore())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("Scores in group 2 do not match:\n want: %v\n  got: %v\n", want, got)
		}
	})

	t.Run("import tutorials", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/05-tutorials")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetEditorials()
		want := []*atlaspb.Editorial{
			{Locale: "en", Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "\\begin{tutorial}{English}\r\nEnglish Editorial\r\n\\end{tutorial}\r\n"}}},
			{Locale: "uk", Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "\\begin{tutorial}{Ukrainian}\r\nUkrainian Editorial\r\n\\end{tutorial}\r\n"}}},
		}

		if len(got) != len(want) {
			t.Fatal("Number of tutorials and editorials does not match")
		}

		for i := range want {
			// erase content if it matches to simplify error output
			if want[i].GetContent().GetLatex() == got[i].GetContent().GetLatex() {
				want[i].Content = nil
				got[i].Content = nil
			}

			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("Problem editorials[%v] do not match:\n want %v\n  got %v", i, want[i], got[i])
			}
		}
	})

	t.Run("import solutions", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/06-solutions")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetSolutions()
		want := []*atlaspb.Solution{
			{Name: "main.cpp", Runtime: "cpp:17-gnu10", Source: "main.cpp content", Type: atlaspb.Solution_CORRECT},
			{Name: "rejected.cpp", Runtime: "cpp:17-gnu10", Source: "rejected.cpp content", Type: atlaspb.Solution_INCORRECT},
			{Name: "accepted.cpp", Runtime: "cpp:17-gnu10", Source: "accepted.cpp content", Type: atlaspb.Solution_CORRECT},
			{Name: "wrong-answer.cpp", Runtime: "cpp:17-gnu10", Source: "wrong-answer.cpp content", Type: atlaspb.Solution_WRONG_ANSWER},
			{Name: "time-limit-exceeded.cpp", Runtime: "cpp:17-gnu10", Source: "time-limit-exceeded.cpp content", Type: atlaspb.Solution_TIMEOUT},
			{Name: "time-limit-exceeded-or-accepted.cpp", Runtime: "cpp:17-gnu10", Source: "time-limit-exceeded-or-accepted.cpp content", Type: atlaspb.Solution_TIMEOUT_OR_ACCEPTED},
			{Name: "time-limit-exceeded-or-memory-limit-exceeded.cpp", Runtime: "cpp:17-gnu10", Source: "time-limit-exceeded-or-memory-limit-exceeded.cpp content", Type: atlaspb.Solution_DONT_RUN},
			{Name: "memory-limit-exceeded.cpp", Runtime: "cpp:17-gnu10", Source: "memory-limit-exceeded.cpp content", Type: atlaspb.Solution_OVERFLOW},
			{Name: "presentation-error.cpp", Runtime: "cpp:17-gnu10", Source: "presentation-error.cpp content", Type: atlaspb.Solution_DONT_RUN},
			{Name: "failed.cpp", Runtime: "cpp:17-gnu10", Source: "failed.cpp content", Type: atlaspb.Solution_FAILURE},
		}

		if len(got) != len(want) {
			t.Fatalf("Number of solutions does not match: want %v, got %v", len(want), len(got))
		}

		for i := range want {
			// erase content if it matches to simplify error output
			if want[i].GetSource() == got[i].GetSource() {
				want[i].Source = ""
				got[i].Source = ""
			}

			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("Problem solutions[%v] do not match:\n want %v\n  got %v", i, want[i], got[i])
			}
		}
	})

	t.Run("import problem with images", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/07-images-in-text")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetStatements()
		want := []*atlaspb.Statement{{
			Locale:  "uk",
			Title:   "Сума масиву",
			Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "Дано $n$ цілих чисел $a_1, a_2, \\ldots, a_n$. Знайдіть їхню суму. \\includegraphics[width=12cm]{https://eolympusercontent.com/image.png} \\includegraphics{https://eolympusercontent.com/image2.png} \n\n\\InputFile\n\nПерший рядок містить ціле число $n$ ($1 \\leq n \\leq 2 \\cdot 10^6$)~--- кількість чисел.\r\n\r\nДругий рядок містить $n$ цілих чисел $a_1, a_2, \\ldots, a_n$ ($0 \\leq a_i \\leq 10^9$)~--- числа масиву.\n\n\\OutputFile\n\nВиведіть одне число~--- суму масиву.\n\n\\Scoring\n\n\\begin{enumerate}\r\n\\item ($10$ балів): $n \\leq 1\\,000$, $a_i \\leq 1\\,000$;\r\n\\item ($10$ балів): $n \\leq 10\\,000$;\r\n\\item ($8$ балів): $n \\leq 200\\,000$;\r\n\\item ($8$ балів): $n \\leq 400\\,000$;\r\n\\item ($8$ балів): $n \\leq 600\\,000$;\r\n\\item ($8$ балів): $n \\leq 800\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,000\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,200\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,400\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,600\\,000$;\r\n\\item ($8$ балів): $n \\leq 1\\,800\\,000$;\r\n\\item ($8$ балів): повні обмеження.\r\n\\end{enumerate}\r\n"}},
			Author:  "Anton Tsypko",
		}}

		if len(got) != len(want) {
			t.Fatalf("Number of solutions does not match: want %v, got %v", len(want), len(got))
		}

		for i := range want {
			// erase content if it matches to simplify error output
			if want[i].GetContent().GetLatex() == got[i].GetContent().GetLatex() {
				want[i].Content = nil
				got[i].Content = nil
			}

			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("Problem statements[%v] do not match:\n want %v\n  got %v", i, want[i], got[i])
			}
		}
	})

	// use `eolymp_tl=` and `eolymp_ml=` tags to override time and memory limits
	t.Run("import problem with custom time and memory limits", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/08-custom-limit")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		testsets := snap.GetTestsets()
		if len(testsets) != 1 {
			t.Fatalf("There must be exactly one test set, got %v instead", len(testsets))
		}

		if want, got := uint32(750), testsets[0].CpuLimit; want != got {
			t.Errorf("Time limit override not applied: want %v, got %v", want, got)
		}

		if want, got := uint64(201326592), testsets[0].MemoryLimit; want != got {
			t.Errorf("Memory limit override not applied: want %v, got %v", want, got)
		}
	})

	t.Run("import editorial with images", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/09-images-in-tutorial")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetEditorials()
		want := []*atlaspb.Editorial{
			{Locale: "en", Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "\\begin{tutorial}{English}\r\nEnglish Editorial\r\n\\includegraphics[width=12cm]{https://eolympusercontent.com/image.png} \\includegraphics{https://eolympusercontent.com/image2.png}\r\n\\end{tutorial}\r\n"}}},
			{Locale: "uk", Content: &ecmpb.Content{Value: &ecmpb.Content_Latex{Latex: "\\begin{tutorial}{Ukrainian}\r\nUkrainian Editorial\r\n\\includegraphics[width=12cm]{https://eolympusercontent.com/image.png}\r\n\\includegraphics{https://eolympusercontent.com/image2.png}\r\n\\end{tutorial}\r\n"}}},
		}

		if len(got) != len(want) {
			t.Fatal("Number of tutorials and editorials does not match")
		}

		for i := range want {
			// erase content if it matches to simplify error output
			if want[i].GetContent().GetLatex() == got[i].GetContent().GetLatex() {
				want[i].Content = nil
				got[i].Content = nil
			}

			if !reflect.DeepEqual(want[i], got[i]) {
				t.Errorf("Problem editorials[%v] do not match:\n want %v\n  got %v", i, want[i], got[i])
			}
		}
	})

	t.Run("import run count", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/10-run-count")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		got := snap.GetTesting()
		want := &atlaspb.TestingConfig{RunCount: 11}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("Problem testing configuration do not match:\n want %v\n  got %v", want, got)
		}
	})

	// importing generated tests without actual files
	t.Run("import tests generator", func(t *testing.T) {
		got, err := loader.Snapshot(ctx, ".testdata/11-tests-generator")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		tid := got.GetTestsets()[0].GetId()

		want := &atlaspb.Snapshot{
			Scripts: []*atlaspb.Script{
				{Name: "gen", Runtime: "cpp:17-gnu10", Source: "#include \"testlib.h\"\n#include <iostream>\nusing ll = long long;\nusing namespace std;\n\nint main(int argc, char* argv[]) {\n    registerGen(argc, argv, 1);\n    cout << 12 << '\\n';\n    return 0;\n}\n"},
				{Name: "solution", Runtime: "cpp:17-gnu10", Source: "#include <bits/stdc++.h>\r\nusing namespace std;\r\n\r\nint32_t main() {\r\n    ios_base::sync_with_stdio(false);\r\n    cin.tie(nullptr);\r\n    cout.tie(nullptr);\r\n\r\n    return 0;\r\n}"},
			},
			Tests: []*atlaspb.Test{
				{TestsetId: tid, Index: 0, Score: 0, Example: true, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"5", "10", "20"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 1, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "10", "100"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 2, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10000"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 3, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10001"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 4, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10002"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 5, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10003"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 6, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10004"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 7, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10005"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 8, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10006"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
				{TestsetId: tid, Index: 9, Score: 4, Example: false, Input: &atlaspb.Test_InputGenerator{InputGenerator: &atlaspb.Test_Generator{ScriptName: "gen", Arguments: []string{"10", "100", "10007"}}}, Answer: &atlaspb.Test_AnswerGenerator{AnswerGenerator: &atlaspb.Test_Generator{ScriptName: "solution"}}},
			},
		}

		// verify scripts
		if len(got.GetScripts()) != len(want.GetScripts()) {
			t.Fatalf("Number of scripts does not match: want %v, got %v", len(want.GetScripts()), len(got.GetScripts()))
		}

		for i := range want.GetScripts() {
			if !reflect.DeepEqual(want.GetScripts()[i], got.GetScripts()[i]) {
				t.Errorf("Problem scripts[%v] do not match:\n want %v\n  got %v", i, want.GetScripts()[i], got.GetScripts()[i])
			}
		}

		// verify tests
		if len(got.GetTests()) != len(want.GetTests()) {
			t.Fatalf("Number of tests does not match: want %v, got %v", len(want.GetTests()), len(got.GetTests()))
		}

		for i := range want.GetTests() {
			if !reflect.DeepEqual(want.GetTests()[i], got.GetTests()[i]) {
				t.Errorf("Problem tests[%v] do not match:\n want %v\n  got %v", i, want.GetTests()[i], got.GetTests()[i])
			}
		}
	})

	// importing generated tests with pre-generated files (ie. full windows package)
	t.Run("import pre generated tests", func(t *testing.T) {
		got, err := loader.Snapshot(ctx, ".testdata/12-tests-generator-pregenerated")
		if err != nil {
			t.Fatal("Problem snapshot has failed:", err)
		}

		tid := got.GetTestsets()[0].GetId()

		want := &atlaspb.Snapshot{
			Scripts: []*atlaspb.Script{
				{Name: "gen", Runtime: "cpp:17-gnu10", Source: "#include \"testlib.h\"\n#include <iostream>\nusing ll = long long;\nusing namespace std;\n\nint main(int argc, char* argv[]) {\n    registerGen(argc, argv, 1);\n    cout << 12 << '\\n';\n    return 0;\n}\n"},
				{Name: "solution", Runtime: "cpp:17-gnu10", Source: "#include <bits/stdc++.h>\r\nusing namespace std;\r\n\r\nint32_t main() {\r\n    ios_base::sync_with_stdio(false);\r\n    cin.tie(nullptr);\r\n    cout.tie(nullptr);\r\n\r\n    return 0;\r\n}"},
			},
			Tests: []*atlaspb.Test{
				{TestsetId: tid, Index: 0, Score: 0, Example: true, Input: &atlaspb.Test_InputUrl{InputUrl: "01"}, Answer: &atlaspb.Test_AnswerUrl{AnswerUrl: "01.a"}},
				{TestsetId: tid, Index: 1, Score: 4, Example: false, Input: &atlaspb.Test_InputUrl{InputUrl: "02"}, Answer: &atlaspb.Test_AnswerUrl{AnswerUrl: "02.a"}},
				{TestsetId: tid, Index: 2, Score: 4, Example: false, Input: &atlaspb.Test_InputUrl{InputUrl: "03"}, Answer: &atlaspb.Test_AnswerUrl{AnswerUrl: "03.a"}},
			},
		}

		// verify scripts
		if len(got.GetScripts()) != len(want.GetScripts()) {
			t.Fatalf("Number of scripts does not match: want %v, got %v", len(want.GetScripts()), len(got.GetScripts()))
		}

		for i := range want.GetScripts() {
			if !reflect.DeepEqual(want.GetScripts()[i], got.GetScripts()[i]) {
				t.Errorf("Problem scripts[%v] do not match:\n want %v\n  got %v", i, want.GetScripts()[i], got.GetScripts()[i])
			}
		}

		// verify tests
		if len(got.GetTests()) != len(want.GetTests()) {
			t.Fatalf("Number of tests does not match: want %v, got %v", len(want.GetTests()), len(got.GetTests()))
		}

		for i := range want.GetTests() {
			if !reflect.DeepEqual(want.GetTests()[i], got.GetTests()[i]) {
				t.Errorf("Problem tests[%v] do not match:\n want %v\n  got %v", i, want.GetTests()[i], got.GetTests()[i])
			}
		}
	})
}
