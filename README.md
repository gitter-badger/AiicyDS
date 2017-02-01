AiicyCMS(with go)
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
## How to build

* get source

  [zip file](https://github.com/Aiicy/AiicyCMS/archive/go.zip)
  
  or
```
git clone https://github.com/Aiicy/AiicyCMS.git -b go
```
* get gom
```
go get github.com/mattn/gom
```
* install the dep package
```
gom install
```
* build the exec
```
gom build
```

##RUN AiicyCMS
on Linux
```
./start.sh
```
on windows
```
./start.bat
```
##Test
assess http://127.0.0.1:8088 with brower

## go-bindata
```shell
	$go get -u github.com/jteeuwen/go-bindata/go-bindata
	$./bindata.sh
```
