I read [this interesting article][ref_art] about choosing primary keys (PK) in Postgres databases. I found there a finding that using UUID as PK
causes performance drop w.r.t. to sequential integer PK.

The package implements [B-tree][ref_btree] with insertion (no deletion).


[ref_btree]:...
