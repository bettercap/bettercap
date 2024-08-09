# build stage
FROM golang:1.22-alpine3.20 AS build-env

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache bash gcc g++ binutils-gold iptables wireless-tools build-base libpcap-dev libusb-dev linux-headers libnetfilter_queue-dev git

WORKDIR $GOPATH/src/github.com/bettercap/bettercap
ADD . $GOPATH/src/github.com/bettercap/bettercap
RUN make

# get caplets
RUN mkdir -p /usr/local/share/bettercap
RUN git clone https://github.com/bettercap/caplets /usr/local/share/bettercap/caplets

# final stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache bash iproute2 libpcap libusb-dev libnetfilter_queue wireless-tools
COPY --from=build-env /go/src/github.com/bettercap/bettercap/bettercap /app/
COPY --from=build-env /usr/local/share/bettercap/caplets /app/
WORKDIR /app

EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["/app/bettercap"]
