// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrNotFound = errors.New("resource not found")
var ErrShieldPlanChangeBlocked = errors.New("The Shield Plan cannot be changed, using features not available in the new plan.")
var ErrShieldWafLimitReached = errors.New("The limit for Custom WAF Rules was reached.")

func extractErrorMessage(response *http.Response) error {
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

func extractDatabaseErrorMessage(response *http.Response) error {
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

func extractMCErrorMessage(response *http.Response) error {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	_ = response.Body.Close()

	var responseObj []struct {
		Message string `json:"message"`
	}

	err = json.Unmarshal(bodyBytes, &responseObj)
	if err != nil {
		var responseMessage string
		err = json.Unmarshal(bodyBytes, &responseMessage)
		if err != nil {
			return nil
		}

		return errors.New(responseMessage)
	}

	if len(responseObj) == 0 {
		return nil
	}

	return errors.New(responseObj[0].Message)
}

func extractShieldErrorMessage(response *http.Response) error {
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
		ErrorResponse struct {
			Message  string `json:"message"`
			ErrorKey string `json:"errorKey"`
		} `json:"errorResponse"`
		ErrorKey string `json:"errorKey"`
	}

	err = json.Unmarshal(bodyBytes, &responseObj)
	if err != nil {
		return nil
	}

	errorKey := responseObj.ErrorKey

	if responseObj.Error.ErrorKey != "" {
		errorKey = responseObj.Error.ErrorKey
	}

	if responseObj.ErrorResponse.ErrorKey != "" {
		errorKey = responseObj.ErrorResponse.ErrorKey
	}

	switch errorKey {
	case "plan_change_blocked.shieldzone":
		return ErrShieldPlanChangeBlocked
	case "limit_reached.waf":
		return ErrShieldWafLimitReached
	case "not_found_or_unauthorised_access.waf_rule":
		return ErrNotFound
	}

	return errors.New(errorKey)
}
