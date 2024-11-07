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

type ComputeScriptVariable struct {
	Id           int64  `json:"Id,omitempty"`
	ScriptId     int64  `json:"-"`
	Name         string `json:"Name"`
	Required     bool   `json:"Required"`
	DefaultValue string `json:"DefaultValue"`
}

func (c *Client) GetComputeScriptVariableByName(scriptId int64, name string) (ComputeScriptVariable, error) {
	var data ComputeScriptVariable

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, scriptId), nil)
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

	var scriptData struct {
		EdgeScriptVariables []ComputeScriptVariable `json:"EdgeScriptVariables"`
	}

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &scriptData)
	if err != nil {
		return data, err
	}

	for _, variable := range scriptData.EdgeScriptVariables {
		if variable.Name == name {
			variable.ScriptId = scriptId
			return variable, nil
		}
	}

	return data, errors.New("variable not found")
}

func (c *Client) GetComputeScriptVariable(scriptId int64, id int64) (ComputeScriptVariable, error) {
	var data ComputeScriptVariable

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d/variables/%d", c.apiUrl, scriptId, id), nil)
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

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	data.ScriptId = scriptId

	return data, nil
}

func (c *Client) CreateComputeScriptVariable(dataApi ComputeScriptVariable) (ComputeScriptVariable, error) {
	scriptId := dataApi.ScriptId
	bodyBytes, err := json.Marshal(dataApi)
	if err != nil {
		return ComputeScriptVariable{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/variables/add", c.apiUrl, scriptId), bytes.NewReader(bodyBytes))
	if err != nil {
		return ComputeScriptVariable{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return ComputeScriptVariable{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return ComputeScriptVariable{}, err
		}

		return ComputeScriptVariable{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComputeScriptVariable{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := ComputeScriptVariable{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return c.GetComputeScriptVariable(scriptId, dataApiResult.Id)
}

func (c *Client) UpdateComputeScriptVariable(dataApi ComputeScriptVariable) (ComputeScriptVariable, error) {
	id := dataApi.Id
	scriptId := dataApi.ScriptId

	// update attributes
	{
		dataApi.Id = 0
		body, err := json.Marshal(dataApi)

		if err != nil {
			return ComputeScriptVariable{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/variables/%d", c.apiUrl, scriptId, id), bytes.NewReader(body))
		if err != nil {
			return ComputeScriptVariable{}, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScriptVariable{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScriptVariable{}, err
			}

			return ComputeScriptVariable{}, errors.New(obj.Message)
		}
	}

	return c.GetComputeScriptVariable(scriptId, id)
}

func (c *Client) DeleteComputeScriptVariable(scriptId int64, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d/variables/%d", c.apiUrl, scriptId, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
