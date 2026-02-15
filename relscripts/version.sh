#!/bin/bash
set -e

VERSION=${1}

if [ -z "$VERSION" ]; then
    echo "Usage: ./relscripts/version.sh v0.1.0"
    exit 1
fi

# Strip 'v' prefix for the version string in main.go
VERSION_STRING=${VERSION#v}

echo "Updating version to ${VERSION}..."

# Update Version in main.go
sed -i "" "s/const Version = \".*\"/const Version = \"${VERSION_STRING}\"/" cmd/petri/main.go

# Check if there are changes
if git diff --quiet cmd/petri/main.go; then
    echo "No version changes needed (already at ${VERSION})"
else
    echo "Committing version change..."
    git add cmd/petri/main.go
    git commit -m "chore: Bump version to ${VERSION}"
fi

# Create tag
if git rev-parse ${VERSION} >/dev/null 2>&1; then
    echo "Tag ${VERSION} already exists"
else
    echo "Creating tag ${VERSION}..."
    git tag -a ${VERSION} -m "Release ${VERSION}"
fi

echo "✓ Version updated to ${VERSION}"
