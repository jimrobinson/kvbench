USAGE:

kvbench OPTIONS

OVERVIEW

kvbench is a simple benchmarking tool to evaluate the read performance
of a key/value store while writes are being applied in parallel.

DETAILS:

Running a benchmark consists of two steps:

(1) Generate a sample data file using the -o option.
    Use the -n, -b[01], -k[01], and -v[01] options to control the
    size of the sample.  Use the -r <seed> option to change the
    pseudo-random data.  Given identical inputs, an identical data
    file should be generated.

(2) Consume a sample data file and execute a benchmark using the
    -i option.  Use the -d[01] options to control the inter-arrival
    rate of new row sets to be written to the collection.  The -p
    option controls how often the benchmark will attempt to iterate
    over the keys.

You may specify both -o and -i options to generate the data and
then immediately run the benchmark.

OUTPUT OPTIONS

- -r n - pseudo-random seed
- -n n - total number of blocks to generate
- -b0 min - minimum number of records per block
- -b1 max - maximum number of records per block
- -k0 min - minimum length of key to generate
- -k1 max - maximum length of key to generate
- -v0 min - minimum length of value to generate
- -v1 max - maximum length of value to generate
-o dat - output path for data file

INPUT OPTIONS

- -r n    - pseudo-random seed
- -d0 dur - minimum inter-arrival rate
- -d1 dur - maximum inter-arrival rate (not guaranteed)
- -p dur  - poll db at this interval and print statistics
- -i dat   - input path for data file
- -b bench - name of the benchmark to run (bolt, kv, kv-mu, leveldb, noop)
- -f path  - path to the database

EXAMPLE

````
$ ./kvbench -n 100 -b0 0 -b1 4000 -k0 34 -k1 34 -v0 70 -v1 78 -o sample.dat
2014/04/16 19:57:46 writing data to sample.dat

$ ./kvbench -i sample.dat -b leveldb -f test/leveldb.db -d0 50ms -d1 175ms -p 10s
2014/04/16 19:57:51 reading data from sample.dat
2014/04/16 19:58:02 leveldb: 178046 ops in 296 ms: 601 ops/ms
2014/04/16 19:58:12 leveldb: 202517 ops in 315 ms: 642 ops/ms
2014/04/16 19:58:12 100 row sets arrived at an average inter-arrival rate of 116.046572ms
2014/04/16 19:58:12 leveldb: 202517 ops in 308 ms: 657 ops/ms

$ ./kvbench -i sample.dat -b bolt -f test/bolt.db -d0 50ms -d1 175ms -p 10s
2014/04/16 19:58:31 reading data from sample.dat
2014/04/16 19:58:41 bolt: 151830 ops in 141 ms: 1076 ops/ms
2014/04/16 19:58:51 bolt: 202517 ops in 210 ms: 964 ops/ms
2014/04/16 19:58:51 100 row sets arrived at an average inter-arrival rate of 134.606206ms
2014/04/16 19:58:51 bolt: 202517 ops in 168 ms: 1205 ops/ms

$ ./kvbench -i sample.dat -b kv -f test/kv.db -d0 50ms -d1 175ms -p 10s
2014/04/16 19:58:53 reading data from sample.dat
^C

$ date
Wed Apr 16 19:59:54 PDT 2014

$ ./kvbench -i sample.dat -b kv-mu -f test/kv-mu.db -d0 50ms -d1 175ms -p 10s
2014/04/16 20:00:01 reading data from sample.dat
2014/04/16 20:00:12 kv-mu: 18575 ops in 378 ms: 49 ops/ms
2014/04/16 20:00:24 kv-mu: 34127 ops in 588 ms: 58 ops/ms
2014/04/16 20:00:36 kv-mu: 50581 ops in 554 ms: 91 ops/ms
2014/04/16 20:00:49 kv-mu: 65373 ops in 1267 ms: 51 ops/ms
2014/04/16 20:01:00 kv-mu: 78226 ops in 1108 ms: 70 ops/ms
2014/04/16 20:01:12 kv-mu: 90815 ops in 1395 ms: 65 ops/ms
2014/04/16 20:01:25 kv-mu: 102478 ops in 2174 ms: 47 ops/ms
2014/04/16 20:01:38 kv-mu: 112228 ops in 1869 ms: 60 ops/ms
2014/04/16 20:01:51 kv-mu: 122214 ops in 2357 ms: 51 ops/ms
2014/04/16 20:02:06 kv-mu: 133389 ops in 1981 ms: 67 ops/ms
2014/04/16 20:02:18 kv-mu: 142738 ops in 1950 ms: 73 ops/ms
2014/04/16 20:02:30 kv-mu: 150676 ops in 1979 ms: 76 ops/ms
2014/04/16 20:02:46 kv-mu: 161695 ops in 2518 ms: 64 ops/ms
2014/04/16 20:03:00 kv-mu: 170495 ops in 2986 ms: 57 ops/ms
2014/04/16 20:03:16 kv-mu: 180825 ops in 3610 ms: 50 ops/ms
2014/04/16 20:03:33 kv-mu: 191325 ops in 3493 ms: 54 ops/ms
2014/04/16 20:03:48 kv-mu: 198455 ops in 4151 ms: 47 ops/ms
2014/04/16 20:04:01 kv-mu: 202517 ops in 3252 ms: 62 ops/ms
2014/04/16 20:04:01 100 row sets arrived at an average inter-arrival rate of 2.300565857s
2014/04/16 20:04:05 kv-mu: 202517 ops in 3538 ms: 57 ops/ms

$ ./kvbench -i sample.dat -b kv -f test/kv-2.db -d0 2s -d1 5s -p 10s
2014/04/16 20:04:40 reading data from sample.dat
2014/04/16 20:04:50 kv: 2936 ops in 39 ms: 75 ops/ms
2014/04/16 20:05:00 kv: 7593 ops in 110 ms: 69 ops/ms
2014/04/16 20:05:11 kv: 14963 ops in 287 ms: 52 ops/ms
2014/04/16 20:05:21 kv: 20289 ops in 274 ms: 74 ops/ms
2014/04/16 20:05:33 kv: 27723 ops in 2334 ms: 11 ops/ms
2014/04/16 20:05:44 kv: 34127 ops in 469 ms: 72 ops/ms
2014/04/16 20:05:54 kv: 40052 ops in 524 ms: 76 ops/ms
2014/04/16 20:06:05 kv: 44990 ops in 672 ms: 66 ops/ms
2014/04/16 20:06:16 kv: 51604 ops in 733 ms: 70 ops/ms
2014/04/16 20:06:27 kv: 58077 ops in 976 ms: 59 ops/ms
2014/04/16 20:06:39 kv: 69760 ops in 2488 ms: 28 ops/ms
2014/04/16 20:06:50 kv: 77416 ops in 1133 ms: 68 ops/ms
2014/04/16 20:07:02 kv: 82087 ops in 1714 ms: 47 ops/ms
2014/04/16 20:07:13 kv: 88134 ops in 1253 ms: 70 ops/ms
2014/04/16 20:07:27 kv: 97417 ops in 3505 ms: 27 ops/ms
2014/04/16 20:07:44 kv: 107525 ops in 7438 ms: 14 ops/ms
2014/04/16 20:08:00 kv: 118256 ops in 6163 ms: 19 ops/ms
2014/04/16 20:08:16 kv: 130346 ops in 5284 ms: 24 ops/ms
2014/04/16 20:08:36 kv: 143446 ops in 10035 ms: 14 ops/ms
2014/04/16 20:08:48 kv: 150676 ops in 2252 ms: 66 ops/ms
2014/04/16 20:09:04 kv: 158712 ops in 6425 ms: 24 ops/ms
2014/04/16 20:09:41 kv: 184613 ops in 26382 ms: 6 ops/ms
2014/04/16 20:10:00 kv: 197430 ops in 9778 ms: 20 ops/ms
2014/04/16 20:10:15 kv: 202517 ops in 4464 ms: 45 ops/ms
2014/04/16 20:10:15 100 row sets arrived at an average inter-arrival rate of 3.28728241s
2014/04/16 20:10:18 kv: 202517 ops in 3482 ms: 58 ops/ms
````
