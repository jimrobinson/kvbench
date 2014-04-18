package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

var usage = `USAGE:

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

-i dat   - input path for data file
-b bench - name of the benchmark to run (bolt, kv, leveldb, noop)
-f path  - path to the database

-p dur  - poll db at this interval and print statistics
-wlog n  - log every N write operations (approximate)
-mr dur - abort if a read takes longer than this duration
-mw dur - abort if a write takes longer than this duration
`
var help bool

var seed int64

var blocks int
var b0 int
var b1 int
var k0 int
var k1 int
var v0 int
var v1 int
var outputDat string

var d0 time.Duration
var d1 time.Duration
var p time.Duration
var benchmarkId string
var databasePath string
var inputDat string

var wlog int
var mr time.Duration
var mw time.Duration

func main() {
	flag.BoolVar(&help, "h", false, "print usage")

	flag.Int64Var(&seed, "r", 0, "pseudo random seed")

	flag.IntVar(&blocks, "n", 200, "number of record blocks")
	flag.IntVar(&b0, "b0", 1, "minimum number of records per block")
	flag.IntVar(&b1, "b1", 1000, "maximum number of records per block")
	flag.IntVar(&k0, "k0", 32, "minimum number of bytes in a key")
	flag.IntVar(&k1, "k1", 32, "maximum number of bytes in a key")
	flag.IntVar(&v0, "v0", 512, "minimum number of bytes in a value")
	flag.IntVar(&v1, "v1", 1024, "maximum number of bytes in a value")
	flag.StringVar(&outputDat, "o", "", "output path for data")

	flag.DurationVar(&d0, "d0", 500*time.Millisecond, "minimum inter-arrival rate")
	flag.DurationVar(&d1, "d1", time.Second, "maximum inter-arrival rate (not guaranteed)")
	flag.DurationVar(&p, "p", 10*time.Second, "poll db at this interval and print statistics")
	flag.StringVar(&benchmarkId, "b", "", "benchmark id: leveldb, kv, kv-mu, bolt")
	flag.StringVar(&databasePath, "f", "", "database path")
	flag.StringVar(&inputDat, "i", "", "input path for data")
	flag.IntVar(&wlog, "wlog", 100000, "log after this many write operations")
	flag.DurationVar(&mw, "mw", time.Minute, "abort if a write takes longer than this duration")
	flag.DurationVar(&mr, "mr", time.Minute, "abort if a read takes longer than this duration")

	flag.Parse()

	if help {
		fmt.Println(usage)
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	rnd := NewRandom(seed)

	if outputDat != "" {
		fh, err := os.Create(outputDat)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("writing %s\n", outputDat)
		err = rnd.Write(fh, blocks, b0, b1, k0, k1, v0, v1)
		if err != nil {
			log.Fatal(err)
		}

		fh.Close()
	}

	if inputDat != "" {
		if benchmarkId == "" {
			fmt.Println("missing required -b <benchmarkId> argument")
			return
		}
		if databasePath == "" {
			fmt.Println("missing required -f <database> argument")
			return
		}

		fh, err := os.Open(inputDat)
		if err != nil {
			log.Fatal(err)
		}

		defer fh.Close()

		// create a new benchmark
		benchmark, err := NewBenchmark(benchmarkId, databasePath)
		if err != nil {
			log.Println(err)
			return
		}

		// feed input data to the channel
		ch := make(chan []*Row, 100)
		go func() {
			log.Printf("reading %s\n", inputDat)
			err = rnd.Send(ch, fh, d0, d1)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
			}

			close(ch)
		}()

		// consume the channel to run the benchmark,
		// log every every wlog writes (approximately),
		// and poll the collection every p duration.
		benchmark.Run(ch, wlog, mw, p, mr)

		// wait for benchmark to complete, checking
		// whether or not it aborted.
		aborted := benchmark.Wait()
		if aborted {
			log.Println("Benchmark aborted")
		}
	}
}
