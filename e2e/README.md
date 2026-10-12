# E2E tests (Playwright)

Browser tests against an isolated SQLite database (`e2e/.data/e2e.db`): class/schedule flows and auth/registration flows.

**Never point `ZION_DB_PATH` at `data/zion.db` in `.env`.** Older e2e-seed versions used that env var and could wipe your dev database. Current seed always resets only `e2e/.data/e2e.db`. If dev data was replaced, restore from backup or Time Machine; the seeded dev DB only contains E2E fixture teachers/students.

## Prerequisites

- Go toolchain and project `.env` (`APP_ENV`, `SECRET`, superuser vars, etc.)
- Node.js 20+

## Credentials

`go run . e2e-seed` (run automatically in Playwright `globalSetup`) creates:

| Field | Value |
|-------|--------|
| Seeded teacher email | `e2e-teacher@zion-english.test` |
| Seeded teacher password | `E2eTeacherPass1!` |
| Admin teacher email / password | `e2e-admin@zion-english.test` / `E2eAdminPass1!` |
| Student | `E2E Test Student` |
| `studentId` / `teacherId` in credentials JSON | Seeded DB ids for direct edit/view routes |
| Sibling student (same parent) | `E2E Sibling Student` |
| Parent name (relationship diagram) | `E2E Parent` |
| Superuser email / password | From `.env` (`SUPERUSER_USERNAME`, `SUPERUSER_PASSWORD`) |
| `overdueScheduledDate` | Yesterday (PHT) ISO date for the seeded overdue scheduled class |

Seeded teacher has **Teacher Pro** (lifetime) for export tests.

The same values are written to `e2e/.auth-credentials.json` for tests.

### Specs

| File | Coverage |
|------|----------|
| `specs/admin-gaps-flows.spec.ts` | Student edit, teacher view, announcement create |
| `specs/auth-flows.spec.ts` | Core auth journey (runs first among auth specs) |
| `specs/auth-flows2.spec.ts` | Invalid login, duplicate registration, superuser register, delete/unapprove, password reset |
| `specs/auth-gaps-flows.spec.ts` | Admin teacher login and students hub |
| `specs/class-flows.spec.ts` | Record/schedule/repeat/series/conduct/cancel/edit/delete class flows |
| `specs/class-gaps-flows.spec.ts` | Trial class, LM on conduct, overdue panel, calendar nav |
| `specs/feature-flags-flows.spec.ts` | Feature flags no-op save |
| `specs/guides-smoke.spec.ts` | Guides index and getting started |
| `specs/insights-gaps-flows.spec.ts` | My earnings, analytics, Pro exports (after `class-flows`) |
| `specs/materials-flows.spec.ts` | Create learning material |
| `specs/nav-smoke.spec.ts` | Teacher nav route smoke |
| `specs/payments-flows.spec.ts` | Reports send payment and teacher confirm (after `class-flows`) |
| `specs/profile-flows.spec.ts` | Profile middle name |
| `specs/resources-gaps-flows.spec.ts` | LM edit/delete, training material, profile password |
| `specs/schedule-flows2.spec.ts` | Overdue conduct, edit move date, series delete (after `class-flows`) |
| `specs/settings-flows.spec.ts` | Settings SECRET reveal (uses `.env` `SECRET`) |
| `specs/student-relationships-flows.spec.ts` | Relationship diagram search and graph |
| `specs/students-flows.spec.ts` | Teacher register student, superuser list |
| `specs/superuser-nav-smoke.spec.ts` | Superuser nav smoke and reports history filter |
| `specs/superuser-smoke.spec.ts` | Students, process, reports |

Playwright runs spec files in **alphabetical order**: `auth-flows.spec.ts` before `auth-flows2.spec.ts` so `e2e-auth-register@zion-english.test` is approved before extended tests. `payments-flows` and `insights-gaps-flows` run after `class-flows` and expect conducted classes from that file.

Password reset in `auth-flows2` only changes the auth-register teacher password, not the seeded class teacher (`e2e-teacher@zion-english.test`).

## Run

From the repo root:

```bash
./run.sh e2e
```

## Pre-commit

Install the repo hook (runs `./run.sh e2e` on every commit):

```bash
./run.sh hooks-install
```

Skip E2E for a single commit:

```bash
SKIP_E2E=1 git commit -m "your message"
```

Debug with a visible browser:

```bash
cd e2e && npx playwright test --headed
```

Environment:

- Project `.env` is loaded automatically from the repo root when Playwright starts (`global-setup` and `playwright.config.ts`). `./run.sh e2e` also sources `.env`.
- `go run . e2e-seed` and `e2e-restore-teacher-password` always use `e2e/.data/e2e.db` (repo root). They ignore `ZION_DB_PATH` so a dev `.env` cannot point seed at `data/zion.db`.
- Playwright `webServer` runs `e2e-seed` before starting the app so the DB is never reset while the server holds the file open.
- The E2E server sets `ZION_DB_PATH` to that same file only on `E2E_PORT`.
- `E2E_PORT` defaults to `18080` so local `8080` dev server can stay up. `./run.sh e2e` stops anything on that port unless `E2E_REUSE_SERVER=1` (also sets `webServer.reuseExistingServer`). Do not use `E2E_REUSE_SERVER=1` with a stale server: `globalSetup` replaces `e2e.db` and needs the port free. Do not `rm` `e2e/.data/e2e.db` while an E2E or dev server on `18080` is running (causes `database disk image is malformed`).
