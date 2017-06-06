#!/usr/bin/env bash
set -e

# Get the version from the environment, or try to figure it out.
if [ -z $VERSION ]; then
	  VERSION=$(awk -F\" '/Version =/ { print $2; exit }' < version/version.go)
fi
if [ -z "$VERSION" ]; then
    echo "Please specify a version."
    exit 1
fi

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd "$DIR"

# Get the version from the environment, or try to figure it out.
if [ -z $VERSION ]; then
	VERSION=$(awk -F\" '/Version =/ { print $2; exit }' < version/version.go)
fi
if [ -z "$VERSION" ]; then
    echo "Please specify a version."
    exit 1
fi

DIST_DIR="build/dist"

# copy to s3
aws s3 cp --recursive ${DIST_DIR} s3://ethermint/${VERSION} --acl public-read --exclude "*" --include "*.zip"
aws s3 cp ${DIST_DIR}/ethermint_${VERSION}_SHA256SUMS s3://ethermint/${VERSION}/${VERSION}_SHA256SUMS --acl public-read

exit 0
