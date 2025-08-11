# Deploying Superwork to Render (Free Tier)

This guide shows how to deploy the Superwork Go web application and its PostgreSQL database to [Render](https://render.com/) using their free tier. Render offers a free web service instance (512 MB RAM) that can run continuously and a managed PostgreSQL database that is free for 30 days. You will receive a public URL to share with your team.

## Prerequisites

- A GitHub repository containing the Superwork code (this repository).  
- A Render account (sign up at https://render.com). Render does **not** require a credit card for the free tier.
- The `psql` CLI or another PostgreSQL client installed locally to initialize the database.

## Steps

### 1. Connect your GitHub repo to Render

1. Log in to your Render dashboard.
2. Click **New > Web Service** and choose **From a Git Repository**.
3. Select your `saulmartinx/superwork` repository. Render will detect the Go project automatically.

### 2. Configure the web service

On the service creation page:

- **Name:** Choose a name for your service.
- **Build Command:**

  ```bash
  go build -o superwork ./
  ```

- **Start Command:**

  ```bash
  ./superwork
  ```

- **Environment Variables:** Add the following key‑value pairs (click **Advanced > Environment** to add them):

  | Key                   | Value                                               | Notes                                            |
  |----------------------|------------------------------------------------------|--------------------------------------------------|
  | `SUPERWORK_PORT`     | `${PORT}`                                            | Render assigns an internal port via `$PORT`; pass it to the app. |
  | `SUPERWORK_PUBLIC`   | `public`                                             | Static assets directory.                         |
  | `SUPERWORK_SECRET`   | _a long random string_                               | Secret for session cookies; change for production. |
  | `DATABASE_URL`       | (leave blank for now)                                | Will be set by Render when you add the database. |

Leave other configuration options at their defaults. Click **Create Web Service** when done. Render will start an initial build (it will fail until the database is configured; that’s okay).

### 3. Add a PostgreSQL database

1. From the Render dashboard, click **New > PostgreSQL**.
2. Give the database a name (e.g., `superwork-db`) and choose the **Free** plan.
3. After creation, go to the database details page and note the **Internal URL** and **External URL**. Render will automatically inject the `DATABASE_URL` environment variable into your web service with the internal connection string.

### 4. Initialize the database schema

To create the tables, run the SQL setup script against the new database. You can do this from your local machine:

```bash
# Replace <EXTERNAL_URL> with the URL shown in the Render dashboard (be sure to include the credentials and database name)
psql "<EXTERNAL_URL>" -f db/setup.sql
```

This loads all necessary tables, triggers, and sample data (and enables the `pgcrypto` extension). If you don't have `psql` installed, you can use a GUI like DBeaver to connect using the External URL and execute `db/setup.sql` there.

### 5. Redeploy the service

Return to your Superwork web service on Render and click **Manual Deploy > Deploy latest commit**. The service will build and start successfully now that the database exists. Within a minute or two, Render will assign a public URL such as `https://superwork.onrender.com`.

### 6. Verify the deployment

Open the provided URL in your browser. You should see the Superwork login page. Sign up or sign in (if you configured OAuth, you can use Google or Facebook). Then create workflows, clients, tasks, etc., as described in the project README.

## Notes

- **Database expiration:** Render’s free PostgreSQL instances last for 30 days. Before the end of the term, you can upgrade to a paid plan or recreate the database (export data if needed). Alternatively, connect an external free Postgres provider such as [Neon](https://neon.tech/) or [Supabase](https://supabase.com/) by setting the `DATABASE_URL` manually.
- **Environment variables:** The application also honors `SUPERWORK_GEOCODE_API_KEY`, `SUPERWORK_BUGSNAG_API_KEY`, `SUPERWORK_ADMIN_EMAIL`, and related variables. Set these if you need geocoding, error monitoring, or outgoing email.
- **Local testing:** To run the app locally, use the provided `Makefile` or run `go build` then `./superwork`. Create a local database (named `superwork`) and load `db/setup.sql`. The local connection string defaults to `user=superwork dbname=superwork sslmode=disable`.

You are now ready to host Superwork with a public URL. Happy managing!