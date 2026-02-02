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

type PullzoneAccessList struct {
	Id         int64
	PullzoneId int64
	Name       string
	IsEnabled  bool
	Type       uint8
	Action     uint8
	Entries    []string
}

type pullzoneAccessListHttpType struct {
	Data struct {
		Id      int64  `json:"id"`
		Name    string `json:"name"`
		Type    uint8  `json:"type"`
		Content string `json:"content"`
	} `json:"data"`
}

type PullzoneAccessListQuery uint8

const (
	PullzoneAccessListQueryAll = iota
	PullzoneAccessListQueryCustom
	PullzoneAccessListQueryCurated
)

func (c *Client) GetPullzoneAccessList(ctx context.Context, pullzoneId int64, listId int64) (PullzoneAccessList, error) {
	var result PullzoneAccessList

	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(pullzoneId)
	if err != nil {
		return result, err
	}

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists/%d", c.apiUrl, shieldZoneId, listId), nil)
	if err != nil {
		return result, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return result, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return result, err
		} else {
			return result, errors.New("get access list failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /shield/shield-zone/%d/access-lists/%d: %s", shieldZoneId, listId, string(bodyResp)))

	var httpResult pullzoneAccessListHttpType
	err = json.Unmarshal(bodyResp, &httpResult)
	if err != nil {
		return result, err
	}

	result.Id = httpResult.Data.Id
	result.PullzoneId = pullzoneId
	result.Name = httpResult.Data.Name
	result.Type = httpResult.Data.Type
	result.Entries = strings.Split(httpResult.Data.Content, "\n")

	{
		v, err := c.getPullzoneAccessListInfo(ctx, shieldZoneId, listId)
		if err != nil {
			return result, err
		}

		result.IsEnabled = v.IsEnabled
		result.Action = v.Action
	}

	return result, nil
}

func (c *Client) CreatePullzoneAccessList(ctx context.Context, data PullzoneAccessList) (PullzoneAccessList, error) {
	var result PullzoneAccessList

	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(data.PullzoneId)
	if err != nil {
		return result, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"name":    data.Name,
		"type":    data.Type,
		"content": convertAccessListContentForApiSave(data.Entries),
	})

	if err != nil {
		return result, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /shield/shield-zone/%d/access-lists: %s", shieldZoneId, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists", c.apiUrl, shieldZoneId), bytes.NewReader(body))
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractShieldErrorMessage(resp)
		if err != nil {
			return result, err
		} else {
			return result, errors.New("create access list failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /shield/shield-zone/%d/access-lists: %s", shieldZoneId, string(bodyResp)))

	var httpResult pullzoneAccessListHttpType
	err = json.Unmarshal(bodyResp, &httpResult)
	if err != nil {
		return result, err
	}

	// save action
	{
		err := c.updatePullzoneAccessListConfiguration(ctx, shieldZoneId, httpResult.Data.Id, data.IsEnabled, data.Action)
		if err != nil {
			return result, err
		}
	}

	return c.GetPullzoneAccessList(ctx, data.PullzoneId, httpResult.Data.Id)
}

func (c *Client) UpdatePullzoneAccessList(ctx context.Context, data PullzoneAccessList) (PullzoneAccessList, error) {
	var result PullzoneAccessList

	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(data.PullzoneId)
	if err != nil {
		return result, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"name":    data.Name,
		"content": convertAccessListContentForApiSave(data.Entries),
	})

	if err != nil {
		return result, err
	}

	resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists/%d", c.apiUrl, shieldZoneId, data.Id), bytes.NewReader(body))
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return result, err
		} else {
			return result, errors.New("create access list failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	var httpResult pullzoneAccessListHttpType
	err = json.Unmarshal(bodyResp, &httpResult)
	if err != nil {
		return result, err
	}

	// save action
	{
		err := c.updatePullzoneAccessListConfiguration(ctx, shieldZoneId, httpResult.Data.Id, data.IsEnabled, data.Action)
		if err != nil {
			return result, err
		}
	}

	return c.GetPullzoneAccessList(ctx, data.PullzoneId, httpResult.Data.Id)
}

func (c *Client) DeletePullzoneAccessList(ctx context.Context, pullzoneId int64, listId int64) error {
	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(pullzoneId)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists/%d", c.apiUrl, shieldZoneId, listId), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return err
		} else {
			return errors.New("delete access list failed with " + resp.Status)
		}
	}

	return nil
}

type pullzoneAccessListInfo struct {
	ListId          int64  `json:"listId"`
	ConfigurationId int64  `json:"configurationId"`
	Name            string `json:"name"`
	Type            uint8  `json:"type"`
	Action          uint8  `json:"action"`
	IsEnabled       bool   `json:"isEnabled"`
	RequiredPlan    uint8  `json:"requiredPlan"`
}

func (c *Client) GetPullzoneAccessLists(ctx context.Context, pullzoneId int64, query PullzoneAccessListQuery) ([]PullzoneAccessList, error) {
	var result []PullzoneAccessList

	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(pullzoneId)
	if err != nil {
		return result, err
	}

	accessLists, err := c.getPullzoneAccessLists(ctx, shieldZoneId, query)
	if err != nil {
		return result, err
	}

	for _, list := range accessLists {
		result = append(result, PullzoneAccessList{
			Id:         list.ListId,
			PullzoneId: pullzoneId,
			Name:       list.Name,
			IsEnabled:  list.IsEnabled,
			Type:       list.Type,
			Action:     list.Action,
			Entries:    nil,
		})
	}

	return result, nil
}

func (c *Client) getPullzoneAccessLists(ctx context.Context, shieldZoneId int64, query PullzoneAccessListQuery) ([]pullzoneAccessListInfo, error) {
	var result []pullzoneAccessListInfo

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists", c.apiUrl, shieldZoneId), nil)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return result, err
		} else {
			return result, errors.New("get access list failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	tflog.Debug(ctx, fmt.Sprintf("GET /shield/shield-zone/%d/access-lists: %s", shieldZoneId, string(bodyResp)))

	var httpResult struct {
		ManagedLists []pullzoneAccessListInfo `json:"managedLists"`
		CustomLists  []pullzoneAccessListInfo `json:"customLists"`
	}

	err = json.Unmarshal(bodyResp, &httpResult)
	if err != nil {
		return result, err
	}

	switch query {
	case PullzoneAccessListQueryAll:
		result = append(result, httpResult.ManagedLists...)
		result = append(result, httpResult.CustomLists...)

	case PullzoneAccessListQueryCustom:
		result = append(result, httpResult.CustomLists...)

	case PullzoneAccessListQueryCurated:
		result = append(result, httpResult.ManagedLists...)

	default:
		return result, fmt.Errorf("unknown query type: %d", query)
	}

	return result, nil
}

func (c *Client) getPullzoneAccessListInfo(ctx context.Context, shieldZoneId int64, listId int64) (pullzoneAccessListInfo, error) {
	var result pullzoneAccessListInfo

	httpResult, err := c.getPullzoneAccessLists(ctx, shieldZoneId, PullzoneAccessListQueryAll)
	if err != nil {
		return result, err
	}

	for _, v := range httpResult {
		if v.ListId != listId {
			continue
		}

		return v, nil
	}

	return result, errors.New("access list not found")
}

func (c *Client) updatePullzoneAccessListConfiguration(ctx context.Context, shieldZoneId int64, listId int64, isEnabled bool, action uint8) error {
	body, err := json.Marshal(map[string]interface{}{
		"action":    action,
		"isEnabled": isEnabled,
	})

	if err != nil {
		return err
	}

	accessListInfo, err := c.getPullzoneAccessListInfo(ctx, shieldZoneId, listId)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("PATCH /shield/shield-zone/%d/access-lists/configurations/%d", shieldZoneId, accessListInfo.ConfigurationId))

	resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/shield-zone/%d/access-lists/configurations/%d", c.apiUrl, shieldZoneId, accessListInfo.ConfigurationId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractShieldErrorMessage(resp)
		if err != nil {
			return err
		} else {
			return errors.New("update access list failed with " + resp.Status)
		}
	}

	_ = resp.Body.Close()
	return nil
}

func convertAccessListContentForApiSave(entries []string) string {
	return strings.Join(entries, "\n")
}
