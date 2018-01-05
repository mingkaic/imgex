#!/usr/bin/env bash

apt-get update
apt-get install -y \
	build-essential g++ \
	flex bison gperf ruby \
	perl  libsqlite3-dev \
	libfontconfig1-dev \
	libicu-dev libfreetype6 \
	libssl-dev libpng-dev \
	libjpeg-dev python \
	libx11-dev libxext-dev

# build grpc service
go get -u github.com/golang/protobuf/protoc-gen-go
