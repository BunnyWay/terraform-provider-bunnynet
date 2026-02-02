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

type Database struct {
	Id              string   `json:"id"`
	Name            string   `json:"name"`
	Url             string   `json:"url"`
	BlockReads      bool     `json:"block_reads"`
	BlockWrites     bool     `json:"block_writes"`
	SizeMax         string   `json:"size_max"`
	CurrentSize     string   `json:"current_size"`
	StorageRegion   string   `json:"storage_region"`
	PrimaryRegions  []string `json:"primary_regions"`
	ReplicasRegions []string `json:"replicas_regions"`
}

func (c *Client) GetDatabase(ctx context.Context, id string) (Database, error) {
	var data struct {
		Database Database `json:"db"`
	}

	resp, err := c.doJWTRequest(http.MethodGet, fmt.Sprintf("%s/edgedb/v2/databases/%s", c.apiUrl, id), nil)
	if err != nil {
		return data.Database, err
	}

	tflog.Info(ctx, fmt.Sprintf("GET /edgedb/v2/databases/%s: %s", id, resp.Status))

	if resp.StatusCode == http.StatusNotFound {
		return data.Database, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return data.Database, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return data.Database, err
	}

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data.Database, err
	}

	return data.Database, nil
}

func (c *Client) CreateDatabase(ctx context.Context, data Database) (Database, error) {
	body, err := json.Marshal(map[string]interface{}{
		"name":             data.Name,
		"primary_regions":  data.PrimaryRegions,
		"replicas_regions": data.ReplicasRegions,
		"storage_region":   "us-east-1",
	})

	if err != nil {
		return Database{}, err
	}

	resp, err := c.doJWTRequest(http.MethodPost, fmt.Sprintf("%s/edgedb/v2/databases", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return Database{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("POST /edgedb/v2/databases: %s", string(body)))

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractDatabaseErrorMessage(resp)
		if err != nil {
			return Database{}, err
		}

		return Database{}, errors.New("create database failed with " + resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Database{}, err
	}
	_ = resp.Body.Close()

	var dataApiResult struct {
		Id string `json:"db_id"`
	}

	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return Database{}, err
	}

	return c.GetDatabase(ctx, dataApiResult.Id)
}

func (c *Client) UpdateDatabase(ctx context.Context, data Database) (Database, error) {
	body, err := json.Marshal(map[string]interface{}{
		"primary_regions":  data.PrimaryRegions,
		"replicas_regions": data.ReplicasRegions,
	})

	if err != nil {
		return Database{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("PATCH /edgedb/v2/databases/%s: %s", data.Id, string(body)))

	resp, err := c.doJWTRequest(http.MethodPatch, fmt.Sprintf("%s/edgedb/v2/databases/%s", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return Database{}, err
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		// @TODO extract error message
		return Database{}, errors.New("update database failed with " + resp.Status)
	}

	return c.GetDatabase(ctx, data.Id)
}

func (c *Client) DeleteDatabase(id string) error {
	resp, err := c.doJWTRequest(http.MethodDelete, fmt.Sprintf("%s/edgedb/v2/databases/%s", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
