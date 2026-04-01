package config

import (
	flag "github.com/spf13/pflag"
)

var flagServerAddr string
var flagBaseURL string
var flagLogLevel string
var flagFileStoragePath string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagServerAddr, "address", "a", "", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "defaults to address value")
	flag.StringVarP(&flagLogLevel, "loglevel", "l", "", "log level")
	flag.StringVarP(&flagFileStoragePath, "filestoragepath", "f", "", "")
	flag.Parse()
}
