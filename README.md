# zeroconfig
A relatively simple declarative config format for [zerolog](https://github.com/rs/zerolog).

Meant to be used as YAML, but JSON struct tags are included as well.

## Config reference
```yaml
# Global minimum log level. Defaults to trace.
min_level: trace

# Should logs include timestamps? Defaults to true.
timestamps: true
# Should logs include the caller function? Defaults to false.
caller: false

# Additional log metadata to add globally. Map from string key to arbitrary value.
metadata: null

# List of writers to output logs to.
# The `type` field is always required. `format`, `min_level` and `max_level` can be specified for any type of writer.
# Some types have additional custom configuration
writers:
# `stdout` and `stderr` write to the corresponding standard IO streams.
# They have no custom configuration fields, the extra fields below showcase the fields that can be added to any writer.
- # The type of writer.
  type: stdout
  # The format to write. Available formats are json, pretty and pretty-colored. Defaults to json.
  format: pretty-colored
  # If format is pretty or pretty-colored, time_format can be used to specify how timestamps are formatted.
  # Uses Go time formatting https://pkg.go.dev/time#pkg-constants and defaults to RFC3339 (2006-01-02T15:04:05Z07:00).
  time_format: 2006-01-02 15:04:05
  # Minimum level for this writer. Defaults to no level (i.e. inherited from root min_level).
  # This can only reduce the amount of logs written to this writer, levels below the global min_level are never logged.
  min_level: info
  # Maximum level for this writer. Defaults to no level (all logs above minimum are logged).
  max_level: warn
# If you want errors in stderr, make a separate writer like this:
# If you want all logs in stdout, just remove this and the max_level above.
- type: stderr
  format: pretty-colored
  min_level: error

# `file` is a rotating file handler (https://github.com/natefinch/lumberjack).
- type: file
  # File name for the current log. Backups will be stored in the same directory, named as name-<timestamp>.ext
  filename: example.log
  # Maximum size in megabytes for the log file before rotating. Defaults to 100 megabytes.
  max_size: 100
  # Maximum age of rotated log files to keep as days. Defaults to no limit.
  max_age: 0
  # Maximum number of rotated log files to keep. Defaults to no limit.
  max_backups: 0
  # Should rotated log file names use local time instead of UTC? Defaults to false.
  local_time: false
  # Should rotated log files be compressed with gzip? Defaults to false.
  compress: false

# `syslog` writes to the system log service using the Go stdlib syslog package.
- type: syslog  # you can also use syslog-cee to add the MITRE CEE prefix.
  # These four parameters are passed to https://pkg.go.dev/log/syslog#Dial directly.
  network: udp
  host: localhost
  # Priority flags as defined in syslog.h
  flags: 8
  tag: zerolog

# `journald` writes to systemd's logging service using https://github.com/coreos/go-systemd.
# It has no custom configuration fields.
- type: journald
```

## Usage example
```go
package main

import (
	"github.com/rs/zerolog"
	"go.mau.fi/zeroconfig"
	"gopkg.in/yaml.v3"
)

func prepareLog(yamlConfig []byte) (*zerolog.Logger, error) {
	var cfg zeroconfig.Config
	err := yaml.Unmarshal(yamlConfig, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg.Compile()
}

// This should obviously be loaded from a file rather than hardcoded
const logConfig = `
min_level: debug
writers:
- type: stdout
  format: pretty-colored
`

func main() {
	log, err := prepareLog([]byte(logConfig))
	if err != nil {
		panic(err)
	}
	log.Info().Msg("Logger initialized")
}
```
