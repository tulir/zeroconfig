//go:build !linux
// +build !linux

package zeroconfig

import (
	"fmt"
	"io"
)

type SyslogConfig struct{}

func (wc *WriterConfig) compileMainForOS() (io.Writer, error) {
	return nil, fmt.Errorf("unknown writer type %q", wc.Type)
}
