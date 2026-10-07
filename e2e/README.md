# E2E tests (Playwright)

Browser tests against an isolated SQLite database (`e2e/.data/e2e.db`): class/schedule flows and auth/registration flows.

## Prerequisites

- Go toolchain and project `.env` (`APP_ENV`, `SECRET`, superuser vars, etc.)
- Node.js 20+

## Credentials

`go run . e2e-seed` (run automatically in Playwright `globalSetup`) creates:

| Field | Value |
|-------|--------|
| Seeded teacher email | `e2e-teacher@zion-english.test` |
| Seeded teacher password | `E2eTeacherPass1!` |
| Student | `E2E Test Student` |
| Superuser email / password | From `.env` (`SUPERUSER_USERNAME`, `SUPERUSER_PASSWORD`) |
| `overdueScheduledDate` | Yesterday (PHT) ISO date for the seeded overdue scheduled class |

The same values are written to `e2e/.auth-credentials.json` for tests.

### Specs

| File | Coverage |
|------|----------|
| `specs/auth-flows.spec.ts` | Core auth journey (runs first among auth specs) |
| `specs/auth-flows2.spec.ts` | Invalid login, duplicate registration, superuser register, delete/unapprove, password reset |
| `specs/class-flows.spec.ts` | Record/schedule/repeat/series/conduct/cancel/edit/delete class flows |
| `specs/materials-flows.spec.ts` | Create learning material |
| `specs/nav-smoke.spec.ts` | Teacher nav route smoke |
| `specs/profile-flows.spec.ts` | Profile middle name |
| `specs/schedule-flows2.spec.ts` | Overdue conduct, edit move date, series delete (after `class-flows`) |
| `specs/students-flows.spec.ts` | Teacher register student, superuser list |
| `specs/superuser-smoke.spec.ts` | Students, process, reports |

Playwright runs spec files in **alphabetical order**: `auth-flows.spec.ts` before `auth-flows2.spec.ts` so `e2e-auth-register@zion-english.test` is approved before extended tests.

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
- `ZION_DB_PATH` defaults to `e2e/.data/e2e.db` during E2E (via Playwright `webServer` and seed).
- `E2E_PORT` defaults to `18080` so local `8080` dev server can stay up. `./run.sh e2e` stops anything on that port unless `E2E_REUSE_SERVER=1` (also sets `webServer.reuseExistingServer`). Do not run a dev server on `18080` during E2E.
