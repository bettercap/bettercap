# iron/go:dev is the alpine image with the go tools added
FROM iron/go:dev

WORKDIR /bettercap-ng

ENV SRC_DIR=/go/src/github.com/evilocket/bettercap-ng

ADD . $SRC_DIR

RUN apk add --update ca-certificates && \
    apk add --no-cache --update \
        bash \
        build-base \
        libpcap-dev;\
    cd $SRC_DIR; \
    make deps; \
    make; \
    cp bettercap-ng /bettercap-ng/

EXPOSE 80 443 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["./bettercap-ng"]
CMD ["-h"]

