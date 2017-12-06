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

# xgo requires GOPATH
if [ ! -z "$GOPATH" ]; then
	GOPATH=$(go env GOPATH)
fi

# Get the git commit
GIT_COMMIT="$(git rev-parse --short HEAD)"
GIT_IMPORT="github.com/tendermint/ethermint/version"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64 arm-5 arm-6 arm-7 mips mipsle mips64 mips64le"}
XC_OS=${XC_OS:-"darwin linux windows"}
IGNORE=("darwin/arm darwin/386")

TARGETS=""
for os in $XC_OS; do
    for arch in $XC_ARCH; do
        target="$os/$arch"

        case ${IGNORE[@]} in *$target*) continue;; esac
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
# ldflags: -s Omit the symbol table and debug information.
#	         -w Omit the DWARF symbol table.
echo "==> Building..."
TARGETS=${TARGETS::${#TARGETS}-1}
xgo -go="1.8.3" \
	-targets="${TARGETS}" \
	-ldflags "-s -w -X ${GIT_IMPORT}.GitCommit=${GIT_COMMIT}" \
	-dest "build/pkg" \
	-tags="${BUILD_TAGS}" \
	"${DIR}/cmd/ethermint"

echo "==> Renaming exe files..."
for FILE in $(ls ./build/pkg); do
    f=${FILE#*-}
    if [[ $f == *"exe" ]]
    then
        f=${f%.*}
    fi
    echo "$f"
    mkdir -p "./build/pkg/$f"

    name="ethermint"
    if [[ $FILE == *"exe" ]]
    then
        name="ethermint.exe"
    fi
    echo $name

    pushd ./build/pkg
    mv "$FILE" "$f/$name"
    popd
done

# Zip all the files.
echo "==> Packaging..."
for PLATFORM in $(find ./build/pkg -mindepth 1 -maxdepth 1 -type d); do
		OSARCH=$(basename "${PLATFORM}")
		echo "--> ${OSARCH}"

		pushd "$PLATFORM" >/dev/null 2>&1
		zip "../${OSARCH}.zip" ./*
		popd >/dev/null 2>&1
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
