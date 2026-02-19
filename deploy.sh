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
SKIP_BUILD=false
SKIP_K8S=false
SKIP_DEPLOY=false
DRY_RUN=false

# ─── Help ─────────────────────────────────────────────────────────────────────
show_help() {
    cat << EOF
Usage: $(basename "$0") --tag=TAG [OPTIONS]

Automated deployment pipeline for Kkal Tracker.

Performs the full cycle:
  1. Validate git state (on develop, clean tree)
  2. Build & push Docker image to Docker Hub
  3. Update k8s deployment manifest with new image tag
  4. Commit version.txt + deployment.yaml, push develop
  5. Merge develop → master, push master
  6. SSH to server, git pull, kubectl apply

REQUIRED:
    --tag=TAG         Version tag (e.g., 5.5.0)

OPTIONS:
    --help            Show this help message
    --skip-build      Skip Docker build & push (steps 2-5)
    --skip-k8s        Skip k8s manifest update (step 3-4)
    --skip-deploy     Skip server deployment (step 6)
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
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-k8s)
            SKIP_K8S=true
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

run() {
    if [ "$DRY_RUN" = true ]; then
        echo -e "  ${YELLOW}[dry-run]${NC} $*"
    else
        "$@"
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
[ "$SKIP_BUILD" = true ]  && warn "Skipping build"
[ "$SKIP_K8S" = true ]    && warn "Skipping k8s manifest update"
[ "$SKIP_DEPLOY" = true ] && warn "Skipping server deploy"
[ "$DRY_RUN" = true ]     && warn "DRY RUN — no changes will be made"
echo ""
read -p "Proceed? [y/N] " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

# ─── Step 1: Validate git state ──────────────────────────────────────────────
step "Validating git state"

cd "$SCRIPT_DIR"

CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "develop" ]; then
    echo -e "${RED}Error: Must be on 'develop' branch (currently on '${CURRENT_BRANCH}')${NC}"
    exit 1
fi
info "On branch: develop"

if [ "$SKIP_BUILD" = false ]; then
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo -e "${RED}Error: Working tree has uncommitted changes. Commit or stash them first.${NC}"
        exit 1
    fi
    info "Working tree is clean"
fi

# ─── Step 2: Build & push Docker image ───────────────────────────────────────
if [ "$SKIP_BUILD" = false ]; then
    step "Building and pushing Docker image (${TAG})"
    cd "$SCRIPT_DIR"
    run ./build-and-push.sh "$TAG" --push $PLATFORM_FLAG
    info "Image pushed: ${IMAGE_NAME}:${TAG}"
fi

# ─── Step 3: Update k8s deployment manifest ──────────────────────────────────
if [ "$SKIP_K8S" = false ]; then
    step "Updating k8s deployment manifest"
    cd "$SCRIPT_DIR"
    if [ "$DRY_RUN" = true ]; then
        echo -e "  ${YELLOW}[dry-run]${NC} sed update deployment.yaml image to ${IMAGE_NAME}:${TAG}"
    else
        sed -i.bak "s|image: ${IMAGE_NAME}:.*|image: ${IMAGE_NAME}:${TAG}|" "$DEPLOYMENT_YAML"
        rm -f "${DEPLOYMENT_YAML}.bak"
    fi
    info "deployment.yaml updated to ${IMAGE_NAME}:${TAG}"
fi

# ─── Step 4: Commit & push develop ───────────────────────────────────────────
if [ "$SKIP_BUILD" = false ] || [ "$SKIP_K8S" = false ]; then
    step "Committing changes and pushing develop"
    cd "$SCRIPT_DIR"

    FILES_TO_ADD=()
    [ "$SKIP_BUILD" = false ] && FILES_TO_ADD+=("version.txt")
    [ "$SKIP_K8S" = false ]   && FILES_TO_ADD+=("kubernetes/base/deployment.yaml")

    run git add "${FILES_TO_ADD[@]}"
    if git diff --cached --quiet 2>/dev/null; then
        warn "No changes to commit"
    else
        run git commit -m "v${TAG}"
        run git push origin develop
        info "Pushed to develop"
    fi
fi

# ─── Step 5: Merge develop → master ──────────────────────────────────────────
if [ "$SKIP_BUILD" = false ]; then
    step "Merging develop → master"
    run git checkout master
    run git merge develop --no-edit
    run git push origin master
    run git checkout develop
    info "Merged and pushed to master, back on develop"
fi

# ─── Step 6: Deploy to server ────────────────────────────────────────────────
if [ "$SKIP_DEPLOY" = false ]; then
    step "Deploying to server (${SSH_HOST})"
    run ssh "$SSH_HOST" "bash -l -c '
set -euo pipefail
export KUBECONFIG=\${KUBECONFIG:-/home/kuber/.kube/config}
echo \"Pulling latest changes...\"
cd ${K8S_REPO_SERVER}
git pull

echo \"Applying k8s manifests...\"
kubectl apply -k kubernetes/overlays/prod

echo \"Checking rollout status...\"
kubectl rollout status deployment/kkal-tracker -n default --timeout=120s || true
'"
    info "Deployed to server"
fi

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  Deployment complete!${NC}"
echo -e "${GREEN}${BOLD}  Kkal Tracker: v${TAG}${NC}"
echo -e "${GREEN}${BOLD}══════════════════════════════════════${NC}"
