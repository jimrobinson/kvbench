package main

import (
	"bytes"
	"testing"
)

var testKey [][]byte
var testValue [][]byte

func init() {
	rnd := NewRandom(99)
	buf := &bytes.Buffer{}
	for i := 1; i <= 4; i++ {
		buf.Reset()
		err := rnd.Bytes(buf, 32*i)
		if err != nil {
			panic(err)
		}

		testKey = append(testKey, buf.Bytes())

		buf.Reset()
		err = rnd.Bytes(buf, 128*i)
		if err != nil {
			panic(err)
		}

		testValue = append(testValue, buf.Bytes())
	}
}

func TestRowKeyDecodeEncode(t *testing.T) {
	if len(testKey) != len(testValue) {
		t.Fatalf("len(testKey) [%d] != len(testValue) [%d]", len(testKey), len(testValue))
	}
	for i := 0; i < len(testKey); i++ {
		rk, err := DecodeRowKey(testKey[i])
		if err != nil {
			t.Error(i, err)
			continue
		}

		bk, err := rk.Bytes()
		if err != nil {
			t.Error(i, err)
			continue
		}

		if bytes.Compare(testKey[i], bk) != 0 {
			t.Errorf("%d: key and re-encoded key produced different values:%v (original)\n%v (re-encoded)",
				i, testKey[i], bk)
		}
	}
}

func TestRowValueDecodeEncode(t *testing.T) {
	if len(testKey) != len(testValue) {
		t.Fatalf("len(testKey) [%d] != len(testValue) [%d]", len(testKey), len(testValue))
	}
	for i := 0; i < len(testValue); i++ {
		rv, err := DecodeRowValue(testValue[i])
		if err != nil {
			t.Error(i, err)
			continue
		}

		bv, err := rv.Bytes()
		if err != nil {
			t.Error(i, err)
			continue
		}

		if bytes.Compare(testValue[i], bv) != 0 {
			t.Errorf("%d: key and re-encoded value produced different values:%v (original)\n%v (re-encoded)",
				i, testValue[i], bv)
		}
	}
}

func BenchmarkRowKeyEncode32(b *testing.B) { benchRowKeyEncode(b, testKey[0]) }
func BenchmarkRowKeyDecode32(b *testing.B) { benchRowKeyDecode(b, testKey[0]) }

func BenchmarkRowKeyEncode64(b *testing.B) { benchRowKeyEncode(b, testKey[1]) }
func BenchmarkRowKeyDecode64(b *testing.B) { benchRowKeyDecode(b, testKey[1]) }

func BenchmarkRowKeyEncode128(b *testing.B) { benchRowKeyEncode(b, testKey[2]) }
func BenchmarkRowKeyDecode128(b *testing.B) { benchRowKeyDecode(b, testKey[2]) }

func BenchmarkRowKeyEncode256(b *testing.B) { benchRowKeyEncode(b, testKey[3]) }
func BenchmarkRowKeyDecode256(b *testing.B) { benchRowKeyDecode(b, testKey[3]) }

func benchRowKeyDecode(b *testing.B, v []byte) {
	for i := 0; i < b.N; i++ {
		_, err := DecodeRowKey(v)
		if err != nil {
			b.Error(err)
		}
	}
}

func benchRowKeyEncode(b *testing.B, v []byte) {
	rk, err := DecodeRowKey(v)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = rk.Bytes()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkRowValueEncode128(b *testing.B) { benchRowValueEncode(b, testValue[0]) }
func BenchmarkRowValueDecode128(b *testing.B) { benchRowValueDecode(b, testValue[0]) }

func BenchmarkRowValueEncode256(b *testing.B) { benchRowValueEncode(b, testValue[1]) }
func BenchmarkRowValueDecode256(b *testing.B) { benchRowValueDecode(b, testValue[1]) }

func BenchmarkRowValueEncode512(b *testing.B) { benchRowValueEncode(b, testValue[2]) }
func BenchmarkRowValueDecode512(b *testing.B) { benchRowValueDecode(b, testValue[2]) }

func BenchmarkRowValueEncode1024(b *testing.B) { benchRowValueEncode(b, testValue[3]) }
func BenchmarkRowValueDecode1024(b *testing.B) { benchRowValueDecode(b, testValue[3]) }

func benchRowValueDecode(b *testing.B, v []byte) {
	for i := 0; i < b.N; i++ {
		_, err := DecodeRowValue(v)
		if err != nil {
			b.Error(err)
		}
	}
}

func benchRowValueEncode(b *testing.B, v []byte) {
	rv, err := DecodeRowValue(v)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = rv.Bytes()
		if err != nil {
			b.Error(err)
		}
	}
}
