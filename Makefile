TARGET=bettercap-ng

all: fmt vet build
	@echo "@ Done"

test: build
	@go test ./...

build: resources
	@echo "@ Building ..."
	@go build -o $(TARGET) .

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
	@rm -rf $(TARGET).*
	@rm -rf $(TARGET)*
	@rm -rf build
