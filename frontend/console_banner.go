package frontend

import (
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/version"
)

type ConsoleBannerData struct {
	AppEnv  string `json:"appEnv"`
	AppName string `json:"appName"`
	Tagline string `json:"tagline"`
	Commit  string `json:"commit"`
}

func BuildConsoleBannerData() ConsoleBannerData {
	return ConsoleBannerData{
		AppEnv:  conf.Conf().AppEnv,
		AppName: constants.AppName,
		Tagline: constants.AppTagline,
		Commit:  version.Get().Commit,
	}
}