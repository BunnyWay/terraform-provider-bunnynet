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

type StorageZone struct {
	Id                 int64    `json:"Id,omitempty"`
	Name               string   `json:"Name,omitempty"`
	Password           string   `json:"Password,omitempty"`
	ReadOnlyPassword   string   `json:"ReadOnlyPassword,omitempty"`
	Region             string   `json:"Region,omitempty"`
	ReplicationRegions []string `json:"ReplicationRegions,omitempty"`
	StorageHostname    string   `json:"StorageHostname,omitempty"`
	ZoneTier           uint8    `json:"ZoneTier,omitempty"`
	Custom404FilePath  string   `json:"Custom404FilePath,omitempty"`
	Rewrite404To200    bool     `json:"Rewrite404To200"`
	DateModified       string   `json:"DateModified,omitempty"`
}

func (c *Client) GetStorageZone(ctx context.Context, id int64) (StorageZone, error) {
	var data StorageZone
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
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

	tflog.Debug(ctx, fmt.Sprintf("GET /storagezone/%d", id), map[string]any{
		"status":   resp.Status,
		"response": string(bodyResp),
	})

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (c *Client) CreateStorageZone(ctx context.Context, data StorageZone) (StorageZone, error) {
	body, err := json.Marshal(map[string]interface{}{
		"Name":               data.Name,
		"ZoneTier":           data.ZoneTier,
		"Region":             data.Region,
		"ReplicationRegions": data.ReplicationRegions,
	})
	if err != nil {
		return StorageZone{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/storagezone", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return StorageZone{}, err
	}

	tflog.Debug(ctx, "POST /storagezone", map[string]interface{}{
		"payload": string(body),
		"status":  resp.Status,
	})

	if resp.StatusCode != http.StatusCreated {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return StorageZone{}, err
		} else {
			return StorageZone{}, errors.New("create storage zone failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return StorageZone{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := StorageZone{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	data.Id = dataApiResult.Id
	return c.UpdateStorageZone(ctx, data)
}

func (c *Client) UpdateStorageZone(ctx context.Context, dataApi StorageZone) (StorageZone, error) {
	id := dataApi.Id

	dataUpdate := map[string]interface{}{
		"Rewrite404To200":   dataApi.Rewrite404To200,
		"ReplicationZones":  dataApi.ReplicationRegions,
		"Custom404FilePath": dataApi.Custom404FilePath,
	}

	body, err := json.Marshal(dataUpdate)
	if err != nil {
		return StorageZone{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return StorageZone{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /storagezone/%d", id), map[string]interface{}{
		"payload": string(body),
		"status":  resp.Status,
	})

	if resp.StatusCode != http.StatusNoContent {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return StorageZone{}, err
		} else {
			return StorageZone{}, errors.New("update storage zone failed with " + resp.Status)
		}
	}

	dataApiResult, err := c.GetStorageZone(ctx, id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStorageZone(ctx context.Context, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("DELETE /storagezone/%d", id), map[string]interface{}{
		"status": resp.Status,
	})

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
