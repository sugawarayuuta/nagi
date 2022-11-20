#!/bin/sh

SRC=.

GOOS=darwin  GOARCH=amd64 go build -o ./dist/nagi-mac-64 ${SRC}
GOOS=darwin  GOARCH=arm64 go build -o ./dist/nagi-mac-arm64 ${SRC}
GOOS=linux   GOARCH=386   go build -o ./dist/nagi-linux-32 ${SRC}
GOOS=linux   GOARCH=amd64 go build -o ./dist/nagi-linux-64 ${SRC}
GOOS=linux   GOARCH=arm   go build -o ./dist/nagi-linux-arm32 ${SRC}
GOOS=linux   GOARCH=arm64 go build -o ./dist/nagi-linux-arm64 ${SRC}
GOOS=windows GOARCH=386   go build -o ./dist/nagi-win-32.exe ${SRC}
GOOS=windows GOARCH=amd64 go build -o ./dist/nagi-win-64.exe ${SRC}
GOOS=windows GOARCH=arm   go build -o ./dist/nagi-win-arm32.exe ${SRC}
GOOS=windows GOARCH=arm64 go build -o ./dist/nagi-win-arm64.exe ${SRC}