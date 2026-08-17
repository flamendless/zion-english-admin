# Changelog YAML schema

File: `data/changelogs.yaml`

## Top-level structure

```yaml
unreleased:
  breaking_changes: []
  new_features: []
  improvements: []
  bug_fixes: []
  removed: []
  security: []
  deprecated: []
  internal: []

versions:
  - version: "2026.08.17"   # calver label shown on page
    date: "2026-08-17"      # ISO date; displayed in PHT on page
    released: true          # false = omitted from public page
    breaking_changes: []
    new_features: []
    improvements: []
    bug_fixes: []
    removed: []
    security: []
    deprecated: []
    internal: []
```

## Entry shape

```yaml
- text: "User-friendly description of the change."
  commit: "abc1234"   # optional short SHA for GitHub commit link
```

## Rules

- `unreleased` is shown on `/changelogs` when it has at least one entry (prepended above released versions).
- Only `versions` with `released: true` appear on the page, sorted newest-first by `date`.
- All eight section keys must exist on `unreleased` and each version (use `[]` when empty).
- Version labels use calver (`YYYY.MM.DD`).
- The `/changelogs` page requires authentication (superuser or teacher role).

## Reader implementation

- Package: `internal/changelog`
- Handler: `cmd/changelogs.go`
- View builder: `internal/changelog/view.go`
- Template: `frontend/changelogs.templ`
