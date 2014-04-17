package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var testRows = make([]*Row, 0, 254)

func init() {
	for i := 0; i < cap(testRows); i++ {
		testRows = append(testRows, &Row{
			Key: RowKey{
				b: []byte{byte(i)},
			},
			Value: &RowValue{
				b: []byte{byte(i + 1)},
			},
		})
	}
}

func TestCollectionLevelDB(t *testing.T) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	c, err := NewLevelDBCollection(path)
	if err != nil {
		t.Fatal(err)
	}

	testCollectionSet(t, "leveldb", c)
	testCollectionRows(t, "leveldb", c)
	testCollectionDelete(t, "leveldb", c)

	err = c.Close(true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCollectionBolt(t *testing.T) {
	fh, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(fh.Name())
	defer fh.Close()

	c, err := NewBoltCollection(fh.Name())
	if err != nil {
		t.Fatal(err)
	}

	testCollectionSet(t, "leveldb", c)
	testCollectionRows(t, "leveldb", c)
	testCollectionDelete(t, "leveldb", c)

	err = c.Close(true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCollectionKV(t *testing.T) {
	path, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	fh, err := ioutil.TempFile(path, "")
	if err != nil {
		t.Fatal(err)
	}
	err = fh.Close()
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.Remove(fh.Name())
	if err != nil {
		t.Fatal(err)
	}

	c, err := NewKVCollection(fh.Name())
	if err != nil {
		t.Fatal(err)
	}

	testCollectionSet(t, "leveldb", c)
	testCollectionRows(t, "leveldb", c)
	testCollectionDelete(t, "leveldb", c)

	err = c.Close(true)
	if err != nil {
		t.Fatal(err)
	}
}

func testCollectionSet(t *testing.T, id string, c Collection) {
	err := c.Set(testRows)
	if err != nil {
		t.Error(id, "Set", err)
	}
}

func testCollectionRows(t *testing.T, id string, c Collection) {
	ch := c.Rows()
	arrived := make([]Row, 0, len(testRows))
	for rows := range ch {
		arrived = append(arrived, rows)
	}
	if len(arrived) != len(testRows) {
		t.Errorf("%s Rows failed: expected %d row sets, got %d",
			id, len(testRows), len(arrived))
	}
	for i, v := range arrived {
		if bytes.Compare(v.Key.b, testRows[i].Key.b) != 0 {
			t.Errorf("%s mismatch on row set %d: key mismatch:\narrived = %v\nexpected = %v",
				id, i, v.Key.b, testRows[i].Key.b)
		}
		if bytes.Compare(v.Value.b, testRows[i].Value.b) != 0 {
			t.Errorf("%s mismatch on row set %d: value mismatch:\narrived = %v\nexpected = %v",
				id, i, v.Value.b, testRows[i].Value.b)
		}
	}
}

func testCollectionDelete(t *testing.T, id string, c Collection) {
	for i, row := range testRows {
		err := c.Delete(row.Key)
		if err != nil {
			t.Errorf(id, "Delete:", err)
			continue
		}

		for v := range c.Rows() {
			if bytes.Compare(v.Key.b, row.Key.b) == 0 {
				t.Errorf("%s Rows returned key %v (#%d) after Delete", id, row.Key.b, i)
			}
		}
	}
}
