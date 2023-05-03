FROM golang:1.20.3-alpine AS builder
LABEL maintainer="mengbin1992@outlook.com"

COPY . /go/src/wechat
ENV GOPROXY="https://goproxy.cn,direct"
ENV CGO_ENABLED=0

# build wechat client
WORKDIR /go/src/wechat
RUN cd /go/src/wechat && go build -ldflags "-s -w" -o wechat

# build watcher
RUN cd /go/src/wechat/cmd/watcher && go build -ldflags "-s -w" -o watcher

FROM alpine:3.17
LABEL maintainer="mengbin1992@outlook.com"

RUN mkdir /app

COPY --from=builder /go/src/wechat/wechat /app
COPY --from=builder /go/src/wechat/cmd/watcher/watcher /app
COPY conf /app/conf

WORKDIR /app

ENTRYPOINT ["/app/watcher"]