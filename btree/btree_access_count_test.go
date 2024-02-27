package btree_test

import (
	"btree-cache-benchmark/btree"
	"btree-cache-benchmark/utils"
	"fmt"
	"io"
	"os"
	"slices"
	"testing"
)

func TestCountAccess(t *testing.T) {
	ac := cacheAccessCounter{
		lastAccess: make(map[any]int),
		hist:       make(map[int]int),
	}
	b := btree.NewWithAccessCounter[int, int](2, ac.count)
	values := utils.GetSequenceRange(100_000)
	utils.Shuffle(values)
	for _, v := range values {
		b.Insert(v, v)
	}
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
	for _, ts := range c.hist {
		timestamps = append(timestamps, ts)
	}
	slices.Sort(timestamps)
	fmt.Fprintf(w, "ts\tcount\n")
	for _, ts := range timestamps {
		fmt.Fprintf(w, "%d\t%d\n", ts, c.hist[ts])
	}
}
