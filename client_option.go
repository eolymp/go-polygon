package polygon

func UseBaseURL(base string) func(*Client) {
	return func(cli *Client) {
		cli.base = base
	}
}

func UseHTTPClient(hc httpClient) func(*Client) {
	return func(cli *Client) {
		cli.cli = hc
	}
}
