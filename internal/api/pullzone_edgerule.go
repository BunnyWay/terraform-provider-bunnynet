package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type PullzoneEdgerule struct {
	Id          string                    `json:"Guid,omitempty"`
	Enabled     bool                      `json:"Enabled"`
	Description string                    `json:"Description"`
	Action      uint8                     `json:"ActionType"`
	MatchType   uint8                     `json:"TriggerMatchingType"`
	Triggers    []PullzoneEdgeruleTrigger `json:"Triggers"`
	PullzoneId  int64                     `json:"-"`
}

type PullzoneEdgeruleTrigger struct {
	Type       uint8    `json:"Type"`
	MatchType  uint8    `json:"PatternMatchingType"`
	Patterns   []string `json:"PatternMatches"`
	Parameter1 string   `json:"Parameter1,omitempty"`
	Parameter2 string   `json:"Parameter2,omitempty"`
}

func (c *Client) CreatePullzoneEdgerule(data PullzoneEdgerule) (PullzoneEdgerule, error) {
	if data.PullzoneId == 0 {
		return PullzoneEdgerule{}, errors.New("pullzone is required")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/edgerules/addOrUpdate", c.apiUrl, data.PullzoneId), bytes.NewReader(body))
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	req.Header.Add("AccessKey", c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return PullzoneEdgerule{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneEdgerule{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := PullzoneEdgerule{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.PullzoneId = data.PullzoneId

	return dataApiResult, nil
}

func (c *Client) GetPullzoneEdgerule(pullzoneId int64, guid string) (PullzoneEdgerule, error) {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	for _, edgerule := range pullzone.Edgerules {
		if edgerule.Id == guid {
			edgerule.PullzoneId = pullzoneId
			return edgerule, nil
		}
	}

	return PullzoneEdgerule{}, errors.New("Edgerule not found")
}

func (c *Client) DeletePullzoneEdgerule(pullzoneId int64, guid string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d/edgerules/%s", c.apiUrl, pullzoneId, guid), nil)
	if err != nil {
		return err
	}

	req.Header.Add("AccessKey", c.apiKey)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
