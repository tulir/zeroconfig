// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package zeroconfig_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"go.mau.fi/zeroconfig"
)

var allLevels = []zerolog.Level{zerolog.TraceLevel, zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel}

type logLine struct {
	Level   zerolog.Level `json:"level"`
	Message string        `json:"message"`
}

func TestMinMaxLevelWriter(t *testing.T) {
	tests := []struct {
		name string
		min  zerolog.Level
		max  zerolog.Level
		out  []zerolog.Level
	}{{
		"No limits",
		zerolog.NoLevel,
		zerolog.NoLevel,
		allLevels,
	}, {
		"Only up to info",
		zerolog.NoLevel,
		zerolog.InfoLevel,
		[]zerolog.Level{zerolog.TraceLevel, zerolog.DebugLevel, zerolog.InfoLevel},
	}, {
		"Only above error",
		zerolog.ErrorLevel,
		zerolog.NoLevel,
		[]zerolog.Level{zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel},
	}, {
		"Only between debug and warn",
		zerolog.DebugLevel,
		zerolog.WarnLevel,
		[]zerolog.Level{zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := zerolog.New(zeroconfig.MinMaxLevelWriter(&buf, test.min, test.max))
			for i, level := range allLevels {
				log.WithLevel(level).Int("meow_num", i).Msg("meow")
			}
			dec := json.NewDecoder(&buf)
			for i, expected := range test.out {
				var ll logLine
				err := dec.Decode(&ll)
				require.NoError(t, err, "Decoding log line should be successful")
				require.Equalf(t, expected, ll.Level, "Log #%d has expected level", i+1)
			}
			require.False(t, dec.More(), "Log output should be empty after reading expected levels")
		})
	}
}
