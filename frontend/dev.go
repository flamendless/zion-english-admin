package frontend

import "zion-english/internal/conf"

func ShowDevBanner() bool {
	return conf.Conf().IsLocal()
}
