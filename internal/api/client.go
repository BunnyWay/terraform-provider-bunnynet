// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type Client struct {
	apiKey       string
	jwtToken     string
	apiUrl       string
	streamApiUrl string
	userAgent    string
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

func (c *Client) doStreamRequest(library StreamLibrary, method string, suffixUrl string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s/library/%d/%s", c.streamApiUrl, library.Id, suffixUrl)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("AccessKey", library.ApiKey)
	req.Header.Add("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	return httpClient.Do(req)
}

func (c *Client) doJWTRequest(method string, url string, body io.Reader) (*http.Response, error) {
	jwtToken, err := c.getJWTToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", jwtToken)
	req.Header.Add("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	return httpClient.Do(req)
}

func (c *Client) getJWTToken() (string, error) {
	if len(c.jwtToken) > 0 {
		return c.jwtToken, nil
	}

	bodyJson, err := json.Marshal(map[string]string{
		"AccessKey": c.apiKey,
	})

	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(http.MethodPost, c.apiUrl+"/apikey/exchange", bytes.NewReader(bodyJson))
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	_ = resp.Body.Close()
	var obj struct {
		Token string `json:"Token"`
	}

	err = json.Unmarshal(bodyResp, &obj)
	if err != nil {
		return "", err
	}

	if len(obj.Token) > 0 {
		c.jwtToken = obj.Token
		return c.jwtToken, nil
	}

	return "", errors.New("Invalid JWT token received")
}

func NewClient(apiKey string, apiUrl string, streamApiUrl string, userAgent string) *Client {
	return &Client{
		apiKey:       apiKey,
		apiUrl:       apiUrl,
		streamApiUrl: streamApiUrl,
		userAgent:    userAgent,
		jwtToken:     "",
	}
}
