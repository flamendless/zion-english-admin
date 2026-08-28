---
name: update-changelog
description: >-
  Append unreleased changelog entries to data/changelogs.yaml from git history.
  Use when the user asks to update the changelog, add release notes, or publish a version.
disable-model-invocation: true
---

# Update Changelog

Maintain `data/changelogs.yaml`: the single source of truth for the `/changelogs` page.

## When to use

- User explicitly asks to update, append, or release the changelog
- User says "add changelog entry", "update changelogs", or "release version"

Do **not** run automatically after feature work.

## File location

- Default: `data/changelogs.yaml`
- Override: `CHANGELOG_PATH` environment variable

Read the schema reference: [references/schema.md](references/schema.md)

## Workflow: append unreleased entry

```
Task Progress:
- [ ] Read data/changelogs.yaml
- [ ] Collect candidate changes (git diff + commits not already referenced)
- [ ] Draft user-friendly bullets with correct section and commit SHA
- [ ] Append under unreleased only (dedupe by commit or near-duplicate text)
- [ ] Run validation checklist below
```

### Collecting changes

1. Read `data/changelogs.yaml` and collect every `commit` value already present (released + unreleased).
2. Run `git diff` (staged and unstaged) for current work context.
3. Run `git log` for commits whose full or short SHA is **not** already in the YAML.
4. Prefer one bullet per **user-visible** feature or fix; skip refactors and test-only commits unless admin-visible.

### Insertion rules

- **Append only** under `unreleased`: never edit released `versions` unless the user explicitly asks.
- Place entries in the correct section key (see schema).
- Each entry: `text` (required), `commit` (short SHA, recommended for GitHub link).
- Dedupe: skip if the same `commit` or substantially the same `text` already exists anywhere in the file.

## Workflow: release version

Only when the user asks to **release**, **publish**, or **ship** a changelog version:

1. Choose calver label `YYYY.MM.DD` (today UTC or user-specified) and ISO `date`.
2. Insert a new entry at the **top** of `versions:` with `released: true`.
3. Copy all `unreleased` section arrays into that version entry.
4. Reset every `unreleased` section to an empty array `[]`.
5. Validate checklist below.

## Section keys → user-facing headings

| YAML key | Page heading | Use for |
|----------|--------------|---------|
| `breaking_changes` | Breaking changes | Admin-impacting breaks |
| `new_features` | New features | New admin-visible capabilities |
| `improvements` | Improvements | UX/clarity/performance improvements |
| `bug_fixes` | Bug fixes | Fixes teachers or superusers would notice |
| `removed` | Removed | Removed admin-facing features |
| `security` | Security | Security fixes (describe impact carefully) |
| `deprecated` | Deprecated | Features marked for removal |
| `internal` | Internal | Dev tooling, ops, non-admin surfaces |

## Writing rules

- **Voice:** user-friendly, admin-focused: "You can now…", "Fixed an issue where…"
- **Granularity:** one user-visible change per bullet
- **Commits:** link via short SHA; use `version.GitHubCommitURL` pattern (`https://github.com/flamendless/zion-english-admin/commit/<sha>`)
- **Dates:** ISO `YYYY-MM-DD` on released versions; page displays PHT
- **Unreleased:** shown at the top of `/changelogs` when any section has entries; cleared on release
- Never use `TEST<FIELD>` tokens in changelog copy

## Validation checklist (no script)

- [ ] Only allowed section keys used
- [ ] Every entry has non-empty `text`
- [ ] Released versions have `version`, `date` (ISO), and `released: true`
- [ ] `unreleased` sections exist (empty arrays if nothing pending)
- [ ] No duplicate commit references for the same change
- [ ] YAML indentation is consistent (2 spaces)
- [ ] User-facing copy; no raw commit subjects pasted as bullets

## Editing rules

1. Use the Read tool before editing; use StrReplace/Write: never `echo >>` in shell.
2. Preserve unrelated entries and formatting.
3. Minimal diff: only touch changelog sections being updated.

## Examples

**Append unreleased improvement:**

```yaml
improvements:
  - text: "Class schedule is easier to read with an updated layout."
    commit: "0d46bad"
```

**Release entry (top of versions):**

```yaml
versions:
  - version: "2026.08.17"
    date: "2026-08-17"
    released: true
    breaking_changes: []
    new_features:
      - text: "You can now record class start and end times when logging a class."
        commit: "48d9c9e"
    improvements: []
    bug_fixes: []
    removed: []
    security: []
    deprecated: []
    internal: []
```
