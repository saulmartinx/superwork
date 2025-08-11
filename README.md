[README.md](https://github.com/user-attachments/files/21710509/README.md)
# Superwork

A web-based **Customer Relationship & Task Management** app with built‑in **time tracking** and light **location (geocoding) support**. The app combines a simple CRM (organizations & contacts), Kanban‑style workflows for tasks/deals, activity scheduling, and a timer for work logs — all in one interface.

> An Estonian user manual is included at `public/documents/kasutusjuhend.pdf`.

---

## Features

- **Tasks & Pipelines:** Customizable workflows (stages) and a Kanban board with drag‑and‑drop between stages.
- **CRM Lite:** Organizations and contact persons that can be linked to tasks and activities.
- **Activities:** Schedule calls/meetings/follow‑ups and mark them done.
- **Team View:** See tasks by assignee and rebalance work with drag‑and‑drop.
- **Time Tracking:** Start/stop timers and export work logs to CSV.
- **Geocoding:** Optional Google Geocoding for addresses (stores coordinates).
- **OAuth login (optional):** Google and Facebook sign‑in.
- **Simple stack:** Go backend + single‑page UI (Backbone.js, jQuery, Bootstrap) served from `public/`.

---

## Quick start

### Prerequisites
- **Go** (modern Go toolchain)
- **PostgreSQL** (with the `pgcrypto` extension available)
- `git` (optional, if cloning)

### 1) Get the code
```bash
git clone <your-repo-url> superwork
cd superwork
```

### 2) Prepare the database
Create a role and database named `superwork`, then load the schema:
```bash
# create role and database (no password expected by default code)
createuser superwork --no-password
createdb superwork -O superwork

# initialize schema & seed data
psql -U superwork -d superwork -f db/setup.sql
```
> The schema uses `gen_random_uuid()` from `pgcrypto`. If the extension is not enabled in your cluster, ensure the `CREATE EXTENSION pgcrypto;` statement in `db/setup.sql` succeeds.

### 3) Configure (environment variables)
The app reads configuration from environment variables (defaults exist for dev). Common ones:

```bash
# Server
export SUPERWORK_PORT=8000
export SUPERWORK_PUBLIC="public"
export SUPERWORK_LOG=true                # write logs to file
export SUPERWORK_LOGFILE="superwork.log"
export SUPERWORK_SECRET="<random-long-string>"

# Integrations (optional)
export SUPERWORK_GEOCODE_API_KEY="<google-geocoding-api-key>"
export SUPERWORK_BUGSNAG_API_KEY="<bugsnag-api-key>"
export SUPERWORK_ADMIN_EMAIL="support@superwork.io"
export SUPERWORK_ZOHO_PASSWORD="<smtp-password>"

# OAuth callbacks (optional; match your OAuth app settings)
export SUPERWORK_GOOGLE_REDIRECT="http://localhost:8000/api/oauth2callback/google"
export SUPERWORK_FACEBOOK_REDIRECT="http://localhost:8000/api/oauth2callback/facebook"
```

> Note: Some legacy configs/scripts use `GOSHAREWORK_*` for OAuth redirect variables. Prefer `SUPERWORK_GOOGLE_REDIRECT`/`SUPERWORK_FACEBOOK_REDIRECT` or update the code to use a single prefix consistently.

### 4) Build & run
```bash
# build
go build              # produces ./superwork (or superwork.exe on Windows)

# run
./superwork
# App will start on http://localhost:8000 (or SUPERWORK_PORT)
```

Open `http://localhost:8000` in your browser.

---

## Production notes

- **Binary for Linux:** `make build_linux` produces `superwork_linux`.
- **Asset minification:** `make minify` (writes `public/js/lib.js` and `public/css/lib.css`), `make dist` bundles optimized assets into `dist/`.
- **Service & proxy examples:** See `config/etc/init/superwork.conf` (Upstart example) and `config/etc/nginx/sites-enabled/default` (Nginx proxy pointing `/api` to the Go app and serving static files from `public/`). Adjust paths, domain and env vars for your server.
- **OAuth credentials:** Replace hard‑coded IDs/secrets in `oauth.go` with your own or move them to environment variables before deploying publicly.
- **Security:** Always set a strong `SUPERWORK_SECRET`. Review any default credentials and email settings before exposing the app.

---

## Folder structure (high level)

```
.
├── db/                      # PostgreSQL schema & seed (uses pgcrypto)
├── public/                  # SPA assets (HTML/CSS/JS, fonts, images)
│   ├── index.html
│   ├── js/                  # Backbone app, libs, i18n (incl. Estonian locale)
│   ├── css/
│   ├── fonts/
│   └── documents/kasutusjuhend.pdf
├── config/                  # Example Upstart & Nginx configs
├── *.go                     # Go backend: routes, handlers, models, oauth, email, geocode
└── Makefile                 # build/minify/dist helpers
```

---

## Usage (basics)

1. **Sign up / Sign in** (email or OAuth if configured).
2. **Create/Invite users** under *Settings → Users*.
3. **Define workflows** (pipelines/stages) under *Settings*.
4. **Add organizations & contacts** (optional CRM).
5. **Create tasks/deals**, assign to team members, move on the **Board**.
6. **Schedule activities** and **track time** in the Work view.
7. **Export CSV** for time entries/reports when needed.

For a step‑by‑step guide with screenshots, see the Estonian manual at `public/documents/kasutusjuhend.pdf`.

---

## Testing

```bash
make coverage   # run tests and open coverage report (generates coverage.out)
```

---

## License

No explicit license file is included. Until a license is added, treat the code as **all rights reserved** by the authors.

---

## Credits

Originally developed by **GTM Solutions / superwork.io**.  
Contact in historical materials: `support@superwork.io`.

