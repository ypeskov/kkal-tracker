#!/usr/bin/env bash
set -euo pipefail

IMAGE_NAME="ypeskov/kcal-tracker"
TAG="latest"
PLATFORM_ARG=""
PUSH=false

# Parse arguments
for arg in "$@"
do
    case $arg in
        --platform=*)
        PLATFORM_ARG="$arg"
        shift
        ;;
        push)
        PUSH=true
        shift
        ;;
        *)
        TAG="$arg"
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

echo "Done."