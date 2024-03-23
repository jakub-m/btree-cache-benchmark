package main

import (
	"btree-cache-benchmark/btree"
	"btree-cache-benchmark/utils"
	"flag"
	"fmt"
)

func main() {
	flagN := 0
	flagShuffle := false
	flagRandom := false
	flagOrder := 2
	flag.IntVar(&flagN, "n", 1000000, "number of values in the sequence")
	flag.BoolVar(&flagShuffle, "shuffle", false, "shuffle, can be used to shuffle sequence of N values")
	flag.BoolVar(&flagRandom, "random", false, "random integers")
	flag.IntVar(&flagOrder, "order", 2, "order of btree")
	flag.Parse()
	rc := counter{}
	b := btree.New[int, int](flagOrder)
	b.SetRebalanceCounter(rc.count)
	var values []int
	summary := ""

	if flagRandom {
		summary = "random"
		values = utils.GetRandomArray(flagN)
	} else {
		summary = "sequence"
		values = utils.GetSequenceRange(flagN)
	}
	if flagShuffle {
		summary = "shuffled"
		utils.Shuffle(values)
	}
	for _, v := range values {
		b.Insert(v, v)
	}
	fmt.Printf("%s\t%d\t%d\t%d\n", summary, flagOrder, flagN, rc.c)
}

type counter struct {
	c int
}

func (c *counter) count() {
	c.c++
}
