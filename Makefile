all: import

import: import.go
	go build -o $@ $^

fmt:
	gofmt -s -w -l .
