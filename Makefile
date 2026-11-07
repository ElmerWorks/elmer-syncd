BINARY:=elmer-syncd
VERSION?=v0.0.0

.PHONY: all build cross release-local release

all: build

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/elmer-syncd

cross:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY)-linux ./cmd/elmer-syncd
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY).exe ./cmd/elmer-syncd

release-local:
	goreleaser --snapshot --rm-dist

release:
	goreleaser release --rm-dist
