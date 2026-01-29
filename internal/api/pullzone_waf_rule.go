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
	"io"
	"net/http"
)

type PullzoneWafRuleConfiguration struct {
	ActionType          uint8                        `json:"actionType"`
	VariableTypes       map[string]string            `json:"variableTypes"`
	OperatorType        int64                        `json:"operatorType"`
	TransformationTypes []int64                      `json:"transformationTypes"`
	Value               string                       `json:"value"`
	RequestCount        int64                        `json:"requestCount"`
	BlockTime           int64                        `json:"blockTime"`
	Timeframe           int64                        `json:"timeframe"`
	ChainedRules        []PullzoneWafRuleChainedRule `json:"chainedRuleConditions"`
}

type PullzoneWafRuleChainedRule struct {
	VariableTypes map[string]string `json:"variableTypes"`
	OperatorType  int64             `json:"operatorType"`
	Value         string            `json:"value"`
}

type PullzoneWafRule struct {
	Id                int64                        `json:"id,omitempty"`
	PullzoneId        int64                        `json:"-"`
	ShieldZoneId      int64                        `json:"shieldZoneId"`
	Name              string                       `json:"ruleName"`
	Description       string                       `json:"ruleDescription"`
	RuleJson          string                       `json:"ruleJson"`
	RuleConfiguration PullzoneWafRuleConfiguration `json:"ruleConfiguration"`
}

func (c *Client) GetPullzoneWafRule(pullzoneId int64, ruleId int64) (PullzoneWafRule, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/waf/custom-rule/%d", c.apiUrl, ruleId), nil)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneWafRule{}, err
		} else {
			return PullzoneWafRule{}, errors.New("get WAF rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	result := PullzoneWafRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	result.PullzoneId = pullzoneId

	return result, nil
}

func (c *Client) CreatePullzoneWafRule(ctx context.Context, data PullzoneWafRule) (PullzoneWafRule, error) {
	shieldZoneId, err := c.GetPullzoneShieldIdByPullzone(data.PullzoneId)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	data.ShieldZoneId = shieldZoneId
	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/shield/waf/custom-rule", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return PullzoneWafRule{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneWafRule{}, err
		} else {
			return PullzoneWafRule{}, errors.New("create WAF rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	result := PullzoneWafRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	return c.GetPullzoneWafRule(data.PullzoneId, result.Id)
}

func (c *Client) UpdatePullzoneWafRule(data PullzoneWafRule) (PullzoneWafRule, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/waf/custom-rule/%d", c.apiUrl, data.Id), bytes.NewReader(body))
	if err != nil {
		return PullzoneWafRule{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneWafRule{}, err
		} else {
			return PullzoneWafRule{}, errors.New("create WAF rule failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	result := PullzoneWafRule{}
	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneWafRule{}, err
	}

	return c.GetPullzoneWafRule(data.PullzoneId, data.Id)
}

func (c *Client) DeletePullzoneWafRule(ruleId int64) error {
	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("%s/shield/waf/custom-rule/%d", c.apiUrl, ruleId), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return err
		} else {
			return errors.New("delete WAF rule failed with " + resp.Status)
		}
	}

	return nil
}
