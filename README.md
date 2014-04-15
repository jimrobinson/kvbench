USAGE:

kvbench OPTIONS

DETAILS:

Running a benchmark consists of two steps:

(1) Generate a sample data file using the -o option.
    Use the -n, -b[01], -k[01], and -v[01] options to control the
    size of the sample.  Use the -r <seed> option to change the
    pseudo-random data.  Given identical inputs, an identical data
    file should be generated.

(2) Consume a sample data file and execute a benchmark using the
    -i option.  Or use both -o and -i options to generate the data
    and then run the benchmark.  Use the -d[01] options to control
    the inter-arrival rate of new row sets to be written to the
    collection.  The -p option controls how often the benchmark
    will attempt to iterate over the keys.

OUTPUT OPTIONS

-r n - pseudo-random seed
-n n - total number of blocks to generate

-b0 min - minimum number of records per block
-b1 max - maximum number of records per block

-k0 min - minimum length of key to generate
-k1 max - maximum length of key to generate

-v0 min - minium length of value to generate
-v1 max - maximum length of value to generate

-o dat - output path for data file

INPUT OPTIONS

-r n    - pseudo-random seed
-d0 dur - minimum inter-arrival rate
-d1 dur - maximum inter-arrival rate (not guaranteed)
-p dur  - poll db at this interval and print statistics

-i dat   - input path for data file
-b bench - name of the benchmark to run (bolt, kv, leveldb, noop)
-f path  - path to the database

EXAMPLE

````
$ ./kvbench -n 100 -b0 0 -b1 4000 -k0 34 -k1 34 -v0 70 -v1 78 -o sample.dat
2014/04/15 02:41:52 writing data to sample.dat

$ rm -rf test; mkdir test; ./kvbench -i sample.dat -b leveldb -f test/leveldb.db -d0 50ms -d1 175ms -p 10s
2014/04/15 02:40:56 reading data from sample.dat
2014/04/15 02:41:06 timing leveldb: 170828 in 286 ms
2014/04/15 02:41:16 timing leveldb: 201825 in 315 ms
2014/04/15 02:41:16 100 row sets arrived at an average inter-arrival rate of 116.036663ms
2014/04/15 02:41:17 leveldb 201825	 304 ms

$ rm -rf test; mkdir test; ./kvbench -i sample.dat -b bolt -f test/bolt.db -d0 50ms -d1 175ms -p 10s
2014/04/15 02:41:26 reading data from sample.dat
2014/04/15 02:41:36 timing bolt: 124452 in 118 ms
2014/04/15 02:41:46 timing bolt: 201825 in 177 ms
2014/04/15 02:41:46 100 row sets arrived at an average inter-arrival rate of 156.916447ms
2014/04/15 02:41:47 bolt    201825	 163 ms

$ rm -rf test; mkdir test; ./kvbench -i sample.dat -b kv -f test/kv.db -d0 50ms -d1 175ms -p 10s
2014/04/15 02:43:12 reading data from sample.dat
^C
$ date
Tue Apr 15 02:43:52 PDT 2014

$ rm -rf test; mkdir test; ./kvbench -i sample.dat -b kv -f test/kv.db -d0 2s -d1 5s -p 10s
2014/04/15 02:34:14 reading data from sample.dat
2014/04/15 02:34:24 timing kv: 5065 in 65 ms
2014/04/15 02:34:34 timing kv: 9102 in 123 ms
2014/04/15 02:34:46 timing kv: 15781 in 2396 ms
2014/04/15 02:34:57 timing kv: 21875 in 287 ms
2014/04/15 02:35:07 timing kv: 26462 in 371 ms
2014/04/15 02:35:17 timing kv: 31663 in 419 ms
2014/04/15 02:35:32 timing kv: 41591 in 4623 ms
2014/04/15 02:35:43 timing kv: 48843 in 640 ms
2014/04/15 02:35:53 timing kv: 55129 in 670 ms
2014/04/15 02:36:04 timing kv: 62049 in 799 ms
2014/04/15 02:36:18 timing kv: 69230 in 3993 ms
2014/04/15 02:36:30 timing kv: 73580 in 1801 ms
2014/04/15 02:36:41 timing kv: 83055 in 1091 ms
2014/04/15 02:36:59 timing kv: 93630 in 7983 ms
2014/04/15 02:37:13 timing kv: 103533 in 3580 ms
2014/04/15 02:37:28 timing kv: 112750 in 5834 ms
2014/04/15 02:37:46 timing kv: 125726 in 7201 ms
2014/04/15 02:37:59 timing kv: 135485 in 3218 ms
2014/04/15 02:38:12 timing kv: 139912 in 3156 ms
2014/04/15 02:38:30 timing kv: 143677 in 8375 ms
2014/04/15 02:38:54 timing kv: 166526 in 13873 ms
2014/04/15 02:39:14 timing kv: 179326 in 9380 ms
2014/04/15 02:39:38 timing kv: 194762 in 14047 ms
2014/04/15 02:39:51 timing kv: 201825 in 2885 ms
2014/04/15 02:39:51 100 row sets arrived at an average inter-arrival rate of 3.287373923s
2014/04/15 02:39:53 kv	201825	 2615 ms
````
