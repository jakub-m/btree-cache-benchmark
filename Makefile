gofiles=$(shell find . -name \*.go)
default: bin/btree_hist bin/count_rebalance
bin/btree_hist: $(gofiles)
	go build -o bin/btree_hist ./cli/bree_hist/main.go
bin/count_rebalance: $(gofiles)
	go build -o bin/count_rebalance cli/count_rebalance/main.go
test:
	go test -tags assertions ./...
benchmark:
	go test -bench=.  ./... -run='^#' | tee benchmark.log
clean:
	rm -rfv bin/ out/
.phony: clean test 
