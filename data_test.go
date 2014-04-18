package main

import (
	"bytes"
	"testing"
	"time"
)

func TestRandomInt(t *testing.T) {
	seed := int64(99)

	rnd1 := NewRandom(seed)
	rnd2 := NewRandom(seed)

	for i := 0; i < 100; i++ {
		x := i
		y := i + 100

		v1 := rnd1.Int(x, y)
		v2 := rnd2.Int(x, y)

		// given the same seed, two Random should produce
		// the same values
		if v1 != v2 {
			t.Errorf("rnd1 and rnd2 produced different values for (%d, %d): %d vs %d",
				x, y, v1, v2)
		}

		// v1 should be greater than or equal to x
		if v1 < x {
			t.Errorf("outside min range: (%d, %d): %d", x, y, v1)
		}

		// v1 should be less than or equal to y
		if v1 > y {
			t.Errorf("outside max range: (%d, %d): %d", x, y, v1)
		}
	}

	x, y := 100, 100
	if n := rnd1.Int(x, y); n != 100 {
		t.Errorf("unexpected result for (%d, %d): wanted %d, got %d",
			x, y, x, n)
	}
}

func TestRandomWriteSend(t *testing.T) {

	seed := int64(99)

	rnd1 := NewRandom(seed)
	buf := &bytes.Buffer{}

	for i := 0; i < 10; i++ {

		n := i
		b0 := i + 10
		b1 := i + 100
		k0 := i + 10
		k1 := i + 100
		v0 := i + 10
		v1 := i + 100

		err := rnd1.Write(buf, n, b0, b1, k0, k1, v0, v1)
		if err != nil {
			t.Error(i, err)
			continue
		}

		ch := make(chan []*Row, n)
		d0 := 10 * time.Millisecond
		d1 := 2 * d0

		go func(ch chan []*Row, d0, d1 time.Duration, n, b0, b1, k0, k1, v0, v1 int) {
			t0 := time.Now()
			i := 0
			for rows := range ch {
				t1 := time.Now()
				delay := t1.Sub(t0)
				t0 = t1

				if i > 0 {
					if delay.Nanoseconds() < d0.Nanoseconds() {
						t.Errorf("rows[%d] arrived after %s, expected it to be >= %s",
							i, delay, d0)
					}
					if n := delay.Nanoseconds() - d1.Nanoseconds(); n > 1e6 {
						t.Errorf("rows[%d] arrived after %s, expected it to be <= %s",
							i, delay, d1)
					}
				}

				var j int
				var row *Row
				for j, row = range rows {
					if x := len(row.Key.b); x < k0 {
						t.Errorf("rows[%d][%d].Key.b was %d, expected it to be >= %d",
							i, j, x, k0)
					}

					if x := len(row.Key.b); x > k1 {
						t.Errorf("rows[%d][%d].Key.b was %d, expected it to be <= %d",
							i, j, x, k1)
					}

					if x := len(row.Value.b); x < v0 {
						t.Errorf("rows[%d][%d].Value.b was %d, expected it to be >= %d",
							i, j, x, v0)
					}

					if x := len(row.Value.b); x > v1 {
						t.Errorf("rows[%d][%d].Value.b was %d, expected it to be <= %d",
							i, j, x, v1)
					}
				}

				if j < b0 {
					t.Errorf("rows[%d] contained %d items, expected it to be >= %d",
						i, j, b0)
				}

				if j > b1 {
					t.Errorf("rows[%d] contained %d items, expected it to be <= %d",
						i, j, b1)
				}

				i++
			}

			if i != n {
				t.Errorf("expected %d row sets, got %d", n, i)
			}

		}(ch, d0, d1, n, b0, b1, k0, k1, v0, v1)

		rnd1.Send(ch, buf, d0, d1)
		close(ch)

		buf.Reset()
	}
}
