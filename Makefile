SHELL := /bin/bash

run:
	@CGO_ENABLED=0 \
	GOAMD64=v3 \
	go build \
	    -ldflags="-s -w" \
	    -gcflags="all=-l=4 -B" \
	    -o 1brc .

	@time ./1brc
