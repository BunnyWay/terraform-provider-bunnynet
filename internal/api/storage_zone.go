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

func (c *Client) GetStorageZone(id int64) (StorageZone, error) {
	var data StorageZone
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
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

	return data, nil
}

func (c *Client) CreateStorageZone(data StorageZone) (StorageZone, error) {
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

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StorageZone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StorageZone{}, err
		}

		return StorageZone{}, errors.New(obj.Message)
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
	dataApiResult, err = c.UpdateStorageZone(data)
	if err != nil {
		_ = c.DeleteStorageZone(data.Id)
	}
	return dataApiResult, err
}

func (c *Client) UpdateStorageZone(dataApi StorageZone) (StorageZone, error) {
	id := dataApi.Id

	dataUpdate := map[string]interface{}{
		"Rewrite404To200":  dataApi.Rewrite404To200,
		"ReplicationZones": dataApi.ReplicationRegions,
	}

	// @TODO API can't unset Custom404FilePath
	if dataApi.Custom404FilePath != "" {
		dataUpdate["Custom404FilePath"] = dataApi.Custom404FilePath
	}

	body, err := json.Marshal(dataUpdate)
	if err != nil {
		return StorageZone{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return StorageZone{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StorageZone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StorageZone{}, err
		}

		return StorageZone{}, errors.New(obj.Message)
	}

	dataApiResult, err := c.GetStorageZone(id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStorageZone(id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
