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
Production is built only from master: develop is merged into master first,
then the image is built from the master checkout.

Performs the full cycle:
  1. Validate git state (on develop, clean tree, develop and master in sync with origin)
  2. Write the version to version.txt and the k8s deployment manifest, commit on develop, push develop
  3. Merge develop → master, push master
  4. Build & push the Docker image from master
  5. SSH to server, git pull (master), kubectl apply
  6. Back on develop

REQUIRED:
    --tag=TAG         Version tag (e.g., 5.5.0)

OPTIONS:
    --help            Show this help message
    --skip-deploy     Stop after the image is pushed (step 5 is skipped)
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

# Whatever happens after the checkout of master, finish on develop
return_to_develop() {
    if [ "$(git -C "$SCRIPT_DIR" branch --show-current)" = "master" ]; then
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

# ─── Step 2: Version commit on develop ───────────────────────────────────────
step "Writing version ${TAG}, committing on develop, pushing develop"

if [ "$DRY_RUN" = true ]; then
    echo -e "  ${YELLOW}[dry-run]${NC} version.txt <- v${TAG}; deployment.yaml image <- ${IMAGE_NAME}:${TAG}"
else
    echo "v${TAG}" > "$VERSION_FILE"
    sed -i.bak "s|image: ${IMAGE_NAME}:.*|image: ${IMAGE_NAME}:${TAG}|" "$DEPLOYMENT_YAML"
    rm -f "${DEPLOYMENT_YAML}.bak"
fi

run git add "$VERSION_FILE" "$DEPLOYMENT_YAML"
if [ "$DRY_RUN" = false ] && git diff --cached --quiet; then
    fail "version.txt and deployment.yaml already say ${TAG}; nothing to release"
fi
run git commit -q -m "v${TAG}"
run git push -q origin develop
info "Pushed to develop"

# ─── Step 3: Merge develop → master ──────────────────────────────────────────
step "Merging develop → master, pushing master"

trap return_to_develop EXIT
run git checkout -q master
run git merge -q --no-edit develop
run git push -q origin master
info "Merged and pushed to master"

# ─── Step 4: Build & push the Docker image from master ───────────────────────
step "Building and pushing Docker image ${IMAGE_NAME}:${TAG} from master"

run ./build-and-push.sh "$TAG" --push $PLATFORM_FLAG
# build-and-push.sh rewrites version.txt with the same content; nothing may be left behind
if [ "$DRY_RUN" = false ] && ! git diff --quiet; then
    fail "The build changed tracked files on master: $(git diff --name-only | tr '\n' ' ')"
fi
run git tag "v${TAG}"
run git push -q origin "v${TAG}"
info "Image pushed, master tagged v${TAG}"

# ─── Step 5: Deploy to server ────────────────────────────────────────────────
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

# ─── Step 6: Back on develop ─────────────────────────────────────────────────
step "Returning to develop"
run git checkout -q develop
trap - EXIT
info "On branch: develop"

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Deployment complete!${NC}"
echo -e "${GREEN}${BOLD}  Kkal Tracker: v${TAG}${NC}"
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
