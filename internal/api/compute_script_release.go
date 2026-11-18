// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"io"
	"net/http"
)

type ComputeScriptRelease struct {
	Id   int64  `json:"Id,omitempty"`
	Uuid string `json:"Uuid"`
	Note string `json:"Note"`
	Code string `json:"Code"`
}

func (c *Client) GetComputeScriptActiveRelease(scriptId int64) (ComputeScriptRelease, error) {
	var response ComputeScriptRelease

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d/releases/active", c.apiUrl, scriptId), nil)
	if err != nil {
		return response, err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return response, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return response, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(bodyResp, &response)
	return response, err
}
