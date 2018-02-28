TARGET=bettercap

all: fmt vet build
	@echo "@ Done"

test: build
	@go test ./...

build: resources
	@echo "@ Building ..."
	@go build -o $(TARGET) .

resources: oui

oui:
	@python ./network/make_oui.py

vet:
	@go vet ./...

fmt:
	@go fmt ./...

lint:
	@golint ./...

deps:
	@go get ./...

clean:
	@rm -rf $(TARGET).*
	@rm -rf $(TARGET)*
	@rm -rf build
