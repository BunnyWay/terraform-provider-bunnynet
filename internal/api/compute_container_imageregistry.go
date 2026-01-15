// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ComputeContainerImageregistry struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	HostName    string `json:"hostName"`
	UserName    string `json:"userName"`
	Token       string `json:"-"`
	IsPublic    bool   `json:"isPublic,omitempty"`
}

func (c *Client) CreateComputeContainerImageregistry(data ComputeContainerImageregistry) (ComputeContainerImageregistry, error) {
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

	resp, err := c.doJWTRequest(http.MethodPost, fmt.Sprintf("%s/v2/user/namespaces/default/container-registries", c.apiUrl), bytes.NewReader(body))
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

	var result struct {
		Id string `json:"id"`
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
	resp, err := c.doJWTRequest(http.MethodGet, fmt.Sprintf("%s/v2/user/namespaces/default/container-registries", c.apiUrl), nil)
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

	var result []ComputeContainerImageregistry
	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetComputeContainerImageregistry(ctx context.Context, id int64) (ComputeContainerImageregistry, error) {
	result, err := c.getAllComputeContainerImageregistries(ctx)
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	idStr := fmt.Sprintf("%d", id)
	for _, item := range result {
		if item.Id == idStr {
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
	idStr := data.Id
	token := data.Token

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"id":          idStr,
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

	resp, err := c.doJWTRequest(http.MethodPut, fmt.Sprintf("%s/v2/user/namespaces/default/container-registries/%s", c.apiUrl, idStr), bytes.NewReader(body))
	if err != nil {
		return ComputeContainerImageregistry{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return ComputeContainerImageregistry{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return ComputeContainerImageregistry{}, err
		}

		return ComputeContainerImageregistry{}, errors.New(obj.Title + ": " + obj.Detail)
	}

	dataApiResult, err := c.GetComputeContainerImageregistry(ctx, id)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.Token = token

	return dataApiResult, nil
}

func (c *Client) DeleteComputeContainerImageregistry(id int64) error {
	resp, err := c.doJWTRequest(http.MethodDelete, fmt.Sprintf("%s/v2/user/namespaces/default/container-registries/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
