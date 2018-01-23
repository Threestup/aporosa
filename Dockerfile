FROM golang:alpine AS build

LABEL maintainer "jeremy@threestup.com"

COPY . /go/src/github.com/threestup/contactification
WORKDIR /go/src/github.com/threestup/contactification

RUN apk update && apk upgrade && apk add git
RUN go get -u ./... && go build

FROM scratch

RUN mkdir /outputs
COPY --from=build /go/src/github.com/threestup/contactification/contactification /contactification

ENTRYPOINT ["/contactification", "--port=1789", "--outDir=/outputs"]
