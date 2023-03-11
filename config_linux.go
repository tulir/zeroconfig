package zeroconfig

import (
	"fmt"
	"io"
	"log/syslog"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/journald"
)

const (
	// WriterTypeSyslog writes to the system logging service.
	// The configuration is stored in the SyslogConfig struct.
	WriterTypeSyslog WriterType = "syslog"
	// WriterTypeSyslogCEE writes to the system logging service with MITRE CEE prefixes.
	WriterTypeSyslogCEE WriterType = "syslog-cee"
	// WriterTypeJournald writes to systemd's logging service.
	WriterTypeJournald WriterType = "journald"
)

// SyslogConfig contains the configuration options for the syslog writer.
//
// See https://pkg.go.dev/log/syslog for exact details.
type SyslogConfig struct {
	// All parameters are passed to https://pkg.go.dev/log/syslog#Dial directly.
	Network string          `json:"network,omitempty" yaml:"network,omitempty"`
	Host    string          `json:"host,omitempty" yaml:"host,omitempty"`
	Flags   syslog.Priority `json:"flags,omitempty" yaml:"flags,omitempty"`
	Tag     string          `json:"tag,omitempty" yaml:"tag,omitempty"`
}

func (wc *WriterConfig) compileMainForOS() (io.Writer, error) {
	switch wc.Type {
	case WriterTypeSyslog, WriterTypeSyslogCEE:
		sl, err := syslog.Dial(wc.Network, wc.Host, wc.Flags, wc.Tag)
		if err != nil {
			return nil, err
		}
		if wc.Type == WriterTypeSyslogCEE {
			return zerolog.SyslogCEEWriter(sl), nil
		} else {
			return zerolog.SyslogLevelWriter(sl), nil
		}
	case WriterTypeJournald:
		return journald.NewJournalDWriter(), nil
	default:
		return nil, fmt.Errorf("unknown writer type %q", wc.Type)
	}
}
