SHELL := /bin/bash

run:
	@go build .
	@time go run .
