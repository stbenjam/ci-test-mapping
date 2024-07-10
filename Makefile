all: test build

verify: lint

build:
	go build .

test:
	go test ./...

mapping: build
	./ci-test-mapping map --mode=local
	./ci-test-mapping map-verify

unmapped:
	jq -r '.[] | select(.Component == "Unknown") | .Name' mapping.json | sort | uniq

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f ci-test-mapping
