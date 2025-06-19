// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package storageplanmodifier

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"strings"
)

func DetectFileContentsChange() planmodifier.String {
	return detectFileContentsChange{}
}

type detectFileContentsChange struct{}

func (m detectFileContentsChange) Description(_ context.Context) string {
	return "Will state as changed if the checksum for the \"content\" or \"source\" attributes differs."
}

func (m detectFileContentsChange) MarkdownDescription(_ context.Context) string {
	return "Will state as changed if the checksum for the `content` or `source` attributes differs."
}

func (m detectFileContentsChange) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}

	var fileContents []byte
	var content string
	req.Plan.GetAttribute(ctx, path.Root("content"), &content)
	if len(content) > 0 {
		fileContents = []byte(content)
	}

	var source string
	req.Plan.GetAttribute(ctx, path.Root("source"), &source)
	if len(source) > 0 {
		var err error
		fileContents, err = os.ReadFile(source)
		if err != nil {
			resp.Diagnostics.AddError("Could not read source file", err.Error())
			return
		}
	}

	checksum := strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256(fileContents)))
	resp.PlanValue = types.StringValue(checksum)
}
