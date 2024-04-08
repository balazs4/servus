servus: clean fmt test build

.PHONY: fmt
fmt:
	go fmt .

.PHONY: fmt
run:
	go run .

.PHONY: build
build:
	go build .

.PHONY: clean
clean:
	go clean .

.PHONY: test
test:
	go test -test.v ./...

.PHONY: release
release:
	gh release create  "v1.$$(date '+%Y%m%d.%H%M')" --generate-notes
