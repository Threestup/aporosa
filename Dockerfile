FROM golang:alpine AS build

LABEL maintainer "contact@threestup.com"

COPY . /go/src/github.com/Threestup/aporosa
WORKDIR /go/src/github.com/Threestup/aporosa
RUN apk update && apk upgrade && apk add curl openssh git
RUN go get -u ./... && go build -o exe

FROM alpine:3.7

RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=build /go/src/github.com/Threestup/aporosa/exe /exe

ENTRYPOINT ["/exe", "--port=1789"]
