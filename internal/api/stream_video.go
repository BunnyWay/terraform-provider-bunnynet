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

type StreamVideo struct {
	Id        string               `json:"guid,omitempty"`
	LibraryId int64                `json:"videoLibraryId"`
	Title     string               `json:"title"`
	MetaTags  []StreamVideoMetaTag `json:"metaTags"`
	Chapters  []StreamVideoChapter `json:"chapters"`
	Moments   []StreamVideoMoment  `json:"moments"`
}

type StreamVideoMetaTag struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

type StreamVideoMoment struct {
	Label     string `json:"label"`
	Timestamp uint64 `json:"timestamp"`
}

type StreamVideoChapter struct {
	Title string `json:"title"`
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

func (c *Client) GetStreamVideo(libraryId int64, id string) (StreamVideo, error) {
	var data StreamVideo

	library, err := c.GetStreamLibrary(libraryId)
	if err != nil {
		return data, err
	}

	resp, err := c.doStreamRequest(library, http.MethodGet, fmt.Sprintf("videos/%s", id), nil)
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

func (c *Client) UpdateStreamVideo(dataApi StreamVideo) (StreamVideo, error) {
	id := dataApi.Id

	body, err := json.Marshal(dataApi)
	if err != nil {
		return StreamVideo{}, err
	}

	library, err := c.GetStreamLibrary(dataApi.LibraryId)
	if err != nil {
		return StreamVideo{}, err
	}

	resp, err := c.doStreamRequest(library, http.MethodPost, fmt.Sprintf("videos/%s", id), bytes.NewReader(body))
	if err != nil {
		return StreamVideo{}, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StreamVideo{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StreamVideo{}, err
		}

		return StreamVideo{}, errors.New(obj.Message)
	}

	dataApiResult, err := c.GetStreamVideo(dataApi.LibraryId, id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStreamVideo(libraryId int64, id string) error {
	library, err := c.GetStreamLibrary(libraryId)
	if err != nil {
		return err
	}

	resp, err := c.doStreamRequest(library, http.MethodDelete, fmt.Sprintf("videos/%s", id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
