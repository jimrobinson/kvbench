package main

import (
	"sync"
	"time"
)

type NoopCollection struct {
	sync.RWMutex
	sync.WaitGroup
	n int
}

func NewNoopCollection() (c Collection, err error) {
	return &NoopCollection{}, nil
}

func (c *NoopCollection) Close(force bool) (err error) {
	return nil
}

func (c *NoopCollection) Rows() (ch chan Row) {
	ch = make(chan Row, 1000)

	go func(ch chan Row) {
		defer close(ch)

		c.RLock()
		n := c.n
		c.RUnlock()

		row := Row{Key: RowKey{}, Value: &RowValue{}}
		for i := 0; i < n; i++ {
			ch <- row
		}
	}(ch)

	return ch
}

func (c *NoopCollection) Set(rows []*Row) (err error) {
	c.Lock()
	c.n += len(rows)
	c.Unlock()
	return nil
}

func (c *NoopCollection) Delete(k RowKey) (err error) {
	c.Lock()
	if c.n > 0 {
		c.n = c.n - 1
	}
	c.Unlock()
	return nil
}

func (c *NoopCollection) Timing() (int, time.Duration) {
	t0 := time.Now()
	n := 0
	for range c.Rows() {
		n++
	}
	t1 := time.Now()
	return n, t1.Sub(t0)
}
