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

- **Student management** — Register students, assign teachers, and track status and contact details
- **Scheduling & classes** — View schedules, record class outcomes, and track weekly totals
- **Sheet processing** — Import class data from Google Drive spreadsheets and export payroll-ready files
- **Role-based access** — Teachers manage their students and classes; superusers manage the full organization
- **Announcements** — Broadcast info, warning, and critical banners to all or selected teachers
- **Guides** — Step-by-step help for everyday admin tasks
- **Audit logs** — System and processing logs for accountability

## License

MIT — see [LICENSE](LICENSE).
