#!/bin/bash

go version

export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin

echo "go build ..."
go build *.go