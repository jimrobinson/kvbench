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

- -i dat   - input path for data file
- -b bench - name of the benchmark to run (bolt, kv, kv-mu, leveldb, noop)
- -f path  - path to the database

- -p dur  - poll db at this interval and print statistics
- -wlog n - log every N write operations (approximate)
- -mr dur - abort if a read takes longer than this duration
- -mw dur - abort if a write takes longer than this duration

EXAMPLE

````
$ ./kvbench -n 100 -b0 0 -b1 4000 -k0 34 -k1 34 -v0 70 -v1 78 -o sample.dat
2014/04/17 17:39:07 writing sample.dat

$ ./kvbench -i sample.dat -b leveldb -f test/leveldb.db -d0 50ms -d1 175ms -p 10s
2014/04/17 17:39:15 reading sample.dat
2014/04/17 17:39:21 102478 rows written
2014/04/17 17:39:25 polling collection
2014/04/17 17:39:27 202517 rows written
2014/04/17 17:39:27 100 row sets arrived at an average inter-arrival rate of 116.007307ms
2014/04/17 17:39:27 leveldb: 202517 write ops in 796.554744ms: 3933 ns/op
2014/04/17 17:39:35 polling collection
2014/04/17 17:39:36 leveldb: 380563 read ops in 605.182374ms: 1590 ns/op

$ ./kvbench -i sample.dat -b bolt -f test/bolt.db -d0 50ms -d1 175ms -p 10s
2014/04/17 17:39:41 reading sample.dat
2014/04/17 17:39:48 102478 rows written
2014/04/17 17:39:51 polling collection
2014/04/17 17:39:55 202517 rows written
2014/04/17 17:39:55 100 row sets arrived at an average inter-arrival rate of 136.228483ms
2014/04/17 17:39:55 bolt: 202517 write ops in 11.890752546s: 58714 ns/op
2014/04/17 17:40:02 polling collection
2014/04/17 17:40:02 bolt: 354347 read ops in 308.743374ms: 871 ns/op

$ ./kvbench -i sample.dat -b kv -f test/kv.db -d0 50ms -d1 175ms -p 10s
2014/04/17 17:40:04 reading sample.dat
2014/04/17 17:40:14 polling collection
2014/04/17 17:41:14 read timeout reached: 1m0s
2014/04/17 17:41:14 Benchmark aborted

$ ./kvbench -i sample.dat -b kv-mu -f test/kv-mu.db -d0 50ms -d1 175ms -p 10s
2014/04/17 17:41:25 reading sample.dat
2014/04/17 17:41:35 polling collection
2014/04/17 17:41:47 polling collection
2014/04/17 17:41:58 polling collection
2014/04/17 17:42:08 polling collection
2014/04/17 17:42:20 polling collection
2014/04/17 17:42:34 polling collection
2014/04/17 17:42:47 polling collection
2014/04/17 17:42:47 102478 rows written
2014/04/17 17:42:59 polling collection
2014/04/17 17:43:12 polling collection
2014/04/17 17:43:24 polling collection
2014/04/17 17:43:38 polling collection
2014/04/17 17:43:53 polling collection
2014/04/17 17:44:08 polling collection
2014/04/17 17:44:23 polling collection
2014/04/17 17:44:37 polling collection
2014/04/17 17:44:52 polling collection
2014/04/17 17:44:53 202517 rows written
2014/04/17 17:44:53 100 row sets arrived at an average inter-arrival rate of 2.053726453s
2014/04/17 17:44:53 kv-mu: 202517 write ops in 3m4.16796948s: 909395 ns/op
2014/04/17 17:44:55 kv-mu: 1877504 read ops in 25.898371881s: 13794 ns/op

$ ./kvbench -i sample.dat -b kv -f test/kv-2.db -d0 2s -d1 5s -p 10s
2014/04/17 17:45:49 reading sample.dat
2014/04/17 17:45:59 polling collection
2014/04/17 17:46:09 polling collection
2014/04/17 17:46:19 polling collection
2014/04/17 17:46:29 polling collection
2014/04/17 17:46:39 polling collection
2014/04/17 17:46:52 polling collection
2014/04/17 17:47:02 polling collection
2014/04/17 17:47:13 polling collection
2014/04/17 17:47:23 polling collection
2014/04/17 17:47:34 polling collection
2014/04/17 17:47:45 polling collection
2014/04/17 17:47:55 polling collection
2014/04/17 17:48:06 polling collection
2014/04/17 17:48:18 polling collection
2014/04/17 17:48:31 polling collection
2014/04/17 17:48:40 102478 rows written
2014/04/17 17:48:44 polling collection
2014/04/17 17:49:01 polling collection
2014/04/17 17:49:14 polling collection
2014/04/17 17:49:29 polling collection
2014/04/17 17:49:50 polling collection
2014/04/17 17:50:06 polling collection
2014/04/17 17:50:22 polling collection
2014/04/17 17:50:54 polling collection
2014/04/17 17:51:17 polling collection
2014/04/17 17:51:19 202517 rows written
2014/04/17 17:51:19 100 row sets arrived at an average inter-arrival rate of 3.287451441s
2014/04/17 17:51:19 kv: 202517 write ops in 2m48.711798874s: 833074 ns/op
2014/04/17 17:51:22 kv: 2088446 read ops in 1m33.002394318s: 44531 ns/op

