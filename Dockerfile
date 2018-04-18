# build stage
FROM golang:1.10-alpine AS build-env

ENV SRC_DIR $GOPATH/src/github.com/bettercap/bettercap

RUN apk add --update ca-certificates 
RUN apk add --no-cache --update bash iptables wireless-tools build-base libpcap-dev linux-headers libnetfilter_queue-dev git

WORKDIR $SRC_DIR
ADD . $SRC_DIR
RUN go get -u github.com/golang/dep/...
RUN make deps
RUN make

# final stage
FROM alpine
RUN apk add --no-cache --update bash iproute2 libpcap libnetfilter_queue
COPY --from=build-env /go/src/github.com/bettercap/bettercap/bettercap /app/
WORKDIR /app
EXPOSE 80 443 53 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["/app/bettercap"]
