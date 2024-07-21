.DEFAULT_GOAL := build

.PHONY: build run clean

build:
	go build -o bloomturtle

run: build
	./bloomturtle

clean:
	rm -f bloomturtle

test:
	go test ./...


