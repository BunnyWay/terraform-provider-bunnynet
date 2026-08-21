// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) UrlPurgeCache(ctx context.Context, urlToPurge string, exactPath bool) error {
	query := url.Values{
		"url":       []string{urlToPurge},
		"async":     []string{"true"},
		"exactPath": []string{strconv.FormatBool(exactPath)},
	}

	queryEncoded := query.Encode()
	tflog.Info(ctx, fmt.Sprintf("POST /purge?%s", queryEncoded))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/purge?%s", c.apiUrl, queryEncoded), nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	err = extractErrorMessage(resp)
	if err != nil {
		return err
	}

	return fmt.Errorf("Purge URL failed with %s", resp.Status)
}
