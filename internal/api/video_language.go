package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type VideoLanguage struct {
	ShortCode                string `json:"ShortCode"`
	Name                     string `json:"Name"`
	SupportPlayerTranslation bool   `json:"SupportPlayerTranslation"`
	SupportTranscribing      bool   `json:"SupportTranscribing"`
	TranscribingAccuracy     int64  `json:"TranscribingAccuracy"`
}

func (c *Client) GetVideoLanguage(code string) (VideoLanguage, error) {
	languages, err := c.GetVideoLanguages()
	if err != nil {
		return VideoLanguage{}, err
	}

	for _, language := range languages {
		if language.ShortCode == code {
			return language, nil
		}
	}

	return VideoLanguage{}, errors.New("language not found")
}

func (c *Client) GetVideoLanguages() ([]VideoLanguage, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/videolibrary/languages", c.apiUrl), nil)
	if err != nil {
		return []VideoLanguage{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []VideoLanguage{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return []VideoLanguage{}, err
	}

	_ = resp.Body.Close()
	var languages []VideoLanguage

	err = json.Unmarshal(bodyResp, &languages)
	if err != nil {
		return []VideoLanguage{}, err
	}

	return languages, nil
}
