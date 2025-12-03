#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME="ypeskov/kcal-tracker"
TAG=""
PLATFORM_ARG=""
PUSH=false

# Display help message
show_help() {
    cat << EOF
Usage: $(basename "$0") TAG [OPTIONS]

Build and optionally push Docker image for Kkal Tracker application.

TAG:
    Required. The tag name for the Docker image (e.g., 1.2.3)
    The image will be tagged as both TAG and 'latest'

OPTIONS:
    --help                  Show this help message and exit
    --push                  Push the built image to Docker registry
    --platform=PLATFORM     Set target platform for the build
                           (e.g., --platform=linux/amd64)

EXAMPLES:
    $(basename "$0") 1.2.3                            # Build with tag 'v1.2.3' and 'latest'
    $(basename "$0") 1.2.3 --push                     # Build and push with tag 'v1.2.3'
    $(basename "$0") 1.2.3 --platform=linux/arm64     # Build for ARM64 platform
    $(basename "$0") 1.0 --push --platform=linux/amd64  # Build, push for specific platform

The built image will be named: ${IMAGE_NAME}:TAG
After build, the tag will be written to version.txt file.
EOF
}

# Check if no arguments provided or help requested
if [ $# -eq 0 ] || [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    show_help
    exit 0
fi

# First argument is always the TAG
TAG="$1"
shift

# Parse remaining arguments (options)
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
        --push)
        PUSH=true
        shift
        ;;
        *)
        echo "Error: Unknown option '$1'"
        echo "Use --help to see available options."
        exit 1
        ;;
    esac
done

# Clean local dist directory to avoid cache issues
echo "Cleaning local web/dist directory..."
rm -rf web/dist

# Build the image without cache (always fresh build)
echo "Building ${IMAGE_NAME}:${TAG} (no cache)..."
docker build --no-cache $PLATFORM_ARG -t "${IMAGE_NAME}:${TAG}" -t "${IMAGE_NAME}:latest" .

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