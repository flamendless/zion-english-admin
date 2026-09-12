package startup

import (
	"fmt"
	"strings"

	"zion-english/internal/conf"
	"zion-english/internal/logs"

	"go.uber.org/zap"
)

const appName = "Zion English Admin"

const dbPath = "data/zion.db"

type IntegrationStatus struct {
	ZoomConfigured           bool
	GoogleCalendarConfigured bool
	MeetingService           string
}

type Options struct {
	Cfg          *conf.Config
	ListenPort   string
	BasePath     string
	HTTPS        bool
	TLSAddress   string
	Integrations IntegrationStatus
}

func LogStartup(opts Options) {
	printStartupBanner(opts)
	if opts.Cfg.IsLocal() {
		return
	}
	cfg := opts.Cfg
	integrations := opts.Integrations
	logs.Log().Info("server starting",
		zap.String("env", cfg.AppEnv),
		zap.String("version", versionSummary()),
		zap.String("listen_addr", listenAddr(opts)),
		zap.String("public_url", ListenURL(opts)),
		zap.Bool("https", opts.HTTPS),
		zap.String("database", dbPath),
		zap.String("superuser", valueOrUnset(cfg.SuperuserUsername)),
		zap.String("meeting_service", integrations.MeetingService),
		zap.Bool("zoom_configured", integrations.ZoomConfigured),
		zap.Bool("google_calendar_configured", integrations.GoogleCalendarConfigured),
	)
	if opts.HTTPS {
		logs.Log().Info("https certificates",
			zap.String("address", opts.TLSAddress),
			zap.String("cert", tlsCertFile(opts.TLSAddress)),
			zap.String("key", tlsKeyFile(opts.TLSAddress)),
		)
	}
}

func LogListening(opts Options) {
	if opts.Cfg.IsLocal() {
		printDevListening(opts)
		return
	}
	logs.Log().Info("server listening", zap.String("addr", listenAddr(opts)))
}

func ListenURL(opts Options) string {
	scheme := "http"
	host := "localhost"
	if opts.HTTPS {
		scheme = "https"
		if opts.TLSAddress != "" {
			host = opts.TLSAddress
		}
	}
	port := strings.TrimPrefix(opts.ListenPort, ":")
	basePath := normalizeBasePath(opts.BasePath)
	return fmt.Sprintf("%s://%s:%s%s", scheme, host, port, basePath)
}

func listenAddr(opts Options) string {
	return opts.ListenPort
}

func normalizeBasePath(basePath string) string {
	if basePath == "" || basePath == "/" {
		return ""
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	return strings.TrimSuffix(basePath, "/")
}

func envLabel(env string) string {
	return "(" + env + ")"
}

func valueOrUnset(value string) string {
	if value == "" {
		return "(unset)"
	}
	return value
}

func tlsCertFile(address string) string {
	return fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", address)
}

func tlsKeyFile(address string) string {
	return fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", address)
}
