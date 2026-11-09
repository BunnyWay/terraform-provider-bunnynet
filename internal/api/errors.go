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

	return messageError(responseObj.Message)
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

	return messageError(responseObj.Message)
}

func extractMCErrorMessage(response *http.Response) error {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil
	}

	_ = response.Body.Close()

	{
		var responseObj []struct {
			Message string `json:"message"`
		}

		err = json.Unmarshal(bodyBytes, &responseObj)
		if err == nil {
			if len(responseObj) == 0 {
				return nil
			}

			return messageError(responseObj[0].Message)
		}
	}

	{
		var responseObj struct {
			Detail string `json:"detail"`
		}

		err = json.Unmarshal(bodyBytes, &responseObj)
		if err == nil && len(responseObj.Detail) > 0 {
			return errors.New(responseObj.Detail)
		}
	}

	var responseMessage string
	err = json.Unmarshal(bodyBytes, &responseMessage)
	if err != nil {
		return nil
	}

	return messageError(responseMessage)
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
	case "not_found.access_list":
		return ErrNotFound
	}

	return messageError(errorKey)
}

// messageError returns nil for an empty message, so callers fall back to
// their "... failed with <status>" error instead of reporting an empty error.
func messageError(message string) error {
	if message == "" {
		return nil
	}

	return errors.New(message)
}
