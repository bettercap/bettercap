#!/bin/bash
BUILD_FOLDER=build
VERSION=$(cat core/banner.go | grep Version | cut -d '"' -f 2)

bin_dep() {
    BIN=$1
    which $BIN > /dev/null || { echo "@ Dependency $BIN not found !"; exit 1; }
}

host_dep() {
    HOST=$1
    ping -c 1 $HOST > /dev/null || { echo "@ Virtual machine host $HOST not visible !"; exit 1; }
}

download_pcap() {
    bin_dep 'wget'
    bin_dep 'tar'

    cd /tmp
    rm -rf libpcap-1.8.1
    if [ ! -f /tmp/libpcap-1.8.1.tar.gz ]; then
        echo "@ Downloading  https://www.tcpdump.org/release/libpcap-1.8.1.tar.gz ..."
        wget -q https://www.tcpdump.org/release/libpcap-1.8.1.tar.gz -O /tmp/libpcap-1.8.1.tar.gz
    fi
    tar xf libpcap-1.8.1.tar.gz
}

xcompile_pcap() {
    ARCH=$1
    HOST=$2
    COMPILER=$3

    bin_dep 'make'
    bin_dep 'yacc'
    bin_dep 'flex'
    bin_dep "$COMPILER"

    echo "@ Cross compiling libpcap for $ARCH with $COMPILER ..."
    cd /tmp/libpcap-1.8.1
    export CC=$COMPILER
    ./configure --host=$HOST --with-pcap=linux > /dev/null
    make CFLAGS='-w' -j4 > /dev/null
}

build_linux_amd64() {
    OUTPUT=$1
    echo "@ Building $OUTPUT ..."
    go build -o "$OUTPUT" ..
}

build_linux_arm7() {
    OUTPUT=$1
    OLD=$(pwd)

    download_pcap
    xcompile_pcap 'arm' 'arm-linux-gnueabi' 'arm-linux-gnueabi-gcc'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_linux_mips() {
    OUTPUT=$1
    OLD=$(pwd)

    download_pcap
    xcompile_pcap 'mips' 'mips-linux-gnu' 'mips-linux-gnu-gcc'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=mips-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=mips CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_linux_mipsle() {
    OUTPUT=$1
    OLD=$(pwd)

    download_pcap
    xcompile_pcap 'mipsel' 'mipsel-linux-gnu' 'mipsel-linux-gnu-gcc'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=mipsel-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=mipsle CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_linux_mips64() {
    OUTPUT=$1
    OLD=$(pwd)

    download_pcap
    xcompile_pcap 'mips64' 'mips64-linux-gnuabi64' 'mips64-linux-gnuabi64-gcc'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=mips64-linux-gnuabi64-gcc CGO_ENABLED=1 GOOS=linux GOARCH=mips64 CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_linux_mips64le() {
    OUTPUT=$1
    OLD=$(pwd)

    download_pcap
    xcompile_pcap 'mips64el' 'mips64el-linux-gnuabi64' 'mips64el-linux-gnuabi64-gcc'

    echo "@ Building $OUTPUT ..."
    cd "$OLD"
    env CC=mips64el-linux-gnuabi64-gcc CGO_ENABLED=1 GOOS=linux GOARCH=mips64le CGO_LDFLAGS="-L/tmp/libpcap-1.8.1" go build -o "$OUTPUT" ..
    rm -rf /tmp/libpcap-1.8.1
}

build_macos_amd64() {
    host_dep 'osxvm'

    DIR=/Users/evilsocket/gocode/src/github.com/bettercap/bettercap
    OUTPUT=$1

    echo "@ Updating repo on MacOS VM ..."
    ssh osxvm "cd $DIR && rm -rf '$OUTPUT' && git checkout . && git pull" > /dev/null

    echo "@ Building $OUTPUT ..."
    ssh osxvm "export GOPATH=/Users/evilsocket/gocode && cd '$DIR' && PATH=$PATH:/usr/local/bin && go get ./... && go build -o $OUTPUT ." > /dev/null

    echo "@ Downloading $OUTPUT ..."
    scp -C osxvm:$DIR/$OUTPUT . > /dev/null
}

build_windows_amd64() {
    host_dep 'winvm'

    DIR=c:/Users/evilsocket/gopath/src/github.com/bettercap/bettercap
    OUTPUT=$1

    echo "@ Updating repo on Windows VM ..."
    ssh winvm "cd $DIR && del *.exe && git pull" > /dev/null

    echo "@ Building $OUTPUT ..."
    ssh winvm "cd $DIR && go build -o $OUTPUT ." > /dev/null

    echo "@ Downloading $OUTPUT ..."
    scp -C winvm:$DIR/$OUTPUT . > /dev/null
}

build_android_arm() {
    host_dep 'shield'

    DIR=/data/data/com.termux/files/home/go/src/github.com/bettercap/bettercap
    OUTPUT=$1

    echo "@ Updating repo on Android host ..."
    ssh -p 8022 root@shield "cd "$DIR" && rm -rf bettercap* && git pull && go get ./..."

    echo "@ Building $OUTPUT ..."
    ssh -p 8022 root@shield "cd $DIR && go build  -o $OUTPUT ."

    echo "@ Downloading $OUTPUT ..."
    scp -C -P 8022 root@shield:$DIR/$OUTPUT . 
}

rm -rf $BUILD_FOLDER
mkdir $BUILD_FOLDER
cd $BUILD_FOLDER

build_android_arm bettercap_android_arm_$VERSION
build_linux_amd64 bettercap_linux_amd64_$VERSION
build_linux_arm7 bettercap_linux_arm7_$VERSION
build_linux_mips bettercap_linux_mips_$VERSION
build_linux_mipsle bettercap_linux_mipsle_$VERSION
build_linux_mips64 bettercap_linux_mips64_$VERSION
build_linux_mips64le bettercap_linux_mips64le_$VERSION
build_macos_amd64 bettercap_macos_amd64_$VERSION
build_windows_amd64 bettercap_windows_amd64_$VERSION.exe
sha256sum * > checksums.txt

echo
echo
du -sh *

cd --



