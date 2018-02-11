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
	@go get ./...

clean:
	@rm -rf bettercap-ng*.*
	@rm -rf bettercap-ng*
	@rm -rf build
