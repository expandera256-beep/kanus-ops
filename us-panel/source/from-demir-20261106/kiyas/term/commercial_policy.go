package main

import (
	"strings"
)

// Set by the release build. Environment variables cannot disable this gate.
var releaseChannel = "development"

func commercialRequired() bool { return releaseChannel == "commercial" }

func (s *session) commercialSessionReady() bool {
	return s.accountGateway != "" && s.accountGateway == strings.TrimRight(activeGateway(), "/") &&
		strings.HasPrefix(s.accountGateway, "https://") && !isLoopbackGateway(s.accountGateway) &&
		validateGateway(s.accountGateway) == nil && len(strings.TrimSpace(s.key)) >= 32 && s.key != localTunnelKey
}
