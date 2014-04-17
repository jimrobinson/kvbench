package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"sync"
	"time"
)

var bucketId = []byte("values")

type BoltCollection struct {
	sync.WaitGroup
	db *bolt.DB
}

func NewBoltCollection(path string) (c Collection, err error) {
	boltc := &BoltCollection{}

	boltc.db, err = bolt.Open(path, 0644)
	if err != nil {
		err = fmt.Errorf("unable to open %s: %v", path, err)
		return
	}
	err = boltc.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketId)
		return err
	})
	return boltc, err
}

func (c *BoltCollection) Close(force bool) (err error) {
	if !force {
		c.Wait()
	}
	err = c.db.Close()
	return
}

func (c *BoltCollection) Rows() (ch chan Row) {
	ch = make(chan Row, 1000)

	go func(ch chan Row) {
		defer close(ch)

		c.db.View(
			func(tx *bolt.Tx) error {

				cursor := tx.Bucket(bucketId).Cursor()

				for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
					row := Row{}

					row.Key, row.Err = DecodeRowKey(k)
					if row.Err != nil {
						ch <- row
						continue
					}

					row.Value, row.Err = DecodeRowValue(v)
					if row.Err != nil {
						ch <- row
						continue
					}

					ch <- row
				}

				return nil
			})

	}(ch)

	return ch
}

func (c *BoltCollection) Set(rows []*Row) (err error) {
	return c.db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketId)

			for _, row := range rows {
				var bk, bv []byte

				bk, err = row.Key.Bytes()
				if err != nil {
					return err
				}

				bv, err = row.Value.Bytes()
				if err != nil {
					return err
				}

				err = b.Put(bk, bv)
				if err != nil {
					return err
				}
			}

			return nil
		})
}

func (c *BoltCollection) Delete(k RowKey) (err error) {
	return c.db.Update(
		func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketId)

			var bk []byte
			bk, err = k.Bytes()
			if err != nil {
				return err
			}

			return b.Delete(bk)
		})
}

func (c *BoltCollection) Timing() (int, time.Duration) {
	t0 := time.Now()
	n := 0
	for _ = range c.Rows() {
		n++
	}
	t1 := time.Now()
	return n, t1.Sub(t0)
}
