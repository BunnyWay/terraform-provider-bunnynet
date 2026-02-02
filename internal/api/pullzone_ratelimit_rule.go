// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
)

type PullzoneRatelimitRuleConfiguration struct {
	ActionType          uint8                              `json:"actionType"`
	VariableTypes       map[string]string                  `json:"variableTypes"`
	OperatorType        int64                              `json:"operatorType"`
	TransformationTypes []int64                            `json:"transformationTypes"`
	Value               string                             `json:"value"`
	RequestCount        int64                              `json:"requestCount"`
	BlockTime           int64                              `json:"blockTime"`
	Timeframe           int64                              `json:"timeframe"`
	ChainedRules        []PullzoneRatelimitRuleChainedRule `json:"chainedRuleConditions"`
}

type PullzoneRatelimitRuleChainedRule struct {
	VariableTypes map[string]string `json:"variableTypes"`
	OperatorType  int64             `json:"operatorType"`
	Value         string            `json:"value"`
}

type PullzoneRatelimitRule struct {
	Id                int64                              `json:"id,omitempty"`
	PullzoneId        int64                              `json:"-"`
	ShieldZoneId      int64                              `json:"shieldZoneId"`
	Name              string                             `json:"ruleName"`
	Description       string                             `json:"ruleDescription"`
	RuleJson          string                             `json:"ruleJson"`
	RuleConfiguration PullzoneRatelimitRuleConfiguration `json:"ruleConfiguration"`
}

func (c *Client) GetPullzoneRatelimitRule(pullzoneId int64, ruleId int64) (PullzoneRatelimitRule, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/rate-limit/%d", c.apiUrl, ruleId), nil)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return PullzoneRatelimitRule{}, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneRatelimitRule{}, err
		} else {
			return PullzoneRatelimitRule{}, errors.New("get ratelimit rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	result := PullzoneRatelimitRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	result.PullzoneId = pullzoneId

	return result, nil
}

func (c *Client) CreatePullzoneRatelimitRule(ctx context.Context, data PullzoneRatelimitRule) (PullzoneRatelimitRule, error) {
	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(data.PullzoneId)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	data.ShieldZoneId = shieldZoneId
	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	tflog.Info(ctx, fmt.Sprintf("POST /shield/rate-limit: %s", string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/shield/rate-limit", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneRatelimitRule{}, err
		} else {
			return PullzoneRatelimitRule{}, errors.New("create ratelimit rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	result := PullzoneRatelimitRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	return c.GetPullzoneRatelimitRule(data.PullzoneId, result.Id)
}

func (c *Client) UpdatePullzoneRatelimitRule(data PullzoneRatelimitRule) (PullzoneRatelimitRule, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/rate-limit/%d", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneRatelimitRule{}, err
		} else {
			return PullzoneRatelimitRule{}, errors.New("create ratelimit rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	result := PullzoneRatelimitRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneRatelimitRule{}, err
	}

	return c.GetPullzoneRatelimitRule(data.PullzoneId, data.Id)
}

func (c *Client) DeletePullzoneRatelimitRule(ruleId int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/shield/rate-limit/%d", c.apiUrl, ruleId), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return err
		} else {
			return errors.New("delete ratelimit rule failed with " + resp.Status)
		}
	}

	return nil
}
