#!/usr/bin/env bash

set -e

if [ ! -f install.sh ]; then
	echo 'install must be run within its container folder' 1>&2
	exit 1
fi

CURDIR=`pwd`
OLDGOPATH="$GOPATH"
export GOPATH="$CURDIR"

if [ ! -d log ]; then
	mkdir log
fi

BUILD="`git symbolic-ref HEAD | cut -b 12-`-`git rev-parse HEAD`"

go build -o bin/AiicyDS  -ldflags "-X modules/global.Build="$BUILD server/AiicyDS
go build -o bin/crawler  server/crawler
go build -o bin/indexer  server/indexer

export GOPATH="$OLDGOPATH"
export PATH="$OLDPATH"

echo 'finished'

