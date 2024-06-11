package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Storagezone struct {
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

func (c *Client) GetStoragezone(id int64) (Storagezone, error) {
	var data Storagezone
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
	if err != nil {
		return data, err
	}

	req.Header.Add("AccessKey", c.apiKey)

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
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

func (c *Client) CreateStoragezone(data Storagezone) (Storagezone, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return Storagezone{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/storagezone", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return Storagezone{}, err
	}

	req.Header.Add("AccessKey", c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return Storagezone{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return Storagezone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return Storagezone{}, err
		}

		return Storagezone{}, errors.New(obj.Message)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Storagezone{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := Storagezone{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdateStoragezone(dataApi Storagezone) (Storagezone, error) {
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
		return Storagezone{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), bytes.NewReader(body))
	if err != nil {
		return Storagezone{}, err
	}

	req.Header.Add("AccessKey", c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return Storagezone{}, err
	}

	if resp.StatusCode != http.StatusNoContent {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return Storagezone{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return Storagezone{}, err
		}

		return Storagezone{}, errors.New(obj.Message)
	}

	dataApiResult, err := c.GetStoragezone(id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStoragezone(id int64) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/storagezone/%d", c.apiUrl, id), nil)
	if err != nil {
		return err
	}

	req.Header.Add("AccessKey", c.apiKey)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
