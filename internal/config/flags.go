package config

import (
	flag "github.com/spf13/pflag"
)

var flagServerAddr string
var flagBaseURL string
var flagLogLevel string
var flagFileStoragePath string
var flagDatabaseDSN string
var flagJWTSecretKey string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagServerAddr, "address", "a", "", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "defaults to address value")
	flag.StringVarP(&flagLogLevel, "loglevel", "l", "", "log level")
	flag.StringVarP(&flagFileStoragePath, "filestoragepath", "f", "", "")
	flag.StringVarP(&flagDatabaseDSN, "database-dsn", "d", "", "database DSN")
	flag.StringVarP(&flagJWTSecretKey, "jwt-secret", "j", "", "JWT signing secret key")
	flag.Parse()
}
