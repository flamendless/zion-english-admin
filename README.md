<p align="center">
  <img src="static/logo.png" alt="Zion English" width="120">
</p>

# Zion English Admin

Admin portal for Zion English teachers and administrators. Manage students, classes, schedules, and payroll in one place.

## Setup

**Requirements:** Go 1.26+

1. Clone the repository and enter the project directory.
2. Copy the environment file and fill in the required values:

   ```bash
   cp .env.sample .env
   ```

   Set `SECRET`, `SUPERUSER_USERNAME`, and `SUPERUSER_PASSWORD` in `.env`.

3. Make the run script executable and start the dev server:

   ```bash
   chmod +x run.sh
   ./run.sh serve
   ```

   This generates SQL and templ code, runs database migrations, and starts the app.

4. Open [http://localhost:8080/zion-english-admin](http://localhost:8080/zion-english-admin).

## Features

- **Student management:** Register students, assign teachers, and track status and contact details
- **Scheduling & classes:** View schedules, record class outcomes, and track weekly totals
- **Sheet processing:** Import class data from Google Drive spreadsheets and export payroll-ready files
- **Role-based access:** Teachers manage their students and classes; superusers manage the full organization
- **Announcements:** Broadcast info, warning, and critical banners to all or selected teachers
- **Guides:** Step-by-step help for everyday admin tasks
- **Audit logs:** System and processing logs for accountability

## LLM use

Disclaimers:

- This repo started without any use of LLM because it was so simple and basic for that time.
- I now use LLM extensively to further improve the system, but with utmost care for code review. LLM use started in [9e438bc](https://github.com/flamendless/zion-english-admin/commit/9e438bcb49bfaa29ddfc41c18589aae1b4f74bdb) (2026-06-17, "UI and auth fixes"), when Cursor rules and agent skills were first added to the repository.

## License

MIT. See [LICENSE](LICENSE).
