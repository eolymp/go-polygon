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

	t.Run("tag to topic mapping", func(t *testing.T) {
		snap, err := loader.Snapshot(ctx, ".testdata/01-tag-to-topic")
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

	t.Run("importing statements", func(t *testing.T) {
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

		if !reflect.DeepEqual(want, got) {
			// erase statements if they match to simplify output
			if len(want) != 0 && len(got) != 0 && want[0].GetContent().GetLatex() == got[0].GetContent().GetLatex() {
				want[0].Content = nil
				got[0].Content = nil
			}

			t.Errorf("Problem statements do not match:\n want %v\n  got %v", want, got)
		}
	})
}
