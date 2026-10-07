---
name: dev-server
description: Start, restart or check the local Kkal Tracker dev server (Air live reload) on the development VM, and find its URL and logs. Use to verify a change in the browser or when the user asks to run the app.
---

# Dev Server

Development happens on a VM; nobody else runs the server there, so it may be started to verify a change or when the
user asks.

## Start
1. Port: `PORT` in `.env` (currently 8081). Check that nothing listens on it: `ss -ltnp | grep :<port>`.
   If an old `air` or `./main` process holds it, stop that process first
2. Start Air in the background, logging to a file:
   ```bash
   PATH="$HOME/go/bin:$PATH" make watch > tmp/dev-server.log 2>&1
   ```
   Run it as a background task of your runtime, not with `&` inside a foreground command
3. The first build takes a while (npm install + vite + go build). Wait until the log shows the server listening
4. Tell the user the URL: `http://<VM IP>:<port>` (`hostname -I | awk '{print $1}'`)

## Behavior
- Air rebuilds on changes to go, tsx, ts, html, css, js, json files (`.air.toml`); a rebuild includes the frontend
  build, so wait for it before checking the browser
- Build errors: `tmp/build-errors.log`
- `listen tcp :<port>: bind: address already in use` after a restart: an orphaned `./main` (the binary Air builds in the project root) survived. Kill it
  (`pkill -x main` only if it is this project's binary, check with `ps`) and restart Air

## Database
The local database is `data/app.db`. A fresh copy of production: the `prod-db-snapshot` skill (live database from
the server) or `prod-db-from-backup` (nightly backup).
