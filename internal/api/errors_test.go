// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusBadGateway,
		Status:     "502 Bad Gateway",
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestExtractErrorMessages(t *testing.T) {
	extractors := map[string]func(*http.Response) error{
		"extractErrorMessage":         extractErrorMessage,
		"extractDatabaseErrorMessage": extractDatabaseErrorMessage,
		"extractMCErrorMessage":       extractMCErrorMessage,
		"extractShieldErrorMessage":   extractShieldErrorMessage,
	}

	// A body that parses as JSON but carries no message must not turn into an
	// empty error; nil lets the caller report the HTTP status instead.
	for name, extract := range extractors {
		for _, body := range []string{`{}`, `{"Message":""}`, `[{"message":""}]`, `""`, `not json`} {
			if err := extract(responseWithBody(body)); err != nil {
				t.Errorf("%s(%q) = %q, want nil", name, body, err.Error())
			}
		}
	}

	cases := []struct {
		name    string
		extract func(*http.Response) error
		body    string
		want    string
	}{
		{"extractErrorMessage", extractErrorMessage, `{"Message":"invalid origin"}`, "invalid origin"},
		{"extractDatabaseErrorMessage", extractDatabaseErrorMessage, `{"error":"db failed"}`, "db failed"},
		{"extractMCErrorMessage", extractMCErrorMessage, `[{"message":"bad image"}]`, "bad image"},
		{"extractMCErrorMessage", extractMCErrorMessage, `"bad registry"`, "bad registry"},
		{"extractShieldErrorMessage", extractShieldErrorMessage, `{"errorKey":"some.error"}`, "some.error"},
	}

	for _, c := range cases {
		err := c.extract(responseWithBody(c.body))
		if err == nil || err.Error() != c.want {
			t.Errorf("%s(%q) = %v, want %q", c.name, c.body, err, c.want)
		}
	}

	if err := extractShieldErrorMessage(responseWithBody(`{"error":{"errorKey":"limit_reached.waf"}}`)); err != ErrShieldWafLimitReached {
		t.Errorf("extractShieldErrorMessage: got %v, want ErrShieldWafLimitReached", err)
	}
}
