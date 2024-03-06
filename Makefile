bin=bin/btree_hist
gofiles=$(shell find . -name \*.go)
$(bin): $(gofiles)
	go build -o $(bin) main.go
test:
	go test -tags assertions ./...
benchmark:
	go test -bench=.  ./... -run='^#' | tee benchmark.log
clean:
	rm -rfv $(bin) out/
.phony: clean test 
