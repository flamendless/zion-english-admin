package utils

import (
	"github.com/medama-io/go-useragent"
	"zion-english/internal/models"
)

var uaParser = useragent.NewParser()

func ParseUserAgent(userAgent string) models.UserAgentInfo {
	if userAgent == "" {
		return models.UserAgentInfo{}
	}

	ua := uaParser.Parse(userAgent)
	return models.UserAgentInfo{
		Browser:        string(ua.Browser()),
		BrowserVersion: ua.BrowserVersion(),
		OS:             string(ua.OS()),
		Device:         string(ua.Device()),
	}
}
