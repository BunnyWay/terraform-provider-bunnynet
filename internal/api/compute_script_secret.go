// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"io"
	"net/http"
)

type ComputeScriptSecret struct {
	Id       int64  `json:"Id,omitempty"`
	ScriptId int64  `json:"-"`
	Name     string `json:"Name"`
	Value    string `json:"Secret,omitempty"`
}

func (c *Client) GetComputeScriptSecretByName(scriptId int64, name string) (ComputeScriptSecret, error) {
	var data ComputeScriptSecret

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d/secrets", c.apiUrl, scriptId), nil)
	if err != nil {
		return data, err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return data, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return data, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	var result struct {
		Secrets []ComputeScriptSecret `json:"Secrets"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return data, err
	}

	for _, secret := range result.Secrets {
		if secret.Name == name {
			secret.ScriptId = scriptId
			return secret, nil
		}
	}

	return data, errors.New("secret not found")
}

func (c *Client) CreateComputeScriptSecret(dataApi ComputeScriptSecret) (ComputeScriptSecret, error) {
	scriptId := dataApi.ScriptId

	dataApiResult, err := c.createComputeScriptSecret(scriptId, dataApi)
	if err != nil {
		return dataApi, err
	}

	return c.GetComputeScriptSecretByName(scriptId, dataApiResult.Name)
}

func (c *Client) createComputeScriptSecret(scriptId int64, dataApi ComputeScriptSecret) (ComputeScriptSecret, error) {
	var dataApiResult ComputeScriptSecret

	bodyBytes, err := json.Marshal(dataApi)
	if err != nil {
		return dataApiResult, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/secrets", c.apiUrl, scriptId), bytes.NewReader(bodyBytes))
	if err != nil {
		return dataApiResult, err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err := extractErrorMessage(resp)
		if err != nil {
			return dataApiResult, err
		}
		return dataApiResult, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return dataApiResult, err
	}

	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdateComputeScriptSecret(dataApi ComputeScriptSecret) (ComputeScriptSecret, error) {
	scriptId := dataApi.ScriptId

	err := c.updateComputeScriptSecret(scriptId, dataApi)
	if err != nil {
		return dataApi, err
	}

	return c.GetComputeScriptSecretByName(scriptId, dataApi.Name)
}

func (c *Client) updateComputeScriptSecret(scriptId int64, dataApi ComputeScriptSecret) error {
	dataApi.Id = 0
	body, err := json.Marshal(dataApi)

	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/compute/script/%d/secrets", c.apiUrl, scriptId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		err := extractErrorMessage(resp)
		if err != nil {
			return err
		}
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) DeleteComputeScriptSecret(scriptId int64, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d/secrets/%d", c.apiUrl, scriptId, id), nil)
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
