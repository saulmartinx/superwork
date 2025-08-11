# Deploying Superwork on Render (Free Tier)

This guide shows how to deploy **Superwork** (Go + PostgreSQL) on **Render** with a public URL.

> TL;DR — Connect your GitHub repo on Render → add a free PostgreSQL → set `DATABASE_URL` and `SUPERWORK_PORT=${PORT}` → run `go build -o superwork ./` and start `./superwork`.

---

## 0) Code note (DATABASE_URL support)
The repository’s `db_connection.go` should read the Postgres DSN from the environment (e.g. `DATABASE_URL`). If you use the patched file provided, no other code changes are needed. Otherwise, update the code to prefer `os.Getenv("DATABASE_URL")` and fall back to the local dev DSN.

---

## 1) Create the Web Service
1. Go to **https://render.com** → **New → Web Service** → connect your GitHub repo (the Superwork repo).
2. **Environment**: keep Docker **off** (native Go build).
3. **Build Command**:  
   ```bash
   go build -o superwork ./
   ```
4. **Start Command**:  
   ```bash
   ./superwork
   ```
Render will provision an URL like `https://<your-app>.onrender.com` with HTTPS.

---

## 2) Add PostgreSQL
1. In Render dashboard: **New → PostgreSQL → Free**.
2. Open the DB and copy **External Connection** (full URL), e.g.  
   `postgresql://USER:PASSWORD@HOST:PORT/DB?sslmode=require`

> The free Postgres on Render is for dev; it **expires after ~30 days**. For long‑term use, consider an external free tier (Neon, Supabase, ElephantSQL) and place its URL into `DATABASE_URL`.

---

## 3) Configure Environment Variables (Web Service → Environment)
Set the following keys:
- `DATABASE_URL=<paste Render DB External Connection string>`
- `SUPERWORK_PORT=${PORT}`  ← **important**, app must bind Render’s assigned port
- `SUPERWORK_PUBLIC=public`
- `SUPERWORK_SECRET=<generate a long random string>`

Optional (if you need these):
- `SUPERWORK_GEOCODE_API_KEY=<google-geocoding-api-key>`
- `SUPERWORK_BUGSNAG_API_KEY=<bugsnag-key>`
- `SUPERWORK_GOOGLE_REDIRECT=https://<your-app>.onrender.com/api/oauth2callback/google`
- `SUPERWORK_FACEBOOK_REDIRECT=https://<your-app>.onrender.com/api/oauth2callback/facebook`

Click **Save**.

---

## 4) Initialize the Database Schema
From your local machine (Windows):
- Install **psql** (PostgreSQL client) or use a GUI (DBeaver/TablePlus).
- Run the schema against the managed DB:
  ```powershell
  psql "<DATABASE_URL_FROM_RENDER>" -f db/setup.sql
  ```
> Ensure the URL includes `sslmode=require` (Render default). If using a GUI, open the connection and execute `db/setup.sql` there.

---

## 5) Deploy & Test
- **Auto deploy:** push to `main` → Render builds & deploys automatically.
- **Manual:** press **Deploy** in Render dashboard.
- Open your URL and sign up/login. You should see the Superwork UI.

---

## Troubleshooting
- **DB connection errors** → verify `DATABASE_URL`, and that `db/setup.sql` was executed.
- **Port/bind error** → make sure `SUPERWORK_PORT=${PORT}` is set.
- **Cold start delay** → free services may sleep after inactivity; first request can be slower.
- **Static files** → `public/` is served by the app; keep `SUPERWORK_PUBLIC=public`.

---

## Notes & Alternatives
- **Security**: always set a strong `SUPERWORK_SECRET` before exposing the app.
- **External DBs**: for persistence beyond Render’s ~30 days, use Neon/Supabase/ElephantSQL and keep the same `DATABASE_URL` interface.
- **Fly.io alternative**: more “always‑on” free resources and managed Postgres, but requires a credit card on signup.

Good luck!