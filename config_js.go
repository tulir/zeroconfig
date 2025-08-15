// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package zeroconfig

import (
	"io"
	"syscall/js"
	"unsafe"

	"github.com/rs/zerolog"
)

type JSWriter struct {
	Console js.Value
}

func (jw *JSWriter) Write(p []byte) (n int, err error) {
	return jw.WriteLevel(zerolog.DebugLevel, p)
}

func (jw *JSWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	var method string
	switch level {
	case zerolog.TraceLevel:
		method = "debug"
	case zerolog.DebugLevel, zerolog.NoLevel:
		method = "log"
	case zerolog.InfoLevel:
		method = "info"
	case zerolog.WarnLevel:
		method = "warn"
	case zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel:
		method = "error"
	case zerolog.Disabled:
		return len(p), nil
	}
	parsedP := js.Global().Get("JSON").Call("parse", unsafe.String(unsafe.SliceData(p), len(p)))
	message := parsedP.Get("message")
	parsedP.Delete("message")
	parsedP.Delete("level")
	jw.Console.Call(method, message, parsedP)
	return len(p), nil
}

var _ zerolog.LevelWriter = (*JSWriter)(nil)

func compileJS(wc *WriterConfig) (io.Writer, error) {
	wc.Format = LogFormatJSON
	return &JSWriter{Console: js.Global().Get("console")}, nil
}

func init() {
	writerCompilers[WriterTypeJS] = compileJS
}
