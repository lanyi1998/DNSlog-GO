test:
	hurl --test --variables-file tests/api/.env tests/api/

test-go:
	go test ./...

build-cli:
	mkdir -p build
	go build -o build/dnslog-cli ./cmd/dnslog-cli
