package main

import (
	"bytes"
	"os"
	"path"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/boltdb/bolt"
)

// Implements dbm.DB
type BoltDb struct {
	db *bolt.DB	
}

// Implements dbm.Iterator
type BoltIterator struct {
	tx *bolt.Tx
	cursor *bolt.Cursor
	start, end []byte
	reverse bool
	valid bool
	key, value []byte // Current key and value
}

func (bit *BoltIterator) Domain() (start []byte, end []byte) {
	return bit.start, bit.end
}

func (bit *BoltIterator) Valid() bool {
	return bit.valid
}

func (bit *BoltIterator) Next() {
	if !bit.valid {
		return
	}
	var k, v []byte
	if bit.reverse {
		k, v = bit.cursor.Prev()
		if k == nil || (bit.end != nil && bytes.Compare(k, bit.end) <= 0) {
			bit.tx.Rollback()
			bit.valid = false
		} else {
			bit.key = make([]byte, len(k))
			copy(bit.key, k)
			bit.value = make([]byte, len(v))
			copy(bit.value, v)
		}
	} else {
		k, v = bit.cursor.Next()
		if k == nil || (bit.end != nil && bytes.Compare(k, bit.end) >= 0) {
			bit.tx.Rollback()
			bit.valid = false
		} else {
			bit.key = make([]byte, len(k))
			copy(bit.key, k)
			bit.value = make([]byte, len(v))
			copy(bit.value, v)
		}
	}
}

func (bit *BoltIterator) Key() (key []byte) {
	return bit.key
}

func (bit *BoltIterator) Value() (value []byte) {
	return bit.value
}

func (bit *BoltIterator) Close() {
	if bit.valid {
		bit.tx.Rollback()
	}
}

// Implements dbm.Batch
type BoltBatch struct {
	db *bolt.DB
	keys, values [][]byte
}

func (bb *BoltBatch) Set(key, value []byte) {
	bb.keys = append(bb.keys, key)
	bb.values = append(bb.values, value)
}

func (bb *BoltBatch) Delete(key []byte) {
	bb.keys = append(bb.keys, key)
	bb.values = append(bb.values, nil)
}

func (bb *BoltBatch) Write() {
	if err := bb.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		for i, k := range bb.keys {
			v := bb.values[i]
			if v == nil {
				if e := b.Delete(k); e != nil {
					return e
				}
			} else {
				if e := b.Put(k, v); e != nil {
					return e
				}
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	bb.keys = nil
	bb.values = nil
}

func (bb *BoltBatch) WriteSync() {
	bb.Write()
}

var bucket = []byte("B")

// Open BoltDB database and wraps it into BoltDb struct
func OpenBoltDb(filename string) (*BoltDb, error) {
	// Create necessary directories
	if err := os.MkdirAll(path.Dir(filename), os.ModePerm); err != nil {
		return nil, err
	}
	db, err := bolt.Open(filename, 0600, &bolt.Options{})
	if err != nil {
		return nil, err
	}
	// Create the bucket if does not exist
	if err := db.Update(func (tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(bucket)
		return e
	}); err != nil {
		db.Close()
		return nil, err
	}
	return &BoltDb{
		db: db,
	}, nil
}

func (bdb *BoltDb) Get(key []byte) []byte {
	var value []byte
	bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		v := b.Get(key)
		if v != nil {
			value = make([]byte, len(v))
			copy(value, v)
		}
		return nil
	})
	return value
}

func (bdb *BoltDb) Has(key []byte) bool {
	var has bool
	bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		v := b.Get(key)
		if v != nil {
			has = true
		}
		return nil
	})
	return has
}

func (bdb *BoltDb) Set(key []byte, value []byte) {
	if err := bdb.db.Update(func (tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Put(key, value)
	}); err != nil {
		panic(err)
	}
}

func (bdb *BoltDb) SetSync(key []byte, value []byte) {
	bdb.Set(key, value)
}

func (bdb *BoltDb) Delete(key []byte) {
	if err := bdb.db.Update(func (tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Delete(key)
	}); err != nil {
		panic(err)
	}	
}

func (bdb *BoltDb) DeleteSync(key []byte) {
	bdb.Delete(key)
}

func createIterator(reverse bool, start, end []byte, db *bolt.DB) *BoltIterator {
	var valid bool
	if reverse {
		valid = bytes.Compare(start, end) > 0
	} else {
		valid = bytes.Compare(start, end) < 0
	}
	if !valid {
		return &BoltIterator{tx: nil, cursor: nil, start: start, end: end, reverse: reverse, valid: false}
	}
	tx, err := db.Begin(false /*writeable*/)
	if err != nil {
		panic(err)
	}
	b := tx.Bucket(bucket)
	cursor := b.Cursor()
	return &BoltIterator{tx: tx, cursor: cursor, start: start, end: end, reverse: reverse, valid: true}
}

func (bdb *BoltDb) Iterator(start, end []byte) dbm.Iterator {
	it := createIterator(false /*reverse*/, start, end, bdb.db)
	if !it.valid {
		return it
	}
	var k, v []byte
	if start == nil {
		k, v = it.cursor.First()
	} else {
		k, v = it.cursor.Seek(start)
	}
	if k == nil || (end != nil && bytes.Compare(k, end) >= 0) {
		it.tx.Rollback()
		it.valid = false
	} else {
		it.key = make([]byte, len(k))
		copy(it.key, k)
		it.value = make([]byte, len(v))
		copy(it.value, v)
	}
	return it
}

func (bdb *BoltDb) ReverseIterator(start, end []byte) dbm.Iterator {
	it := createIterator(true /*reverse*/, start, end, bdb.db)
	if !it.valid {
		return it
	}
	var k, v []byte
	if start == nil {
		k, v = it.cursor.Last()
	} else {
		k, v = it.cursor.Seek(start)
		// Move backwards if needed
		for k != nil && bytes.Compare(k, start) > 0 {
			k, v = it.cursor.Prev()
		}
	}
	if k == nil || (end != nil && bytes.Compare(k, end) <= 0) {
		it.tx.Rollback()
		it.valid = false
	} else {
		it.key = make([]byte, len(k))
		copy(it.key, k)
		it.value = make([]byte, len(v))
		copy(it.value, v)
	}
	return it
}

func (bdb *BoltDb) Close() {
	bdb.db.Close()
}

func (bdb *BoltDb) NewBatch() dbm.Batch {
	return &BoltBatch{db: bdb.db}
}

func (bdb *BoltDb) Print() {
}

func (bdb *BoltDb) Stats() map[string]string {
	return make(map[string]string)
}
