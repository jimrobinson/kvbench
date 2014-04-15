package main

import (
	"fmt"
	"github.com/cznic/kv"
	"io"
	"sync"
	"time"
)

type KVCollection struct {
	sync.WaitGroup
	db *kv.DB
}

func NewKVCollection(path string) (c Collection, err error) {
	kvdb := &KVCollection{}

	opts := &kv.Options{}

	kvdb.db, err = kv.Create(path, opts)
	if err != nil {
		err = fmt.Errorf("unable to open %s: %v", path, err)
		return
	}

	return kvdb, err
}

func (c *KVCollection) Close(force bool) (err error) {
	if !force {
		c.Wait()
	}
	err = c.db.Close()
	return
}

func (c *KVCollection) Rows() (ch chan Row) {
	ch = make(chan Row, 1000)

	go func(ch chan Row) {
		defer close(ch)

		enum, err := c.db.SeekFirst()
		if err != nil {
			if err != io.EOF {
				ch <- Row{Err: err}
			}
			return
		}

		for {
			kb, vb, err := enum.Next()
			if err != nil {
				if err != io.EOF {
					ch <- Row{Err: err}
				}
				return
			}

			row := Row{}

			row.Key, row.Err = DecodeRowKey(kb)
			if row.Err != nil {
				ch <- row
				continue
			}

			row.Value, row.Err = DecodeRowValue(vb)
			if row.Err != nil {
				ch <- row
				continue
			}

			ch <- row
		}
	}(ch)

	return ch
}

func (c *KVCollection) Set(rows []*Row) (err error) {
	for _, row := range rows {

		var bk, bv []byte

		bk, err = row.Key.Bytes()
		if err != nil {
			return
		}

		bv, err = row.Value.Bytes()
		if err != nil {
			return
		}

		err = c.db.Set(bk, bv)
		if err != nil {
			return
		}
	}
	return
}

func (c *KVCollection) Delete(k RowKey) (err error) {
	var bk []byte
	bk, err = k.Bytes()
	if err != nil {
		return
	}

	err = c.db.Delete(bk)
	return
}

func (c *KVCollection) Timing() (int, time.Duration) {
	t0 := time.Now()
	n := 0
	for _ = range c.Rows() {
		n++
	}
	t1 := time.Now()
	return n, t1.Sub(t0)
}
