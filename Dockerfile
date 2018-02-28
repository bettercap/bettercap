FROM iron/go:dev
MAINTAINER Simone Margaritelli <https://evilsocket.net/>

ENV GOPATH=/gocode
ENV SRC_DIR=/gocode/src/github.com/bettercap/bettercap
COPY . $SRC_DIR

WORKDIR $SRC_DIR

RUN apk add --update ca-certificates
RUN apk add --no-cache --update bash iptables build-base libpcap-dev

# As Alpine Linux uses a different folder, we need this
# ugly hack in order to compile gopacket statically
# https://github.com/bettercap/bettercap/issues/106
RUN mkdir -p /usr/lib/x86_64-linux-gnu/
RUN cp /usr/lib/libpcap.a /usr/lib/x86_64-linux-gnu/libpcap.a

RUN make deps
RUN make

EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["./bettercap"]
