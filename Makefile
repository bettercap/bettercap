TARGET=bettercap-ng
PCAP=http://www.tcpdump.org/release/libpcap-1.8.1.tar.gz
LDFLAGS='-linkmode external -extldflags "-static -s -w"'

all: fmt vet build
	@echo "@ Done"

test: build
	@go test ./...

build: resources
	@echo "@ Building ..."
	@go build $(FLAGS) -o $(TARGET) .

linux_arm:
	@echo "@ Cross compiling for linux/arm-7"
	@xgo -ldflags=$(LDFLAGS) --deps=$(PCAP) --depsargs="--with-pcap=linux" --targets=linux/arm-7 .

linux_arm64:
	@echo "@ Cross compiling for linux/arm64"
	@xgo -ldflags=$(LDFLAGS) --deps=$(PCAP) --depsargs="--with-pcap=linux" --targets=linux/arm64 .

linux_mips:
	@echo "@ Cross compiling for linux/mips"
	@xgo -ldflags=$(LDFLAGS) --deps=$(PCAP) --depsargs="--with-pcap=linux" --targets=linux/mips .

linux_mips64:
	@echo "@ Cross compiling for linux/mips64"
	@xgo -ldflags=$(LDFLAGS) --deps=$(PCAP) --depsargs="--with-pcap=linux" --targets=linux/mips64 .

resources: oui

oui:
	@$(GOPATH)/bin/go-bindata -o net/oui_compiled.go -pkg net net/oui.dat

vet:
	@go vet ./...

fmt:
	@go fmt ./...

lint:
	@golint ./...

deps:
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get github.com/ykyuen/xgo
	@go get ./...

clean:
	@rm -rf bettercap-ng*.*
	@rm -rf bettercap-ng*

clear_arp:
	@ip -s -s neigh flush all

bcast_ping:
	@ping -b 255.255.255.255

release:
	@./new_release.sh

deadlock_detect_build:
	@go get github.com/sasha-s/go-deadlock/...
	@find . -name "*.go" | xargs sed -i "s/sync.Mutex/deadlock.Mutex/"
	@goimports -w .
	@git status

