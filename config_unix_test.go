// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// go:build unix

package zeroconfig_test

import (
	"testing"
)

func TestWriterConfig_Compile_Journald(t *testing.T) {
	compile(t, `{"writers": [{"type": "journald"}]}`)
}
