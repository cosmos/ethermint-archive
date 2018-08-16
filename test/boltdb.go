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
}

func (bit *BoltIterator) Domain() (start []byte, end []byte) {
	return bit.start, bit.end
}

func (bit *BoltIterator) Valid() bool {
	if bit.reverse {
		return bytes.Compare(bit.start, bit.end) > 0
	} else {
		return bytes.Compare(bit.start, bit.end) < 0
	}
}

func (bit *BoltIterator) Next() {
}

func (bit *BoltIterator) Key() (key []byte) {
	return nil
}

func (bit *BoltIterator) Value() (value []byte) {
	return nil
}

func (bit *BoltIterator) Close() {
	bit.tx.Rollback()
}

// Implements dbm.Batch
type BoltBatch struct {
}

func (bb *BoltBatch) Set(key, value []byte) {
}

func (bb *BoltBatch) Delete(key []byte) {
}

func (bb *BoltBatch) Write() {
}

func (bb *BoltBatch) WriteSync() {
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

func (bdb *BoltDb) Get([]byte) []byte {
	return nil
}

func (bdb *BoltDb) Has(key []byte) bool {
	return true
}

func (bdb *BoltDb) Set([]byte, []byte) {
}

func (bdb *BoltDb) SetSync([]byte, []byte) {
}

func (bdb *BoltDb) Delete([]byte) {
}

func (bdb *BoltDb) DeleteSync([]byte) {
}

func (bdb *BoltDb) Iterator(start, end []byte) dbm.Iterator {
	tx, err := bdb.db.Begin(false /*writeable*/)
	if err != nil {
		panic(err)
	}
	b := tx.Bucket(bucket)
	cursor := b.Cursor()
	return &BoltIterator{tx: tx, cursor: cursor, start: start, end: end, reverse: false}
}

func (bdb *BoltDb) ReverseIterator(start, end []byte) dbm.Iterator {
	tx, err := bdb.db.Begin(false /*writeable*/)
	if err != nil {
		panic(err)
	}
	b := tx.Bucket(bucket)
	cursor := b.Cursor()
	return &BoltIterator{tx: tx, cursor: cursor, start: start, end: end, reverse: true}
}

func (bdb *BoltDb) Close() {
	bdb.db.Close()
}

func (bdb *BoltDb) NewBatch() dbm.Batch {
	return &BoltBatch{}
}

func (bdb *BoltDb) Print() {
}

func (bdb *BoltDb) Stats() map[string]string {
	return make(map[string]string)
}
