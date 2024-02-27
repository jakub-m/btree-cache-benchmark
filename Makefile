bin=bin/main
gofiles=$(shell find . -name \*.go)
$(bin): $(gofiles)
	go build -o $(bin) main.go
test:
	go test -tags assertions ./...
benchmark:
	go test -bench=.  ./... -run='^#'
clean:
	rm -fv $(bin)
.phony: clean test
