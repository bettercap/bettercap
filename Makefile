TARGET=bettercap-ng
BUILD_DATE=`date +%Y-%m-%d\ %H:%M`
BUILD_FILE=core/build.go

all: build
	@echo "@ Done"
	@echo -n "\n"

test: build
	@go test ./...

build: build_file
	@echo "@ Building ..."
	@go build $(FLAGS) -o $(TARGET) .

build_file: resources
	@rm -f $(BUILD_FILE)
	@echo "package core" > $(BUILD_FILE)
	@echo "const (" >> $(BUILD_FILE)
	@echo "  BuildDate = \"$(BUILD_DATE)\"" >> $(BUILD_FILE)
	@echo ")" >> $(BUILD_FILE)

resources: deps
	@echo "@ Compiling resources into go files ..."
	@go-bindata -o net/oui_compiled.go -pkg net net/

deps:
	@echo "@ Installing dependencies ..."
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get github.com/elazarl/goproxy
	@go get github.com/google/gopacket 
	@go get github.com/malfunkt/iprange
	@go get github.com/rogpeppe/go-charset/charset
	@go get github.com/chzyer/readline
	@go get github.com/op/go-logging


clean:
	@rm -rf $(TARGET) net/oui_compiled.go

clear_arp:
	@ip -s -s neigh flush all

bcast_ping:
	@ping -b 255.255.255.255
