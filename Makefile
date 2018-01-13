TARGET=bettercap-ng

all: fmt vet build
	@echo "@ Done"

test: build
	@go test ./...

build: resources
	@echo "@ Building ..."
	@go build $(FLAGS) -o $(TARGET) .

resources:
	@echo "@ Compiling resources into go files ..."
	@$(GOPATH)/bin/go-bindata -o net/oui_compiled.go -pkg net net/oui.dat

vet:
	@echo "@ Running VET ..."
	@go vet ./...

fmt:
	@echo "@ Formatting ..."
	@go fmt ./...

lint:
	@echo "@ Running LINT ..."
	@golint ./...

deps:
	@echo "@ Installing dependencies ..."
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get github.com/elazarl/goproxy
	@go get github.com/google/gopacket 
	@go get github.com/mdlayher/dhcp6
	@go get github.com/malfunkt/iprange
	@go get github.com/rogpeppe/go-charset/charset
	@go get github.com/chzyer/readline
	@go get github.com/robertkrimen/otto
	@go get github.com/dustin/go-humanize
	@go get github.com/olekukonko/tablewriter

clean:
	@rm -rf $(TARGET) net/oui_compiled.go

clear_arp:
	@ip -s -s neigh flush all

bcast_ping:
	@ping -b 255.255.255.255
