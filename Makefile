TARGET=bettercap
PACKAGES=core firewall log modules network packets session tls

all: deps build

deps: godep golint gofmt
	@dep ensure

build_with_race_detector: resources
	@go build -race -o $(TARGET) .

build: resources
	@go build -o $(TARGET) .

resources: network/manuf.go

network/manuf.go:
	@python ./network/make_manuf.py

clean:
	@rm -rf $(TARGET)
	@rm -rf build

install:
	@mkdir -p /usr/local/share/bettercap/caplets
	@cp bettercap /usr/local/bin/

docker:
	@docker build -t bettercap:latest .

# Go 1.9 doesn't support test coverage on multiple packages, while
# Go 1.10 does, let's keep it 1.9 compatible in order not to break
# travis
test: deps
	@echo "mode: atomic" > coverage.profile
	@for pkg in $(PACKAGES); do \
		go fmt ./$$pkg ; \
		go vet ./$$pkg ; \
		touch $$pkg.profile ; \
		go test -race ./$$pkg -coverprofile=$$pkg.profile -covermode=atomic; \
		tail -n +2 $$pkg.profile >> coverage.profile && rm -rf $$pkg.profile ; \
	done

html_coverage: test
	@go tool cover -html=coverage.profile -o coverage.profile.html

benchmark: server_deps
	@go test ./... -v -run=doNotRunTests -bench=. -benchmem

# tools
godep:
	@go get -u github.com/golang/dep/...

golint:
	@go get -u golang.org/x/lint/golint

gofmt:
	gofmt -s -w $(PACKAGES)
