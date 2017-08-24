#! /bin/bash
set -eu

ID=$1
echo "172.58.0.$((100+$ID))"


