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
			if want[i].GetContent().GetLatex() == want[i].GetContent().GetLatex() {
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
}
