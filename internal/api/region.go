package api

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (c *Client) GetRegion(regionCode string) (Region, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/region", c.apiUrl), nil)
	if err != nil {
		return Region{}, err
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return Region{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Region{}, errors.New(resp.Status)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Region{}, err
	}

	_ = resp.Body.Close()
	var regions []Region

	err = json.Unmarshal(bodyResp, &regions)
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
