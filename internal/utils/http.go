// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"io"
)

func Drain(c io.ReadCloser) {
	_, _ = io.Copy(io.Discard, c)
	_ = c.Close()
}
