#!/usr/bin/env bash

pushd /tmp
export PHANTOM_JS="phantomjs-1.9.8-linux-x86_64"
wget https://bitbucket.org/ariya/phantomjs/downloads/$PHANTOM_JS.tar.bz2
tar xvjf $PHANTOM_JS.tar.bz2
mv $PHANTOM_JS/bin $GOPATH/src/github.com/mingkaic/imgex
popd
