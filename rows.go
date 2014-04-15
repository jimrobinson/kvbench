package main

type Row struct {
	Key   RowKey
	Value *RowValue
	Err   error
}

type RowKey struct {
	b []byte
}

func DecodeRowKey(b []byte) (rk RowKey, err error) {
	rk = RowKey{b: make([]byte, len(b))}
	for i, v := range b { // simulate something expensive
		rk.b[i] = v
	}
	return rk, nil
}

func (rk RowKey) Bytes() (b []byte, err error) {
	b = make([]byte, len(rk.b))
	for i, v := range rk.b { // simulate something expensive
		b[i] = v
	}
	return b, nil
}

type RowValue struct {
	b []byte
}

func (v *RowValue) Merge(m *RowValue) {
	return
}

func NewRowValue() *RowValue {
	return &RowValue{}
}

func DecodeRowValue(b []byte) (rv *RowValue, err error) {
	rv = &RowValue{b: make([]byte, len(b))}
	for i, v := range b { // simulate something expensive
		rv.b[i] = v
	}
	return rv, nil
}

func (rv *RowValue) Bytes() (b []byte, err error) {
	b = make([]byte, len(rv.b))
	for i, v := range rv.b { // simulate something expensive
		b[i] = v
	}
	return b, nil
}
