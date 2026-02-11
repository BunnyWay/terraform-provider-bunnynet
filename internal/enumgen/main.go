package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

func main() {
	if os.Getenv("BUNNYNET_API_KEY") == "" {
		_, err := fmt.Fprintln(os.Stderr, "[WARN] enumgen: BUNNYNET_API_KEY environment variable not set, skipping")
		if err != nil {
			panic(err)
		}

		// @TODO we don't want to expose BUNNYNET_API_KEY during the build process, nor break GHA builds
		os.Exit(0)
	}

	var result []GenResult
	result = append(result, generateFromOpenApiSchema()...)
	result = append(result, generatePullzoneShield()...)
	result = append(result, generatePullzoneShieldAccessList()...)

	slices.SortStableFunc(result, func(a, b GenResult) int {
		fileCmp := strings.Compare(a.File.File, b.File.File)
		if fileCmp != 0 {
			return fileCmp
		}

		return strings.Compare(a.Variable, b.Variable)
	})

	files := map[*Fileinfo]string{}
	for _, r := range result {
		files[r.File] += r.Contents + "\n\n"
	}

	for fileinfo, content := range files {
		prefix := "// This file was generated via \"go generate\". DO NOT EDIT.\npackage " + fileinfo.Package + "\n\n"
		contentsFmt, err := format.Source([]byte(prefix + content))
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(fileinfo.File, contentsFmt, 0o644)
		if err != nil {
			panic(err)
		}
	}

	os.Exit(0)
}

type Fileinfo struct {
	File    string
	Package string
}

type GenResult struct {
	File     *Fileinfo
	Variable string
	Contents string
}

var FileinfoProvider = &Fileinfo{File: "internal/provider/enums.go", Package: "provider"}
var FileinfoEdgeruleValidator = &Fileinfo{File: "internal/pullzoneedgeruleresourcevalidator/types.go", Package: "pullzoneedgeruleresourcevalidator"}

func generateFromOpenApiSchema() []GenResult {
	resp, err := http.Get("https://core-api-public-docs.b-cdn.net/docs/v3/public.json")
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	_ = resp.Body.Close()

	var schema struct {
		Components struct {
			Schemas map[string]struct {
				Type       string        `json:"type"`
				EnumValues []interface{} `json:"enum"`
				EnumNames  []string      `json:"x-enumNames"`
			} `json:"schemas"`
		} `json:"components"`
	}

	err = json.Unmarshal(body, &schema)
	if err != nil {
		panic(err)
	}

	var variableMap = []struct {
		File      *Fileinfo
		Variable  string
		SchemaKey string
		Type      string
	}{
		{
			File:      FileinfoProvider,
			Variable:  "dnsRecordMonitorTypeMap",
			SchemaKey: "DnsMonitoringType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "dnsRecordSmartRoutingTypeMap",
			SchemaKey: "DnsSmartRoutingType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "dnsRecordTypeMap",
			SchemaKey: "DnsRecordTypes",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "dnsZoneLogAnonymizedStyleMap",
			SchemaKey: "LogAnonymizationType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "pullzoneLogAnonymizedStyleMap",
			SchemaKey: "LogAnonymizationType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "pullzoneLogForwardFormatMap",
			SchemaKey: "PullZoneLogFormat",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "pullzoneLogForwardProtocolMap",
			SchemaKey: "PullZoneLogForwarderProtocolType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "pullzoneOptimizerWatermarkPositionMap",
			SchemaKey: "OptimizerWatermarkPosition",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "storageZoneTierMap",
			SchemaKey: "StorageZoneTier",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoProvider,
			Variable:  "streamLibraryEncodingTierMap",
			SchemaKey: "EncodingTier",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoEdgeruleValidator,
			Variable:  "ActionMap",
			SchemaKey: "EdgeRuleActionType",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoEdgeruleValidator,
			Variable:  "TriggerTypeMap",
			SchemaKey: "TriggerTypes",
			Type:      "map[uint8]string",
		},
		{
			File:      FileinfoEdgeruleValidator,
			Variable:  "TriggerMatchTypeMap",
			SchemaKey: "PatternMatchingTypes",
			Type:      "map[uint8]string",
		},
	}

	result := []GenResult{}

	for _, v := range variableMap {
		s, ok := schema.Components.Schemas[v.SchemaKey]
		if !ok {
			continue
		}

		contents := fmt.Sprintf("var %s = %s{\n", v.Variable, v.Type)

		for enumIdx, enumValue := range s.EnumValues {
			switch v.Type {
			case "map[int64]string":
				fallthrough
			case "map[uint64]string":
				fallthrough
			case "map[int8]string":
				fallthrough
			case "map[uint8]string":
				contents += fmt.Sprintf("\t%0.f: \"%s\",\n", enumValue, s.EnumNames[enumIdx])

			case "[]int8":
				fallthrough
			case "[]uint8":
				fallthrough
			case "[]int64":
				fallthrough
			case "[]uint64":
				contents += fmt.Sprintf("\t%d,\n", enumValue)
			}
		}

		contents += "}"

		result = append(result, GenResult{
			File:     v.File,
			Variable: v.Variable,
			Contents: contents,
		})
	}

	return result
}

func generatePullzoneShield() []GenResult {
	req, err := http.NewRequest(http.MethodGet, "https://api.bunny.net/shield/waf/enums", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("AccessKey", os.Getenv("BUNNYNET_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_, err := fmt.Fprintf(os.Stderr, "failed to fetch enums from API: %s\n", resp.Status)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("failed to fetch enums from API: %s", resp.Status))
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("[WARN] failed to close response body: %s\n", err)
		}
	}()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data struct {
		Data []struct {
			Name   string `json:"enumName"`
			Values []struct {
				Name  string `json:"name"`
				Value int64  `json:"value"`
			} `json:"enumValues"`
		} `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		panic(err)
	}

	var variableMap = map[string]struct {
		File     *Fileinfo
		Variable string
		Type     string
	}{
		"WafRuleOperatorType":       {File: FileinfoProvider, Variable: "pullzoneShieldRuleConditionOperationMap", Type: "map[int64]string"},
		"WafRuleTransformationType": {File: FileinfoProvider, Variable: "pullzoneShieldRuleTransformationMap", Type: "map[int64]string"},
		"WafRuleVariableType":       {File: FileinfoProvider, Variable: "pullzoneShieldRuleConditionVariableMap", Type: "map[uint8]string"},
		"WafRateLimitTimeframeType": {File: FileinfoProvider, Variable: "pullzoneShieldRatelimitRuleLimitTimeframeOptions", Type: "[]int64"},
		"WafRatelimitBlockType":     {File: FileinfoProvider, Variable: "pullzoneShieldRatelimitRuleResponseTimeframeOptions", Type: "[]int64"},
		"WAFPayloadLimitAction":     {File: FileinfoProvider, Variable: "pullzoneShieldWafBodyLimitMap", Type: "map[uint8]string"},
		"WafRuleActionType":         {File: FileinfoProvider, Variable: "pullzoneShieldWafRuleResponseActionMap", Type: "map[uint8]string"},
	}

	result := []GenResult{}

	for _, value := range data.Data {
		variable, ok := variableMap[value.Name]
		if !ok {
			continue
		}

		contents := fmt.Sprintf("var %s = %s{\n", variable.Variable, variable.Type)

		for _, value := range value.Values {
			switch variable.Type {
			case "map[int64]string":
				fallthrough
			case "map[uint64]string":
				fallthrough
			case "map[int8]string":
				fallthrough
			case "map[uint8]string":
				contents += fmt.Sprintf("\t%d: \"%s\",\n", value.Value, value.Name)

			case "[]int8":
				fallthrough
			case "[]uint8":
				fallthrough
			case "[]int64":
				fallthrough
			case "[]uint64":
				contents += fmt.Sprintf("\t%d,\n", value.Value)
			}
		}

		contents += "}"

		result = append(result, GenResult{
			File:     variable.File,
			Variable: variable.Variable,
			Contents: contents,
		})
	}

	return result
}

func generatePullzoneShieldAccessList() []GenResult {
	// @TODO url must be dynamic
	req, err := http.NewRequest(http.MethodGet, "https://api.bunny.net/shield/shield-zone/13939/access-lists/enums", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("AccessKey", os.Getenv("BUNNYNET_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_, err := fmt.Fprintf(os.Stderr, "failed to fetch enums from API: %s\n", resp.Status)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("failed to fetch enums from API: %s", resp.Status))
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("[WARN] failed to close response body: %s\n", err)
		}
	}()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data map[string]map[string]string

	err = json.Unmarshal(bodyResp, &data)
	if err != nil {
		panic(err)
	}

	var variableMap = map[string]struct {
		File     *Fileinfo
		Variable string
		Type     string
	}{
		"AccessListType":   {File: FileinfoProvider, Variable: "pullzoneAccessListTypeMap", Type: "map[uint8]string"},
		"AccessListAction": {File: FileinfoProvider, Variable: "pullzoneAccessListActionMap", Type: "map[uint8]string"},
	}

	result := []GenResult{}

	for mapKey, variable := range variableMap {
		mapValue, ok := data[mapKey]
		if !ok {
			panic(fmt.Sprintf("Missing value for %s", mapKey))
		}

		contents := fmt.Sprintf("var %s = %s{\n", variable.Variable, variable.Type)
		contentsVars := []string{}

		for k, v := range mapValue {
			keyInt, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				panic(err)
			}

			switch variable.Type {
			case "map[int64]string":
				fallthrough
			case "map[uint64]string":
				fallthrough
			case "map[int8]string":
				fallthrough
			case "map[uint8]string":
				contentsVars = append(contentsVars, fmt.Sprintf("\t%d: \"%s\",", keyInt, v))
			}
		}

		slices.Sort(contentsVars)
		contents += strings.Join(contentsVars, "\n") + "\n}"

		result = append(result, GenResult{
			File:     variable.File,
			Variable: variable.Variable,
			Contents: contents,
		})
	}

	return result
}
