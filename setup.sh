#!/usr/bin/env bash

apt-get install -y build-essential g++ flex bison gperf ruby perl \
  libsqlite3-dev libfontconfig1-dev libicu-dev libfreetype6 libssl-dev \
  libpng-dev libjpeg-dev python libx11-dev libxext-dev

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )";

# download source
git clone git://github.com/ariya/phantomjs.git
cd phantomjs
git checkout 2.1.1
git submodule init
git submodule update

# build
python build.py
