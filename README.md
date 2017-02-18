AiicyDS(with go)
===========
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

./aiicyds
```
