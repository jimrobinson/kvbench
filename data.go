package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"
)

// Random provides a source for pseudo-random
// numbers and bytes.
type Random struct {
	sync.Mutex
	r *rand.Rand
}

// NewRandom returns an initialized Random seeded
// with the provided seed.
func NewRandom(seed int64) *Random {
	return &Random{
		r: rand.New(rand.NewSource(seed)),
	}
}

// Int returns a pseudo-random int between min
// and max.  If max-min is < 1, the result will be min.
func (rnd *Random) Int(min, max int) int {
	diff := max - min
	if diff < 1 {
		return min
	}
	rnd.Lock()
	defer rnd.Unlock()
	return min + rnd.r.Intn(max-min)
}

// randBytes writes n pseudo random bytes,
// in the range 0 through 255, to w.
func (rnd *Random) Bytes(w io.Writer, n int) error {
	rnd.Lock()
	defer rnd.Unlock()
	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(rnd.r.Intn(255))
	}
	return binary.Write(w, binary.LittleEndian, buf)
}

// Write writes pseudo-random benchmark data
// to w based on the parameters provided
// - rnd provides a source of pseudo-random data
// - n indicates the total number of record blocks to write
// - b0 indicates the minimum number of records per block
// - b1 indicates the maximum number of records per block
// - k0 indicates the minimum key length
// - k1 indicates the maximum key length
// - v0 indicates the minimum value length
// - v1 indicates the maximum value length
func (rnd *Random) Write(w io.Writer, n, b0, b1, k0, k1, v0, v1 int) (err error) {
	for i := 0; i < n; i++ {
		x := rnd.Int(b0, b1)
		if err = binary.Write(w, binary.LittleEndian, int64(x)); err != nil {
			return
		}

		for j := 0; j < x; j++ {
			k := rnd.Int(k0, k1)
			if err = binary.Write(w, binary.LittleEndian, int64(k)); err != nil {
				return
			}
			if err = rnd.Bytes(w, k); err != nil {
				return
			}

			v := rnd.Int(v0, v1)
			if err = binary.Write(w, binary.LittleEndian, int64(v)); err != nil {
				return
			}
			if err = rnd.Bytes(w, v); err != nil {
				return
			}
		}
	}
	return
}

// Send reads record blocks from r and sends them to ch.
// At least d0 duration will pass in-between sends on ch, and
// an attempt will be made to send within d1 duration.  Note
// that d1 is not guaranteed, as there are external factors
// that will affect how quickly each row can be prepared.
func (rnd *Random) Send(ch chan []*Row, r io.Reader, d0, d1 time.Duration) (err error) {
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
			ns := int64(rnd.Int(int(d0), int(d1))) - t1.Nanoseconds()
			if ns > 0 {
				time.Sleep(time.Duration(ns))
			}
		}

		t0 = time.Now()
		ch <- rows
	}
}
