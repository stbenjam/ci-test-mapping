all: test build

verify: lint

build:
	go build .

test:
	go test ./...

mapping: build
	./ci-test-mapping map --mode=local
	git diff

unmapped:
	jq '.[] | select(.Component == "Unknown") | .Name' mapping.json | sort | uniq

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f ci-test-mapping
