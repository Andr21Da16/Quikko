package redirect

import (
	"strings"

	"github.com/mssola/user_agent"
)

// Valores normalizados de tipo de dispositivo.
const (
	deviceMobile   = "mobile"
	deviceDesktop  = "desktop"
	deviceTablet   = "tablet"
	deviceUnknown  = "unknown"
	browserUnknown = "unknown"
)

// ParseUserAgent extrae el tipo de dispositivo y el navegador de un User-Agent.
// mssola/user_agent no distingue tablet de mobile, así que detectamos tablet por
// heurística sobre el string crudo (iPad / "Tablet").
func ParseUserAgent(uaString string) (deviceType, browser string) {
	if strings.TrimSpace(uaString) == "" {
		return deviceUnknown, browserUnknown
	}

	ua := user_agent.New(uaString)

	name, _ := ua.Browser()
	if name == "" {
		name = browserUnknown
	}

	return classifyDevice(uaString, ua), name
}

func classifyDevice(uaString string, ua *user_agent.UserAgent) string {
	lower := strings.ToLower(uaString)
	if strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet") {
		return deviceTablet
	}
	if ua.Mobile() {
		return deviceMobile
	}
	if ua.Bot() {
		return deviceUnknown
	}
	return deviceDesktop
}
