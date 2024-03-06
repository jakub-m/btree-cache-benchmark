# Comparing performance of sequential and random insertions to a B-tree

I read [this interesting article][ref_art] about choosing primary keys (PK) in Postgres. The article mentions that using UUID as PK causes performance drop w.r.t. to sequential integer PK:

> [...] when you use Postgres native UUID v4 type instead of bigserial table size grows by 25% and insert rate drops to 25% of bigserial.

Back when I worked in Amazon I recall a tech talk that mentioned that using a UUID keys instead of serial keys causes performance drops for databases
that use B-tree, because insertions of UUID to a B-tree are less cache-friendly than inserting serial integers.

I wanted to explore that further. I implemented a B-tree and run some benchmarks. While I can measure performance directly (with [Go benchmark][ref_go_bench]), I wanted also to somehow measure "cache friendliness". To measure cache friendliness, I counted time when a node of the B-tree was last accessed (where "time" is just a tick of a counter). If a node is accessed more frequently (less "ticks" between the accesses), there is a larger chance that the node is cached. The more "ticks" between the subsequent accesses of the mode, the more chance the node was evicted from the cache.

The "experiment" is very rough and the implementation is far from perfect.

[ref_go_bench]: https://pkg.go.dev/testing#hdr-Benchmarks
[ref_btree]: https://en.wikipedia.org/wiki/B-tree#Insertion
[ref_art]: https://shekhargulati.com/2022/07/08/my-notes-on-gitlabs-postgres-schema-design/

# Performance

First I compared insertion of a sequence of integers, versus inserting the same sequence but in a random order (shuffled). Inserting shuffled sequence should somehow emulate the scenario when one inserts UUID that is later hashed to an integer key.

The table summarises the time it takes to insert 100k keys to B-trees of different B-tree order. The "relative time" is the time it takes to insert the shuffled sequence `t_shuf` compared to time it takes to insert the straight sequence `t_seq`:

```
t_relative = t_shuf/t_seq - 1
```

The larger relative time, the more performance degradation when inserting the shuffled sequence.

| btree order | relative time (degradation) |
| ----------- | --------------------------- |
| 2           | 156.6%                      |
| 3           | 17.5%                       |
| 6           | 17.0%                       |
| 10          | 20.9%                       |
| 23          | 29.0%                       |

Apart of the tree of order of 2 (a binary tree), the performance is consistently ~20% worse for insertion of the shuffled sequence.

# TODO cache friendlieness

# TODO to check

- Check if there is a "sharp drop" of performance. Check the benchmarks while increasing n from 1000 to 1M, assume 2MB cache will kick
