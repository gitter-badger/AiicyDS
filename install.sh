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

if [  ! -d "vendor/src" ]; then
	mkdir vendor/src
fi


if [ -d vendor/github.com ]; then
	mv vendor/github.com vendor/src
fi

if [ -d vendor/golang.org ]; then
	mv vendor/golang.org vendor/src
fi

BUILD="`git symbolic-ref HEAD | cut -b 12-`-`git rev-parse HEAD`"

gom build -ldflags "-X global.Build="$BUILD -o bin/AiicyDS

#go install server/indexer
#go install server/crawler

export GOPATH="$OLDGOPATH"
export PATH="$OLDPATH"

echo 'finished'

