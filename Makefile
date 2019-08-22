export GO111MODULE=on

all: deps
test:
	go test -v ./...
clean:
	go clean
deps:
	go build -v .
upgrade:
	go get -u