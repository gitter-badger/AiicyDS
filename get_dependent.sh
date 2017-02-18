#!/usr/bin/env bash

go get github.com/go-macaron/binding

go get github.com/go-macaron/cache

go get github.com/go-macaron/csrf

go get gopkg.in/clog.v1

go get github.com/jaytaylor/html2text

go get gopkg.in/gomail.v2

go get gopkg.in/editorconf

go get gopkg.in/editorconfig/editorconfig-core-go.v1

go get github.com/golang/net/html

mv $GOPATH/src/github.com/golang/net $GOPATH/src/golang/x/
