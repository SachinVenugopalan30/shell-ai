VERSION ?= dev

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o shellai ./cmd/shellai

install:
	go install -ldflags "-s -w -X main.version=$(VERSION)" ./cmd/shellai

clean:
	rm -f shellai
