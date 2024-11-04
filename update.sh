#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tmp/main InstanceTypAZCheck.go
zip -j /tmp/main.zip /tmp/main

aws lambda update-function-code \
    --function-name InstanceTypAZCheck \
    --zip-file fileb:///tmp/main.zip
