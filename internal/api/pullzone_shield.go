// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzoneshieldresourcevalidator"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type PullzoneShieldWafEngineConfigItem struct {
	Name         string `json:"name"`
	ValueEncoded string `json:"valueEncoded"`
}

type PullzoneShieldAccessList struct {
	Id        int64  `json:"-"`
	Name      string `json:"-"`
	Action    uint8  `json:"-"`
	IsEnabled bool   `json:"-"`
}

type PullzoneShield struct {
	Id                                   int64    `json:"shieldZoneId,omitempty"`
	PullzoneId                           int64    `json:"pullZoneId"`
	PlanType                             uint8    `json:"planType"`
	BotDetectionMode                     uint8    `json:"-"`
	BotDetectionFingerprintSensitivity   uint8    `json:"-"`
	BotDetectionFingerprintAggression    uint8    `json:"-"`
	BotDetectionIPSensitivity            uint8    `json:"-"`
	BotDetectionRequestIntegrity         uint8    `json:"-"`
	BotDetectionComplexFingerprinting    bool     `json:"-"`
	WhiteLabelResponsePages              bool     `json:"whitelabelResponsePages"`
	DDoSMode                             uint8    `json:"dDoSExecutionMode"`
	DDoSLevel                            uint8    `json:"dDoSShieldSensitivity"`
	DDosChallengeWindow                  int64    `json:"dDoSChallengeWindow"`
	WafEnabled                           bool     `json:"wafEnabled"`
	WafMode                              uint8    `json:"wafExecutionMode"`
	WafRealtimeThreatIntelligenceEnabled bool     `json:"wafRealtimeThreatIntelligenceEnabled"`
	WafLogHeaders                        bool     `json:"wafRequestHeaderLoggingEnabled"`
	WafLogHeadersExcluded                []string `json:"wafRequestIgnoredHeaders"`
	WafAllowedHttpVersions               []string `json:"-"`
	WafAllowedHttpMethods                []string `json:"-"`
	WafAllowedRequestContentTypes        []string `json:"-"`
	WafRuleSensitivityBlocking           uint8    `json:"-"`
	WafRuleSensitivityDetection          uint8    `json:"-"`
	WafRuleSensitivityExecution          uint8    `json:"-"`
	WafRequestBodyLimitAction            uint8    `json:"wafRequestBodyLimitAction"`
	WafResponseBodyLimitAction           uint8    `json:"wafResponseBodyLimitAction"`
	WafRulesDisabled                     []string `json:"wafDisabledRules"`
	WafRulesLogonly                      []string `json:"wafLogOnlyRules"`

	AccessLists     []PullzoneShieldAccessList          `json:"-"`
	WafEngineConfig []PullzoneShieldWafEngineConfigItem `json:"wafEngineConfig,omitempty"`
}

type PullzoneShieldWafEngineConfig struct {
	AllowedHttpVersions        map[string]struct{}
	AllowedHttpMethods         map[string]struct{}
	AllowedRequestContentTypes map[string]struct{}
	RuleSensitivityBlocking    uint8
	RuleSensitivityDetection   uint8
	RuleSensitivityExecution   uint8
}

func (c *Client) GetPullzoneShieldDefaultWafEngineConfig() (PullzoneShieldWafEngineConfig, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/waf/engine-config", c.apiUrl), nil)
	if err != nil {
		return PullzoneShieldWafEngineConfig{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneShieldWafEngineConfig{}, err
		} else {
			return PullzoneShieldWafEngineConfig{}, errors.New("get shield engine config failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneShieldWafEngineConfig{}, err
	}

	var result struct {
		Data []PullzoneShieldWafEngineConfigItem `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneShieldWafEngineConfig{}, err
	}

	config := PullzoneShieldWafEngineConfig{}

	for _, item := range result.Data {
		if item.Name == "allowed_http_versions" {
			config.AllowedHttpVersions = utils.SliceToSet(strings.Split(item.ValueEncoded, " "))
		}

		if item.Name == "allowed_methods" {
			config.AllowedHttpMethods = utils.SliceToSet(strings.Split(item.ValueEncoded, " "))
		}

		if item.Name == "allowed_request_content_type" {
			var items []string
			re := regexp.MustCompile(`(\|([^\|]+)\|)`)
			matches := re.FindAllStringSubmatch(item.ValueEncoded, -1)
			for _, item := range matches {
				items = append(items, item[2])
			}

			config.AllowedRequestContentTypes = utils.SliceToSet(items)
		}

		if item.Name == "detection_paranoia_level" {
			level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
			if err != nil {
				return config, err
			}

			config.RuleSensitivityDetection = uint8(level)
		}

		if item.Name == "executing_paranoia_level" {
			level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
			if err != nil {
				return config, err
			}

			config.RuleSensitivityExecution = uint8(level)
		}

		if item.Name == "blocking_paranoia_level" {
			level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
			if err != nil {
				return config, err
			}

			config.RuleSensitivityBlocking = uint8(level)
		}
	}

	return config, nil
}

func (c *Client) GetPullzoneShieldIdByPullzone(pullzoneId int64) (int64, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/get-by-pullzone/%d", c.apiUrl, pullzoneId), nil)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return 0, err
		} else {
			return 0, errors.New("get shieldzone for pullzone failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Data struct {
			Id         int64 `json:"shieldZoneId"`
			PullzoneId int64 `json:"pullZoneId"`
		} `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return 0, err
	}

	return result.Data.Id, nil
}

func (c *Client) GetPullzoneShield(ctx context.Context, id int64) (PullzoneShield, error) {
	var result struct {
		Data PullzoneShield `json:"data"`
	}

	// fetch basic config
	{
		resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/%d", c.apiUrl, id), nil)
		if err != nil {
			return PullzoneShield{}, err
		}

		if resp.StatusCode != http.StatusOK {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneShield{}, err
			} else {
				return PullzoneShield{}, errors.New("get shieldzone for pullzone failed with " + resp.Status)
			}
		}

		bodyResp, err := io.ReadAll(resp.Body)
		if err != nil {
			return PullzoneShield{}, err
		}

		tflog.Warn(ctx, fmt.Sprintf("GET /shield/shield-zone/%d: %+v", id, string(bodyResp)))

		err = json.Unmarshal(bodyResp, &result)
		if err != nil {
			return PullzoneShield{}, err
		}

		engineConfigDefaults, err := c.GetPullzoneShieldDefaultWafEngineConfig()
		if err != nil {
			return PullzoneShield{}, err
		}

		// copy defaults into result
		result.Data.WafAllowedHttpMethods = utils.SetToSlice(engineConfigDefaults.AllowedHttpMethods)
		result.Data.WafAllowedHttpVersions = utils.SetToSlice(engineConfigDefaults.AllowedHttpVersions)
		result.Data.WafAllowedRequestContentTypes = utils.SetToSlice(engineConfigDefaults.AllowedRequestContentTypes)

		// override defaults with pullzone data
		for _, item := range result.Data.WafEngineConfig {
			if item.Name == "allowed_http_versions" {
				result.Data.WafAllowedHttpVersions = strings.Split(item.ValueEncoded, " ")
			}

			if item.Name == "allowed_methods" {
				result.Data.WafAllowedHttpMethods = strings.Split(item.ValueEncoded, " ")
			}

			if item.Name == "allowed_request_content_type" {
				var items []string
				re := regexp.MustCompile(`(\|([^\|]+)\|)`)
				matches := re.FindAllStringSubmatch(item.ValueEncoded, -1)
				for _, item := range matches {
					items = append(items, item[2])
				}

				result.Data.WafAllowedRequestContentTypes = items
			}

			if item.Name == "detection_paranoia_level" {
				level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
				if err != nil {
					return PullzoneShield{}, err
				}

				result.Data.WafRuleSensitivityDetection = uint8(level)
			}

			if item.Name == "executing_paranoia_level" {
				level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
				if err != nil {
					return PullzoneShield{}, err
				}

				result.Data.WafRuleSensitivityExecution = uint8(level)
			}

			if item.Name == "blocking_paranoia_level" {
				level, err := strconv.ParseInt(item.ValueEncoded, 10, 64)
				if err != nil {
					return PullzoneShield{}, err
				}

				result.Data.WafRuleSensitivityBlocking = uint8(level)
			}
		}
	}

	// fetch managed access lists
	{
		lists, err := c.getPullzoneAccessLists(ctx, id, PullzoneAccessListQueryCurated)
		if err != nil {
			return result.Data, err
		}

		var resultLists []PullzoneShieldAccessList
		for _, list := range lists {
			if !list.IsEnabled {
				continue
			}

			resultLists = append(resultLists, PullzoneShieldAccessList{
				Id:        list.ListId,
				Action:    list.Action,
				Name:      list.Name,
				IsEnabled: list.IsEnabled,
			})
		}

		result.Data.AccessLists = resultLists
	}

	// fetch bot-detection config
	{
		botDetectionResult, err := c.fetchBotDetection(ctx, id)
		if err != nil {
			return PullzoneShield{}, err
		}

		result.Data.BotDetectionMode = botDetectionResult.Mode
		result.Data.BotDetectionFingerprintSensitivity = botDetectionResult.FingerprintSensitivity
		result.Data.BotDetectionFingerprintAggression = botDetectionResult.FingerprintAggression
		result.Data.BotDetectionIPSensitivity = botDetectionResult.IPSensitivity
		result.Data.BotDetectionRequestIntegrity = botDetectionResult.RequestIntegrity
		result.Data.BotDetectionComplexFingerprinting = botDetectionResult.ComplexFingerprinting
	}

	return result.Data, nil
}

type fetchBotDetectionResult struct {
	Mode                   uint8
	FingerprintSensitivity uint8
	FingerprintAggression  uint8
	IPSensitivity          uint8
	RequestIntegrity       uint8
	ComplexFingerprinting  bool
}

func (c *Client) fetchBotDetection(ctx context.Context, shieldZoneId int64) (fetchBotDetectionResult, error) {
	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("%s/shield/shield-zone/%d/bot-detection", c.apiUrl, shieldZoneId), nil)
	if err != nil {
		return fetchBotDetectionResult{}, err
	}

	if resp.StatusCode == http.StatusAccepted {
		err := utils.ExtractShieldErrorMessage(resp)
		if err != nil {
			if err.Error() == "invalid_plan_type.bot_detection" {
				return fetchBotDetectionResult{
					Mode:                   0,
					FingerprintSensitivity: 0,
					FingerprintAggression:  1,
					IPSensitivity:          0,
					RequestIntegrity:       0,
					ComplexFingerprinting:  false,
				}, nil
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return fetchBotDetectionResult{}, err
		} else {
			return fetchBotDetectionResult{}, errors.New("get shieldzone/bot-detection for pullzone failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return fetchBotDetectionResult{}, err
	}

	tflog.Warn(ctx, fmt.Sprintf("GET /shield/shield-zone/%d/bot-detection: %+v", shieldZoneId, string(bodyResp)))

	var result struct {
		Data struct {
			ShieldZoneId     int64 `json:"shieldZoneId"`
			ExecutionMode    uint8 `json:"executionMode"`
			RequestIntegrity struct {
				Sensitivity uint8 `json:"sensitivity"`
			} `json:"requestIntegrity"`
			IpAddress struct {
				Sensitivity uint8 `json:"sensitivity"`
			} `json:"ipAddress"`
			BrowserFingerprint struct {
				Sensitivity    uint8 `json:"sensitivity"`
				Aggression     uint8 `json:"aggression"`
				ComplexEnabled bool  `json:"complexEnabled"`
			} `json:"browserFingerprint"`
		} `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return fetchBotDetectionResult{}, err
	}

	return fetchBotDetectionResult{
		Mode:                   result.Data.ExecutionMode,
		FingerprintSensitivity: result.Data.BrowserFingerprint.Sensitivity,
		FingerprintAggression:  result.Data.BrowserFingerprint.Aggression,
		IPSensitivity:          result.Data.IpAddress.Sensitivity,
		RequestIntegrity:       result.Data.RequestIntegrity.Sensitivity,
		ComplexFingerprinting:  result.Data.BrowserFingerprint.ComplexEnabled,
	}, nil
}

func (c *Client) CreatePullzoneShield(ctx context.Context, data PullzoneShield) (PullzoneShield, error) {
	id, err := c.GetPullzoneShieldIdByPullzone(data.PullzoneId)
	if err == nil {
		data.Id = id
		return c.UpdatePullzoneShield(ctx, data)
	}

	wafEngineConfig, err := c.convertPullzoneShieldWafEngineConfigToBody(data)
	if err != nil {
		return PullzoneShield{}, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"pullZoneId": data.PullzoneId,
		"shieldZone": map[string]interface{}{
			"planType":                             data.PlanType,
			"dDoSShieldSensitivity":                data.DDoSLevel,
			"dDoSExecutionMode":                    data.DDoSMode,
			"dDoSChallengeWindow":                  data.DDosChallengeWindow,
			"wafEnabled":                           data.WafEnabled,
			"wafExecutionMode":                     data.WafMode,
			"wafRealtimeThreatIntelligenceEnabled": data.WafRealtimeThreatIntelligenceEnabled,
			"wafRequestHeaderLoggingEnabled":       data.WafLogHeaders,
			"wafRequestIgnoredHeaders":             data.WafLogHeadersExcluded,
			"wafEngineConfig":                      wafEngineConfig,
			"wafDisabledRules":                     data.WafRulesDisabled,
			"wafLogOnlyRules":                      data.WafRulesLogonly,
			"whitelabelResponsePages":              data.WhiteLabelResponsePages,
			"learningMode":                         false,
			"wafRequestBodyLimitAction":            data.WafRequestBodyLimitAction,
			"wafResponseBodyLimitAction":           data.WafResponseBodyLimitAction,
		},
	})

	if err != nil {
		return PullzoneShield{}, err
	}

	tflog.Warn(ctx, fmt.Sprintf("POST /shield/shield-zone: %+v", string(body)))

	resp, err := c.doRequest(http.MethodPost, fmt.Sprintf("%s/shield/shield-zone", c.apiUrl), bytes.NewReader(body))
	if err != nil {
		return PullzoneShield{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := utils.ExtractErrorMessage(resp)
		if err != nil {
			return PullzoneShield{}, err
		} else {
			return PullzoneShield{}, errors.New("create pullzone shield failed with " + resp.Status)
		}
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullzoneShield{}, err
	}

	var result struct {
		Data struct {
			ShieldZone struct {
				Id int64 `json:"shieldZoneId"`
			} `json:"shieldZone"`
		} `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		return PullzoneShield{}, err
	}

	data.Id = result.Data.ShieldZone.Id

	return c.UpdatePullzoneShield(ctx, data)
}

func (c *Client) UpdatePullzoneShield(ctx context.Context, data PullzoneShield) (PullzoneShield, error) {
	// general settings
	{
		wafEngineConfig, err := c.convertPullzoneShieldWafEngineConfigToBody(data)
		if err != nil {
			return PullzoneShield{}, err
		}

		body, err := json.Marshal(map[string]interface{}{
			"shieldZoneId": data.Id,
			"shieldZone": map[string]interface{}{
				"planType":                             data.PlanType,
				"dDoSShieldSensitivity":                data.DDoSLevel,
				"dDoSExecutionMode":                    data.DDoSMode,
				"dDoSChallengeWindow":                  data.DDosChallengeWindow,
				"wafEnabled":                           data.WafEnabled,
				"wafExecutionMode":                     data.WafMode,
				"wafRealtimeThreatIntelligenceEnabled": data.WafRealtimeThreatIntelligenceEnabled,
				"wafRequestHeaderLoggingEnabled":       data.WafLogHeaders,
				"wafRequestIgnoredHeaders":             data.WafLogHeadersExcluded,
				"wafEngineConfig":                      wafEngineConfig,
				"wafDisabledRules":                     data.WafRulesDisabled,
				"wafLogOnlyRules":                      data.WafRulesLogonly,
				"whitelabelResponsePages":              data.WhiteLabelResponsePages,
				"learningMode":                         false,
				"wafRequestBodyLimitAction":            data.WafRequestBodyLimitAction,
				"wafResponseBodyLimitAction":           data.WafResponseBodyLimitAction,
			},
		})

		if err != nil {
			return PullzoneShield{}, err
		}

		tflog.Warn(ctx, fmt.Sprintf("PATCH /shield/shield-zone: %+v", string(body)))

		resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/shield-zone", c.apiUrl), bytes.NewReader(body))
		if err != nil {
			return PullzoneShield{}, err
		}

		if resp.StatusCode != http.StatusOK {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneShield{}, err
			} else {
				return PullzoneShield{}, errors.New("update pullzone shield failed with " + resp.Status)
			}
		}
	}

	// managed access lists
	{
		managedLists, err := c.getPullzoneAccessLists(ctx, data.Id, PullzoneAccessListQueryCurated)
		if err != nil {
			return PullzoneShield{}, err
		}

		type accessListConfig struct {
			Action    uint8
			IsEnabled bool
		}

		listConfigMap := make(map[int64]accessListConfig, len(data.AccessLists))
		managedListsMap := make(map[int64]pullzoneAccessListInfo, len(managedLists))
		managedListsNameMap := make(map[string]int64, len(managedLists))

		for _, managedList := range managedLists {
			managedListsMap[managedList.ListId] = managedList
			managedListsNameMap[managedList.Name] = managedList.ListId
		}

		for _, list := range data.AccessLists {
			listId := list.Id
			if listId == 0 {
				var ok bool
				listId, ok = managedListsNameMap[list.Name]
				if !ok {
					return PullzoneShield{}, fmt.Errorf("access list named \"%s\" does not exist", list.Name)
				}
			}

			if mappedList, ok := managedListsMap[listId]; !ok {
				return PullzoneShield{}, fmt.Errorf("access list id %d does not exist", listId)
			} else {
				// this comparison assumes the pullzoneshieldresourcevalidator.PlanTypeMap IDs are sorted from lower to higher
				if mappedList.RequiredPlan > 0 && data.PlanType < mappedList.RequiredPlan {
					tier, ok := pullzoneshieldresourcevalidator.PlanTypeMap[mappedList.RequiredPlan]
					if !ok {
						return PullzoneShield{}, fmt.Errorf("Unexpected Bunny Shield plan: %d", mappedList.RequiredPlan)
					}

					return PullzoneShield{}, fmt.Errorf("The access list \"%s\" requires the \"%s\" Bunny Shield plan.", mappedList.Name, tier)
				}
			}

			listConfigMap[listId] = accessListConfig{
				Action:    list.Action,
				IsEnabled: list.IsEnabled,
			}
		}

		for _, list := range managedLists {
			listConfig, ok := listConfigMap[list.ListId]
			if !ok {
				if list.IsEnabled {
					err := c.updatePullzoneAccessListConfiguration(ctx, data.Id, list.ListId, false, list.Action)
					if err != nil {
						if err.Error() == "not_available.access_list" {
							continue
						}

						return PullzoneShield{}, err
					}
				}

				continue
			}

			if listConfig.Action == list.Action && listConfig.IsEnabled == list.IsEnabled {
				// no changes
				continue
			}

			err := c.updatePullzoneAccessListConfiguration(ctx, data.Id, list.ListId, listConfig.IsEnabled, listConfig.Action)
			if err != nil {
				if err.Error() == "not_available.access_list" {
					continue
				}

				return PullzoneShield{}, err
			}
		}
	}

	// bot-detection fields
	{
		body, err := json.Marshal(map[string]interface{}{
			"shieldZoneId":  data.Id,
			"executionMode": data.BotDetectionMode,
			"requestIntegrity": map[string]interface{}{
				"sensitivity": data.BotDetectionRequestIntegrity,
			},
			"ipAddress": map[string]interface{}{
				"sensitivity": data.BotDetectionIPSensitivity,
			},
			"browserFingerprint": map[string]interface{}{
				"sensitivity":    data.BotDetectionFingerprintSensitivity,
				"aggression":     data.BotDetectionFingerprintAggression,
				"complexEnabled": data.BotDetectionComplexFingerprinting,
			},
		})

		if err != nil {
			return PullzoneShield{}, err
		}

		tflog.Warn(ctx, fmt.Sprintf("POST /shield/shield-zone/%d/bot-detection: %+v", data.Id, string(body)))

		resp, err := c.doRequest(http.MethodPatch, fmt.Sprintf("%s/shield/shield-zone/%d/bot-detection", c.apiUrl, data.Id), bytes.NewReader(body))
		if err != nil {
			return PullzoneShield{}, err
		}

		if resp.StatusCode != http.StatusOK {
			err := utils.ExtractErrorMessage(resp)
			if err != nil {
				return PullzoneShield{}, err
			} else {
				return PullzoneShield{}, errors.New("update pullzone shield/bot-detection failed with " + resp.Status)
			}
		}
	}

	return c.GetPullzoneShield(ctx, data.Id)
}

func (c *Client) DeletePullzoneShield(ctx context.Context, id int64) error {
	// @TODO set all values to default
	data := PullzoneShield{
		Id:                                   id,
		PlanType:                             0,
		DDoSLevel:                            0,
		DDoSMode:                             0,
		DDosChallengeWindow:                  3600,
		WafEnabled:                           false,
		WafMode:                              0,
		WafRealtimeThreatIntelligenceEnabled: false,
		WafLogHeaders:                        true,
		WafLogHeadersExcluded:                []string{"Cookie", "Authorization", "Signature", "Credential", "AccessKey"},
		WafAllowedHttpVersions:               []string{"HTTP/1.0", "HTTP/1.1", "HTTP/2", "HTTP/2.0"},
		WafAllowedHttpMethods:                []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"},
		WafAllowedRequestContentTypes: []string{
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"multipart/related",
			"text/xml",
			"application/xml",
			"application/soap+xml",
			"application/x-amf",
			"application/json",
			"application/octet-stream",
			"application/csp-report",
			"application/xss-auditor-report",
			"text/plain",
		},
		WafRulesDisabled:           []string{},
		WafRulesLogonly:            []string{},
		WhiteLabelResponsePages:    false,
		WafRequestBodyLimitAction:  1, // Log
		WafResponseBodyLimitAction: 2, // Ignore
	}

	_, err := c.UpdatePullzoneShield(ctx, data)
	return err
}

func (c *Client) convertPullzoneShieldWafEngineConfigToBody(data PullzoneShield) ([]map[string]string, error) {
	result := []map[string]string{
		{
			"name":         "detection_paranoia_level",
			"valueEncoded": fmt.Sprintf("%d", data.WafRuleSensitivityDetection),
		},
		{
			"name":         "executing_paranoia_level",
			"valueEncoded": fmt.Sprintf("%d", data.WafRuleSensitivityExecution),
		},
		{
			"name":         "blocking_paranoia_level",
			"valueEncoded": fmt.Sprintf("%d", data.WafRuleSensitivityBlocking),
		},
		{
			"name":         "allowed_methods",
			"valueEncoded": strings.Join(data.WafAllowedHttpMethods, " "),
		},
		{
			"name":         "allowed_http_versions",
			"valueEncoded": strings.Join(data.WafAllowedHttpVersions, " "),
		},
	}

	// AllowedRequestContentTypes
	{
		allowedRequestContentTypeStr := ""
		for _, item := range data.WafAllowedRequestContentTypes {
			allowedRequestContentTypeStr += fmt.Sprintf("|%s| ", item)
		}

		// removes trailing space
		allowedRequestContentTypeStr = allowedRequestContentTypeStr[0 : len(allowedRequestContentTypeStr)-1]

		result = append(result, map[string]string{
			"name":         "allowed_request_content_type",
			"valueEncoded": allowedRequestContentTypeStr,
		})
	}

	return result, nil
}
