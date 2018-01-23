FROM golang:alpine AS build

LABEL maintainer "jeremy@threestup.com"

COPY . /go/src/github.com/Threestup/contactifications
WORKDIR /go/src/github.com/Threestup/contactifications

RUN apk update && apk upgrade && apk add curl openssh git
RUN curl https://glide.sh/get | sh && glide install && go build

FROM scratch

RUN mkdir /outputs
COPY --from=build /go/src/github.com/Threestup/contactifications/contactifications /contactifications

ENTRYPOINT ["/contactifications", "--port=1789", "--outDir=/outputs"]
