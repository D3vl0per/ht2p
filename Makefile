lint: 
	golangci-lint run --fix

test:
	go clean -testcache && go test -race -cover ./...

test-v:
	go clean -testcache && go test ./... -v

benchmark:
	go test -benchmem -bench BenchmarkRequest github.com/D3vl0per/ht2p -benchtime=10s -count=6 | tee "ht2p-$(shell date --iso-8601=seconds).out"

golangci-lint-install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm coverage.out
	google-chrome-stable coverage.html