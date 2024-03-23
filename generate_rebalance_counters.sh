#!/bin/bash

set -eu
set -o pipefail

rm -fv rebalance_counters.tsv
for order in 2 3 5 13; do
	for n in 1000 10000 100000; do
		for mode in "" "-shuffle"; do
			set -x
			bin/count_rebalance ${mode} -order ${order} -n ${n} | tee -a rebalance_counters.tsv
			set +x
		done
	done
done

