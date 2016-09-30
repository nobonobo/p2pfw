FROM alpine:3.4

ADD turnserver-4.5.0.4.tar.gz /root
RUN apk add -U --no-cache --virtual=build-essentials \
 g++ make autoconf libevent-dev \
 linux-headers procps openssl-dev && \
 cd /root/turnserver-4.5.0.4 && \
 ./configure && make && make install && \
 apk del --purge build-essentials && \
 cd /root && rm -rf /root/turnserver-4.5.0.4
RUN apk add --no-cache libevent openssl curl
ADD ./turnserver /server
EXPOSE 8080 3478 3478/udp
ENTRYPOINT /server
