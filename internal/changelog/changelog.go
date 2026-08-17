package changelog

type Document struct {
	Unreleased Sections  `yaml:"unreleased"`
	Versions   []Version `yaml:"versions"`
}

type Version struct {
	Version         string  `yaml:"version"`
	Date            string  `yaml:"date"`
	Released        bool    `yaml:"released"`
	BreakingChanges []Entry `yaml:"breaking_changes"`
	NewFeatures     []Entry `yaml:"new_features"`
	Improvements    []Entry `yaml:"improvements"`
	BugFixes        []Entry `yaml:"bug_fixes"`
	Removed         []Entry `yaml:"removed"`
	Security        []Entry `yaml:"security"`
	Deprecated      []Entry `yaml:"deprecated"`
	Internal        []Entry `yaml:"internal"`
}

type Sections struct {
	BreakingChanges []Entry `yaml:"breaking_changes"`
	NewFeatures     []Entry `yaml:"new_features"`
	Improvements    []Entry `yaml:"improvements"`
	BugFixes        []Entry `yaml:"bug_fixes"`
	Removed         []Entry `yaml:"removed"`
	Security        []Entry `yaml:"security"`
	Deprecated      []Entry `yaml:"deprecated"`
	Internal        []Entry `yaml:"internal"`
}

type Entry struct {
	Text   string `yaml:"text"`
	Commit string `yaml:"commit"`
}

type SectionKey string

const (
	SectionBreakingChanges SectionKey = "breaking_changes"
	SectionNewFeatures     SectionKey = "new_features"
	SectionImprovements    SectionKey = "improvements"
	SectionBugFixes        SectionKey = "bug_fixes"
	SectionRemoved         SectionKey = "removed"
	SectionSecurity        SectionKey = "security"
	SectionDeprecated      SectionKey = "deprecated"
	SectionInternal        SectionKey = "internal"
)

func OrderedSectionKeys() []SectionKey {
	return []SectionKey{
		SectionBreakingChanges,
		SectionNewFeatures,
		SectionImprovements,
		SectionBugFixes,
		SectionRemoved,
		SectionSecurity,
		SectionDeprecated,
		SectionInternal,
	}
}

func (v Version) Sections() Sections {
	return Sections{
		BreakingChanges: v.BreakingChanges,
		NewFeatures:     v.NewFeatures,
		Improvements:    v.Improvements,
		BugFixes:        v.BugFixes,
		Removed:         v.Removed,
		Security:        v.Security,
		Deprecated:      v.Deprecated,
		Internal:        v.Internal,
	}
}

func (s Sections) Entries(key SectionKey) []Entry {
	switch key {
	case SectionBreakingChanges:
		return s.BreakingChanges
	case SectionNewFeatures:
		return s.NewFeatures
	case SectionImprovements:
		return s.Improvements
	case SectionBugFixes:
		return s.BugFixes
	case SectionRemoved:
		return s.Removed
	case SectionSecurity:
		return s.Security
	case SectionDeprecated:
		return s.Deprecated
	case SectionInternal:
		return s.Internal
	default:
		return nil
	}
}

func (s Sections) HasEntries() bool {
	for _, key := range OrderedSectionKeys() {
		if len(s.Entries(key)) > 0 {
			return true
		}
	}
	return false
}

func (d *Document) ReleasedVersions() []Version {
	if d == nil {
		return nil
	}
	out := make([]Version, 0, len(d.Versions))
	for _, v := range d.Versions {
		if v.Released {
			out = append(out, v)
		}
	}
	sortVersions(out)
	return out
}
