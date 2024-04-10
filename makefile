servus: fmt test build

.PHONY: fmt
fmt:
	go fmt .

.PHONY: run
run:
	go run .

.PHONY: build
build:
	go clean .
	go build .

.PHONY: test
test:
	go test -test.v ./...

.PHONY: release
release:
	gh release create $(v) --generate-notes --prerelease
