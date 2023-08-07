package polygon

import (
	"context"
	"net/url"
	"os"
	"testing"
)

func TestProblemLoader_FetchProblem(t *testing.T) {
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

	_, err := loader.FetchProblem(ctx, link.String())
	if err != nil {
		t.Fatal(err)
	}

	// todo: make some assertions
}
