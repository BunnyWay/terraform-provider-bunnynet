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

type DnsZone struct {
	Id                            int64       `json:"Id,omitempty"`
	Domain                        string      `json:"Domain"`
	CustomNameserversEnabled      bool        `json:"CustomNameserversEnabled"`
	Nameserver1                   string      `json:"Nameserver1"`
	Nameserver2                   string      `json:"Nameserver2"`
	SoaEmail                      string      `json:"SoaEmail"`
	LoggingEnabled                bool        `json:"LoggingEnabled"`
	LoggingIPAnonymizationEnabled bool        `json:"LoggingIPAnonymizationEnabled"`
	LogAnonymizationType          uint8       `json:"LogAnonymizationType"`
	Records                       []DnsRecord `json:"Records"`
}

func (c *Client) GetDnsZone(ctx context.Context, id int64) (DnsZone, error) {
	var data DnsZone
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/dnszone/%d", c.apiUrl, id), nil)
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

	tflog.Debug(ctx, fmt.Sprintf("GET /dnszone/%d: %+v", id, string(bodyResp)))

	_ = resp.Body.Close()
	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func (c *Client) GetDnsZoneByDomain(ctx context.Context, domain string) (DnsZone, error) {
	var data DnsZone
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/dnszone", c.apiUrl), nil)
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
	var result struct {
		Items []DnsZone `json:"Items"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return data, err
	}

	for _, record := range result.Items {
		if record.Domain == domain {
			return record, nil
		}
	}

	return data, fmt.Errorf("DNS zone \"%s\" not found", domain)
}

func (c *Client) CreateDnsZone(ctx context.Context, data DnsZone) (DnsZone, error) {
	body, err := json.Marshal(map[string]string{
		"Domain": data.Domain,
	})
	if err != nil {
		return DnsZone{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /dnszone: %+v", string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/dnszone", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return DnsZone{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return DnsZone{}, err
		} else {
			return DnsZone{}, errors.New("create DNS zone failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return DnsZone{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := DnsZone{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	data.Id = dataApiResult.Id
	dataApiResult, err = c.UpdateDnsZone(ctx, data)
	if err != nil {
		_ = c.DeleteDnsZone(ctx, data.Id)
	}

	return dataApiResult, err
}

func (c *Client) UpdateDnsZone(ctx context.Context, dataApi DnsZone) (DnsZone, error) {
	id := dataApi.Id

	body, err := json.Marshal(dataApi)
	if err != nil {
		return DnsZone{}, err
	}

	tflog.Debug(ctx, fmt.Sprintf("POST /dnszone/%d: %+v", id, string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/dnszone/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return DnsZone{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return DnsZone{}, err
		} else {
			return DnsZone{}, errors.New("update DNS zone failed with " + resp.Status)
		}
	}

	dataApiResult, err := c.GetDnsZone(ctx, id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteDnsZone(ctx context.Context, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/dnszone/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	tflog.Debug(ctx, fmt.Sprintf("DELETE /dnszone/%d: %s", id, resp.Status))

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
