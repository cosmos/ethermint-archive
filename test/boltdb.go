package main

import (
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
}

func (bit *BoltIterator) Domain() (start []byte, end []byte) {
	return nil, nil
}

func (bit *BoltIterator) Valid() bool {
	return false
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
	return &BoltIterator{}
}

func (bdb *BoltDb) ReverseIterator(start, end []byte) dbm.Iterator {
	return nil
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
