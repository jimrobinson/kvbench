package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type LevelDBCollection struct {
	sync.WaitGroup
	db        *leveldb.DB
	cleanupFn func()
}

func NewLevelDBCollection(path string) (c Collection, err error) {
	ldb := &LevelDBCollection{}

	if path == "" {
		path, err = ioutil.TempDir("", "blockstatus.")
		if err != nil {
			err = fmt.Errorf("unable to create temporary directory for datastore: %v", err)
			return nil, err
		}
		ldb.cleanupFn = func() {
			os.RemoveAll(path)
		}
	}

	ldb.db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		err = fmt.Errorf("unable to open %s: %v", path, err)
		return
	}

	return ldb, err
}

func (c *LevelDBCollection) Close(force bool) (err error) {
	if !force {
		c.Wait()
	}
	err = c.db.Close()
	if c.cleanupFn != nil {
		c.cleanupFn()
	}
	return
}

func (c *LevelDBCollection) Rows() (ch chan Row) {
	ch = make(chan Row, 1000)

	go func(ch chan Row) {
		defer close(ch)

		iter := c.db.NewIterator(nil, nil)
		for iter.Next() {
			row := Row{}

			row.Key, row.Err = DecodeRowKey(iter.Key())
			if row.Err != nil {
				ch <- row
				continue
			}

			row.Value, row.Err = DecodeRowValue(iter.Value())
			if row.Err != nil {
				ch <- row
				continue
			}

			ch <- row
		}
	}(ch)

	return ch
}

func (c *LevelDBCollection) Set(rows []*Row) (err error) {
	batch := &leveldb.Batch{}
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

		batch.Put(bk, bv)
	}
	err = c.db.Write(batch, nil)
	return
}

func (c *LevelDBCollection) Delete(k RowKey) (err error) {
	var bk []byte
	bk, err = k.Bytes()
	if err != nil {
		return
	}

	err = c.db.Delete(bk, nil)
	return
}

func (c *LevelDBCollection) Timing() (int, time.Duration) {
	t0 := time.Now()
	n := 0
	for range c.Rows() {
		n++
	}
	t1 := time.Now()
	return n, t1.Sub(t0)
}
