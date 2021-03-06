#Building
FROM golang:alpine AS build-env

RUN apk --no-cache add build-base git mercurial gcc

COPY . /src

ENV GO111MODULE=on

RUN cd /src && go build -o whaling

#Final image
FROM alpine:latest

RUN apk --no-cache add curl

ENV URL_LABEL=""

WORKDIR /app
COPY --from=build-env /src/whaling /app/
ENTRYPOINT ./whaling
