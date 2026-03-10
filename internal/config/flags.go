package config

import (
	flag "github.com/spf13/pflag"
)

var flagRunAddr string
var flagBaseURL string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagRunAddr, "address", "a", "localhost:8080", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "defaults to address value")
	flag.Parse()
}
