gofiles=$(shell find . -name \*.go)
bin/btree_hist: $(gofiles)
	go build -o bin/btree_hist ./cli/bree_hist/main.go
test:
	go test -tags assertions ./...
benchmark:
	go test -bench=.  ./... -run='^#' | tee benchmark.log
clean:
	rm -rfv bin/ out/
.phony: clean test 
