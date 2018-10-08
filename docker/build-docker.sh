#!/bin/bash

set -e -x

go tool dist env > /tmp/goenv.tmp
. /tmp/goenv.tmp


VERSION=`cat package.json| grep version | awk -F':' '{print $2}'| tr -d "\", "`
COMMIT=`git rev-parse --short HEAD`

if [ ! -f dist/pseriescollector-${VERSION}-${COMMIT}.tar.gz ]
then
    echo "building binary...."
    npm run build:static
    go run build.go pkg-min-tar
else
    echo "skiping build..."
fi

export VERSION
export COMMIT

cp dist/pseriescollector-${VERSION}-${COMMIT}_${GOOS:-linux}_${GOARCH:-amd64}.tar.gz  docker/pseriescollector-last.tar.gz
cp conf/sample.pseriescollector.toml docker/pseriescollector.toml

cd docker

sudo docker build --label version="${VERSION}" --label commitid="${COMMIT}" -t tonimoreno/pseriescollector:${VERSION} -t tonimoreno/pseriescollector:latest .
rm pseriescollector-last.tar.gz
rm pseriescollector.toml

rm -f /tmp/goenv.tmp
