AiicyDS(with go)
===========

[![Build Status](https://travis-ci.org/Aiicy/AiicyDS.svg?branch=master)](https://travis-ci.org/Aiicy/AiicyDS)
[![Build status](https://ci.appveyor.com/api/projects/status/9s99gcvieye3v3eo7cyw/branch/master?svg=true)](https://ci.appveyor.com/project/countstarlight/aiicyds/branch/master)
[![Sourcegraph](https://sourcegraph.com/github.com/Aiicy/AiicyDS/-/badge.svg)](https://sourcegraph.com/github.com/Aiicy/AiicyDS?badge)
##Environment
* go version >=1.6
* system Linux or windows

## Install golang on Linux amd64
```
wget -c https://storage.googleapis.com/golang/go1.8rc3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.8rc3.linux-amd64.tar.gz
nano ~/.bashrc
```
Write and save the following
```
export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/.go
```
## Install AiicyDS
```bash
go get https://github.com/Aiicy/AiicyDS

cd $GOPATH/src/github.com/Aiicy/AiicyDS

go build

./AiicyDS -h
```

## Run AiicyDS
```
./AiicyDS web
```
