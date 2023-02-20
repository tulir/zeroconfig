// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package zeroconfig

import (
	"io"

	"github.com/rs/zerolog"
)

type levelWriterAdapter struct {
	io.Writer
}

func (lw levelWriterAdapter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	return lw.Write(p)
}

type minMaxLevelWriter struct {
	zerolog.LevelWriter
	MinLevel zerolog.Level
	MaxLevel zerolog.Level
}

// MinMaxLevelWriter wraps a writer in a zerolog.LevelWriter, but limits the log levels that can pass through.
func MinMaxLevelWriter(writer io.Writer, minLevel, maxLevel zerolog.Level) zerolog.LevelWriter {
	lw, ok := writer.(zerolog.LevelWriter)
	if !ok {
		lw = levelWriterAdapter{writer}
	}
	return minMaxLevelWriter{LevelWriter: lw, MinLevel: minLevel, MaxLevel: maxLevel}
}

func (mlw minMaxLevelWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if (mlw.MinLevel == zerolog.NoLevel || l >= mlw.MinLevel) && (mlw.MaxLevel == zerolog.NoLevel || l <= mlw.MaxLevel) {
		return mlw.LevelWriter.WriteLevel(l, p)
	}
	return len(p), nil
}
