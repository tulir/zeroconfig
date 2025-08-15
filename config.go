// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package zeroconfig

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SyslogConfig contains the configuration options for the syslog writer.
//
// See https://pkg.go.dev/log/syslog for exact details.
type SyslogConfig struct {
	// All parameters are passed to https://pkg.go.dev/log/syslog#Dial directly.
	Network string `json:"network,omitempty" yaml:"network,omitempty"`
	Host    string `json:"host,omitempty" yaml:"host,omitempty"`
	Flags   int    `json:"flags,omitempty" yaml:"flags,omitempty"`
	Tag     string `json:"tag,omitempty" yaml:"tag,omitempty"`
}

// FileConfig contains the configuration options for the file writer.
//
// See https://github.com/natefinch/lumberjack for exact details.
type FileConfig struct {
	// File name for the current log. Backups will be stored in the same directory, named as name-<timestamp>.ext
	Filename string `json:"filename,omitempty" yaml:"filename,omitempty"`
	// Maximum size in megabytes for the log file before rotating. Defaults to 100 megabytes.
	MaxSize int `json:"max_size,omitempty" yaml:"max_size,omitempty"`
	// Maximum age of rotated log files to keep as days. Defaults to no limit.
	MaxAge int `json:"max_age,omitempty" yaml:"max_age,omitempty"`
	// Maximum number of rotated log files to keep. Defaults to no limit.
	MaxBackups int `json:"max_backups,omitempty" yaml:"max_backups,omitempty"`
	// Should rotated log file names use local time instead of UTC? Defaults to false.
	LocalTime bool `json:"local_time,omitempty" yaml:"local_time,omitempty"`
	// Should rotated log files be compressed with gzip? Defaults to false.
	Compress bool `json:"compress,omitempty" yaml:"compress,omitempty"`
}

// WriterType is a type of writer.
type WriterType string

const (
	// WriterTypeStdout writes to stdout.
	WriterTypeStdout WriterType = "stdout"
	// WriterTypeStderr writes to stderr.
	WriterTypeStderr WriterType = "stderr"
	// WriterTypeFile writes to a file, including rotating the file when it gets big.
	// The configuration is stored in the FileConfig struct.
	WriterTypeFile WriterType = "file"
	// WriterTypeSyslog writes to the system logging service.
	// The configuration is stored in the SyslogConfig struct.
	WriterTypeSyslog WriterType = "syslog"
	// WriterTypeSyslogCEE writes to the system logging service with MITRE CEE prefixes.
	WriterTypeSyslogCEE WriterType = "syslog-cee"
	// WriterTypeJournald writes to systemd's logging service.
	WriterTypeJournald WriterType = "journald"
	// WriterTypeJS writes to a JavaScript console.
	// Only usable in environments where syscall/js is available (i.e. GOOS=js/GOARCH=wasm).
	WriterTypeJS WriterType = "js"
)

// LogFormat describes how logs should be formatted for a writer.
type LogFormat string

const (
	// LogFormatJSON outputs logs as the raw JSON that zerolog produces.
	LogFormatJSON LogFormat = "json"
	// LogFormatPretty uses zerolog's console writer, but disables color.
	LogFormatPretty LogFormat = "pretty"
	// LogFormatPrettyColored uses zerolog's console writer including color.
	LogFormatPrettyColored LogFormat = "pretty-colored"
)

// WriterConfig contains the configuration for an individual log writer.
type WriterConfig struct {
	// The type of writer.
	Type   WriterType `json:"type" yaml:"type"`
	Format LogFormat  `json:"format,omitempty" yaml:"format,omitempty"`

	MinLevel *zerolog.Level `json:"min_level,omitempty" yaml:"min_level,omitempty"`
	MaxLevel *zerolog.Level `json:"max_level,omitempty" yaml:"max_level,omitempty"`

	// Only applies when format=console or format=console-colored
	TimeFormat string `json:"time_format,omitempty" yaml:"time_format,omitempty"`

	SyslogConfig `json:",inline,omitempty" yaml:",inline,omitempty"`
	FileConfig   `json:",inline,omitempty" yaml:",inline,omitempty"`
}

// Config contains all the configuration to create a zerolog logger.
type Config struct {
	Writers  []WriterConfig `json:"writers,omitempty" yaml:"writers,omitempty"`
	MinLevel *zerolog.Level `json:"min_level,omitempty" yaml:"min_level,omitempty"`

	Timestamp *bool `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Caller    bool  `json:"caller,omitempty" yaml:"caller,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Outputs used for the stdout and stderr writer types.
var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

func compileUnsupported(wc *WriterConfig) (io.Writer, error) {
	return nil, fmt.Errorf("writer type %q not supported on this OS", wc.Type)
}

type WriterCompiler = func(*WriterConfig) (io.Writer, error)

var writerCompilers = map[WriterType]WriterCompiler{
	WriterTypeStdout:    func(_ *WriterConfig) (io.Writer, error) { return Stdout, nil },
	WriterTypeStderr:    func(_ *WriterConfig) (io.Writer, error) { return Stderr, nil },
	WriterTypeFile:      compileFile,
	WriterTypeJournald:  compileUnsupported,
	WriterTypeSyslog:    compileUnsupported,
	WriterTypeSyslogCEE: compileUnsupported,
}

func RegisterWriter(wt WriterType, compiler WriterCompiler) {
	writerCompilers[wt] = compiler
}

func compileFile(wc *WriterConfig) (io.Writer, error) {
	writer := &lumberjack.Logger{
		Filename:   wc.Filename,
		MaxSize:    wc.MaxSize,
		MaxAge:     wc.MaxAge,
		MaxBackups: wc.MaxBackups,
		LocalTime:  wc.LocalTime,
		Compress:   wc.Compress,
	}
	err := writer.Rotate()
	if err != nil {
		return nil, err
	}
	return writer, nil
}

func (wc *WriterConfig) compileMain() (io.Writer, error) {
	compiler, ok := writerCompilers[wc.Type]
	if !ok {
		return nil, fmt.Errorf("unknown writer type %q", wc.Type)
	}
	return compiler(wc)
}

func levelPtr(ptr *zerolog.Level) zerolog.Level {
	if ptr == nil {
		return zerolog.NoLevel
	}
	return *ptr
}

// Compile creates an io.Writer instance out of the configuration in this struct.
func (wc *WriterConfig) Compile() (io.Writer, error) {
	output, err := wc.compileMain()
	if err != nil {
		return nil, err
	}
	switch wc.Format {
	case "", LogFormatJSON:
		// output directly
	case LogFormatPretty, LogFormatPrettyColored:
		wrapper := zerolog.ConsoleWriter{
			Out: output,
		}
		if wc.Format == LogFormatPretty {
			wrapper.NoColor = true
		}
		if wc.TimeFormat != "" {
			wrapper.TimeFormat = wc.TimeFormat
		} else {
			wrapper.TimeFormat = "2006-01-02T15:04:05.999Z07:00"
		}
		output = wrapper
	default:
		return nil, fmt.Errorf("unknown format %q", wc.Format)
	}
	if wc.MinLevel != nil || wc.MaxLevel != nil {
		output = MinMaxLevelWriter(output, levelPtr(wc.MinLevel), levelPtr(wc.MaxLevel))
	}
	return output, nil
}

// Compile creates a zerolog.Logger instance out of the configuration in this struct.
func (c *Config) Compile() (*zerolog.Logger, error) {
	if len(c.Writers) == 0 || (c.MinLevel != nil && *c.MinLevel == zerolog.Disabled) {
		log := zerolog.Nop()
		return &log, nil
	}
	writers := make([]io.Writer, len(c.Writers))
	for i, wc := range c.Writers {
		writer, err := wc.Compile()
		if err != nil {
			return nil, fmt.Errorf("failed to parse config for writer #%d: %w", i+1, err)
		}
		writers[i] = writer
	}
	var realWriter io.Writer
	if len(writers) == 1 {
		realWriter = writers[0]
	} else if len(writers) > 1 {
		realWriter = zerolog.MultiLevelWriter(writers...)
	}
	with := zerolog.New(realWriter).With()
	if c.Timestamp == nil || *c.Timestamp {
		with = with.Timestamp()
	}
	if c.Caller {
		with = with.Caller()
	}
	if len(c.Metadata) > 0 {
		keys := make([]string, len(c.Metadata))
		i := 0
		for key := range c.Metadata {
			keys[i] = key
			i++
		}
		sort.Strings(keys)
		for _, key := range keys {
			with = with.Interface(key, c.Metadata[key])
		}
	}
	log := with.Logger()
	if c.MinLevel != nil {
		log = log.Level(*c.MinLevel)
	}
	return &log, nil
}
