#!/bin/bash

### This script performs compilation and docker build.  When building for musl based distros (like Alpine) on a glibc
### based distro (like RedHat, Debian, Centos, Ubuntu, etc...) we need to set CGO_ENABLED=0 in order to prevent dynamic
### linking errors

if [ "$1" == "" ]; then
  echo Usage: ./build-docker-image.sh \[image-tag]
  echo \[image-tag] is the docker tag to build, e.g. myuser/purge-dockerhub:DEV
  exit 1
fi

rm -f purge-dockerhub
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
docker build . -t "$1"