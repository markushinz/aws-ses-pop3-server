#!/usr/bin/env bash
version=${1:-"1.0.0"}

set -e

go test -race -v ./...
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o "dist/$version/aws-ses-pop3-server-Linux"
CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o "dist/$version/aws-ses-pop3-server-Darwin"

git fetch
git tag "v$version" -m "v$version"

docker build -t "docker.pkg.github.com/markushinz/aws-ses-pop3-server/aws-ses-pop3-server:$version" \
  -t "docker.pkg.github.com/markushinz/aws-ses-pop3-server/aws-ses-pop3-server:latest" .
docker push "docker.pkg.github.com/markushinz/aws-ses-pop3-server/aws-ses-pop3-server:$version"
docker push "docker.pkg.github.com/markushinz/aws-ses-pop3-server/aws-ses-pop3-server:latest"

git push origin "v$version"
