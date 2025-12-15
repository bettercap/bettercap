TARGET   ?= bettercap
PACKAGES ?= core firewall log modules network packets session tls
PREFIX   ?= /usr/local
GO       ?= go
GOFMT    ?= gofmt

all: build

build: resources
	$(GO) build $(GOFLAGS) -o $(TARGET) .

build_with_race_detector: resources
	$(GO) build $(GOFLAGS) -race -o $(TARGET) .

resources: network/manuf.go

network/manuf.go:
	@python3 ./network/make_manuf.py

install:
	@mkdir -p $(DESTDIR)$(PREFIX)/share/bettercap/caplets
	@cp bettercap $(DESTDIR)$(PREFIX)/bin/

docker:
	@docker build -t bettercap:latest .

test:
	$(GO) test -covermode=atomic -coverprofile=cover.out ./...

html_coverage: test
	$(GO) tool cover -html=cover.out -o cover.out.html

benchmark: server_deps
	$(GO) test -v -run=doNotRunTests -bench=. -benchmem ./...

fmt:
	$(GOFMT) -s -w $(PACKAGES)

clean:
	$(RM) $(TARGET)
	$(RM) -r build

.PHONY: all build build_with_race_detector resources install docker test html_coverage benchmark fmt clean
