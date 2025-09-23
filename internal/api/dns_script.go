package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

func (c *Client) UpdateDnsScript(ctx context.Context, data ComputeScript, previousData ComputeScript) (ComputeScript, error) {
	id := data.Id

	dataApiResult, err := c.UpdateComputeScript(ctx, data, previousData)
	if err != nil {
		return dataApiResult, err
	}

	// publish script
	{
		body, err := json.Marshal(map[string]string{
			"Note": "",
		})

		if err != nil {
			return ComputeScript{}, err
		}

		tflog.Debug(ctx, fmt.Sprintf("POST /compute/script/%d/publish: %+v", id, string(body)))

		resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/compute/script/%d/publish", c.apiUrl, id), bytes.NewReader(body))
		if err != nil {
			return ComputeScript{}, err
		}

		if resp.StatusCode != http.StatusNoContent {
			bodyResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return ComputeScript{}, err
			}

			_ = resp.Body.Close()
			var obj struct {
				Message string `json:"Message"`
			}

			err = json.Unmarshal(bodyResp, &obj)
			if err != nil {
				return ComputeScript{}, err
			}

			return ComputeScript{}, errors.New(obj.Message)
		}

		_ = resp.Body.Close()
	}

	return c.GetComputeScript(ctx, id)
}
