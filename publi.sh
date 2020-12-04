#!/usr/bin/env bash

set -e

git fetch
tag=$(git describe --abbrev=0)

major=0
minor=0
patch=0
if [[ ${tag:1} =~ ([0-9]+).([0-9]+).([0-9]+) ]]; then
  major="${BASH_REMATCH[1]}"
  minor="${BASH_REMATCH[2]}"
  patch="${BASH_REMATCH[3]}"
fi
patch=$(echo "${patch} + 1" | bc)
version=${1:-"${major}.${minor}.${patch}"}

echo "${version}"

go test -race -v ./...
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o "dist/$version/aws-ses-pop3-server-Linux"
CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o "dist/$version/aws-ses-pop3-server-Darwin"

docker build -t "markushinz/aws-ses-pop3-server:$version" \
  -t "markushinz/aws-ses-pop3-server:latest" .

git tag "v$version" -m "v$version"

docker push "markushinz/aws-ses-pop3-server:$version"
docker push "markushinz/aws-ses-pop3-server:latest"

git push origin "v$version"
