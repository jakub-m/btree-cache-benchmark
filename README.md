I read [this interesting article][ref_art] about choosing primary keys (PK) in Postgres. The article mentions that using UUID as PK causes performance drop w.r.t. to sequential integer PK:

> [...] when you use Postgres native UUID v4 type instead of bigserial table size grows by 25% and insert rate drops to 25% of bigserial.

Back when I worked in Amazon I recall a tech talk that mentioned that using a UUID keys instead of serial keys causes performance drops for databases
that use B-tree, because insertions of UUID to a B-tree are less cache-friendly than inserting serial integers.

I wanted to explore that further. I implemented a B-tree and run some benchmarks. While I can measure performance directly (with [Go benchmark][ref_go_bench]), I wanted also to somehow measure "cache friendliness". To measure cache friendliness, I counted time when a node of the B-tree was last accessed (where "time" is just a tick of a counter). If a node is accessed more frequently (less "ticks" between the accesses), there is a larger chance that the node is cached. The more "ticks" between the subsequent accesses of the mode, the more chance the node was evicted from the cache.

[ref_go_bench]: https://pkg.go.dev/testing#hdr-Benchmarks
[ref_btree]: https://en.wikipedia.org/wiki/B-tree#Insertion
[ref_art]: https://shekhargulati.com/2022/07/08/my-notes-on-gitlabs-postgres-schema-design/

| btree order | relative time (degradation) |
| ----------- | --------------------------- |
| 2           | 156.6%                      |
| 3           | 17.5%                       |
| 6           | 17.0%                       |
| 10          | 20.9%                       |
| 23          | 29.0%                       |

# TODO

- Check if there is a "sharp drop" of performance.
