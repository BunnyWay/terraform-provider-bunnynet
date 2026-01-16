// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func SliceDiff[T comparable](s1 []T, s2 []T) []T {
	diff := make([]T, 0)

	for _, v1 := range s1 {
		found := false
		for _, v2 := range s2 {
			if v1 == v2 {
				found = true
				break
			}
		}

		if !found {
			diff = append(diff, v1)
		}
	}

	return diff
}

func ExtractErrorMessage(response *http.Response) error {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	_ = response.Body.Close()
	var responseObj struct {
		Message string `json:"Message"`
	}

	err = json.Unmarshal(bodyBytes, &responseObj)
	if err != nil {
		return nil
	}

	return errors.New(responseObj.Message)
}

func ExtractDatabaseErrorMessage(response *http.Response) error {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	_ = response.Body.Close()
	var responseObj struct {
		Message string `json:"error"`
	}

	err = json.Unmarshal(bodyBytes, &responseObj)
	if err != nil {
		return nil
	}

	return errors.New(responseObj.Message)
}

func ExtractShieldErrorMessage(response *http.Response) error {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	_ = response.Body.Close()
	var responseObj struct {
		Error struct {
			Message  string `json:"message"`
			ErrorKey string `json:"errorKey"`
		} `json:"error"`
	}

	err = json.Unmarshal(bodyBytes, &responseObj)
	if err != nil {
		return nil
	}

	return errors.New(responseObj.Error.ErrorKey)
}

func MapInvert[k comparable, v comparable](m map[k]v) map[v]k {
	result := make(map[v]k, len(m))
	for key, value := range m {
		result[value] = key
	}
	return result
}
