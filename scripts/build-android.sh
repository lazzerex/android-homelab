#!/usr/bin/env bash
set -e
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -o server ./cmd/server
echo Built server
