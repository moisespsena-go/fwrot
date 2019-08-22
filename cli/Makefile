export GO111MODULE=on

all: deps
spt: build_snapshot

test:
	go test -v ./...
clean:
	go clean
deps:
	go build -v ./...
upgrade:
	go get -u
build_snapshot:
	goreleaser --snapshot --skip-publish --rm-dist
build:
	goreleaser