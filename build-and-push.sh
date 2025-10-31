#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME="ypeskov/kcal-tracker"
TAG="latest"
PLATFORM_ARG=""
PUSH=false

# Display help message
show_help() {
    cat << EOF
Usage: $(basename "$0") [OPTIONS] [TAG]

Build and optionally push Docker image for Kkal Tracker application.

OPTIONS:
    --help                  Show this help message and exit
    --platform=PLATFORM     Set target platform for the build
                           (e.g., --platform=linux/amd64)
    push                    Push the built image to Docker registry

TAG:
    Optional tag name (default: latest)
    If provided, the image will be tagged as both TAG and 'latest'

EXAMPLES:
    $(basename "$0")                    # Build with tag 'latest'
    $(basename "$0") 1.2.3              # Build with tag 'v1.2.3' and 'latest'
    $(basename "$0") push                # Build and push with tag 'latest'
    $(basename "$0") 1.2.3 push         # Build with tag 'v1.2.3' and push
    $(basename "$0") --platform=linux/arm64 1.0  # Build for ARM64 platform

The built image will be named: ${IMAGE_NAME}:TAG
After build, the tag will be written to version.txt file.
EOF
}

# Parse arguments
while [ $# -gt 0 ]; do
    case $1 in
        --help|-h)
        show_help
        exit 0
        ;;
        --platform=*)
        PLATFORM_ARG="$1"
        shift
        ;;
        push)
        PUSH=true
        shift
        ;;
        *)
        TAG="$1"
        shift
        ;;
    esac
done

# Build the image
echo "Building ${IMAGE_NAME}:${TAG}..."
docker build $PLATFORM_ARG -t "${IMAGE_NAME}:${TAG}" -t "${IMAGE_NAME}:latest" .

# Push the image if requested
if [ "$PUSH" == true ]; then
    echo "Pushing ${IMAGE_NAME}:${TAG} and ${IMAGE_NAME}:latest..."
    docker push "${IMAGE_NAME}:${TAG}"
    echo "Successfully pushed."
fi

# Write the tag to version.txt
echo "Writing tag v${TAG} to version.txt..."
echo "v${TAG}" > version.txt

echo "Done."