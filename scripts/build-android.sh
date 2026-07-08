#!/bin/bash

export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

go build -o server ./cmd/server