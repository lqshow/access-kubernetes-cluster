#!/usr/bin/env bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  bin="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$bin/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
bin="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
cd $bin

readonly PROJECT_ROOT=$(pwd)/..
readonly PROJECT_MODULE=github.com/lqshow/access-kubernetes-cluster
readonly SOURCE_DATE_EPOCH=$(git show -s --format=format:%ct HEAD)

cd ${PROJECT_ROOT}
. ./script/version.sh
GOBIN=$(mkdir -p ./bin && cd ./bin && pwd) go install -ldflags "$(version::ldflags)" ./...
