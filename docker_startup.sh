#!/usr/bin/env bash

wget https://raw.githubusercontent.com/mingkaic/wait-for-it/master/wait-for-it.sh
chmod u+x wait-for-it.sh

if [ -z "$POSTGRES_HOST" ]; then
    export POSTGRES_HOST="172.0.0.1";
fi
if [ -z "$POSTGRES_PORT" ]; then
    export POSTGRES_PORT="5432";
fi
./wait-for-it.sh $POSTGRES_HOST:$POSTGRES_PORT -- go run server/main.go -download=/data/imgexdb
