package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"runtime"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	db "github.com/tendermint/go-db"
	merkle "github.com/tendermint/go-merkle"
)

const reportInterval = 100

var (
	initSize  = kingpin.Flag("initsize", "Initial DB Size").Short('i').Default("100000").Int()
	keySize   = kingpin.Flag("keysize", "Length of keys (in bytes)").Short('k').Default("16").Int()
	dataSize  = kingpin.Flag("valuesize", "Length of values (in bytes)").Short('v').Default("100").Int()
	blockSize = kingpin.Flag("blocksize", "Number of Txs per block").Short('b').Default("200").Int()
	dbType    = kingpin.Flag("db", "type of backing db").Short('d').Default("goleveldb").String()
)

// blatently copied from benchmarks/bench_test.go
func randBytes(length int) []byte {
	key := make([]byte, length)
	// math.rand.Read always returns err=nil
	rand.Read(key)
	return key
}

// blatently copied from benchmarks/bench_test.go
func prepareTree(db db.DB, size, keyLen, dataLen int) (merkle.Tree, [][]byte) {
	t := merkle.NewIAVLTree(size, db)
	keys := make([][]byte, size)

	for i := 0; i < size; i++ {
		key := randBytes(keyLen)
		t.Set(key, randBytes(dataLen))
		keys[i] = key
	}
	t.Hash()
	t.Save()
	runtime.GC()
	return t, keys
}

func runBlock(t merkle.Tree, dataLen, blockSize int, keys [][]byte) merkle.Tree {
	l := int32(len(keys))

	real := t.Copy()
	check := t.Copy()

	for j := 0; j < blockSize; j++ {
		// always update to avoid changing size
		key := keys[rand.Int31n(l)]
		data := randBytes(dataLen)

		// perform query and write on check and then real
		check.Get(key)
		check.Set(key, data)
		real.Get(key)
		real.Set(key, data)
	}

	// at the end of a block, move it all along....
	real.Hash()
	real.Save()
	return real
}

func loopForever(t merkle.Tree, dataLen, blockSize int, keys [][]byte, initMB float64) {
	for {
		start := time.Now()
		for i := 0; i < reportInterval; i++ {
			t = runBlock(t, dataLen, blockSize, keys)
		}
		// now report
		end := time.Now()
		delta := end.Sub(start)
		timing := delta.Seconds() / reportInterval
		usedMB := memUseMB() - initMB
		fmt.Printf("%s: blocks of %d tx: %0.3f s/block, using %0.2f MB\n",
			end.Format("Jan 2 15:04:05"), blockSize, timing, usedMB)
	}
}

// returns number of MB in use
func memUseMB() float64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	asize := mem.Alloc
	mb := float64(asize) / 1000000
	return mb
}

func main() {
	kingpin.Parse()

	tmpDir, err := ioutil.TempDir("", "loadtest-")
	if err != nil {
		kingpin.Fatalf("Cannot create temp dir: %s", err)
	}

	initMB := memUseMB()
	start := time.Now()

	fmt.Printf("Preparing DB (%s with %d keys)...\n", *dbType, *initSize)
	d := db.NewDB("loadtest", *dbType, tmpDir)
	tree, keys := prepareTree(d, *initSize, *keySize, *dataSize)

	delta := time.Now().Sub(start)
	fmt.Printf("Initialization took %0.3f s, used %0.2f MB\n",
		delta.Seconds(), memUseMB()-initMB)
	fmt.Printf("Keysize: %d, Datasize: %d\n", *keySize, *dataSize)

	fmt.Printf("Starting loadtest (blocks of %d tx)...\n", *blockSize)
	loopForever(tree, *dataSize, *blockSize, keys, initMB)
}
