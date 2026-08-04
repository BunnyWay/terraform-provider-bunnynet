// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) PurgeUrl(purgeUrl string, async bool, exactPath bool) error {
	query := url.Values{}
	query.Set("url", purgeUrl)
	query.Set("async", strconv.FormatBool(async))
	query.Set("exactPath", strconv.FormatBool(exactPath))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/purge?%s", c.apiUrl, query.Encode()), nil)
	if err != nil {
		return err
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}
