package changelog_test

import (
	"testing"

	"zion-english/internal/changelog"
	"zion-english/internal/constants"
	"zion-english/internal/version"

	"github.com/stretchr/testify/require"
)

func TestBuildChangelogsViewReleasedVersions(t *testing.T) {
	t.Parallel()

	doc := &changelog.Document{
		Versions: []changelog.Version{
			{
				Version:  "2026.08.13",
				Date:     "2026-08-13",
				Released: true,
				NewFeatures: []changelog.Entry{
					{Text: "You can now record class times.", Commit: "abc1234"},
				},
			},
			{
				Version:  "2026.07.15",
				Date:     "2026-07-15",
				Released: false,
			},
		},
	}

	v := changelog.BuildChangelogsView(doc, version.Info{
		Commit: "fullcommitsha",
		Tag:    "",
	})

	require.False(t, v.Empty)
	require.Len(t, v.Versions, 1)
	require.Equal(t, "2026.08.13", v.Versions[0].Label)
	require.Equal(t, "2026-08-13", v.Versions[0].DateLabel)
	require.False(t, v.Versions[0].Unreleased)
	require.Len(t, v.Versions[0].Sections, 1)
	require.Equal(t, constants.ChangelogSectionNewFeatures, v.Versions[0].Sections[0].Title)
	require.Equal(t, "You can now record class times.", v.Versions[0].Sections[0].Entries[0].Text)
	require.Equal(t, "abc1234", v.Versions[0].Sections[0].Entries[0].CommitLabel)
	require.Contains(t, v.Versions[0].Sections[0].Entries[0].CommitURL, "abc1234")

	require.Equal(t, "fullcom", v.CurrentVersion)
	require.Contains(t, v.CurrentURL, "fullcommitsha")
}

func TestBuildChangelogsViewUnreleased(t *testing.T) {
	t.Parallel()

	doc := &changelog.Document{
		Unreleased: changelog.Sections{
			Improvements: []changelog.Entry{
				{Text: "TESTIMPROVEMENT"},
			},
		},
		Versions: []changelog.Version{
			{
				Version:  "2026.08.13",
				Date:     "2026-08-13",
				Released: true,
			},
		},
	}

	v := changelog.BuildChangelogsView(doc, version.Info{})

	require.False(t, v.Empty)
	require.Len(t, v.Versions, 2)
	require.Equal(t, constants.ChangelogsUnreleasedVersionLabel, v.Versions[0].Label)
	require.Equal(t, constants.ChangelogsUnreleasedDateLabel, v.Versions[0].DateLabel)
	require.True(t, v.Versions[0].Unreleased)
	require.Equal(t, "TESTIMPROVEMENT", v.Versions[0].Sections[0].Entries[0].Text)
	require.Equal(t, "2026.08.13", v.Versions[1].Label)
}

func TestBuildChangelogsViewEmpty(t *testing.T) {
	t.Parallel()

	v := changelog.BuildChangelogsView(nil, version.Info{})
	require.True(t, v.Empty)
	require.Equal(t, constants.ChangelogsEmptyMessage, v.EmptyMessage)
}
