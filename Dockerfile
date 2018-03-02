# build stage
FROM golang:1.10-alpine AS build-env
ENV GOPATH=/gocode
ENV SRC_DIR=/gocode/src/github.com/bettercap/bettercap

# As Alpine Linux uses a different folder, we need this
# ugly hack in order to compile gopacket statically
# https://github.com/bettercap/bettercap/issues/106
RUN apk add --update ca-certificates && \
apk add --no-cache --update bash iptables wireless-tools build-base libpcap-dev git python && \
mkdir -p /usr/lib/x86_64-linux-gnu/ && \
cp /usr/lib/libpcap.a /usr/lib/x86_64-linux-gnu/libpcap.a

WORKDIR $SRC_DIR
ADD . $SRC_DIR
RUN make deps
RUN make

# final stage
FROM alpine
RUN apk add --no-cache --update bash libpcap-dev 
COPY --from=build-env /gocode/src/github.com/bettercap/bettercap/bettercap /app/
WORKDIR /app
EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["/app/bettercap"]
