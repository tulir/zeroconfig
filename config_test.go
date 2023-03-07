// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package zeroconfig_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mau.fi/zeroconfig"
)

func compile(t *testing.T, cfg string) *zerolog.Logger {
	var parsed zeroconfig.Config
	err := json.Unmarshal([]byte(cfg), &parsed)
	if err != nil {
		require.NoError(t, err, "Unmarshaling config should be successful")
	}
	logger, err := parsed.Compile()
	if err != nil {
		require.NoError(t, err, "Compiling config should be successful")
	}
	return logger
}

func TestWriterConfig_Compile_Stdout(t *testing.T) {
	var out bytes.Buffer
	zeroconfig.Stdout = &out
	log := compile(t, `{
	  "writers": [
	    {"type": "stdout", "format": "pretty"}
	  ],
	  "min_level": "debug",
	  "timestamp": false
	}`)

	log.Trace().Msg("meow")
	require.Empty(t, out.String(), "Output should be empty after logging below minimum level")

	log.Debug().Int("cats", 5).Msg("meow")
	require.NotEmpty(t, out.String(), "Output should not be empty after logging at minimum level")
	require.Equal(t, "<nil> DBG meow cats=5\n", out.String(), "Output is formatted prettily")
	out.Reset()
}

func TestWriterConfig_Compile_Metadata(t *testing.T) {
	var out bytes.Buffer
	zeroconfig.Stdout = &out
	log := compile(t, `{
	  "writers": [
	    {"type": "stdout", "format": "pretty"}
	  ],
	  "metadata": {
	    "meow": 5,
	    "foo": {"bar": "asd"}
	  },
	  "timestamp": false
	}`)

	log.Debug().Msg("meow")
	require.NotEmpty(t, out.String(), "Output should not be empty after logging")
	require.Equal(t, `<nil> DBG meow foo={"bar":"asd"} meow=5`+"\n", out.String(), "Output contains global metadata fields")
	out.Reset()
}

func TestWriterConfig_Compile_TimeFormat(t *testing.T) {
	var out bytes.Buffer
	zeroconfig.Stdout = &out
	log := compile(t, `{
	  "writers": [
	    {"type": "stdout", "format": "pretty", "time_format": "2006"}
	  ],
	  "min_level": "debug"
	}`)

	require.Equal(t, time.Now().Year(), time.Now().Add(5*time.Second).Year(), "This test can't be ran at midnight right before the new year")

	log.Debug().Int("cats", 5).Msg("meow")
	require.Equal(t, strconv.Itoa(time.Now().Year())+" DBG meow cats=5\n", out.String(), "Output has current year as date")
}

func TestWriterConfig_Compile_MultiLevel_Stdio(t *testing.T) {
	var stderr, stdout bytes.Buffer
	zeroconfig.Stdout = &stdout
	zeroconfig.Stderr = &stderr
	log := compile(t, `{
	  "writers": [
	    {"type": "stdout", "format": "pretty", "max_level": "warn"},
        {"type": "stderr", "format": "pretty", "min_level": "error"}
	  ],
	  "min_level": "trace",
	  "timestamp": false
	}`)

	log.Info().Msg("meow")
	assert.Empty(t, stderr.String(), "Stderr should not have info log")
	assert.Equal(t, "<nil> INF meow\n", stdout.String(), "Stdout should have info log")
	stdout.Reset()
	stderr.Reset()

	log.Error().Msg("meow #2")
	assert.Empty(t, stdout.String(), "Stdout should not have error log")
	assert.Equal(t, "<nil> ERR meow #2\n", stderr.String(), "Stderr should have error log")
	stdout.Reset()
	stderr.Reset()

	log.Trace().Msg("meow #3")
	assert.Empty(t, stderr.String(), "Stderr should not have trace log")
	assert.Equal(t, "<nil> TRC meow #3\n", stdout.String(), "Stdout should have trace log")
	stdout.Reset()
	stderr.Reset()
}

func TestWriterConfig_Compile_StdoutAndFile(t *testing.T) {
	dir := t.TempDir()
	var stdout bytes.Buffer
	zeroconfig.Stdout = &stdout
	log := compile(t, fmt.Sprintf(`{
	  "writers": [
	    {"type": "stdout", "format": "pretty", "min_level": "info"},
        {"type": "file", "filename": "%s/test.log"}
	  ],
	  "min_level": "trace",
	  "timestamp": false
	}`, dir))
	fileName := filepath.Join(dir, "test.log")
	file, err := os.Open(fileName)
	require.NoError(t, err, "Opening log file should be successful")
	dec := json.NewDecoder(file)
	var ll logLine

	log.Debug().Msg("meow")
	assert.Empty(t, stdout.String(), "Stdout is empty when logging debug")
	assert.NoError(t, dec.Decode(&ll), "File should contain JSON log line")
	assert.Equal(t, ll.Level, zerolog.DebugLevel)
	assert.Equal(t, ll.Message, "meow")

	log.Error().Msg("meow #2")
	assert.Equal(t, "<nil> ERR meow #2\n", stdout.String(), "Stdout should have error log")
	assert.NoError(t, dec.Decode(&ll), "File should contain JSON log line")
	assert.Equal(t, ll.Level, zerolog.ErrorLevel)
	assert.Equal(t, ll.Message, "meow #2")
}
