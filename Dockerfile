# iron/go:dev is the alpine image with the go tools added
FROM iron/go:dev

MAINTAINER Simone Margaritelli <https://evilsocket.net/>

LABEL Package="BetterCAP" \
      Version="Latest-Stable" \
      Description="BetterCAP the state of the art, modular, portable and easily extensible MITM framework in a Container" \
      Destro="Alpine Linux" \
      GitHub="https://github.com/evilsocket/bettercap-ng" \
      DockerHub="https://hub.docker.com/r/evilsocket/bettercap-ng/" \
      Maintainer="Simone Margaritelli"


ENV SRC_DIR=/gocode/src/github.com/evilsocket/bettercap-ng
WORKDIR $SRC_DIR
ADD . $SRC_DIR

RUN apk add --update ca-certificates; \
    apk add --no-cache --update \
        bash \
        build-base \
        libpcap-dev;\
    cd $SRC_DIR; \
    make deps; \
    make

EXPOSE 80 443 5300 8080 8081 8082 8083 8000
ENTRYPOINT ["./bettercap-ng"]
