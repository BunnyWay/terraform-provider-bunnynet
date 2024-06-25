package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type PullzoneHostname struct {
	Id         int64  `json:"Id,omitempty"`
	ForceSSL   bool   `json:"ForceSSL"`
	Name       string `json:"Value"`
	PullzoneId int64  `json:"-"`
}

func (c *Client) CreatePullzoneHostname(data PullzoneHostname) (PullzoneHostname, error) {
	pullzoneId := data.PullzoneId
	if pullzoneId == 0 {
		return PullzoneHostname{}, errors.New("pullzone is required")
	}

	body, err := json.Marshal(map[string]string{
		"Hostname": data.Name,
	})

	if err != nil {
		return PullzoneHostname{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/addHostname", c.apiUrl, pullzoneId), bytes.NewReader(body))
	if err != nil {
		return PullzoneHostname{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		return PullzoneHostname{}, errors.New(resp.Status)
	}

	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	for _, hostname := range pullzone.Hostnames {
		if hostname.Name == data.Name {
			hostname.PullzoneId = pullzoneId
			hostname.ForceSSL = data.ForceSSL

			return c.UpdatePullzoneHostname(hostname)
		}
	}

	return PullzoneHostname{}, errors.New("Hostname not found")
}

func (c *Client) UpdatePullzoneHostname(data PullzoneHostname) (PullzoneHostname, error) {
	pullzoneId := data.PullzoneId
	if pullzoneId == 0 {
		return PullzoneHostname{}, errors.New("pullzone is required")
	}

	body, err := json.Marshal(map[string]interface{}{
		"ForceSSL": data.ForceSSL,
		"Hostname": data.Name,
	})

	if err != nil {
		return PullzoneHostname{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/setForceSSL", c.apiUrl, pullzoneId), bytes.NewReader(body))
	if err != nil {
		return PullzoneHostname{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		return PullzoneHostname{}, errors.New(resp.Status)
	}

	return c.GetPullzoneHostname(pullzoneId, data.Id)
}

func (c *Client) GetPullzoneHostname(pullzoneId int64, id int64) (PullzoneHostname, error) {
	pullzone, err := c.GetPullzone(pullzoneId)
	if err != nil {
		return PullzoneHostname{}, err
	}

	for _, hostname := range pullzone.Hostnames {
		if hostname.Id == id {
			hostname.PullzoneId = pullzoneId
			return hostname, nil
		}
	}

	return PullzoneHostname{}, errors.New("Hostname not found")
}

func (c *Client) DeletePullzoneHostname(pullzoneId int64, hostname string) error {
	body, err := json.Marshal(map[string]interface{}{
		"Hostname": hostname,
	})

	if err != nil {
		return err
	}

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/pullzone/%d/removeHostname", c.apiUrl, pullzoneId), bytes.NewReader(body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
