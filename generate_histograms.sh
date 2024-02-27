#!/bin/bash
set -eux
mkdir -p out
order=3
bin/btree_hist -order $order > out/hist_m${order}.txt
#bin/btree_hist -order $order -random > out/hist_m${order}_rand.txt
bin/btree_hist -order $order -shuffle > out/hist_m${order}_shuff.txt

