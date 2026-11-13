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
	data, err := c.getComputeScriptInfo(ctx, id)
	if err != nil {
		return data, err
	}

	// code
	code, err := c.getComputeScriptCode(ctx, id)
	if err != nil {
		return data, err
	}

	data.Content = code

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

func (c *Client) getComputeScriptInfo(ctx context.Context, id int64) (ComputeScript, error) {
	var data ComputeScript

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return data, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err = extractErrorMessage(resp)
		if err != nil {
			return data, err
		}
		return data, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /compute/script/%d: %+v", id, string(bodyResp)))

	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (c *Client) getComputeScriptCode(ctx context.Context, id int64) (string, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/compute/script/%d/code", c.apiUrl, id), nil)
	if err != nil {
		return "", err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = extractErrorMessage(resp)
		if err != nil {
			return "", err
		}
		return "", errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /compute/script/%d/code: %+v", id, string(bodyResp)))

	var data struct {
		Code string `json:"code"`
	}

	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return "", err
	}

	return data.Code, nil
}

func (c *Client) CreateComputeScript(ctx context.Context, dataApi ComputeScript) (ComputeScript, error) {
	dataApiResult, err := c.createComputeScript(ctx, dataApi)
	if err != nil {
		return dataApiResult, err
	}

	err = c.setComputeScriptCode(ctx, dataApiResult.Id, dataApi.Content)
	if err != nil {
		return ComputeScript{}, err
	}

	// @TODO should we publish the script on creation?

	return c.GetComputeScript(ctx, dataApiResult.Id)
}

func (c *Client) createComputeScript(ctx context.Context, data ComputeScript) (ComputeScript, error) {
	var dataApiResult ComputeScript

	body := map[string]interface{}{
		"Name":                 data.Name,
		"ScriptType":           data.ScriptType,
		"CreateLinkedPullZone": false,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return dataApiResult, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /compute/script: %+v", string(bodyBytes)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script", c.apiUrl), bytes.NewReader(bodyBytes))
	if err != nil {
		return dataApiResult, err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		err = extractErrorMessage(resp)
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

func (c *Client) setComputeScriptCode(ctx context.Context, id int64, content string) error {
	body := map[string]string{
		"Code": content,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /compute/script/%d/code: %+v", id, string(bodyBytes)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/code", c.apiUrl, id), bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		err = extractErrorMessage(resp)
		if err != nil {
			return err
		}
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) UpdateComputeScript(ctx context.Context, data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	id := data.Id

	// update info
	if data.Name != previousData.Name {
		err := c.updateComputeScript(ctx, id, data)
		if err != nil {
			return data, err
		}
	}

	// update code
	if data.Content != previousData.Content {
		err := c.setComputeScriptCode(ctx, id, data.Content)
		if err != nil {
			return ComputeScript{}, err
		}

		// publish script
		if previousData.CurrentReleaseId > 0 {
			err = c.publishComputeScript(data)
			if err != nil {
				return data, err
			}
		}
	}

	return c.GetComputeScript(ctx, id)
}

func (c *Client) updateComputeScript(ctx context.Context, id int64, data ComputeScript) error {
	body, err := json.Marshal(map[string]string{
		"Name": data.Name,
	})

	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /compute/script/%d: %+v", id, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err = extractErrorMessage(resp)
		if err != nil {
			return err
		}
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) DeleteComputeScript(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/compute/script/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusNoContent {
		err = extractErrorMessage(resp)
		if err != nil {
			return err
		}
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) publishComputeScript(data ComputeScript) error {
	body, err := json.Marshal(map[string]string{
		"Note": "",
	})

	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/publish", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer utils.Drain(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		err = extractErrorMessage(resp)
		if err != nil {
			return err
		}
		return errors.New(resp.Status)
	}

	return nil
}
