package changelog_test

import (
	"os"
	"path/filepath"
	"testing"

	"zion-english/internal/changelog"

	"github.com/stretchr/testify/require"
)

func TestReadParsesFixture(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "changelogs.yaml")
	doc, err := changelog.Read(path)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Len(t, doc.Versions, 2)
	require.Len(t, doc.Unreleased.NewFeatures, 1)
}

func TestReleasedVersionsFiltersAndSorts(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "changelogs.yaml")
	doc, requireErr := changelog.Read(path)
	require.NoError(t, requireErr)

	released := doc.ReleasedVersions()
	require.Len(t, released, 1)
	require.Equal(t, "2026.08.13", released[0].Version)
	require.Equal(t, "2026-08-13", released[0].Date)
	require.Len(t, released[0].NewFeatures, 1)
}

func TestSectionsHasEntries(t *testing.T) {
	t.Parallel()

	empty := changelog.Sections{}
	require.False(t, empty.HasEntries())

	withEntries := changelog.Sections{
		BugFixes: []changelog.Entry{{Text: "TESTFIX"}},
	}
	require.True(t, withEntries.HasEntries())
}

func TestReadMissingFile(t *testing.T) {
	t.Parallel()

	_, err := changelog.Read(filepath.Join(t.TempDir(), "missing.yaml"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestResolvePathEnvOverride(t *testing.T) {
	t.Setenv(changelog.EnvChangelogPath, "/tmp/custom-changelogs.yaml")
	require.Equal(t, "/tmp/custom-changelogs.yaml", changelog.ResolvePath())
}
