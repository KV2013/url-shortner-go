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
var flagAuditFile string
var flagAuditURL string
var flagEnablePprof bool
var flagEnableHTTPS bool
var flagTrustedSubnet string
var flagConfigPath string

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
	flag.StringVarP(&flagAuditFile, "audit-file", "", "", "path to audit log file")
	flag.StringVarP(&flagAuditURL, "audit-url", "", "", "URL of remote audit server")
	flag.BoolVarP(&flagEnablePprof, "enable-pprof", "p", false, "enable pprof debug server on :8082")
	flag.BoolVarP(&flagEnableHTTPS, "enable-https", "s", false, "enable HTTPS with self-signed certificate")
	flag.StringVarP(&flagTrustedSubnet, "trusted-subnet", "t", "", "trusted subnet in CIDR notation")
	flag.StringVarP(&flagConfigPath, "config", "c", "", "path to JSON config file")
	flag.Parse()
}
