BINARY_NAME=watcher
MAIN_PACKAGE=cmd/watcher/watcher.go
GOOS=linux
GOARCH=amd64

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
