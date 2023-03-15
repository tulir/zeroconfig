// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build unix

package zeroconfig

import (
	"io"
	"log/syslog"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/journald"
)

func compileSyslog(wc *WriterConfig) (io.Writer, error) {
	sl, err := syslog.Dial(wc.Network, wc.Host, wc.Flags, wc.Tag)
	if err != nil {
		return nil, err
	}
	if wc.Type == WriterTypeSyslogCEE {
		return zerolog.SyslogCEEWriter(sl), nil
	} else {
		return zerolog.SyslogLevelWriter(sl), nil
	}
}

func compileJournald(_ *WriterConfig) (io.Writer, error) {
	return journald.NewJournalDWriter(), nil
}

func init() {
	writerCompilers[WriterTypeSyslog] = compileSyslog
	writerCompilers[WriterTypeSyslogCEE] = compileSyslog
	writerCompilers[WriterTypeJournald] = compileJournald
}
