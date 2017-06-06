#!/usr/bin/env bash


DIR=$(dirname ${BASH_SOURCE[0]})
source $DIR/vars.sh

echo "removing root directory: $TMPROOT"
rm -rvf /tmp/$TMPROOT/*
