FROM alpine:3.4
ADD ./server /signaling/server
EXPOSE 8080
WORKDIR /signaling
ENV STUN=stun:stun.l.google.com:19302
ENTRYPOINT ./server
