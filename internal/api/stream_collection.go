package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StreamCollection{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StreamCollection{}, err
		}

		return StreamCollection{}, errors.New(obj.Message)
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
		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return StreamCollection{}, err
		}

		_ = resp.Body.Close()
		var obj struct {
			Message string `json:"Message"`
		}

		err = json.Unmarshal(bodyResp, &obj)
		if err != nil {
			return StreamCollection{}, err
		}

		return StreamCollection{}, errors.New(obj.Message)
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

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
