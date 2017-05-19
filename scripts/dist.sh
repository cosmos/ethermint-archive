#!/usr/bin/env bash
set -e

XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}

IGNORE=("darwin/arm solaris/amd64 freebsd/amd64")

# Get the version from the environment, or try to figure it out.
if [ -z $VERSION ]; then
  MAJOR=$(awk '/Major = / { print $3; exit }' < version/version.go)
  MINOR=$(awk '/Minor = / { print $3; exit }' < version/version.go)
  PATCH=$(awk '/Patch = / { print $3; exit }' < version/version.go)
  #META=$(awk -F\" '/Meta  =/ { print $2; exit }' < version/version.go)

  VERSION="$MAJOR.$MINOR.$PATCH" # This is current tag format
fi

if [ -z "$VERSION" ]; then
    echo "Please specify a version."
    exit 1
fi
echo "==> Building version $VERSION..."

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd "$DIR"

echo "==> Building..."
GIT_COMMIT=`git rev-parse HEAD`
GIT_IMPORT="github.com/tendermint/ethermint/version"
TARGETS=""

for os in $XC_OS; do
  for arch in $XC_ARCH; do
    target="$os/$arch"

    case ${IGNORE[@]} in *$target*) continue;; esac
    TARGETS="$os/$arch,$TARGETS"
  done
done

# Remove last comma
TARGETS=${TARGETS::${#TARGETS}-1}

# static build doesn't work with Darwin
xgo --go="latest" \
  --targets="${TARGETS}" \
  --dest build/pkg/ \
  --ldflags "-X ${GIT_IMPORT}.GitCommit=${GIT_COMMIT}" \
  "${DIR}/cmd/ethermint"

echo "==> Packaging..."
for FILE in $(ls ./build/pkg); do
  pushd ./build/pkg
  zip "${FILE}.zip" $FILE
  popd
done

echo "==> Moving to dist"
# Add "ethermint" and $VERSION prefix to package name.
mkdir -p ./build/dist
for FILE in $(ls ./build/pkg/*.zip); do
  FILENAME=$(basename $FILE)
	mv $FILE "./build/dist/${FILENAME/ethermint/ethermint-${VERSION}}"
done

# Make the checksums.
pushd ./build/dist
shasum -a256 ./* > "./ethermint_${VERSION}_SHA256SUMS"
popd

# Done
echo
echo "==> Results:"
ls -hl ./build/dist

exit 0

