FROM golang:alpine

RUN apk add --no-cache --virtual .build-deps git build-base

COPY go.mod /akyuu/
COPY go.sum /akyuu/
COPY main.go /akyuu/
COPY src /akyuu/src
COPY config.toml /akyuu/

VOLUME [ "/akyuu/akyuu" ]

RUN cd /akyuu && \
    go build main.go && \
    apk del .build-deps

WORKDIR /akyuu
ENTRYPOINT [ "./main" ]