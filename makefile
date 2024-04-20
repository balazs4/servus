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
	go test ./...

.PHONY: release
release:
	gh release create $(version) --generate-notes --prerelease
