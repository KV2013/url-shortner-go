package config

import (
	flag "github.com/spf13/pflag"
)

var flagRunAddr string
var flagBaseURL string
var flagLogLevel string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagRunAddr, "address", "a", "localhost:8080", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "defaults to address value")
	flag.StringVarP(&flagLogLevel, "loglevel", "l", "info", "log level")
	flag.Parse()
}
