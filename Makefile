TARGET=bettercap
PACKAGES=core firewall log modules network packets session tls

all: deps build

deps: godep golint gomegacheck
	@dep ensure

build: resources
	@go build -o $(TARGET) .

resources: network/oui.go

network/oui.go:
	@python ./network/make_oui.py

clean:
	@rm -rf $(TARGET).*
	@rm -rf $(TARGET)*
	@rm -rf build

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
		megacheck ./$$pkg ; \
		go test -race ./$$pkg -coverprofile=$$pkg.profile -covermode=atomic; \
		tail -n +2 $$pkg.profile >> coverage.profile && rm -rf $$pkg.profile ; \
	done

html_coverage: test
	@go tool cover -html=coverage.profile -o coverage.profile.html

codecov: test
	@bash <(curl -s https://codecov.io/bash)

benchmark: server_deps
	@go test ./... -v -run=doNotRunTests -bench=. -benchmem

# tools
godep:
	@go get -u github.com/golang/dep/...

golint:
	@go get github.com/golang/lint/golint

gomegacheck:
	@go get honnef.co/go/tools/cmd/megacheck
