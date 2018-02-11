#!/bin/bash
BUILD_FOLDER=build
VERSION=$(cat core/banner.go | grep Version | cut -d '"' -f 2)

bin_dep() {
    BIN=$1
    which $BIN > /dev/null || { echo "@ Dependency $BIN not found !"; exit 1; }
}

vm_dep() {
    HOST=$1
    ping -c 1 $HOST > /dev/null || { echo "@ Virtual machine host $HOST not visible !"; exit 1; }
}

build_linux_amd64() {
    OUTPUT=$1
    echo "@ Building $OUTPUT ..."
    go build -o "$OUTPUT" ..
}

download_pcap() {
    bin_dep 'wget'
    bin_dep 'tar'

    cd /tmp
    rm -rf libpcap-1.8.1
    if [ ! -f /tmp/libpcap-1.8.1.tar.gz ]; then
        echo "@ Downloading  http://www.tcpdump.org/release/libpcap-1.8.1.tar.gz ..."
        wget -q http://www.tcpdump.org/release/libpcap-1.8.1.tar.gz -O /tmp/libpcap-1.8.1.tar.gz
    fi
    tar xf libpcap-1.8.1.tar.gz
}

xcompile_pcap() {
    ARCH=$1

    bin_dep 'make'
    bin_dep 'yacc'
    bin_dep 'flex'
    bin_dep "$ARCH-linux-gnueabi-gcc"

    echo "@ Cross compiling libpcap for $ARCH ..."
    cd /tmp/libpcap-1.8.1
    export CC=$ARCH-linux-gnueabi-gcc
    ./configure --host=$ARCH-linux-gnueabi --with-pcap=linux > /dev/null
    make CFLAGS='-w' -j4 > /dev/null
}

build_linux_arm7() {

    OUTPUT=$1
    OLD=$(pwd)
    
    download_pcap
    xcompile_pcap 'arm'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_macos_amd64() {
    vm_dep 'osxvm'

    DIR=/Users/evilsocket/gocode/src/github.com/evilsocket/bettercap-ng
    OUTPUT=$1

    echo "@ Updating repo on MacOS VM ..."
    ssh osxvm "cd $DIR && rm -rf '$OUTPUT' && git checkout . && git pull" > /dev/null

    echo "@ Building $OUTPUT ..."
    ssh osxvm "export GOPATH=/Users/evilsocket/gocode && cd '$DIR' && PATH=$PATH:/usr/local/bin && go build -o $OUTPUT ." > /dev/null

    echo "@ Downloading $OUTPUT ..."
    scp -C osxvm:$DIR/$OUTPUT . > /dev/null
}

build_windows_amd64() {
    vm_dep 'winvm'

    DIR=c:/Users/evilsocket/gopath/src/github.com/evilsocket/bettercap-ng
    OUTPUT=$1

    echo "@ Updating repo on Windows VM ..."
    ssh winvm "cd $DIR && del *.exe && git pull" > /dev/null

    echo "@ Building $OUTPUT ..."
    ssh winvm "cd $DIR && go build -o $OUTPUT ." > /dev/null

    echo "@ Downloading $OUTPUT ..."
    scp -C winvm:$DIR/$OUTPUT . > /dev/null
}

rm -rf $BUILD_FOLDER
mkdir $BUILD_FOLDER
cd $BUILD_FOLDER

build_linux_amd64 bettercap-ng_linux_amd64_$VERSION
build_linux_arm7 bettercap-ng_linux_arm7_$VERSION
build_macos_amd64 bettercap-ng_macos_amd64_$VERSION
build_windows_amd64 bettercap-ng_windows_amd64_$VERSION.exe

echo
echo
du -sh *

cd --



