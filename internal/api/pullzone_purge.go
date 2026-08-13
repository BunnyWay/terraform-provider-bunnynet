// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

func (c *Client) PullzonePurgeCache(ctx context.Context, id int64, tag string) error {
	var body io.Reader
	var bodyBytes []byte
	var err error

	if tag != "" {
		bodyBytes, err = json.Marshal(map[string]string{
			"CacheTag": tag,
		})

		if err != nil {
			return err
		}

		body = bytes.NewReader(bodyBytes)
	} else {
		bodyBytes = []byte("null")
	}

	tflog.Info(ctx, fmt.Sprintf("POST /pullzone/%d/purgeCache: %s", id, string(bodyBytes)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/pullzone/%d/purgeCache", c.apiUrl, id), body)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	err = utils.ExtractErrorMessage(resp)
	if err != nil {
		return err
	}

	return fmt.Errorf("PurgeCache failed with %s", resp.Status)
}
