FROM golang:alpine AS build
RUN apk add -U git
RUN go get github.com/rs/cors
RUN go get golang.org/x/net/websocket
RUN go get -d github.com/nobonobo/p2pfw/signaling/server
RUN go install -tags netgo github.com/nobonobo/p2pfw/signaling/server

FROM alpine:3.7
COPY --from=build /go/bin/server /signaling/server
EXPOSE 8080
WORKDIR /signaling
ENV STUN=stun:stun.l.google.com:19302
ENTRYPOINT ./server
