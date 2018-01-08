#!/usr/bin/env bash

pushd /tmp
export PHANTOM_JS="phantomjs-2.1.1-linux-x86_64"
wget https://bitbucket.org/ariya/phantomjs/downloads/$PHANTOM_JS.tar.bz2
tar xvjf $PHANTOM_JS.tar.bz2
mv $PHANTOM_JS/bin $GOPATH/src/github.com/mingkaic/imgex
popd
