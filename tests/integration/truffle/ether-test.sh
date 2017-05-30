#!/usr/bin/env bash

FILE=$1
FILEPATH=$2
TMPFILES=/tmp/truffle-tests
SCRIPTDIR=$(dirname ${BASH_SOURCE[0]})

pushd $SCRIPTDIR > /dev/null
SCRIPTDIR=`pwd -P`
popd > /dev/null

if [ ! -f $FILE ]; then
  echo "File $FILE does not exist"
  exit 1
fi

REPOS=($(cat $FILE))

mkdir -pv $TMPFILES

cd $TMPFILES
echo -e '\c' > $FILEPATH
for ((i=0; i < $((${#REPOS[@]})); i=$i+2)); do
  repo=${REPOS[$i]}
  folder=${REPOS[$((i+1))]}
  name=$(basename $repo)

  (
    if [ ! -d $name ]; then
      git clone $repo || {
        echo "Failed to clone repo $repo"
        echo "Skipping.."
        exit 2
      }
    fi

    cd $name
    cd $folder

    echo "NPM Install"
    npm i
    
    echo "Create migrations folder"
    mkdir -p migrations

    echo "Run Truffle Tests"
    TESTS_FILE=$FILEPATH truffle test --mocha.reporter=$SCRIPTDIR/test-reporter.js
    echo ',' >> $FILEPATH
  )
done
