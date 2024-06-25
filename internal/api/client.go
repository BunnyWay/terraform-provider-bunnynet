package api

import (
	"io"
	"net/http"
)

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type Client struct {
	apiKey    string
	apiUrl    string
	userAgent string
}

func (c *Client) doRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("AccessKey", c.apiKey)
	req.Header.Add("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	return httpClient.Do(req)
}

func NewClient(apiKey string, apiUrl string, userAgent string) *Client {
	return &Client{
		apiKey:    apiKey,
		apiUrl:    apiUrl,
		userAgent: userAgent,
	}
}
