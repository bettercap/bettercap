# nothing to see here, just what i use to build for Windows
#
# -> https://stackoverflow.com/questions/38047858/compile-gopacket-on-windows-64bit
DIR=c:/Users/evilsocket/gopath/src/github.com/evilsocket/bettercap-ng
AMD64_EXE=bettercap-ng_win32_amd64.exe
I386_EXE=bettercap-ng_win32_i386.exe

echo "@ Updating repo ..."
ssh winvm "cd $DIR && del *.exe && git pull"

echo "@ Building $AMD64_EXE ..."
ssh winvm "cd $DIR && go build -o $AMD64_EXE ."

# not worth the effort of installing the 32bit toolchain to be honest ...
# echo "@ Building $I386_EXE ..."
# ssh winvm "cd $DIR && set GOARCH=386 && go build -o $I386_EXE . && dir *.exe"

scp winvm:$DIR/$AMD64_EXE .
