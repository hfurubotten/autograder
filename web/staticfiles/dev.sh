#!/bin/sh

cd $GOPATH/src/github.com/hfurubotten/autograder/web/
$GOPATH/bin/go-bindata -o=staticfiles/staticfiles.go -pkg=staticfiles -debug css/ fonts/ html/... img/... js/
