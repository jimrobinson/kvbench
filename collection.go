package main

import (
	"time"
)

type Collection interface {
	Close(force bool) (err error)
	Delete(k RowKey) (err error)
	Rows() (ch chan Row)
	Set(rows []*Row) (err error)
	Timing() (int, time.Duration)
}
