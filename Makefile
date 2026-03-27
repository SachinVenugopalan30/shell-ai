VERSION ?= dev

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o shellai .

install:
	go install -ldflags "-s -w -X main.version=$(VERSION)" .

clean:
	rm -f shellai
