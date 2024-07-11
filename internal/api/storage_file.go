// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type StorageFile struct {
	Id          string
	Zone        int64
	Path        string
	Length      uint64
	ContentType string
	DateCreated string
	LastChanged string
	Checksum    string

	FileContents io.Reader
}

func (c *Client) GetStorageFile(zoneId int64, path string) (StorageFile, error) {
	zone, err := c.GetStorageZone(zoneId)
	if err != nil {
		return StorageFile{}, err
	}

	info, err := getStorageFileInfo(zone, path)
	if err != nil {
		return StorageFile{}, err
	}

	return info, nil
}

func (c *Client) CreateStorageFile(data StorageFile) (StorageFile, error) {
	zone, err := c.GetStorageZone(data.Zone)
	if err != nil {
		return StorageFile{}, err
	}

	body, err := io.ReadAll(data.FileContents)
	if err != nil {
		return StorageFile{}, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s/%s/%s", zone.StorageHostname, zone.Name, data.Path), bytes.NewReader(body))
	if err != nil {
		return StorageFile{}, err
	}

	checksum, err := storageFileGenerateChecksum(body)
	if err != nil {
		return StorageFile{}, err
	}

	req.Header.Add("AccessKey", zone.Password)
	req.Header.Add("Checksum", checksum)
	if len(data.ContentType) > 0 {
		req.Header.Add("Override-Content-Type", data.ContentType)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return StorageFile{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return StorageFile{}, errors.New(resp.Status)
	}

	return c.GetStorageFile(data.Zone, data.Path)
}

func (c *Client) UpdateStorageFile(data StorageFile) (StorageFile, error) {
	return c.CreateStorageFile(data)
}

func (c *Client) DeleteStorageFile(zoneId int64, path string) error {
	zone, err := c.GetStorageZone(zoneId)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://%s/%s/%s", zone.StorageHostname, zone.Name, path), nil)
	if err != nil {
		return err
	}

	req.Header.Add("AccessKey", zone.Password)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func getStorageFileContents(zone StorageZone, path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/%s/%s", zone.StorageHostname, zone.Name, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("AccessKey", zone.Password)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_ = resp.Body.Close()
	return bodyResp, nil
}

func getStorageFileInfo(zone StorageZone, path string) (StorageFile, error) {
	req, err := http.NewRequest("DESCRIBE", fmt.Sprintf("https://%s/%s/%s", zone.StorageHostname, zone.Name, path), nil)
	if err != nil {
		return StorageFile{}, err
	}

	req.Header.Add("AccessKey", zone.Password)
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return StorageFile{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return StorageFile{}, err
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return StorageFile{}, err
	}

	_ = resp.Body.Close()
	var obj struct {
		Guid            string `json:"Guid"`
		StorageZoneName string `json:"StorageZoneName"`
		Path            string `json:"Path"`
		ObjectName      string `json:"ObjectName"`
		Length          int    `json:"Length"`
		LastChanged     string `json:"LastChanged"`
		ServerId        int    `json:"ServerId"`
		ArrayNumber     int    `json:"ArrayNumber"`
		IsDirectory     bool   `json:"IsDirectory"`
		UserId          string `json:"UserId"`
		ContentType     string `json:"ContentType"`
		DateCreated     string `json:"DateCreated"`
		StorageZoneId   int    `json:"StorageZoneId"`
		Checksum        string `json:"Checksum"`
		ReplicatedZones string `json:"ReplicatedZones"`
	}

	err = json.Unmarshal(bodyResp, &obj)
	if err != nil {
		return StorageFile{}, err
	}

	dataResult := StorageFile{
		Id:          obj.Guid,
		Zone:        int64(obj.StorageZoneId),
		Path:        path,
		Length:      uint64(obj.Length),
		ContentType: obj.ContentType,
		DateCreated: obj.DateCreated,
		LastChanged: obj.LastChanged,
		Checksum:    obj.Checksum,
	}

	return dataResult, nil
}

func storageFileGenerateChecksum(content []byte) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write(content)
	if err != nil {
		return "", err
	}

	return strings.ToUpper(fmt.Sprintf("%x", hasher.Sum(nil))), nil
}
