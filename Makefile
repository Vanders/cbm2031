.PHONY: all

all: cbm2031 

cbm2031: *.go
	go build -o cbm2031 *.go

test:
	go test -v ./...
