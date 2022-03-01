#!/bin/bash

mkdir -p Build

env GOOS=linux GOARCH=amd64 go build -o Build/hasher .
env GOOS=windows GOARCH=amd64 go build -o Build/hasher.exe .
