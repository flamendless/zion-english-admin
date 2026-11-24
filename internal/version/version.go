package version

import (
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	Commit = "unknown"
	Tag    = ""

	cachedCommit     string
	cachedCommitOnce sync.Once
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

func (i Info) Summary() string {
	if i.Tag != "" {
		return fmt.Sprintf("tag=%s  commit=%s", i.Tag, i.Commit)
	}
	return fmt.Sprintf("commit=%s", i.Commit)
}

func resolveCommit() string {
	if Commit != "" && Commit != "unknown" {
		return Commit
	}
	cachedCommitOnce.Do(func() {
		cachedCommit = resolveCommitFromBuildInfoOrGit()
	})
	if cachedCommit != "" {
		return cachedCommit
	}
	return "unknown"
}

func resolveCommitFromBuildInfoOrGit() string {
	if rev := vcsRevisionFromBuildInfo(); rev != "" {
		return rev
	}
	return gitHeadSHA()
}

func vcsRevisionFromBuildInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			return setting.Value
		}
	}
	return ""
}

func gitHeadSHA() string {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	sha := strings.TrimSpace(string(out))
	if len(sha) < 7 {
		return ""
	}
	return sha
}
