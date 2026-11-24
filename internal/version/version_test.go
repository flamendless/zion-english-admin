package version

import (
	"os/exec"
	"sync"
	"testing"
)

func TestResolveCommitPrefersInjectedCommit(t *testing.T) {
	oldCommit := Commit
	t.Cleanup(func() {
		Commit = oldCommit
		cachedCommit = ""
		cachedCommitOnce = sync.Once{}
	})

	Commit = "injecteddeadbeef"
	if got := resolveCommit(); got != "injecteddeadbeef" {
		t.Fatalf("got %q want injecteddeadbeef", got)
	}
}

func TestResolveCommitFromGitWhenAvailable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	if _, err := exec.Command("git", "rev-parse", "HEAD").Output(); err != nil {
		t.Skip("not in a git work tree")
	}

	oldCommit := Commit
	t.Cleanup(func() {
		Commit = oldCommit
		cachedCommit = ""
		cachedCommitOnce = sync.Once{}
	})

	Commit = "unknown"
	cachedCommit = ""
	cachedCommitOnce = sync.Once{}

	got := resolveCommit()
	if got == "" || got == "unknown" {
		t.Fatalf("expected git HEAD, got %q", got)
	}
	if len(got) < 40 {
		t.Fatalf("expected full SHA, got %q", got)
	}
}