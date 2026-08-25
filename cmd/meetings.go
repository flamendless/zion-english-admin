package cmd

import (
	"zion-english/internal/conf"
	"zion-english/internal/meetings"
)

var meetingSvc *meetings.Service

func initMeetingService() {
	cfg := conf.Conf()
	meetingSvc = meetings.New(dbRW.GetQueries(), meetings.Config{
		Secret:         cfg.Secret,
		DefaultService: cfg.Meeting.Service,
		Zoom: meetings.ZoomConfig{
			ClientID:     cfg.Meeting.Zoom.ClientID,
			ClientSecret: cfg.Meeting.Zoom.ClientSecret,
			RedirectURI:  cfg.Meeting.Zoom.RedirectURI,
			AuthorizeURL: cfg.Meeting.Zoom.AuthorizeURL,
		},
	})
}
