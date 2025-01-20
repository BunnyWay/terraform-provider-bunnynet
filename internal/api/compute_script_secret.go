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

	_ = resp.Body.Close()
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
	bodyBytes, err := json.Marshal(dataApi)
	if err != nil {
		return ComputeScriptSecret{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/secrets", c.apiUrl, scriptId), bytes.NewReader(bodyBytes))
	if err != nil {
		return ComputeScriptSecret{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return ComputeScriptSecret{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return ComputeScriptSecret{}, err
		}

		return ComputeScriptSecret{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComputeScriptSecret{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := ComputeScriptSecret{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return c.GetComputeScriptSecretByName(scriptId, dataApiResult.Name)
}

func (c *Client) UpdateComputeScriptSecret(dataApi ComputeScriptSecret) (ComputeScriptSecret, error) {
	scriptId := dataApi.ScriptId

	// update attributes
	{
		dataApi.Id = 0
		body, err := json.Marshal(dataApi)

		if err != nil {
			return ComputeScriptSecret{}, err
		}

		resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/compute/script/%d/secrets", c.apiUrl, scriptId), bytes.NewReader(body))
		if err != nil {
			return ComputeScriptSecret{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScriptSecret{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScriptSecret{}, err
			}

			return ComputeScriptSecret{}, errors.New(obj.Message)
		}
	}

	return c.GetComputeScriptSecretByName(scriptId, dataApi.Name)
}

func (c *Client) DeleteComputeScriptSecret(scriptId int64, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d/secrets/%d", c.apiUrl, scriptId, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
