package main

import (
	"btree-cache-benchmark/btree"
	"os"
)

func main() {
	b := btree.New[int, int](3)
	b.Insert(10, 10)
	b.Insert(20, 20)
	b.Insert(30, 30)
	b.Insert(40, 40)
	b.Print(os.Stdout)
}
