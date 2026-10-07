# Infrastructure

## Build
1. The frontend builds to `web/dist/` (Vite); Go embeds it via `web/embed.go` (`//go:embed`)
2. A single binary serves both the API and the static files
3. Air (`.air.toml`) watches go, tsx, ts, html, css, js, json; excludes `web/dist`, `web/node_modules`, `tmp`, `bin`,
   `data` and `web/package-lock.json` (npm install touches it on every build). Build errors go to `tmp/build-errors.log`

## Docker
- `Dockerfile`: multi-stage (node → golang → distroless static, non-root); no `.env` in the image
- `build-and-push.sh TAG [--push] [--platform=PLATFORM]`: always `--no-cache`, removes `web/dist/` first, tags `TAG`
  and `latest`, writes `vTAG` to `version.txt`. Called by `deploy.sh`; production images are built only from `master`
- Image: `ypeskov/kcal-tracker`. Build for the host architecture, without `--platform`

## Kubernetes
- Ingress controller: Traefik v3.3, cluster-level, deployed separately via the k8s infra repo and shared with other
  apps (IngressClass `traefik`, controller `traefik.io/ingress-controller`)
- `kubernetes/base/`: Deployment, Service, Ingress (`ingressClassName: nginx` for local dev), PV/PVC, backup ConfigMap
  and CronJob
- `kubernetes/overlays/prod/`: ingress patched to `ingressClassName: traefik` with TLS
  (`traefik.ingress.kubernetes.io/router.tls: "true"`); ConfigMap `kkal-tracker-env` generated from `.env`
  (non-sensitive settings) and Secret `kkal-tracker-secrets` from `.env.secret` (`JWT_SECRET`, `SMTP_PASSWORD`,
  `OPENAI_API_KEY`, `GDRIVE_OAUTH_TOKEN`); both files exist only on the server, see the `*.sample` files
- Server: `SSH_HOST` and `K8S_REPO_SERVER` from `.deploy.env` (git-ignored, see `.deploy.env.sample`); the server repo
  is on `master`; kubectl needs `KUBECONFIG=/home/kuber/.kube/config`
- The PV is a hostPath: the production database is `<K8S_REPO_SERVER>/data/app.db` on the server (`/data/app.db` in
  the pod)

## Backups
- CronJob (image `ypeskov/kkal-tracker-backup`, built from `kubernetes/backup/Dockerfile`: pinned rclone + sqlite3;
  receives only the `GDRIVE_*` settings) runs nightly at 02:00 Europe/Sofia (file names carry the UTC time):
  `sqlite3 .backup` → gzip →
  `<K8S_REPO_SERVER>/data/backups/kkal_tracker_backup_YYYYMMDD_HHMMSS.db.gz` → upload to Google Drive
  `services/kkal-tracker/backups`
- Retention: `RETENTION_DAYS` (7): files older than that are deleted on the server (`find -mtime +7`, so 8-9 files
  remain) and on Google Drive

## Environment variables
See `.env.sample`. The main ones:
```
PORT=8080
JWT_SECRET=...                 # unset ENVIRONMENT means production: a strong JWT_SECRET (32+ chars) is then required
LOG_LEVEL=debug
ENVIRONMENT=development
# TRUSTED_PROXIES=10.42.0.0/16 # CIDR ranges whose X-Forwarded-For is trusted (default: loopback and private networks)
DATABASE_PATH=./data/app.db
OPENAI_API_KEY=sk-...          # AI providers are activated when their keys are set; OPENAI_BASE_URL is optional
GDRIVE_OAUTH_TOKEN={...}       # backups; rclone OAuth2 token
GDRIVE_FOLDER_PATH=/services/kkal-tracker/backups
```
