SHELL := /bin/bash

run:
	@CGO_ENABLED=0 \
	GOAMD64=v3 \
	go build \
	    -ldflags="-s -w" \
	    -gcflags="all=-l=4 -B" \
	    -o 1brc .

	@time ./1brc

profile:
	@CGO_ENABLED=0 \
	GOAMD64=v3 \
	go test -run=^\$$ -bench=BenchmarkRun -benchtime=3s -count=1 -cpuprofile cpu.out
	@go tool pprof -top cpu.out
