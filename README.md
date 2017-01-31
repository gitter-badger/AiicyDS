AiicyCMS(with go)
===========
##Environment
* go version >=1.6
* system Linux or windows

## How to build

* get source

  [zip file](https://github.com/Aiicy/AiicyCMS/archive/go.zip)
  
  or
  ```
  git clone https://github.com/Aiicy/AiicyCMS.git -b go
  ```
* get gvt
```
go get github.com/polaris1119/gvt

PATH=$PATH:~/polaris1119/gvt
```
* get pkg
on unix
```
./getpkg.sh
```
on windows
```
./getpkg.bat
```
* build AiicyCMS
on unix
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
##test
assess http://127.0.0.1:8088 with brower
