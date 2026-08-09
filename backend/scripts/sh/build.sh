#!/bin/bash
set -e

# Usage:
# ./build.sh image-name version

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <image-name> <version>"
  exit 1
fi

IMAGE_NAME=$1
VERSION=$2
TAG="${IMAGE_NAME}:${VERSION}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "Building Docker image: $TAG"

docker build -f "$BACKEND_DIR/Dockerfile" \
  --build-arg APP_NAME="$IMAGE_NAME" \
  -t "$TAG" \
  "$BACKEND_DIR"

if [ $? -eq 0 ]; then
  echo "Image $TAG built successfully!"
else
  echo "Failed to build image $TAG"
  exit 1
fi
