#!/usr/bin/env bash

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )";

if [ -d $THIS_DIR/../download ]; then
    rm -rf $THIS_DIR/../download;
fi;

if [ -f $THIS_DIR/../imgex.db ]; then
    rm $THIS_DIR/../imgex.db;
fi;
