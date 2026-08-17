package version

import (
	"fmt"
	"runtime/debug"
)

var (
	Commit = "unknown"
	Tag    = ""
)

const GitHubRepo = "https://github.com/flamendless/zion-english-admin"

type Info struct {
	Commit string
	Tag    string
}

func Get() Info {
	return Info{
		Commit: resolveCommit(),
		Tag:    Tag,
	}
}

func GitHubCommitURL(commit string) string {
	return fmt.Sprintf("%s/commit/%s", GitHubRepo, commit)
}

func resolveCommit() string {
	if Commit != "" && Commit != "unknown" {
		return Commit
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				return setting.Value
			}
		}
	}
	return "unknown"
}
