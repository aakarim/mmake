SHELL := /bin/bash

test:
	go build ./... -o /dev/null
	go test ./...
dev_install:
	go install .
	source <(mmake completion)
readme:
	# Copy the usage output of the mmake to the README.md
	go run ./scripts/update_readme