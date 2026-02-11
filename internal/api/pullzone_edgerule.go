// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

type PullzoneEdgerule struct {
	Id               string                        `json:"Guid,omitempty"`
	Enabled          bool                          `json:"Enabled"`
	Description      string                        `json:"Description"`
	Action           uint8                         `json:"ActionType"`
	ActionParameter1 string                        `json:"ActionParameter1"`
	ActionParameter2 string                        `json:"ActionParameter2"`
	ActionParameter3 string                        `json:"ActionParameter3"`
	ExtraActions     []PullzoneEdgeruleExtraAction `json:"ExtraActions"`
	MatchType        uint8                         `json:"TriggerMatchingType"`
	Triggers         []PullzoneEdgeruleTrigger     `json:"Triggers"`
	OrderIndex       int64                         `json:"OrderIndex,omitempty"`
	PullzoneId       int64                         `json:"-"`
}

type PullzoneEdgeruleExtraAction struct {
	ActionType       uint8  `json:"ActionType"`
	ActionParameter1 string `json:"ActionParameter1"`
	ActionParameter2 string `json:"ActionParameter2"`
	ActionParameter3 string `json:"ActionParameter3"`
}

type PullzoneEdgeruleTrigger struct {
	Type       uint8    `json:"Type"`
	MatchType  uint8    `json:"PatternMatchingType"`
	Patterns   []string `json:"PatternMatches"`
	Parameter1 string   `json:"Parameter1,omitempty"`
	Parameter2 string   `json:"Parameter2,omitempty"`
}

func (c *Client) CreatePullzoneEdgerule(ctx context.Context, data PullzoneEdgerule) (PullzoneEdgerule, error) {
	if data.PullzoneId == 0 {
		return PullzoneEdgerule{}, errors.New("pullzone is required")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /pullzone/%d/edgerules/addOrUpdate: %+v", data.PullzoneId, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/edgerules/addOrUpdate", c.apiUrl, data.PullzoneId), bytes.NewReader(body))
	if err != nil {
		return PullzoneEdgerule{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneEdgerule{}, err
		} else {
			return PullzoneEdgerule{}, errors.New(resp.Status)
		}
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

	return PullzoneEdgerule{}, ErrNotFound
}

func (c *Client) DeletePullzoneEdgerule(pullzoneId int64, guid string) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d/edgerules/%s", c.apiUrl, pullzoneId, guid), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
