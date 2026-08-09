package httpserver

import (
	"time"

	"github.com/raykavin/gobox/httpserver"
)

// NewGin builds a Gin server configuration from a ServerProvider.
func NewGin(
	listenPort uint16,
	tlsEnabled bool,
	tlsCertPath string,
	tlsKeyPath string,
	noRouteRedir string,
	writeTimeout time.Duration,
	readTimeout time.Duration,
	idleTimeout time.Duration,
) *httpserver.GinConfig {
	defconfig := httpserver.DefaultGinConfig()

	defconfig.UseSSL = tlsEnabled
	defconfig.SSLCert = tlsCertPath
	defconfig.SSLKey = tlsKeyPath
	defconfig.Port = listenPort
	defconfig.NoRouteTo = noRouteRedir

	if rt := readTimeout; rt > 0 {
		defconfig.ReadTimeout = rt
	}
	if wt := writeTimeout; wt > 0 {
		defconfig.WriteTimeout = wt
	}
	if it := idleTimeout; it > 0 {
		defconfig.IdleTimeout = it
	}

	return defconfig
}
