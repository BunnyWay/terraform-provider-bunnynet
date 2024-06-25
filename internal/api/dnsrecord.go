package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type DnsRecord struct {
	Zone                  int64   `json:"-"`
	Id                    int64   `json:"Id,omitempty"`
	Type                  uint8   `json:"Type"`
	Ttl                   int64   `json:"Ttl"`
	Value                 string  `json:"Value"`
	Name                  string  `json:"Name"`
	Weight                int64   `json:"Weight"`
	Priority              int64   `json:"Priority"`
	Port                  int64   `json:"Port"`
	Flags                 int64   `json:"Flags"`
	Tag                   string  `json:"Tag"`
	Accelerated           bool    `json:"Accelerated"`
	AcceleratedPullZoneId int64   `json:"AcceleratedPullZoneId"`
	LinkName              string  `json:"LinkName"`
	MonitorType           uint8   `json:"MonitorType"`
	GeolocationLatitude   float64 `json:"GeolocationLatitude"`
	GeolocationLongitude  float64 `json:"GeolocationLongitude"`
	LatencyZone           string  `json:"LatencyZone"`
	SmartRoutingType      uint8   `json:"SmartRoutingType"`
	Disabled              bool    `json:"Disabled"`
	Comment               string  `json:"Comment"`
}

func (c *Client) GetDnsRecord(zoneId int64, id int64) (DnsRecord, error) {
	zone, err := c.GetDnsZone(zoneId)
	if err != nil {
		return DnsRecord{}, err
	}

	for _, record := range zone.Records {
		if record.Id == id {
			record.Zone = zoneId
			return record, nil
		}
	}

	return DnsRecord{}, errors.New("DNS record not found")

}

func (c *Client) CreateDnsRecord(data DnsRecord) (DnsRecord, error) {
	dnsZoneId := data.Zone
	if dnsZoneId == 0 {
		return DnsRecord{}, errors.New("zone is required")
	}

	body, err := json.Marshal(data)
	if err != nil {
		return DnsRecord{}, err
	}

	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/dnszone/%d/records", c.apiUrl, dnsZoneId), bytes.NewReader(body))
	if err != nil {
		return DnsRecord{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return DnsRecord{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return DnsRecord{}, err
		}

		return DnsRecord{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return DnsRecord{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := DnsRecord{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	dataApiResult.Zone = dnsZoneId

	return dataApiResult, nil
}

func (c *Client) UpdateDnsRecord(dataApi DnsRecord) (DnsRecord, error) {
	id := dataApi.Id
	zoneId := dataApi.Zone

	body, err := json.Marshal(dataApi)
	if err != nil {
		return DnsRecord{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/dnszone/%d/records/%d", c.apiUrl, zoneId, id), bytes.NewReader(body))
	if err != nil {
		return DnsRecord{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return DnsRecord{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return DnsRecord{}, err
		}

		return DnsRecord{}, errors.New(obj.Message)
	}

	dataApiResult, err := c.GetDnsRecord(dataApi.Zone, id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteDnsRecord(zoneId int64, id int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/dnszone/%d/records/%d", c.apiUrl, zoneId, id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
