FROM iron/go:dev
MAINTAINER Simone Margaritelli <https://evilsocket.net/>

ENV GOPATH=/gocode
ENV SRC_DIR=/gocode/src/github.com/evilsocket/bettercap-ng
COPY . $SRC_DIR

WORKDIR $SRC_DIR

RUN apk add --update ca-certificates
RUN apk add --no-cache --update bash iptables build-base libpcap-dev
RUN make deps
RUN make

EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["./bettercap-ng"]
