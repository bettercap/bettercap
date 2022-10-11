# build stage
FROM golang:1.16-alpine3.15 AS build-env

ENV SRC_DIR $GOPATH/src/github.com/bettercap/bettercap

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache bash iptables wireless-tools build-base libpcap-dev libusb-dev linux-headers libnetfilter_queue-dev git

WORKDIR $SRC_DIR
ADD . $SRC_DIR
RUN make

# get caplets
RUN mkdir -p /usr/local/share/bettercap
RUN git clone https://github.com/bettercap/caplets /usr/local/share/bettercap/caplets

# final stage
FROM alpine:3.15
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache bash iproute2 libpcap libusb-dev libnetfilter_queue wireless-tools
COPY --from=build-env /go/src/github.com/bettercap/bettercap/bettercap /app/
COPY --from=build-env /usr/local/share/bettercap/caplets /app/
WORKDIR /app

EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["/app/bettercap"]
