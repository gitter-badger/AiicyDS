#!/usr/bin/env bash

set -e

RELEASE_ROOT="release"
NOW=$(date -u '+%Y%m%d%I%M%S')

OS=$(uname -s)
OS_TYPE=$(uname -m)

ZIP_FILE_PREFIX=AiicyDS-${OS}-${OS_TYPE}

if [ ! -f release_tool.sh ]; then
    echo 'release_tool.sh must be run within its container folder' 1>&2
    exit 1
fi

# Create release dir, if it not exists
if [ ! -d $RELEASE_ROOT ]; then
	mkdir $RELEASE_ROOT
	mkdir  -p $RELEASE_ROOT/license
fi

# test bin/AiicyDS exists or not
if [ ! -f bin/AiicyDS ]; then
	echo 'you need to run install.sh first, to build the application' 1>&2
	exit 1
fi

# copy the file to release folder
function CopyFiles() {
	cp -rp ./bin/ ${RELEASE_ROOT}/
	cp -rp ./config/ ${RELEASE_ROOT}/
	cp -rp ./static/ ${RELEASE_ROOT}/
	cp -rp ./template/ ${RELEASE_ROOT}/
	cp start.sh ${RELEASE_ROOT}/
	cp stop.sh ${RELEASE_ROOT}/
	cp AiicyDSInitDB.py ${RELEASE_ROOT}/
	cp LICENSE ${RELEASE_ROOT}/license/
}

# pack all the file in a zip file
function Pack() {
	cd ${RELEASE_ROOT} && zip -r ../${ZIP_FILE_PREFIX}.zip .
}

# print the zip file name
function PrintZipName() {
	echo ${ZIP_FILE_PREFIX}.zip
}

# write the md5sum info to a file
function WriteMD5SUM() {
	cd ../
	echo "generate the md5sum for " ${ZIP_FILE_PREFIX}.zip
	md5sum ${ZIP_FILE_PREFIX}.zip >> ${ZIP_FILE_PREFIX}.md5sum
}

CopyFiles
Pack
WriteMD5SUM
