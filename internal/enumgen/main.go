package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Sprintf("Usage: %s -o output_file.go", os.Args[0]))
	}

	if os.Getenv("BUNNYNET_API_KEY") == "" {
		_, err := fmt.Fprintln(os.Stderr, "[WARN] enumgen: BUNNYNET_API_KEY environment variable not set, skipping")
		if err != nil {
			panic(err)
		}

		// @TODO we don't want to expose BUNNYNET_API_KEY during the build process, nor break GHA builds
		os.Exit(0)
	}

	contents := "// This file was generated via \"go generate\". DO NOT EDIT.\npackage provider\n\n"
	contents += generatePullzoneShield()
	contents += "\n"
	contents += generatePullzoneShieldAccessList()

	contentsFmt, err := format.Source([]byte(contents))
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(os.Args[2], contentsFmt, 0o644)
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}

func generatePullzoneShield() string {
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

	var result struct {
		Data []struct {
			Name   string `json:"enumName"`
			Values []struct {
				Name  string `json:"name"`
				Value int64  `json:"value"`
			} `json:"enumValues"`
		} `json:"data"`
	}

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		panic(err)
	}

	var variableMap = map[string]struct {
		Variable string
		Type     string
	}{
		"WafRuleOperatorType":       {Variable: "pullzoneShieldRuleConditionOperationMap", Type: "map[int64]string"},
		"WafRuleTransformationType": {Variable: "pullzoneShieldRuleTransformationMap", Type: "map[int64]string"},
		"WafRuleVariableType":       {Variable: "pullzoneShieldRuleConditionVariableMap", Type: "map[uint8]string"},
		"WafRateLimitTimeframeType": {Variable: "pullzoneShieldRatelimitRuleLimitTimeframeOptions", Type: "[]int64"},
		"WafRatelimitBlockType":     {Variable: "pullzoneShieldRatelimitRuleResponseTimeframeOptions", Type: "[]int64"},
		"WAFPayloadLimitAction":     {Variable: "pullzoneShieldWafBodyLimitMap", Type: "map[uint8]string"},
		"WafRuleActionType":         {Variable: "pullzoneShieldWafRuleResponseActionMap", Type: "map[uint8]string"},
	}

	contents := map[string]string{}

	for _, value := range result.Data {
		variable, ok := variableMap[value.Name]
		if !ok {
			continue
		}

		contents[variable.Variable] = fmt.Sprintf("var %s = %s{\n", variable.Variable, variable.Type)
		for _, value := range value.Values {
			switch variable.Type {
			case "map[int64]string":
				fallthrough
			case "map[uint64]string":
				fallthrough
			case "map[int8]string":
				fallthrough
			case "map[uint8]string":
				contents[variable.Variable] += fmt.Sprintf("\t%d: \"%s\",\n", value.Value, value.Name)

			case "[]int8":
				fallthrough
			case "[]uint8":
				fallthrough
			case "[]int64":
				fallthrough
			case "[]uint64":
				contents[variable.Variable] += fmt.Sprintf("\t%d,\n", value.Value)
			}
		}

		contents[variable.Variable] += "}\n"
	}

	keys := maps.Keys(contents)
	slices.Sort(keys)

	contentStr := []string{}
	for _, key := range keys {
		value, ok := contents[key]
		if !ok {
			panic(fmt.Sprintf("Missing value for %s", key))
		}

		contentStr = append(contentStr, value)
	}

	return strings.Join(contentStr, "\n")
}

func generatePullzoneShieldAccessList() string {
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

	var result map[string]map[string]string

	err = json.Unmarshal(bodyResp, &result)
	if err != nil {
		panic(err)
	}

	var variableMap = map[string]struct {
		Variable string
		Type     string
	}{
		"AccessListType":   {Variable: "pullzoneAccessListTypeMap", Type: "map[uint8]string"},
		"AccessListAction": {Variable: "pullzoneAccessListActionMap", Type: "map[uint8]string"},
	}

	contents := map[string]string{}

	for mapKey, variable := range variableMap {
		mapValue, ok := result[mapKey]
		if !ok {
			panic(fmt.Sprintf("Missing value for %s", mapKey))
		}

		contents[variable.Variable] = fmt.Sprintf("var %s = %s{\n", variable.Variable, variable.Type)
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
		contents[variable.Variable] += strings.Join(contentsVars, "\n") + "\n}\n"
	}

	keys := maps.Keys(contents)
	slices.Sort(keys)

	contentStr := []string{}
	for _, key := range keys {
		value, ok := contents[key]
		if !ok {
			panic(fmt.Sprintf("Missing value for %s", key))
		}

		contentStr = append(contentStr, value)
	}

	return strings.Join(contentStr, "\n")
}
