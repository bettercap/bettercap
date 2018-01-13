FROM iron/go:dev
MAINTAINER Simone Margaritelli <https://evilsocket.net/>

ENV SRC_DIR=/gocode/src/github.com/evilsocket/bettercap-ng
WORKDIR $SRC_DIR
ADD . $SRC_DIR

RUN apk add --update ca-certificates
RUN apk add --no-cache --update bash build-base libpcap-dev

RUN cd $SRC_DIR
RUN make deps
RUN make

EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["./bettercap"]
