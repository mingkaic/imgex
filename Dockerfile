FROM golang:latest

ENV APP_DIR $GOPATH/src/github.com/mingkaic/imgex
RUN mkdir -p $APP_DIR
WORKDIR $APP_DIR

COPY ./scripts /tmp
RUN bash /tmp/docker_setup.sh
RUN bash /tmp/phantom_setup.sh
RUN bash /tmp/protobuf_setup.sh

COPY . $APP_DIR
RUN go get ./...
RUN make

CMD [ "bash", "docker_startup.sh" ]
