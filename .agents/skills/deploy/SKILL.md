---
name: deploy
description: Release Kkal Tracker to production - checks, commits on develop, deploy.sh (merge into master, image from master, version commit, server rollout) and verification. Use when the user asks to release, deploy or ship to prod.
---

# Deploy to Production

Production runs ONLY what is on `master`, and nothing is pushed to git before the artifact it describes exists:
the image goes to Docker Hub before the version commit, the version commit before the deploy. The order is fixed:
checks → commits on `develop` → `deploy.sh` (push develop → merge into master → image from master → version
commit on master → sync develop → deploy). Never build a production image from `develop` or from an uncommitted tree.

## 1. Checks (feature is done, not committed yet)
```bash
go vet ./... && go test ./...        # tests need web/dist: run `make build-frontend` first
cd web && npx tsc --noEmit -p tsconfig.app.json && npm run lint
```
Verify the change in the running dev server when it has a visible side (the `dev-server` skill).

## 2. Commits on develop
- Work on `develop`, one commit per logical change, message in English, no AI attribution
- `git push origin develop` only when the user asked to push/deploy

## 3. Release with deploy.sh
```bash
./deploy.sh --tag=X.Y.Z            # asks for confirmation; `--dry-run` prints the steps
```
The script reads from stdin for the confirmation: pipe `y` (`echo y | ./deploy.sh --tag=X.Y.Z`) when running it
non-interactively, and only after the user agreed to the release.

Pick the next version from `version.txt` (`vX.Y.Z`): patch for fixes, minor for features. The script:
1. Checks: on `develop`, clean tree, `develop`/`master` in sync with origin, `master` contains nothing that `develop` lacks, `vX.Y.Z` is new
2. Pushes `develop`
3. Merges `develop` into `master` locally (not pushed)
4. Builds the image from the `master` checkout (`build-and-push.sh X.Y.Z --push`: removes `web/dist`, `--no-cache`,
   tags `X.Y.Z` and `latest`) and pushes it to Docker Hub. If this fails, local `master` is reset to `origin/master`:
   nothing has left the machine and the same tag can be retried
5. Writes `version.txt` and `kubernetes/base/deployment.yaml` (`image: ypeskov/kcal-tracker:X.Y.Z`), commits
   `vX.Y.Z` on `master`, tags it, pushes `master` and the tag
6. Merges `master` back into `develop`, pushes `develop`
7. SSH to the server (`.deploy.env`: `SSH_HOST`, `K8S_REPO_SERVER`), checks the repo is on `master`, `git pull`,
   `kubectl apply -k kubernetes/overlays/prod`, waits for the rollout
8. Ends on `develop`

Do NOT pass `--platform`: the image is built for the host architecture. `--skip-deploy` stops after the image is pushed.
The build takes several minutes: run it in the background and wait for it instead of polling.

## 4. Verify
```bash
ssh $SSH_HOST  # then, with KUBECONFIG=/home/kuber/.kube/config:
kubectl get pods -l app=kkal-tracker
kubectl get deployment kkal-tracker -o jsonpath='{.spec.template.spec.containers[0].image}'
kubectl logs POD_NAME            # --previous: logs from the previous run if it crashed
```
Then open https://kcal.peskov.info and check that it responds.

Fallbacks when the script cannot be used: `kubectl set image deployment/kkal-tracker kkal-tracker=ypeskov/kcal-tracker:X.Y.Z`,
or `kubectl rollout restart deployment kkal-tracker` when the tag did not change.

## Migrations
Neither the deploy nor the server applies migrations (only `cmd/migrate` does, and it is not in the image). If the
release contains new files in `migrations/` since the last `vX.Y.Z` tag, stop before the release and agree with the
user how they are applied to the production database.

## Common errors
| Error | Cause | Solution |
|-------|-------|----------|
| `exec format error` | Image built for the wrong architecture | Rebuild without `--platform` |
| `CrashLoopBackOff` | Application error | `kubectl logs --previous` |
| Pod doesn't update | Kubernetes caches an image with the same tag | Use a new version tag |
| `permission denied` on k3s.yaml | kubectl uses the root kubeconfig | `export KUBECONFIG=/home/kuber/.kube/config` |
