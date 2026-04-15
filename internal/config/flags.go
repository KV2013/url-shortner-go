package config

import (
	flag "github.com/spf13/pflag"
)

var flagServerAddr string
var flagBaseURL string
var flagLogLevel string
var flagFileStoragePath string
var flagDbHost string
var flagDbPort int
var flagDbUser string
var flagDbPassword string
var flagDbName string
var flagDbSSLMode string

func parseFlags() {
	if flag.Parsed() {
		return
	}

	flag.StringVarP(&flagServerAddr, "address", "a", "", "")
	flag.StringVarP(&flagBaseURL, "baseurl", "b", "", "defaults to address value")
	flag.StringVarP(&flagLogLevel, "loglevel", "l", "", "log level")
	flag.StringVarP(&flagFileStoragePath, "filestoragepath", "f", "", "")
	flag.StringVar(&flagDbHost, "db-host", "", "database host")
	flag.IntVar(&flagDbPort, "db-port", 0, "database port")
	flag.StringVar(&flagDbUser, "db-user", "", "database user")
	flag.StringVar(&flagDbPassword, "db-password", "", "database password")
	flag.StringVar(&flagDbName, "db-name", "", "database name")
	flag.StringVar(&flagDbSSLMode, "db-sslmode", "", "database ssl mode")
	flag.Parse()
}
