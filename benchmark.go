package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Benchmark struct {
	id string
	c  Collection
	mu *sync.RWMutex
	wg *sync.WaitGroup
}

func NewBenchmark(id string, path string) (b *Benchmark, err error) {
	b = &Benchmark{
		id: id,
		wg: &sync.WaitGroup{},
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

// Run launches two goroutines, one to read row sets
// from ch, writing those rows to the collection, and another
// to wake up every dur interval and poll the collection for
// number of records and the time it took to iterate over
// those records.
func (b *Benchmark) Run(ch chan []*Row, dur time.Duration) {
	done := make(chan bool)

	go func(b *Benchmark, ch chan []*Row) {
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

		done <- true
		if n > 0 {
			log.Printf("%d row sets arrived at an average inter-arrival rate of %s",
				n, time.Duration(ns/n))
		}
	}(b, ch)

	go func(b *Benchmark, done <-chan bool) {
		defer b.wg.Done()
		for {
			select {
			case _ = <-done:
				if b.mu != nil {
					b.mu.RLock()
				}
				n, t := b.c.Timing()
				if b.mu != nil {
					b.mu.RUnlock()
				}
				log.Printf("%s\t%d\t%d ms\n",
					b.id, n, t.Nanoseconds()/1e6)
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
				log.Printf("timing %s: %d in %d ms\n", b.id, n, t.Nanoseconds()/1e6)
			}
		}
	}(b, done)
}
