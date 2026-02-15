#!/bin/bash
set -e

VERSION=${1:-$(git describe --tags --always --dirty)}
PROJECT="petri"
DIST_DIR="dist"

echo "Building ${PROJECT} ${VERSION}..."

# Clean dist directory
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}

    OUTPUT_NAME="${PROJECT}-${VERSION}-${GOOS}-${GOARCH}"

    if [ $GOOS = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "Building ${GOOS}/${GOARCH}..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X main.version=${VERSION}" \
        -o "${DIST_DIR}/${OUTPUT_NAME}" \
        ./cmd/petri

    echo "  ✓ ${OUTPUT_NAME}"
done

echo ""
echo "Build complete! Binaries in ${DIST_DIR}/"
ls -lh ${DIST_DIR}/
