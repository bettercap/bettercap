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

create_exe_archive() {
    bin_dep 'zip'

    OUTPUT=$1

    echo "@ Creating archive $OUTPUT ..."
    zip -j "$OUTPUT" bettercap.exe ../README.md ../LICENSE.md > /dev/null
    rm -rf bettercap bettercap.exe
}

create_archive() {
    bin_dep 'zip'

    OUTPUT=$1

    echo "@ Creating archive $OUTPUT ..."
    zip -j "$OUTPUT" bettercap ../README.md ../LICENSE.md > /dev/null
    rm -rf bettercap bettercap.exe
}

build_linux_amd64() {
    echo "@ Building linux/amd64 ..."
    go build -o bettercap ..
}


build_linux_armv6l() {
    host_dep 'arc.local'

    DIR=/home/pi/gocode/src/github.com/bettercap/bettercap

    echo "@ Updating repo on arm6l host ..."
    ssh pi@arc.local "cd $DIR && rm -rf '$OUTPUT' && git checkout . && git checkout master && git pull" > /dev/null

    echo "@ Building linux/armv6l ..."
    ssh pi@arc.local "export GOPATH=/home/pi/gocode && cd '$DIR' && PATH=$PATH:/usr/local/bin && go get ./... && go build -o bettercap ." > /dev/null

    scp -C pi@arc.local:$DIR/bettercap . > /dev/null
}

build_macos_amd64() {
    host_dep 'osxvm'

    DIR=/Users/evilsocket/gocode/src/github.com/bettercap/bettercap

    echo "@ Updating repo on MacOS VM ..."
    ssh osxvm "cd $DIR && rm -rf '$OUTPUT' && git checkout . && git checkout master && git pull" > /dev/null

    echo "@ Building darwin/amd64 ..."
    ssh osxvm "export GOPATH=/Users/evilsocket/gocode && cd '$DIR' && PATH=$PATH:/usr/local/bin && go get ./... && go build -o bettercap ." > /dev/null

    scp -C osxvm:$DIR/bettercap . > /dev/null
}

build_windows_amd64() {
    host_dep 'winvm'

    DIR=c:/Users/evilsocket/gopath/src/github.com/bettercap/bettercap

    echo "@ Updating repo on Windows VM ..."
    ssh winvm "cd $DIR && git checkout . && git checkout master && git pull && go get ./..." > /dev/null

    echo "@ Building windows/amd64 ..."
    ssh winvm "cd $DIR && go build -o bettercap.exe ." > /dev/null

    scp -C winvm:$DIR/bettercap.exe . > /dev/null
}

build_android_arm() {
    host_dep 'shield'
    
    BASE=/data/data/com.termux/files
    THEPATH="$BASE/usr/bin:$BASE/usr/bin/applets:/system/xbin:/system/bin"
    LPATH="$BASE/usr/lib"
    GPATH=$BASE/home/go
    DIR=$GPATH/src/github.com/bettercap/bettercap

    echo "@ Updating repo on Android host ..."
    ssh -p 8022 root@shield "su -c 'export PATH=$THEPATH && export LD_LIBRARY_PATH="$LPATH" && cd "$DIR" && rm -rf bettercap* && git pull && export GOPATH=$GPATH && go get ./...'"

    echo "@ Building android/arm ..."
    ssh -p 8022 root@shield "su -c 'export PATH=$THEPATH && export LD_LIBRARY_PATH="$LPATH" && cd "$DIR" && export GOPATH=$GPATH && go build -o bettercap . && setenforce 0'"

    echo "@ Downloading bettercap ..."
    scp -C -P 8022 root@shield:$DIR/bettercap . 
}

rm -rf $BUILD_FOLDER
mkdir $BUILD_FOLDER
cd $BUILD_FOLDER

if [ -z "$1" ]
  then
      WHAT=all
  else
      WHAT="$1"
fi

printf "@ Building for $WHAT ...\n\n"

if [[ "$WHAT" == "all" || "$WHAT" == "linux_amd64" ]]; then
    build_linux_amd64 && create_archive bettercap_linux_amd64_$VERSION.zip
fi

if [[ "$WHAT" == "all" || "$WHAT" == "linux_armv6l" ]]; then
    build_linux_armv6l && create_archive bettercap_linux_armv6l_$VERSION.zip
fi

if [[ "$WHAT" == "all" || "$WHAT" == "osx" || "$WHAT" == "mac" || "$WHAT" == "macos" ]]; then
    build_macos_amd64 && create_archive bettercap_macos_amd64_$VERSION.zip
fi

if [[ "$WHAT" == "all" || "$WHAT" == "win" || "$WHAT" == "windows" ]]; then
    build_windows_amd64 && create_exe_archive bettercap_windows_amd64_$VERSION.zip
fi 

if [[ "$WHAT" == "all" || "$WHAT" == "android" ]]; then
    build_android_arm && create_archive bettercap_android_armv7l_$VERSION.zip
fi

sha256sum * > checksums.txt

echo
echo
du -sh *

cd --
