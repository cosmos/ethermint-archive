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
echo "==> Building version $VERSION..."

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that.
cd "$DIR"

# Generate the tag.
# if [ -z "$NOTAG" ]; then
#     echo "==> Tagging..."
#     git commit --allow-empty -a -m "Release v$VERSION"
#     git tag -a -m "Version $VERSION" "v${VERSION}" master
# fi

# Do a hermetic build inside a Docker container.
# docker build -t ethermint/ethermint-builder scripts/ethermint-builder/
# docker run --rm -v "$(pwd)":/go/src/github.com/tendermint/ethermint ethermint/ethermint-builder ./scripts/dist_build.sh
# Get the git commit
GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DESCRIBE="$(git describe --tags --always)"
GIT_IMPORT="github.com/tendermint/ethermint/version"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-"solaris darwin freebsd linux windows"}

IGNORE=("darwin/arm solaris/amd64 freebsd/amd64")
NON_STATIC=("darwin/386 darwin/amd64")

TARGETS=""
NON_STATIC_TARGETS=""

for os in $XC_OS; do
    for arch in $XC_ARCH; do
        target="$os/$arch"

        case ${IGNORE[@]} in *$target*) continue;; esac
        case ${NON_STATIC[@]} in *$target*)
          NON_STATIC_TARGETS="$target,$NON_STATIC_TARGETS"
          continue;;
        esac

        TARGETS="$target,$TARGETS"
    done
done

# Delete the old dir
echo "==> Removing old directory..."
rm -rf build/pkg
mkdir -p build/pkg

# Make sure build tools are available.
make tools

# Get VENDORED dependencies
make get_vendor_deps

# Build!
if [ ! -z "$TARGETS" ]; then
  echo "==> Building Static Binaries..."
  TARGETS=${TARGETS::${#TARGETS}-1}
  xgo -go="latest" \
    -targets="${TARGETS}" \
    -ldflags "-extldflags '-static' -X ${GIT_IMPORT}.GitCommit=${GIT_COMMIT} -X ${GIT_IMPORT}.GitDescribe=${GIT_DESCRIBE}" \
    -dest "build/pkg" \
    -tags="${BUILD_TAGS}" \
    ${DIR}/cmd/ethermint
fi

if [ ! -z "$NON_STATIC_TARGETS" ]; then
  echo "==> Building Non-Static Binaries..."
  NON_STATIC_TARGETS=${NON_STATIC_TARGETS::${#NON_STATIC_TARGETS}-1}
  xgo -go="latest" \
    -targets="${NON_STATIC_TARGETS}" \
    -ldflags "-X ${GIT_IMPORT}.GitCommit=${GIT_COMMIT} -X ${GIT_IMPORT}.GitDescribe=${GIT_DESCRIBE}" \
    -dest "build/pkg" \
    -tags="${BUILD_TAGS}" \
    ${DIR}/cmd/ethermint
fi

echo "==> Packaging..."
for FILE in $(ls ./build/pkg); do
    pushd ./build/pkg
    zip "${FILE}.zip" $FILE
    popd
done

# Add "ethermint" and $VERSION prefix to package name.
rm -rf ./build/dist
mkdir -p ./build/dist
for FILENAME in $(find ./build/pkg -mindepth 1 -maxdepth 1 -type f); do
    FILENAME=$(basename "$FILENAME")
	  cp "./build/pkg/${FILENAME}" "./build/dist/ethermint_${VERSION}_${FILENAME}"
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
