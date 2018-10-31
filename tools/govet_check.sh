#!/bin/bash

rootdir=$(cd $(dirname $0)/..; pwd)
tmpfile=$(mktemp)

cd $rootdir
go vet ./... >> $tmpfile

size=$(du $tmpfile | awk '{print $1}')
if [[ $size != "0" ]]; then
	echo "please format the following file(s):"
	cat $tmpfile
	exit 1
fi
