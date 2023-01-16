lint:
	golangci-lint run -vvv -c  golangci.yaml

fmt:
	gofumpt -w .
	goimports -w .