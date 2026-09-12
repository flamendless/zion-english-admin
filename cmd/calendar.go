package cmd

import (
	"zion-english/internal/calendar"
	"zion-english/internal/conf"
)

var calendarSvc *calendar.Service

func initCalendarService() {
	cfg := conf.Conf()
	calendarSvc = calendar.New(dbRW.GetQueries(), calendar.Config{
		Secret: cfg.Secret,
		Google: calendar.GoogleConfig{
			ClientID:     cfg.Calendar.Google.ClientID,
			ClientSecret: cfg.Calendar.Google.ClientSecret,
			RedirectURI:  cfg.Calendar.Google.RedirectURI,
		},
	})
}
