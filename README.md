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
* get gvt
```
go get github.com/FiloSottile/gvt
```
If you can not access Google
```
go get github.com/polaris1119/gvt
```
in ~/.bashrc
```
export PATH=$PATH:~/.go/bin
```
```
cd AiicyCMS/
```
* get pkg

on Linux
```
./getpkg.sh
```
on windows
```
./getpkg.bat
```
* build AiicyCMS
on Linux
```
./install.sh
```
on windows
```
./install.bat
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
