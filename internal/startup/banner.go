package startup

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"zion-english/internal/conf"
	"zion-english/internal/version"
)

const bannerListPrefix = "  - "

const (
	ansiReset = "\033[0m"
	ansiDim   = "\033[2m"
	ansiBold  = "\033[1m"
	ansiCyan  = "\033[36m"
	ansiGreen = "\033[32m"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func printStartupBanner(opts Options) {
	printBox(startupBannerLines(opts))
}

func startupBannerLines(opts Options) []string {
	innerWidth := bannerInnerWidth(opts)
	lines := []string{
		"",
		centerBannerLine(ansiBold+appName+ansiReset, innerWidth),
		centerBannerLine(ansiDim+envLabel(opts.Cfg.AppEnv)+ansiReset, innerWidth),
		"",
		bannerHeading("Configuration"),
	}
	for _, line := range startupConfigLines(opts) {
		lines = append(lines, bannerListPrefix+line)
	}

	lines = append(lines, "", bannerHeading("Integrations"))
	for _, line := range startupIntegrationLines(opts) {
		lines = append(lines, bannerListPrefix+line)
	}

	lines = append(lines, "")
	return lines
}

func bannerInnerWidth(opts Options) int {
	width := visibleLen(appName)
	for _, line := range startupConfigLines(opts) {
		width = max(width, visibleLen(bannerListPrefix+line))
	}
	for _, line := range startupIntegrationLines(opts) {
		width = max(width, visibleLen(bannerListPrefix+line))
	}
	for _, heading := range []string{"Configuration", "Integrations"} {
		width = max(width, visibleLen(heading))
	}
	width = max(width, visibleLen(envLabel(opts.Cfg.AppEnv)))
	if width < 48 {
		return 48
	}
	return width
}

func startupConfigLines(opts Options) []string {
	cfg := opts.Cfg
	lines := []string{
		fmt.Sprintf("version: %s", versionSummary()),
		fmt.Sprintf("env: %s", cfg.AppEnv),
		fmt.Sprintf("http: %s  listen=%s  https=%t", ListenURL(opts), listenAddr(opts), opts.HTTPS),
		fmt.Sprintf("database: sqlite  path=%s", dbPath),
		fmt.Sprintf("secret: %s", conf.MaskSecret(cfg.Secret)),
		fmt.Sprintf("superuser: %s", valueOrUnset(cfg.SuperuserUsername)),
	}
	if opts.HTTPS && opts.TLSAddress != "" {
		lines = append(lines,
			fmt.Sprintf("tls: address=%s  cert=%s", opts.TLSAddress, tlsCertFile(opts.TLSAddress)),
		)
	}
	return lines
}

func startupIntegrationLines(opts Options) []string {
	cfg := opts.Cfg
	integrations := opts.Integrations
	meetingService := integrations.MeetingService
	if meetingService == "" {
		meetingService = "zoom"
	}
	lines := []string{
		fmt.Sprintf("Meeting service: %s", meetingService),
		zoomIntegrationLine(cfg, integrations.ZoomConfigured),
		googleCalendarIntegrationLine(cfg, integrations.GoogleCalendarConfigured),
	}
	return lines
}

func zoomIntegrationLine(cfg *conf.Config, configured bool) string {
	if configured {
		return "Zoom: configured"
	}
	missing := missingZoomFields(cfg)
	if len(missing) == 0 {
		return "Zoom: not configured"
	}
	return fmt.Sprintf("Zoom: not configured (missing: %s)", strings.Join(missing, ", "))
}

func googleCalendarIntegrationLine(cfg *conf.Config, configured bool) string {
	if configured {
		return "Google Calendar: configured"
	}
	missing := missingGoogleCalendarFields(cfg)
	if len(missing) == 0 {
		return "Google Calendar: not configured"
	}
	return fmt.Sprintf("Google Calendar: not configured (missing: %s)", strings.Join(missing, ", "))
}

func missingZoomFields(cfg *conf.Config) []string {
	zoom := cfg.Meeting.Zoom
	missing := make([]string, 0, 4)
	if zoom.ClientID == "" {
		missing = append(missing, "ZOOM_CLIENT_ID")
	}
	if zoom.ClientSecret == "" {
		missing = append(missing, "ZOOM_CLIENT_SECRET")
	}
	if zoom.RedirectURI == "" {
		missing = append(missing, "ZOOM_REDIRECT_URI")
	}
	if zoom.AuthorizeURL == "" {
		missing = append(missing, "ZOOM_AUTHORIZE_URL")
	}
	return missing
}

func missingGoogleCalendarFields(cfg *conf.Config) []string {
	gcal := cfg.Calendar.Google
	missing := make([]string, 0, 3)
	if gcal.ClientID == "" {
		missing = append(missing, "GOOGLE_CALENDAR_CLIENT_ID")
	}
	if gcal.ClientSecret == "" {
		missing = append(missing, "GOOGLE_CALENDAR_CLIENT_SECRET")
	}
	if gcal.RedirectURI == "" {
		missing = append(missing, "GOOGLE_CALENDAR_REDIRECT_URI")
	}
	return missing
}

func versionSummary() string {
	return version.Get().Summary()
}

func printBox(lines []string) {
	innerWidth := 48
	for _, line := range lines {
		innerWidth = max(innerWidth, visibleLen(line))
	}

	top := ansiCyan + strings.Repeat("=", innerWidth+4) + ansiReset
	writeStdout("\n%s\n", top)
	for _, line := range lines {
		padding := innerWidth - visibleLen(line)
		if padding < 0 {
			padding = 0
		}
		writeStdout("%s| %s%s |%s\n", ansiCyan, line, strings.Repeat(" ", padding), ansiReset)
	}
	writeStdout("%s\n\n", top)
}

func printDevListening(opts Options) {
	writeStdout(" %s→ Listening on %s%s\n\n", ansiGreen, ListenURL(opts), ansiReset)
}

func bannerHeading(title string) string {
	return ansiBold + title + ansiReset
}

func centerBannerLine(line string, width int) string {
	padding := width - visibleLen(line)
	if padding < 0 {
		return line
	}
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + line + strings.Repeat(" ", right)
}

func visibleLen(s string) int {
	return len(ansiEscape.ReplaceAllString(s, ""))
}

func writeStdout(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}
