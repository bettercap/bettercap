# nothing to see here, just what i use to build for OS X
DIR=/Users/evilsocket/gocode/src/github.com/evilsocket/bettercap-ng
AMD64_EXE=bettercap-ng_darwin_x86_64

echo "@ Updating repo ..."
ssh osxvm "cd $DIR && rm -rf $AMD64_EXE && git pull"

echo "@ Building $AMD64_EXE ..."
ssh osxvm "export GOPATH=/Users/evilsocket/gocode && cd $DIR && PATH=$PATH:/usr/local/bin && go build -o $AMD64_EXE ."

scp osxvm:$DIR/$AMD64_EXE .
