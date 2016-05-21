package gorocks

import (
	"bytes"
	"testing"
)

func TestWriteBatchIterator(t *testing.T) {
	wb := NewWriteBatch()
	defer wb.Close()

	it := wb.NewIterator()
	if it.Next() {
		t.Fatal("Next on empty iterator")
	}

	wb.Clear()

	key := []byte("key")
	value := []byte("value")
	wb.Put(key, value)

	it = wb.NewIterator()
	it.Next()
	record := it.Record()

	if bytes.Compare(key, record.Key) != 0 {
		t.Fatal("invalid key")
	}

	if bytes.Compare(value, record.Value) != 0 {
		t.Fatal("invalid value")
	}

	if it.Error() != nil {
		t.Fatal("Error on iterator")
	}

	wb.Clear()
	for i := 0; i < 512; i++ {
		key := make([]byte, i)
		for j := 0; j < 512; j++ {
			value := make([]byte, j)
			wb.Put(key, value)
		}
	}
	it = wb.NewIterator()
	var count int
	var kb, vb int
	for count = 0; it.Next(); count++ {
		rec := it.Record()
		if rec.Type != RecordTypeValue {
			t.Fatal("expected value record")
		}
		kb += len(rec.Key)
		vb += len(rec.Value)
	}

	if count != 512*512 {
		t.Fatal("records missing")
	}

	const n = 512 * (511 * 512 / 2)
	if kb != n {
		t.Fatalf("key bytes missing: expected %v, got %v", n, kb)
	}
	if vb != n {
		t.Fatalf("value bytes missing: expected %v, got %v", n, vb)
	}
}
