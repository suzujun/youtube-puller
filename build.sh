#!/bin/sh

PREFIX=youtube-poller
VERSION=v0.0.2

rm -f release/*
mkdir -p release

GOOS=darwin  GOARCH=amd64 go build -o release/${PREFIX}_darwin_amd64_${VERSION} main.go
GOOS=darwin  GOARCH=arm64 go build -o release/${PREFIX}_darwin_arm64_${VERSION} main.go
GOOS=linux   GOARCH=386   go build -o release/${PREFIX}_linux_386_${VERSION} main.go
GOOS=linux   GOARCH=amd64 go build -o release/${PREFIX}_linux_amd64_${VERSION} main.go
GOOS=linux   GOARCH=arm64 go build -o release/${PREFIX}_linux_arm64_${VERSION} main.go
GOOS=windows GOARCH=386   go build -o release/${PREFIX}_windows_386_${VERSION}.exe main.go
GOOS=windows GOARCH=amd64 go build -o release/${PREFIX}_windows_amd64_${VERSION}.exe main.go
GOOS=windows GOARCH=arm64 go build -o release/${PREFIX}_windows_arm64_${VERSION}.exe main.go
GOOS=windows GOARCH=arm   go build -o release/${PREFIX}_windows_arm_${VERSION}.exe main.go
