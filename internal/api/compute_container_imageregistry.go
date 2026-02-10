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
	"strings"
)

type ComputeContainerImageregistry struct {
	Id                   int64  `json:"id"`
	AccountId            string `json:"accountId"`
	UserId               string `json:"userId"`
	NamespaceId          string `json:"namespaceId"`
	IsPublic             bool   `json:"isPublic"`
	DisplayName          string `json:"displayName"`
	HostName             string `json:"hostName"`
	UserName             string `json:"userName"`
	FirstPasswordSymbols string `json:"firstPasswordSymbols"`
	LastPasswordSymbols  string `json:"lastPasswordSymbols"`
	CreatedAt            string `json:"createdAt"`
	LastUpdatedAt        string `json:"lastUpdatedAt"`

	Token string `json:"-"`
}

func (c *Client) CreateComputeContainerImageregistry(ctx context.Context, data ComputeContainerImageregistry) (ComputeContainerImageregistry, error) {
	body, err := json.Marshal(map[string]interface{}{
		"displayName": data.DisplayName,
		"type":        data.DisplayName,
		"passwordCredentials": map[string]string{
			"userName": data.UserName,
			"password": data.Token,
		},
	})

	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /mc/registries: %s", string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/mc/registries", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return ComputeContainerImageregistry{}, err
		}

		return ComputeContainerImageregistry{}, errors.New(string(bodyResp))
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /mc/registries response: %s", string(bodyResp)))

	var result struct {
		Id int64 `json:"id"`
	}

	err = json.Unmarshal(bodyResp, &result)
	_ = resp.Body.Close()

	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	data.Id = result.Id

	return data, nil
}

func (c *Client) getAllComputeContainerImageregistries(ctx context.Context) ([]ComputeContainerImageregistry, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/mc/registries", c.apiUrl), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// @TODO handle pagination once API supports it (see MC-1667)
	var result struct {
		Items []ComputeContainerImageregistry `json:"items"`
	}

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) GetComputeContainerImageregistry(ctx context.Context, id int64) (ComputeContainerImageregistry, error) {
	result, err := c.getAllComputeContainerImageregistries(ctx)
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	for _, item := range result {
		if id == item.Id {
			return item, nil
		}
	}

	return ComputeContainerImageregistry{}, errors.New("Could not get compute container imageregistry")
}

func (c *Client) FindComputeContainerImageregistry(ctx context.Context, registry string, username string) (ComputeContainerImageregistry, error) {
	result, err := c.getAllComputeContainerImageregistries(ctx)
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	if username == "" {
		registry = fmt.Sprintf("%s Public", registry)
	}

	for _, item := range result {
		if item.DisplayName != registry {
			continue
		}

		if item.UserName != username {
			continue
		}

		if item.IsPublic && strings.HasSuffix(item.DisplayName, " Public") {
			item.DisplayName = item.DisplayName[:len(item.DisplayName)-7]
		}

		return item, nil
	}

	return ComputeContainerImageregistry{}, errors.New("Could not find compute container imageregistry")
}

func (c *Client) UpdateComputeContainerImageregistry(ctx context.Context, data ComputeContainerImageregistry) (ComputeContainerImageregistry, error) {
	token := data.Token

	body, err := json.Marshal(map[string]interface{}{
		"id":          data.Id,
		"displayName": data.DisplayName,
		"type":        data.DisplayName,
		"passwordCredentials": map[string]string{
			"userName": data.UserName,
			"password": token,
		},
	})

	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/mc/registries/%d", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractMCErrorMessage(resp)
		if err != nil {
			return ComputeContainerImageregistry{}, err
		} else {
			return ComputeContainerImageregistry{}, fmt.Errorf("Error: HTTP Status: %s", resp.Status)
		}
	}

	dataApiResult, err := c.GetComputeContainerImageregistry(ctx, data.Id)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.Token = token

	return dataApiResult, nil
}

func (c *Client) DeleteComputeContainerImageregistry(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/mc/registries/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
