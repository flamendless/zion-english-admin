package changelog

import (
	"strings"

	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
	"zion-english/internal/version"
)

func BuildChangelogsView(doc *Document, info version.Info) frontend.ChangelogsData {
	v := frontend.ChangelogsData{
		Title:          constants.ChangelogsPageTitle,
		Subtitle:       constants.ChangelogsPageSubtitle,
		EmptyMessage:   constants.ChangelogsEmptyMessage,
		CurrentLabel:   constants.ChangelogsCurrentVersionLabel,
		CurrentVersion: buildCurrentVersion(info),
		CurrentURL:     buildCurrentURL(info),
	}

	if doc == nil {
		v.Empty = true
		return v
	}

	v.Versions = make([]frontend.ChangelogVersionView, 0)
	if doc.Unreleased.HasEntries() {
		v.Versions = append(v.Versions, buildUnreleasedView(doc.Unreleased))
	}

	for _, versionEntry := range doc.ReleasedVersions() {
		v.Versions = append(v.Versions, buildVersionView(versionEntry))
	}

	if len(v.Versions) == 0 {
		v.Empty = true
	}
	return v
}

func buildUnreleasedView(sections Sections) frontend.ChangelogVersionView {
	out := frontend.ChangelogVersionView{
		Label:      constants.ChangelogsUnreleasedVersionLabel,
		DateLabel:  constants.ChangelogsUnreleasedDateLabel,
		Unreleased: true,
	}
	out.Sections = buildSectionViews(sections)
	return out
}

func buildVersionView(v Version) frontend.ChangelogVersionView {
	out := frontend.ChangelogVersionView{
		Label:     v.Version,
		DateLabel: formatChangelogDate(v.Date),
	}
	out.Sections = buildSectionViews(v.Sections())
	return out
}

func buildSectionViews(sections Sections) []frontend.ChangelogSectionView {
	out := make([]frontend.ChangelogSectionView, 0)
	for _, key := range OrderedSectionKeys() {
		entries := sections.Entries(key)
		if len(entries) == 0 {
			continue
		}
		section := frontend.ChangelogSectionView{
			Title:   changelogSectionTitle(key),
			Entries: make([]frontend.ChangelogEntryView, 0, len(entries)),
		}
		for _, entry := range entries {
			section.Entries = append(section.Entries, buildEntryView(entry))
		}
		out = append(out, section)
	}
	return out
}

func buildCurrentVersion(info version.Info) string {
	if tag := strings.TrimSpace(info.Tag); tag != "" {
		return tag
	}
	commit := strings.TrimSpace(info.Commit)
	if commit == "" || commit == "unknown" {
		return "unknown"
	}
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func buildCurrentURL(info version.Info) string {
	commit := strings.TrimSpace(info.Commit)
	if commit == "" || commit == "unknown" {
		return ""
	}
	return version.GitHubCommitURL(commit)
}

func formatChangelogDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	t, err := utils.ParseDatePHT(value)
	if err != nil || t == nil {
		return value
	}
	return utils.DatePHT(*t)
}

func buildEntryView(entry Entry) frontend.ChangelogEntryView {
	out := frontend.ChangelogEntryView{
		Text: strings.TrimSpace(entry.Text),
	}
	commit := strings.TrimSpace(entry.Commit)
	if commit == "" {
		return out
	}
	if len(commit) > 7 {
		out.CommitLabel = commit[:7]
	} else {
		out.CommitLabel = commit
	}
	out.CommitURL = version.GitHubCommitURL(commit)
	return out
}

func changelogSectionTitle(key SectionKey) string {
	switch key {
	case SectionBreakingChanges:
		return constants.ChangelogSectionBreakingChanges
	case SectionNewFeatures:
		return constants.ChangelogSectionNewFeatures
	case SectionImprovements:
		return constants.ChangelogSectionImprovements
	case SectionBugFixes:
		return constants.ChangelogSectionBugFixes
	case SectionRemoved:
		return constants.ChangelogSectionRemoved
	case SectionSecurity:
		return constants.ChangelogSectionSecurity
	case SectionDeprecated:
		return constants.ChangelogSectionDeprecated
	case SectionInternal:
		return constants.ChangelogSectionInternal
	default:
		return string(key)
	}
}
