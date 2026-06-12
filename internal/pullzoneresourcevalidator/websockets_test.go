package pullzoneresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

type testWebsocketsMaxConnectionsTestCase struct {
	ExpectedError bool
	Value         int64
}

func TestWebsocketsMaxConnections(t *testing.T) {
	testCases := []testWebsocketsMaxConnectionsTestCase{
		{true, 0},
		{true, -1},
		{true, -99},
		{true, -100},
		{true, 1},
		{true, 99},
		{true, 199},
		{true, 12345},
		{false, 100},
		{false, 1000},
		{false, 25000},
		{false, 2147483600},
		{true, 2147483647},
	}

	for _, testCase := range testCases {
		request := validator.Int64Request{
			ConfigValue: types.Int64Value(testCase.Value),
		}

		response := validator.Int64Response{}
		websocketsMaxConnectionsValidator{}.ValidateInt64(context.Background(), request, &response)

		if testCase.ExpectedError && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if !testCase.ExpectedError && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}
	}
}
