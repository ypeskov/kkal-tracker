#!/usr/bin/env bash
set -euo pipefail

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# ─── Paths ────────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_ENV="${SCRIPT_DIR}/.deploy.env"
DEPLOYMENT_YAML="${SCRIPT_DIR}/kubernetes/base/deployment.yaml"
VERSION_FILE="${SCRIPT_DIR}/version.txt"
IMAGE_NAME="ypeskov/kcal-tracker"

# ─── Load config ──────────────────────────────────────────────────────────────
if [ ! -f "$DEPLOY_ENV" ]; then
    echo -e "${RED}Error: .deploy.env not found${NC}"
    echo ""
    echo "Create ${DEPLOY_ENV} with:"
    echo "  SSH_HOST=kuber@your-server-ip"
    echo "  K8S_REPO_SERVER=/home/kuber/path-to-repo"
    echo "  PLATFORM=linux/arm64"
    exit 1
fi

source "$DEPLOY_ENV"

# Validate required config
for var in SSH_HOST K8S_REPO_SERVER; do
    if [ -z "${!var:-}" ]; then
        echo -e "${RED}Error: ${var} is not set in .deploy.env${NC}"
        exit 1
    fi
done

PLATFORM="${PLATFORM:-}"

# ─── Defaults ─────────────────────────────────────────────────────────────────
TAG=""
SKIP_DEPLOY=false
DRY_RUN=false

# ─── Help ─────────────────────────────────────────────────────────────────────
show_help() {
    cat << EOF
Usage: $(basename "$0") --tag=TAG [OPTIONS]

Automated deployment pipeline for Kkal Tracker.
Production is built only from master, and nothing is pushed to git before the
artifact it describes exists: the image is pushed before the version commit,
the version commit before the deploy.

Performs the full cycle:
  1. Validate git state (on develop, clean tree, develop and master in sync with origin)
  2. Push develop
  3. Merge develop into master locally (not pushed yet)
  4. Build & push the Docker image from the master checkout
     (on failure master is reset to origin/master and nothing has left the machine)
  5. Write the version to version.txt and the k8s deployment manifest, commit vX.Y.Z on master,
     tag it, push master and the tag
  6. Merge master back into develop, push develop
  7. SSH to server, git pull (master), kubectl apply
  8. Back on develop

REQUIRED:
    --tag=TAG         Version tag (e.g., 5.5.0)

OPTIONS:
    --help            Show this help message
    --skip-deploy     Do everything except the server deploy (step 7)
    --dry-run         Show what would be done without executing
    --platform=PLAT   Override Docker platform (default: host architecture)

EXAMPLES:
    $(basename "$0") --tag=5.5.0
    $(basename "$0") --tag=5.5.0 --skip-deploy
    $(basename "$0") --tag=5.5.0 --dry-run
    $(basename "$0") --tag=5.5.0 --platform=linux/arm64
EOF
}

# ─── Parse arguments ──────────────────────────────────────────────────────────
if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

while [ $# -gt 0 ]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --tag=*)
            TAG="${1#--tag=}"
            shift
            ;;
        --platform=*)
            PLATFORM="${1#--platform=}"
            shift
            ;;
        --skip-deploy)
            SKIP_DEPLOY=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo -e "${RED}Error: Unknown option '$1'${NC}"
            echo "Use --help to see available options."
            exit 1
            ;;
    esac
done

# Validate required tag
if [ -z "$TAG" ]; then
    echo -e "${RED}Error: --tag=TAG is required${NC}"
    echo "Use --help to see usage."
    exit 1
fi

if ! [[ "$TAG" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: tag must look like X.Y.Z (got '${TAG}')${NC}"
    exit 1
fi

# ─── Helpers ──────────────────────────────────────────────────────────────────
STEP_NUM=0

step() {
    STEP_NUM=$((STEP_NUM + 1))
    echo ""
    echo -e "${BLUE}${BOLD}═══ Step ${STEP_NUM}: $1 ═══${NC}"
}

info() {
    echo -e "  ${GREEN}✓${NC} $1"
}

warn() {
    echo -e "  ${YELLOW}⚠${NC} $1"
}

fail() {
    echo -e "${RED}Error: $1${NC}"
    exit 1
}

run() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "  ${YELLOW}[dry-run]${NC} $*"
    else
        "$@"
    fi
}

# Whatever happens after the checkout of master, finish on develop. Until master is pushed
# (MASTER_PUSHED), the local merge and the version commit are discarded, so a failed build
# leaves master exactly as on origin and the same tag can be released again
MASTER_PUSHED=false
return_to_develop() {
    if [ "$(git -C "$SCRIPT_DIR" branch --show-current)" = "master" ]; then
        if [ "$MASTER_PUSHED" = false ] && [ "$DRY_RUN" = false ]; then
            warn "Resetting local master to origin/master"
            git -C "$SCRIPT_DIR" tag -d "v${TAG}" > /dev/null 2>&1 || true
            git -C "$SCRIPT_DIR" reset -q --hard origin/master
        fi
        git -C "$SCRIPT_DIR" checkout -q develop || warn "Could not switch back to develop"
    fi
}

# ─── Build platform flag ─────────────────────────────────────────────────────
PLATFORM_FLAG=""
if [ -n "$PLATFORM" ]; then
    PLATFORM_FLAG="--platform=${PLATFORM}"
fi

# ─── Summary ──────────────────────────────────────────────────────────────────
echo -e "${BOLD}Kkal Tracker Deploy${NC}"
echo -e "  Tag:      ${GREEN}${TAG}${NC}"
echo -e "  Image:    ${IMAGE_NAME}:${TAG}"
[ -n "$PLATFORM" ] && echo -e "  Platform: ${PLATFORM}"
echo -e "  Server:   ${SSH_HOST}"
[ "$SKIP_DEPLOY" = true ] && warn "Skipping server deploy"
[ "$DRY_RUN" = true ]     && warn "DRY RUN — no changes will be made"
echo ""
read -p "Proceed? [y/N] " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

cd "$SCRIPT_DIR"

# ─── Step 1: Validate git state ──────────────────────────────────────────────
step "Validating git state"

CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "develop" ]; then
    fail "Must be on 'develop' branch (currently on '${CURRENT_BRANCH}')"
fi
info "On branch: develop"

if ! git diff --quiet || ! git diff --cached --quiet; then
    fail "Working tree has uncommitted changes. Commit or stash them first."
fi
info "Working tree is clean"

git fetch -q origin develop master
if [ -n "$(git log --oneline develop..origin/develop)" ]; then
    fail "origin/develop has commits that are not in the local develop. Pull first."
fi
if [ -n "$(git log --oneline master..origin/master)" ]; then
    fail "origin/master has commits that are not in the local master. Pull first."
fi
if [ -n "$(git log --oneline develop..master)" ]; then
    fail "master has commits that are not in develop. Merge master into develop first."
fi
info "develop and master are in sync with origin"

if git rev-parse -q --verify "refs/tags/v${TAG}" > /dev/null; then
    fail "Tag v${TAG} already exists"
fi
if [ "$(cat "$VERSION_FILE")" = "v${TAG}" ]; then
    fail "version.txt already says v${TAG}; pick the next version"
fi
info "Version v${TAG} is new"

# ─── Step 2: Push develop ────────────────────────────────────────────────────
step "Pushing develop"
run git push -q origin develop
info "Pushed to develop"

# ─── Step 3: Merge develop → master (locally) ────────────────────────────────
step "Merging develop → master (not pushed yet)"
trap return_to_develop EXIT
run git checkout -q master
run git merge -q --no-edit develop
info "master = develop, nothing pushed"

# ─── Step 4: Build & push the Docker image from master ───────────────────────
step "Building and pushing Docker image ${IMAGE_NAME}:${TAG} from master"
run ./build-and-push.sh "$TAG" --push $PLATFORM_FLAG
info "Image pushed: ${IMAGE_NAME}:${TAG}"

# ─── Step 5: Version commit on master ────────────────────────────────────────
step "Writing version ${TAG}, committing on master, pushing master"
if [ "$DRY_RUN" = true ]; then
    echo -e "  ${YELLOW}[dry-run]${NC} version.txt <- v${TAG}; deployment.yaml image <- ${IMAGE_NAME}:${TAG}"
else
    echo "v${TAG}" > "$VERSION_FILE"
    sed -i.bak "s|image: ${IMAGE_NAME}:.*|image: ${IMAGE_NAME}:${TAG}|" "$DEPLOYMENT_YAML"
    rm -f "${DEPLOYMENT_YAML}.bak"
fi
run git add "$VERSION_FILE" "$DEPLOYMENT_YAML"
run git commit -q -m "v${TAG}"
run git tag "v${TAG}"
run git push -q origin master "v${TAG}"
MASTER_PUSHED=true
info "Pushed master, tagged v${TAG}"

# ─── Step 6: Sync develop with the version ───────────────────────────────────
step "Merging master → develop, pushing develop"
run git checkout -q develop
trap - EXIT
run git merge -q --no-edit master
run git push -q origin develop
info "develop has v${TAG}"

# ─── Step 7: Deploy to server ────────────────────────────────────────────────
if [ "$SKIP_DEPLOY" = false ]; then
    step "Deploying to server (${SSH_HOST})"
    run ssh "$SSH_HOST" "bash -l -c '
set -euo pipefail
export KUBECONFIG=\${KUBECONFIG:-/home/kuber/.kube/config}
cd ${K8S_REPO_SERVER}
if [ \"\$(git branch --show-current)\" != master ]; then
    echo \"The server repo is not on master\" >&2
    exit 1
fi
echo \"Pulling master...\"
git pull -q

echo \"Applying k8s manifests...\"
kubectl apply -k kubernetes/overlays/prod

echo \"Checking rollout status...\"
kubectl rollout status deployment/kkal-tracker -n default --timeout=120s
'"
    info "Deployed to server"
fi

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Deployment complete!${NC}"
echo -e "${GREEN}${BOLD}  Kkal Tracker: v${TAG}${NC}"
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
