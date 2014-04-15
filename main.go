package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
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
-p dur  - poll db at this interval and print statistics

-i dat   - input path for data file
-b bench - name of the benchmark to run (bolt, kv, leveldb, noop)
-f path  - path to the database
`

var rnd *rand.Rand

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

func NewBenchmark(id string, path string) (c Collection, err error) {
	switch id {
	case "bolt":
		c, err = NewBoltCollection(path)
	case "kv":
		c, err = NewKVCollection(path)
	case "leveldb":
		c, err = NewLevelDBCollection(path)
	case "noop":
		c, err = NewNoopCollection()
	default:
		err = fmt.Errorf("unknown benchmark id: %s", id)
	}
	return
}

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
	flag.StringVar(&benchmarkId, "b", "", "benchmark id: leveldb, kv, bolt")
	flag.StringVar(&databasePath, "f", "", "database path")
	flag.StringVar(&inputDat, "i", "", "input path for data")

	flag.Parse()

	if help {
		fmt.Println(usage)
		return
	}

	rnd = rand.New(rand.NewSource(seed))

	if outputDat != "" {
		fh, err := os.Create(outputDat)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("writing data to %s\n", outputDat)
		err = generateData(fh, blocks, b0, b1, k0, k1, v0, v1)
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

		collection, err := NewBenchmark(benchmarkId, databasePath)
		if err != nil {
			log.Println(err)
			return
		}

		wg := &sync.WaitGroup{}

		fh, err := os.Open(inputDat)
		if err != nil {
			log.Fatal(err)
		}

		defer fh.Close()

		ch := make(chan []*Row, 100)

		wg.Add(1)
		go runBenchmark(ch, benchmarkId, collection, p, wg)

		log.Printf("reading data from %s\n", inputDat)
		err = sendData(fh, ch, d0, d1)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
		}

		close(ch)

		wg.Wait()
	}
}

// runBenchmark runs two goroutines, one to read row sets
// from ch, writing those rows to the collection, and another
// to wake up every dur interval and poll the collection for
// number of records and the time it took to iterate over
// those records.
func runBenchmark(ch chan []*Row, id string, c Collection, dur time.Duration, wg *sync.WaitGroup) {
	done := make(chan bool)

	go func(id string, c Collection, ch chan []*Row) {
		n := int64(0)        // row set counter
		ns := int64(0)       // elapsed time in nanoseconds
		var t0, t1 time.Time // time between arrivals
		for rows := range ch {
			n, t1 = n+1, time.Now()
			if n > 1 {
				ns += t1.Sub(t0).Nanoseconds()
			}
			t0 = t1

			err := c.Set(rows)
			if err != nil {
				log.Println(err)
				return
			}
		}

		done <- true
		if n > 0 {
			log.Printf("%d row sets arrived at an average inter-arrival rate of %s", n, time.Duration(ns/n))
		}
	}(id, c, ch)

	go func(id string, c Collection, done <-chan bool, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case _ = <-done:
				n, t := c.Timing()
				log.Printf("%s\t%d\t%d ms\n", id, n, t.Nanoseconds()/1e6)
				return
			default:
				time.Sleep(dur)
				n, t := c.Timing()
				log.Printf("timing %s: %d in %d ms\n", id, n, t.Nanoseconds()/1e6)
			}
		}
	}(id, c, done, wg)
}

// sendData reads record blocks from r and sends them to ch.
// At least d0 duration will pass in-between sends on ch, and
// an attempt will be made to send within d1 duration.  Note
// that d1 is not guaranteed, as there are external factors
// that will affect how quickly each row can be prepared.
func sendData(r io.Reader, ch chan []*Row, d0, d1 time.Duration) (err error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	// t0 will be the last send time
	var t0 time.Time
	for {
		var x int64
		if err = binary.Read(br, binary.LittleEndian, &x); err != nil {
			return
		}

		rows := make([]*Row, 0, int(x))
		for i := 0; i < int(x); i++ {
			var k int64
			if err = binary.Read(br, binary.LittleEndian, &k); err != nil {
				err = fmt.Errorf("error reading key length: %v", err)
				return
			}

			kbuf := make([]byte, int(k))
			if err = binary.Read(br, binary.LittleEndian, kbuf); err != nil {
				return
			}

			var v int64
			if err = binary.Read(br, binary.LittleEndian, &v); err != nil {
				return
			}

			vbuf := make([]byte, int(v))
			if err = binary.Read(br, binary.LittleEndian, vbuf); err != nil {
				return
			}

			var rk RowKey
			rk, err = DecodeRowKey(kbuf)
			if err != nil {
				return
			}

			var rv *RowValue
			rv, err = DecodeRowValue(vbuf)
			if err != nil {
				return
			}

			rows = append(rows, &Row{Key: rk, Value: rv})
		}

		if !t0.IsZero() {
			// t1 is the elapsed time since the last send
			// if it is greater than our randomly computed
			// delay, sleep for the difference
			t1 := time.Now().Sub(t0)
			ns := int64(randN(int(d0), int(d1))) - t1.Nanoseconds()
			if ns > 0 {
				time.Sleep(time.Duration(ns))
			}
		}

		t0 = time.Now()
		ch <- rows
	}
	return
}

// generateData writes pseudo-random benchmark data
// to w based on the parameters provided
// - n indicates the total number of record blocks to write
// - b0 indicates the minimum number of records per block
// - b1 indicates the maximum number of records per block
// - k0 indicates the minimum key length
// - k1 indicates the maximum key length
// - v0 indicates the minimum value length
// - v1 indicates the maximum value length
func generateData(w io.Writer, n, b0, b1, k0, k1, v0, v1 int) (err error) {
	for i := 0; i < n; i++ {
		x := randN(b0, b1)
		if err = binary.Write(w, binary.LittleEndian, int64(x)); err != nil {
			return
		}

		for j := 0; j < x; j++ {
			k := randN(k0, k1)
			if err = binary.Write(w, binary.LittleEndian, int64(k)); err != nil {
				return
			}
			if err = randBytes(w, k); err != nil {
				return
			}

			v := randN(v0, v1)
			if err = binary.Write(w, binary.LittleEndian, int64(v)); err != nil {
				return
			}
			if err = randBytes(w, v); err != nil {
				return
			}
		}
	}
	return
}

// randN returns a pseudo-random int between min
// and max.  If max-min is < 1, the result will be min.
func randN(min, max int) int {
	diff := max - min
	if diff < 1 {
		return min
	}
	return min + rnd.Intn(max-min)
}

// randBytes writes n pseudo random bytes,
// in the range 0 through 255, to w.
func randBytes(w io.Writer, n int) error {
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(rand.Intn(255))
	}
	return binary.Write(w, binary.LittleEndian, buf)
}
