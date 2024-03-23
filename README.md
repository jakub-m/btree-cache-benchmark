# Comparing performance of sequential and random insertions to a B-tree

Described in [the blog post][ref_blog_post].

[ref_blog_post]:https://jakub-m.github.io/2024/03/12/btree.html

To reproduce
1. Run `make benchmark`, which will generate `benchmark.log`
2. Run `bash generate_histograms.sh` to generate histograms for node access times.
3. Run `jupyter lab` and run `plot_benchmark.ipynb` to get the table with benchmark timings.
4. Run `plot_histograms.ipynb` to plot the "cache friendlieness" histograms.
5. Run `generate_rebalance_counters.sh` to generate rebalance counts.
6. Run `plot_rebalance_cnt.ipynb` to plot the counts.
