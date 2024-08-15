// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"math"
	"strconv"
	"strings"
	"sync"
)

func mapValueToKey[T constraints.Integer](mapped map[T]string, value string) T {
	for k, v := range mapped {
		if v == value {
			return k
		}
	}

	panic("value not found in map")
}

func mapKeyToValue[T constraints.Integer](mapped map[T]string, value T) string {
	if v, ok := mapped[value]; ok {
		return v
	}

	panic("key not found in map")
}

// Mutex to manage concurrent changes to pullzone sub-resources (i.e. bunnynet_pullzone_edgerule and bunnynet_pullzone_optimizer_class)
// Based on https://discuss.hashicorp.com/t/cooping-with-parallelism-is-there-a-way-to-prioritise-resource-types/55690
var pzMutex *pullzoneMutex

func init() {
	pzMutex = &pullzoneMutex{mu: sync.Mutex{}, pullzones: map[int64]*sync.Mutex{}}
}

type pullzoneMutex struct {
	mu        sync.Mutex
	pullzones map[int64]*sync.Mutex
}

func (p *pullzoneMutex) Lock(id int64) {
	p.mu.Lock()
	if v, ok := p.pullzones[id]; ok {
		p.mu.Unlock()
		v.Lock()
	} else {
		p.pullzones[id] = &sync.Mutex{}
		p.mu.Unlock()
		p.pullzones[id].Lock()
	}
}

func (p *pullzoneMutex) Unlock(id int64) {
	p.mu.Lock()
	if _, ok := p.pullzones[id]; ok {
		p.pullzones[id].Unlock()
		delete(p.pullzones, id)
	}
	p.mu.Unlock()
}

func convertTimestampToSeconds(timestamp string) (uint64, error) {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 2 {
		return 0, errors.New("Invalid timestamp format, expected \"00:00\"")
	}

	minutes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return uint64(seconds + minutes*60), nil
}

func convertSecondsToTimestamp(seconds uint64) string {
	minutes := int64(seconds / 60)
	remainder := math.Mod(float64(seconds), 60)

	return fmt.Sprintf("%02d:%02d", minutes, int64(remainder))
}

func generateMarkdownMapOptions[T comparable](m map[T]string) string {
	s := maps.Values(m)
	return generateMarkdownSliceOptions(s)
}

func generateMarkdownSliceOptions(s []string) string {
	slices.Sort(s)
	return "Options: `" + strings.Join(s, "`, `") + "`"
}

const enforceDefaultDiagSummary = "Attribute must use default value during creating"
const enforceDefaultDiagError = "Unfortunately our API does not support setting this attribute at creation time. Create the resource using default values, then change them and apply your changes once again."

func planAttrBoolEnforceDefault(ctx context.Context, plan tfsdk.Plan, attrName string) diag.Diagnostics {
	attrPath := path.Root(attrName)
	attrAtPath, diags := plan.Schema.AttributeAtPath(ctx, attrPath)
	if diags.HasError() {
		return diags
	}

	attr := attrAtPath.(schema.BoolAttribute)
	req := defaults.BoolRequest{Path: attrPath}
	resp := defaults.BoolResponse{}
	attr.Default.DefaultBool(ctx, req, &resp)

	var attrValue bool
	plan.GetAttribute(ctx, attrPath, &attrValue)
	defaultValue := resp.PlanValue.ValueBool()

	if attrValue != defaultValue {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attrPath, enforceDefaultDiagSummary, enforceDefaultDiagError)}
	}

	return nil
}

func planAttrStringEnforceDefault(ctx context.Context, plan tfsdk.Plan, attrName string) diag.Diagnostics {
	attrPath := path.Root(attrName)
	attrAtPath, diags := plan.Schema.AttributeAtPath(ctx, attrPath)
	if diags.HasError() {
		return diags
	}

	attr := attrAtPath.(schema.StringAttribute)
	req := defaults.StringRequest{Path: attrPath}
	resp := defaults.StringResponse{}
	attr.Default.DefaultString(ctx, req, &resp)

	var attrValue string
	plan.GetAttribute(ctx, attrPath, &attrValue)
	defaultValue := resp.PlanValue.ValueString()

	if attrValue != defaultValue {
		return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(attrPath, enforceDefaultDiagSummary, enforceDefaultDiagError)}
	}

	return nil
}
