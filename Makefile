test:
	@go test ./...

publish:
	@GOPROXY=proxy.golang.org go list -m github.com/roma-glushko/testcrd@v0.0.1
