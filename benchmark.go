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
	id   string
	c    Collection
	mu   *sync.RWMutex
	wg   *sync.WaitGroup
	done chan bool
}

// NewBenchmark returns a initialized Benchmark
// with an underlying Collection based on the specified
// id and database path.
func NewBenchmark(id string, path string) (b *Benchmark, err error) {
	b = &Benchmark{
		id:   id,
		wg:   &sync.WaitGroup{},
		done: make(chan bool),
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

// Wait blocks until the Run method has completed.
func (b *Benchmark) Wait() {
	b.wg.Wait()
}

// Run launches two goroutines, one to read row sets
// from ch, writing those rows to the collection, and another
// to wake up every dur interval and poll the collection for
// number of records and the time it took to iterate over
// those records.
func (b *Benchmark) Run(ch chan []*Row, dur time.Duration) {
	b.wg.Add(1)
	go b.Writer(ch)
	go b.Poll(dur)
}

// Writer reads rows from ch and writes them to
// the underlying Collection
func (b *Benchmark) Writer(ch chan []*Row) {
	n := int64(0)        // row set counter
	ns := int64(0)       // elapsed time in nanoseconds
	var t0, t1 time.Time // time between arrivals
	for rows := range ch {
		n, t1 = n+1, time.Now()
		if n > 1 {
			ns += t1.Sub(t0).Nanoseconds()
		}
		t0 = t1

		if b.mu != nil {
			b.mu.Lock()
		}
		err := b.c.Set(rows)
		if b.mu != nil {
			b.mu.Unlock()
		}

		if err != nil {
			log.Println(err)
			return
		}
	}

	b.done <- true
	if n > 0 {
		log.Printf("%d row sets arrived at an average inter-arrival rate of %s",
			n, time.Duration(ns/n))
	}
}

// Poll wakes up every dur duration and polls the
// underlying Collection Timer.
func (b *Benchmark) Poll(dur time.Duration) {
	defer b.wg.Done()
	for {
		select {
		case _ = <-b.done:
			if b.mu != nil {
				b.mu.RLock()
			}
			n, t := b.c.Timing()
			if b.mu != nil {
				b.mu.RUnlock()
			}
			ms := t.Nanoseconds() / 1e6
			opsms := int64(n) / ms
			log.Printf("%s: %d ops in %d ms: %d ops/ms\n",
				b.id, n, ms, opsms)
			return
		default:
			time.Sleep(dur)
			if b.mu != nil {
				b.mu.RLock()
			}
			n, t := b.c.Timing()
			if b.mu != nil {
				b.mu.RUnlock()
			}
			ms := t.Nanoseconds() / 1e6
			opsms := int64(n) / ms
			log.Printf("%s: %d ops in %d ms: %d ops/ms\n",
				b.id, n, ms, opsms)
		}
	}
}
