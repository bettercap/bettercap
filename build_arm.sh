# nothing to see here, just what i use to cross compile for ARM
DIR=/Users/evilsocket/gocode/src/github.com/evilsocket/bettercap-ng
EXE=bettercap-ng_arm7

echo "@ Updating repo ..."
rm -rf $EXE && git pull

echo "@ Configuring libpcap ..."
rm -rf libpcap-*.*
rm -rf libpcap*
wget http://www.tcpdump.org/release/libpcap-1.8.1.tar.gz
tar xvf libpcap-1.8.1.tar.gz
cd libpcap-1.8.1
export CC=arm-linux-gnueabi-gcc
./configure --host=arm-linux --with-pcap=linux
make
 
echo "@ Building $EXE ..."
cd ..
env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CGO_LDFLAGS="-Llibpcap-1.8.1" go build -o $EXE .
rm -rf libpcap-1.8.1
