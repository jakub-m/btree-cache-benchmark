bin=bin/main
gofiles=$(shell find . -name \*.go)
$(bin): $(gofiles)
	go build -o $(bin) main.go
test:
	go test -tags assertions ./...
clean:
	rm -fv $(bin)
.phony: clean test
