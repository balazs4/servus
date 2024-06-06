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
	@gh release list --exclude-drafts --json 'tagName' --jq '.[].tagName' \
		| sort -h \
		| tail -1 \
		| xargs -t -I{} bun x semver --inc ${level} {} \
		| xargs -t -I{} gh release create v{} --generate-notes --prerelease
