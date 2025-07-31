package customtype

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestPullzoneOriginUrl(t *testing.T) {
	type testCase struct {
		ExpectedError  bool
		ExpectedResult bool
		VOld           string
		VNew           string
	}

	dataProvider := []testCase{
		// same case
		{false, true, "https://BuNnY.net", "https://bunny.net"},
		{false, false, "https://sub.bunny.net", "https://bunny.net"},

		// end slash
		{false, true, "https://bunny.net", "https://bunny.net/"},
		{false, true, "https://bunny.net/blog", "https://bunny.net/blog/"},

		// scheme/port
		{false, true, "https://bunny.net:443", "https://bunny.net"},
		{false, false, "https://bunny.net:4433", "https://bunny.net"},
		{false, true, "http://bunny.net:80", "http://bunny.net"},
		{false, false, "http://bunny.net:443", "https://bunny.net"},

		// no user
		{false, true, "https://user@bunny.net", "https://bunny.net"},
		{false, true, "https://user:pw@bunny.net", "https://bunny.net/"},
		{false, true, "https://user:pw@bunny.net", "https://user:aa@bunny.net/"},

		// no fragment
		{false, true, "https://bunny.net/#abc", "https://bunny.net/#def"},
		{false, true, "https://bunny.net/", "https://bunny.net/#def"},
		{false, true, "https://bunny.net/#abc", "https://bunny.net:443"},
	}

	for _, tc := range dataProvider {
		newValue := PullzoneOriginUrlValue{StringValue: types.StringValue(tc.VOld)}

		result, diags := PullzoneOriginUrlValue{
			StringValue: types.StringValue(tc.VNew),
		}.StringSemanticEquals(context.Background(), newValue)

		if !tc.ExpectedError && diags.HasError() {
			t.Errorf("expected no error, got %s", diags.Errors())
		}

		if tc.ExpectedError && !diags.HasError() {
			t.Errorf("expected no error, got none")
		}

		if tc.ExpectedResult != result {
			t.Errorf("expected %s == %s to be %t, got %t", tc.VOld, tc.VNew, tc.ExpectedResult, result)
		}
	}
}
