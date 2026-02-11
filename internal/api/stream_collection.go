// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"io"
	"net/http"
)

type StreamCollection struct {
	Id        string `json:"guid,omitempty"`
	Name      string `json:"Name"`
	LibraryId int64  `json:"videoLibraryId"`
}

func (c *Client) GetStreamCollection(libraryId int64, id string) (StreamCollection, error) {
	var data StreamCollection

	library, err := c.GetStreamLibrary(libraryId)
	if err != nil {
		return data, err
	}

	resp, err := c.doStreamRequest(library, http.MethodGet, fmt.Sprintf("collections/%s", id), nil)
	if err != nil {
		return data, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return data, ErrNotFound
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

func (c *Client) CreateStreamCollection(dataApi StreamCollection) (StreamCollection, error) {
	body, err := json.Marshal(dataApi)
	if err != nil {
		return StreamCollection{}, err
	}

	library, err := c.GetStreamLibrary(dataApi.LibraryId)
	if err != nil {
		return StreamCollection{}, err
	}

	resp, err := c.doStreamRequest(library, http.MethodPost, "collections", bytes.NewReader(body))
	if err != nil {
		return StreamCollection{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return StreamCollection{}, err
		} else {
			return StreamCollection{}, errors.New("create stream collection failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return StreamCollection{}, err
	}
	_ = resp.Body.Close()

	dataApiResult := StreamCollection{}
	err = json.Unmarshal(bodyResp, &dataApiResult)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) UpdateStreamCollection(dataApi StreamCollection) (StreamCollection, error) {
	id := dataApi.Id

	body, err := json.Marshal(dataApi)
	if err != nil {
		return StreamCollection{}, err
	}

	library, err := c.GetStreamLibrary(dataApi.LibraryId)
	if err != nil {
		return StreamCollection{}, err
	}

	resp, err := c.doStreamRequest(library, http.MethodPost, fmt.Sprintf("collections/%s", id), bytes.NewReader(body))
	if err != nil {
		return StreamCollection{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return StreamCollection{}, err
		} else {
			return StreamCollection{}, errors.New("update stream collection failed with " + resp.Status)
		}
	}

	dataApiResult, err := c.GetStreamCollection(dataApi.LibraryId, id)
	if err != nil {
		return dataApiResult, err
	}

	return dataApiResult, nil
}

func (c *Client) DeleteStreamCollection(libraryId int64, id string) error {
	library, err := c.GetStreamLibrary(libraryId)
	if err != nil {
		return err
	}

	resp, err := c.doStreamRequest(library, http.MethodDelete, fmt.Sprintf("collections/%s", id), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
