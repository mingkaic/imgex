FROM golang:latest

ENV APP_DIR $GOPATH/src/github.com/mingkaic/imgex

RUN mkdir -p $APP_DIR
COPY . $APP_DIR

WORKDIR $APP_DIR

RUN apt-get update
RUN apt-get install -y \
    build-essential g++ \
    flex bison gperf ruby \
    perl  libsqlite3-dev \
    libfontconfig1-dev \
    libicu-dev libfreetype6 \
    libssl-dev libpng-dev \
    libjpeg-dev python \
    libx11-dev libxext-dev
RUN bash setup.sh
RUN go get ./...

CMD [ "go", "run", "server/main.go" ]
