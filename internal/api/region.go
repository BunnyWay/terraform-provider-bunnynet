// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
)

type Region struct {
	Id                  int64   `json:"Id"`
	Name                string  `json:"Name"`
	PricePerGigabyte    float64 `json:"PricePerGigabyte"`
	RegionCode          string  `json:"RegionCode"`
	ContinentCode       string  `json:"ContinentCode"`
	CountryCode         string  `json:"CountryCode"`
	Latitude            float64 `json:"Latitude"`
	Longitude           float64 `json:"Longitude"`
	AllowLatencyRouting bool    `json:"AllowLatencyRouting"`
}

func (c *Client) GetRegions() ([]Region, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/region", c.apiUrl), nil)
	if err != nil {
		return []Region{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []Region{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Region{}, err
	}

	_ = resp.Body.Close()
	var regions []Region

	err = json.Unmarshal(bodyResp, &regions)
	if err != nil {
		return []Region{}, err
	}

	slices.SortStableFunc(regions, func(a, b Region) int {
		return int(a.Id - b.Id)
	})

	return regions, nil
}

func (c *Client) GetRegion(regionCode string) (Region, error) {
	regions, err := c.GetRegions()
	if err != nil {
		return Region{}, err
	}

	for _, region := range regions {
		if region.RegionCode == regionCode {
			return region, nil
		}
	}

	return Region{}, errors.New("region not found")
}
