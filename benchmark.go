package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Benchmark wraps a Collection and evaluates how
// quickly the collection can iterate over its data
// set while writes are being applied.
type Benchmark struct {
	id    string
	c     Collection
	mu    *sync.RWMutex
	wg    *sync.WaitGroup
	guard *WatchDog
	done  chan bool
}

// NewBenchmark returns a initialized Benchmark
// with an underlying Collection based on the specified
// id and database path.
func NewBenchmark(id string, path string) (b *Benchmark, err error) {
	b = &Benchmark{
		id:    id,
		wg:    &sync.WaitGroup{},
		guard: NewWatchDog(),
		done:  make(chan bool),
	}

	switch id {
	case "bolt":
		b.c, err = NewBoltCollection(path)
	case "kv":
		b.c, err = NewKVCollection(path)
	case "kv-mu":
		b.c, err = NewKVCollection(path)
		b.mu = &sync.RWMutex{}
	case "leveldb":
		b.c, err = NewLevelDBCollection(path)
	case "noop":
		b.c, err = NewNoopCollection()
	default:
		err = fmt.Errorf("unknown benchmark id: %s", id)
	}

	return
}

// Wait blocks until the Run method has completed,
// and will return true if the benchmark was aborted.
func (b *Benchmark) Wait() (aborted bool) {
	b.wg.Wait()
	return b.guard.IsAborted()
}

// Run launches two goroutines, one to read row sets
// from ch, writing those rows to the collection, and another
// to wake up every dur interval and poll the collection for
// number of records and the time it took to iterate over
// those records.  If logevery is positive, the writer will
// log a message every time approximately that many
// rows have been written.
func (b *Benchmark) Run(ch chan []*Row, logevery int, mw, dur, mr time.Duration) {
	b.wg.Wait()
	b.wg.Add(1)
	go b.writer(ch, logevery, mw)
	go b.poller(dur, mr)
}

// writer reads rows from ch and writes them to
// the underlying Collection
func (b *Benchmark) writer(ch chan []*Row, logevery int, timeout time.Duration) {
	defer func() {
		b.done <- true
	}()

	var sets_ttl int // row set counter
	var rows_ttl int // row counter
	var logged int   // rows_ttl we last logged

	var a0, a1 time.Time        // time between arrivals
	var arrival_t time.Duration // cumulative arrival times
	var write_t time.Duration   // cumulative write time

	for rows := range ch {
		if b.guard.IsAborted() {
			return
		}

		// a new row set has arrived, record the time
		// and compute the elapsed time since the
		// previous arrival, if applicable
		a1 = time.Now()
		sets_ttl += 1
		rows_ttl += len(rows)

		if sets_ttl > 1 {
			arrival_t += time.Duration(a1.Sub(a0))
		}

		a0 = a1

		// write the row set to the collection, using
		// a goroutine to handle the case where the
		// collection takes too long to return and we
		// hit our WatchDog timeout.
		var err error
		ch := make(chan bool)
		go func(write_t *time.Duration, err *error) {

			if b.mu != nil {
				b.mu.Lock()
			}

			w0 := time.Now()
			*err = b.c.Set(rows)
			w1 := time.Now()

			// add to the elapsed time counter
			*write_t += time.Duration(w1.Sub(w0))

			if b.mu != nil {
				b.mu.Unlock()
			}

			ch <- true
		}(&write_t, &err)

		select {
		case _ = <-ch:
			// success
		case _ = <-b.guard.Timer(mw):
			b.guard.Abort(fmt.Sprintf("write timeout reached: %s", mw))
			return
		}

		if err != nil {
			log.Println(err)
			return
		}

		if logevery > 0 && rows_ttl-logged > logevery {
			log.Printf("%d rows written\n", rows_ttl)
			logged = rows_ttl
		}
	}

	if sets_ttl > 0 {
		log.Printf("%d row sets arrived at an average inter-arrival rate of %s",
			sets_ttl, time.Duration(arrival_t.Nanoseconds()/int64(sets_ttl)))

		log.Printf("%s: %d write ops in %s: %d ns/op\n",
			b.id, rows_ttl, write_t, write_t.Nanoseconds()/int64(rows_ttl))
	}
}

// poller wakes up every cycle duration and polls the
// underlying Collection Timer.  If a read takes more
// than timeout, abort.
func (b *Benchmark) poller(cycle, timeout time.Duration) {
	defer b.wg.Done()

	var n int                // total rows read
	var read_t time.Duration // cumulative read time

	for {
		if b.guard.IsAborted() {
			return
		}

		select {
		case _ = <-b.done:
			log.Printf("%s: %d read ops in %s: %d ns/op\n",
				b.id, n, read_t, (read_t.Nanoseconds() / int64(n)))
			return
		default:
			time.Sleep(cycle)

			ch := make(chan bool)
			go func(n *int, read_t *time.Duration) {
				log.Println("polling collection")

				if b.mu != nil {
					b.mu.RLock()
				}

				i, t := b.c.Timing()

				*n += i
				*read_t += t

				if b.mu != nil {
					b.mu.RUnlock()
				}
				ch <- true
			}(&n, &read_t)

			select {
			case _ = <-ch:
				// success
			case _ = <-b.guard.Timer(mr):
				b.guard.Abort(fmt.Sprintf("read timeout reached: %s", mr))
				return
			}
		}
	}
}
