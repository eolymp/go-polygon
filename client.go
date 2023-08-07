package polygon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	key    string
	secret string
	base   string
	cli    httpClient
}

func New(key, secret string, opts ...func(*Client)) *Client {
	cli := &Client{
		key:    key,
		secret: secret,
		base:   "https://polygon.codeforces.com/api/",
		cli:    http.DefaultClient,
	}

	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

func (c *Client) request(ctx context.Context, method string, params map[string]string) (*http.Response, error) {
	base, err := url.Parse(c.base)
	if err != nil {
		return nil, fmt.Errorf("base URL %#v is corrupted: %w", c.base, err)
	}

	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}

	query.Set("time", fmt.Sprint(time.Now().Unix()))
	query.Set("apiKey", c.key)
	query.Set("apiSig", Signature(method, c.secret, query))

	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + url.PathEscape(method)

	req, err := http.NewRequest(http.MethodGet, base.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to compose HTTP request: %w", err)
	}

	resp, err := c.cli.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("response status code is not OK: %v", resp.Status)
	}

	return resp, nil
}

func (c *Client) call(ctx context.Context, method string, params map[string]string) (*Envelop, error) {
	resp, err := c.request(ctx, method, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	envelop := &Envelop{}

	if err := json.NewDecoder(resp.Body).Decode(envelop); err != nil {
		return nil, err
	}

	return envelop, nil
}

func (c *Client) ListPackages(ctx context.Context, in ListPackagesInput) ([]Package, error) {
	env, err := c.call(ctx, "problem.packages", map[string]string{"problemId": fmt.Sprint(in.ProblemID)})
	if err != nil {
		return nil, err
	}

	var packages []Package

	if err := env.Unmarshal(&packages); err != nil {
		return nil, err
	}

	return packages, nil
}

func (c *Client) DownloadPackage(ctx context.Context, in DownloadPackageInput) (io.ReadCloser, error) {
	resp, err := c.request(ctx, "problem.package", map[string]string{
		"problemId": fmt.Sprint(in.ProblemID),
		"packageId": fmt.Sprint(in.PackageID),
		"type":      in.Type,
	})

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
