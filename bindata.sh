go-bindata -o=modules/bindata/bindata.go -ignore="\\.DS_Store|config.codekit|less|db.sql|init.sql|setup_root_user.sql" -pkg=bindata template/... config/... static/...
