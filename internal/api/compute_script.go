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

type ComputeScript struct {
	Id         int64  `json:"Id,omitempty"`
	ScriptType uint8  `json:"ScriptType"`
	Name       string `json:"Name"`
	Content    string `json:"Content"`
}

func (c *Client) GetComputeScript(id int64) (ComputeScript, error) {
	var data ComputeScript

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
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

	// code
	{
		codeResp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d/code", c.apiUrl, id), nil)
		if err != nil {
			return data, err
		}

		if codeResp.StatusCode != http.StatusOK {
			return data, errors.New(codeResp.Status)
		}

		codeBodyResp, err := io.ReadAll(codeResp.Body)
		if err != nil {
			return data, err
		}
		_ = codeResp.Body.Close()

		var codeData map[string]string
		err = json.Unmarshal(codeBodyResp, &codeData)
		if err != nil {
			return data, err
		}

		data.Content = codeData["Code"]
	}

	return data, nil
}

func (c *Client) CreateComputeScript(dataApi ComputeScript) (ComputeScript, error) {
	body := map[string]interface{}{
		"Name":                 dataApi.Name,
		"ScriptType":           dataApi.ScriptType,
		"CreateLinkedPullZone": false,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return ComputeScript{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script", c.apiUrl), bytes.NewReader(bodyBytes))
	if err != nil {
		return ComputeScript{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return ComputeScript{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return ComputeScript{}, err
		}

		return ComputeScript{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComputeScript{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := ComputeScript{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	// set code
	{
		codeBody := map[string]interface{}{
			"Code": dataApi.Content,
		}

		codeBodyBytes, err := json.Marshal(codeBody)
		if err != nil {
			return ComputeScript{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/code", c.apiUrl, dataApiResult.Id), bytes.NewReader(codeBodyBytes))
		if err != nil {
			return ComputeScript{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScript{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScript{}, err
			}

			return ComputeScript{}, errors.New(obj.Message)
		}

		_ = resp.Body.Close()
	}

	return c.GetComputeScript(dataApiResult.Id)
}

func (c *Client) UpdateComputeScript(dataApi ComputeScript) (ComputeScript, error) {
	id := dataApi.Id

	// update attributes
	{
		body, err := json.Marshal(map[string]string{
			"Name": dataApi.Name,
		})

		if err != nil {
			return ComputeScript{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), bytes.NewReader(body))
		if err != nil {
			return ComputeScript{}, err
		}

		if resp.StatusCode != http.StatusOK {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScript{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScript{}, err
			}

			return ComputeScript{}, errors.New(obj.Message)
		}
	}

	// update code
	{
		body, err := json.Marshal(map[string]string{
			"Code": dataApi.Content,
		})

		if err != nil {
			return ComputeScript{}, err
		}

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/code", c.apiUrl, id), bytes.NewReader(body))
		if err != nil {
			return ComputeScript{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScript{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScript{}, err
			}

			return ComputeScript{}, errors.New(obj.Message)
		}
	}

	return c.GetComputeScript(id)
}

func (c *Client) DeleteComputeScript(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
