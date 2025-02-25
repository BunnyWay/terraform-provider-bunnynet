// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	rand "math/rand/v2"
	"testing"
)

func TestConvertTimestampToSeconds(t *testing.T) {
	type dataType struct {
		Expected  uint64
		Timestamp string
	}

	dataProvider := []dataType{
		{0, "00:00"},
		{1, "00:01"},
		{2, "00:02"},
		{60, "01:00"},
		{61, "01:01"},
		{3683, "61:23"},
		{7425, "123:45"},
	}

	for _, data := range dataProvider {
		result, err := convertTimestampToSeconds(data.Timestamp)

		if err != nil {
			t.Errorf("Expected no error for %s, got %s", data.Timestamp, err)
		}

		if result != data.Expected {
			t.Errorf("Expected %s to return %d, got %d", data.Timestamp, data.Expected, result)
		}
	}
}

func TestConvertSecondsToTimestamp(t *testing.T) {
	type dataType struct {
		Expected string
		Seconds  uint64
	}

	dataProvider := []dataType{
		{"00:00", 0},
		{"00:01", 1},
		{"00:02", 2},
		{"00:59", 59},
		{"01:00", 60},
		{"01:01", 61},
		{"01:39", 99},
		{"01:40", 100},
		{"61:23", 3683},
		{"123:45", 7425},
	}

	for _, data := range dataProvider {
		result := convertSecondsToTimestamp(data.Seconds)

		if result != data.Expected {
			t.Errorf("Expected %d to return %s, got %s", data.Seconds, data.Expected, result)
		}
	}
}

const randomStringOptions = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateRandomString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = randomStringOptions[rand.IntN(len(randomStringOptions))]
	}

	return string(result)
}
