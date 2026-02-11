// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

const ScriptTypeDns = 0
const ScriptTypeStandalone = 1
const ScriptTypeMiddleware = 2

type ComputeScript struct {
	Id               int64  `json:"Id,omitempty"`
	ScriptType       uint8  `json:"ScriptType"`
	Name             string `json:"Name"`
	Content          string `json:"Content"`
	DeploymentKey    string `json:"DeploymentKey,omitempty"`
	Release          string `json:"-"`
	CurrentReleaseId int64  `json:"CurrentReleaseId,omitempty"`
}

func (c *Client) GetComputeScript(ctx context.Context, id int64) (ComputeScript, error) {
	var data ComputeScript

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

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

	tflog.Debug(ctx, fmt.Sprintf("GET /compute/script/%d: %+v", id, string(bodyResp)))

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

		tflog.Debug(ctx, fmt.Sprintf("GET /compute/script/%d/code: %+v", id, string(codeBodyResp)))

		var codeData map[string]string
		err = json.Unmarshal(codeBodyResp, &codeData)
		if err != nil {
			return data, err
		}

		data.Content = codeData["Code"]
	}

	// current release
	if data.CurrentReleaseId > 0 {
		release, err := c.GetComputeScriptActiveRelease(data.Id)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return data, err
			}
		} else {
			data.Release = release.Uuid
		}
	}

	return data, nil
}

func (c *Client) CreateComputeScript(ctx context.Context, dataApi ComputeScript) (ComputeScript, error) {
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

	return c.GetComputeScript(ctx, dataApiResult.Id)
}

func (c *Client) UpdateComputeScript(ctx context.Context, data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	id := data.Id

	data, err := c.UpdateComputeScriptWithoutGet(data, previousData)
	if err != nil {
		return data, err
	}

	return c.GetComputeScript(ctx, id)
}

func (c *Client) UpdateComputeScriptWithoutGet(data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	id := data.Id

	// update attributes
	if data.Name != previousData.Name {
		body, err := json.Marshal(map[string]string{
			"Name": data.Name,
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
	if data.Content != previousData.Content {
		body, err := json.Marshal(map[string]string{
			"Code": data.Content,
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

	return data, nil
}

func (c *Client) DeleteComputeScript(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
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
