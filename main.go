package main

import (
	"btree-cache-benchmark/btree"
	"btree-cache-benchmark/utils"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
)

func main() {
	flagN := 0
	flagShuffle := false
	flagRandom := false
	flagOrder := 2
	flag.IntVar(&flagN, "n", 1000000, "number of values in the sequence")
	flag.BoolVar(&flagShuffle, "shuffle", false, "shuffle, can be used to shuffle sequence of N values")
	flag.BoolVar(&flagRandom, "r", false, "random integers")
	flag.IntVar(&flagOrder, "m", 2, "order of btree")
	flag.Parse()
	ac := cacheAccessCounter{
		lastAccess: make(map[any]int),
		hist:       make(map[int]int),
	}
	b := btree.NewWithAccessCounter[int, int](flagOrder, ac.count)
	var values []int
	summary := "#"
	summary += fmt.Sprint(" n=", flagN)

	if flagRandom {
		summary += " random"
		values = utils.GetRandomArray(flagN)
	} else {
		summary += " sequence"
		values = utils.GetSequenceRange(flagN)
	}
	if flagShuffle {
		summary += " shuffled"
		utils.Shuffle(values)
	}
	for _, v := range values {
		b.Insert(v, v)
	}
	fmt.Fprintln(os.Stderr, summary)
	ac.writeHistogram(os.Stdout)
}

type cacheAccessCounter struct {
	ts         int
	lastAccess map[any]int
	hist       map[int]int
}

func (c *cacheAccessCounter) count(n any) {
	c.ts++
	if prevTs, ok := c.lastAccess[n]; ok {
		dt := c.ts - prevTs
		c.hist[dt] = c.hist[dt] + 1
		// if !ok then the object is a cache miss for sure (never accessed). Don't add it to stats since it
		// will be once per object.
	}
	c.lastAccess[n] = c.ts
}

func (c *cacheAccessCounter) writeHistogram(w io.Writer) {
	timestamps := []int{}
	for ts := range c.hist {
		timestamps = append(timestamps, ts)
	}
	slices.Sort(timestamps)
	fmt.Fprintf(w, "ts\tcount\n")
	for _, ts := range timestamps {
		cnt := c.hist[ts]
		fmt.Fprintf(w, "%d\t%d\n", ts, cnt)
	}
}
