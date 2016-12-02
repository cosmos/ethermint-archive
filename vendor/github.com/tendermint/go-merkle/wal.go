package merkle

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"

	auto "github.com/tendermint/go-autofile"
	. "github.com/tendermint/go-common"
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/go-wire"
)

/*
WAL messages are always appended to an AutoFile.  The AutoFile Group manages
rolling of WAL messages.

Actions to be performed are written as WAL messages, zero or more messages per
batch. Operational messages within a batch are idempotent.

Before the Merkle tree is modified at the underlying DB layer with Save(), new
nodes and orphaned nodes are written to the WAL.  The producer of these
messages writes messages like so:

> OpenBatch      height
> AddNode        key value
> AddNode        key value
> ...
> DelNode        key
> DelNode        key
> ...
> CloseBatch     height

Failures may happen at any time, so it is possible for some OpenBatch's to be
unmatched and repeated (restarted at the same height).

> OpenBatch    	 height
> ...
> OpenBatch      height
> ...
> CloseBatch     height

After a batch is closed, we process them.

> OpenBatch      height
> ...
> CloseBatch     height
> BatchStarted   height
> BatchEnded     height
> OpenBatch      height+1

It's possible for BatchStarted to also be unmatched and repeated.
The next batch should not open until the previous batch is closed.
*/

//--------------------------------------------------------------------------------
// types and functions for savings consensus messages

type WALMessage interface{}

var _ = wire.RegisterInterface(
	struct{ WALMessage }{},
	wire.ConcreteType{walMsgOpenBatch{}, 0x01},    // Beginning of msgs for batch
	wire.ConcreteType{walMsgCloseBatch{}, 0x02},   // End of msgs for batch
	wire.ConcreteType{walMsgAddNode{}, 0x03},      // Operation: Create a node
	wire.ConcreteType{walMsgDelNode{}, 0x04},      // Operation: Delete a node
	wire.ConcreteType{walMsgBatchStarted{}, 0x05}, // Processing of batch started
	wire.ConcreteType{walMsgBatchEnded{}, 0x06},   // Processing of batch ended
)

type walMsgOpenBatch struct {
	Height int
}

type walMsgCloseBatch struct {
	Height int
}

type walMsgAddNode struct {
	Key   []byte
	Value []byte
}

type walMsgDelNode struct {
	Key []byte
}

type walMsgBatchStarted struct {
	Height int
}

type walMsgBatchEnded struct {
	Height int
}

//----------------------------------------

const (
	walMarkerOpenBatch  = "#OB:"
	walMarkerCloseBatch = "#CB:"
	walMarkerBatchEnded = "#BE:"
)

type BatchStatus int

const (
	batchStatusPending = 1
	batchStatusOpened  = 2
	batchStatusClosed  = 3
	batchStatusStarted = 4
	batchStatusEnded   = 5
)

//--------------------------------------------------------------------------------

type WAL struct {
	BaseService

	group       *auto.Group
	db          dbm.DB
	batchHeight int
	batchStatus BatchStatus
	batchMsgs   []WALMessage
}

func NewWAL(walDir string, db dbm.DB) (*WAL, error) {
	group, err := auto.OpenGroup(walDir + "/wal")
	if err != nil {
		return nil, err
	}
	wal := &WAL{
		group: group,
		db:    db,
	}
	wal.BaseService = *NewBaseService(nil, "WAL", wal)
	return wal, nil
}

func (wal *WAL) OnStart() error {
	wal.doRecover()   // NOTE: must happen before wal.group.Start()
	wal.group.Start() // NOTE: group cleanup starts after .Start()
	return nil
}

func (wal *WAL) OnStop() {
	wal.BaseService.OnStop()
	wal.group.Stop()
	return
}

func (wal *WAL) doRecover() {
	closeBatchHeight := wal.findLastCloseBatchHeight()
	batchEndedHeight := wal.findLastBatchEndedHeight()
	if closeBatchHeight == batchEndedHeight {
		wal.batchHeight = closeBatchHeight + 1
		wal.batchStatus = batchStatusPending
		return
	}
	if closeBatchHeight == batchEndedHeight+1 {
		wal.batchHeight = closeBatchHeight
		wal.batchStatus = batchStatusClosed
		msgs := wal.getMsgsForBatch(closeBatchHeight)
		wal.processWALMessages(closeBatchHeight, msgs)
		return
	}
	panic(Fmt("Inconsistent closeBatch height %v vs batchEnded height %v", closeBatchHeight, batchEndedHeight))
}

func (wal *WAL) findLastCloseBatchHeight() int {
	match, found, err := wal.group.FindLast(walMarkerCloseBatch)
	if err != nil {
		panic(Fmt("Could not find last CloseBatch: %v", err))
	}
	if !found {
		return 0
	}
	height, err := strconv.Atoi(match[len(walMarkerCloseBatch):])
	if err != nil {
		panic(Fmt("Error parsing CloseBatch marker: %v", err))
	}
	return height
}

func (wal *WAL) findLastBatchEndedHeight() int {
	match, found, err := wal.group.FindLast(walMarkerBatchEnded)
	if err != nil {
		panic(Fmt("Could not find last BatchEnded: %v", err))
	}
	if !found {
		return 0
	}
	height, err := strconv.Atoi(match[len(walMarkerBatchEnded):])
	if err != nil {
		panic(Fmt("Error parsing BatchEnded marker: %v", err))
	}
	return height
}

func (wal *WAL) getMsgsForBatch(height int) []WALMessage {
	gr, found, err := wal.group.Search(walMarkerOpenBatch,
		auto.MakeSimpleSearchFunc(walMarkerOpenBatch, height),
	)
	if err != nil {
		panic(Fmt("Error searching for OpenBatch@%v: %v", height, err))
	}
	if !found {
		// No OpenBatch has been found, so presumably this is a new WAL.
		gr, err = wal.group.NewReader(0)
		if err != nil {
			panic(Fmt("Error loading WAL file index 0: %v", err))
		}
	}
	defer gr.Close()

	// Read lines from gr for batch @ height
	var batchMsgs = []WALMessage{}
	for {
		line, err := gr.ReadLine()
		if err != nil {
			panic(Fmt("Error reading line from WAL: %v", err))
		}

		// If line is a marker, ignore it.
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse message
		msgBytes, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			panic(Fmt("Error parsing line from WAL: %v", err))
		}
		var msg WALMessage
		err = wire.ReadBinaryBytes(msgBytes, &msg)
		if err != nil {
			panic(Fmt("Error parsing line from WAL: %v", err))
		}

		switch msg := msg.(type) {
		case walMsgCloseBatch:
			if height != msg.Height {
				panic(Fmt("Unexpected CloseBatch height %v, expected %v", msg.Height, height))
			}
			return batchMsgs
		case walMsgAddNode, walMsgDelNode:
			batchMsgs = append(batchMsgs, msg)
		default:
			panic(Fmt("Unexpected message %v (%v)", msg, reflect.TypeOf(msg)))
		}
	}
}

func (wal *WAL) write(msg WALMessage) {
	// Write markers for certain messages
	switch msg := msg.(type) {
	case walMsgOpenBatch:
		err := wal.group.WriteLine(walMarkerOpenBatch + Fmt("%v", msg.Height))
		if err != nil {
			panic(Fmt("Error writing OpenBatch marker to WAL: %v", err))
		}
	}

	// We need to b64 encode them, newlines not allowed within message
	var msgBytes = wire.BinaryBytes(struct{ WALMessage }{msg})
	var msgBytesB64 = base64.StdEncoding.EncodeToString(msgBytes)
	err := wal.group.WriteLine(string(msgBytesB64))
	if err != nil {
		panic(Fmt("Error writing msg to WAL: %v \n\nMessage: %v", err, msg))
	}

	// Write markers for certain messages
	switch msg := msg.(type) {
	case walMsgCloseBatch:
		err := wal.group.WriteLine(walMarkerCloseBatch + Fmt("%v", msg.Height))
		if err != nil {
			panic(Fmt("Error writing CloseBatch marker to WAL: %v", err))
		}
	case walMsgBatchEnded:
		err := wal.group.WriteLine(walMarkerBatchEnded + Fmt("%v", msg.Height))
		if err != nil {
			panic(Fmt("Error writing BatchEnded marker to WAL: %v", err))
		}
	}
}

func (wal *WAL) processWALMessages(height int, msgs []WALMessage) {
	if wal.batchHeight != height {
		panic(Fmt("Unexpected height in processWALMessages. Got %v, expected %v", height, wal.batchHeight))
	}
	if wal.batchStatus != batchStatusClosed {
		panic(Fmt("Unexpected status in processWALMessages. Got %v, expected %v", wal.batchStatus, batchStatusClosed))
	}

	// Write BatchStarted
	wal.write(walMsgBatchStarted{height})

	// TODO: consider compressing add/dels,
	// Sometimes we delete and add the same ndoe in the same batch.

	// Process messages
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case walMsgAddNode:
			wal.db.Set(msg.Key, msg.Value)
		case walMsgDelNode:
			wal.db.Delete(msg.Key)
		}
	}

	// Flush db
	// TODO: need an official API
	// TODO: verify that this does what we think it does.
	wal.db.SetSync(nil, nil)

	// Write BatchEnded
	wal.write(walMsgBatchEnded{height})
	wal.group.Head.Sync()

	wal.batchHeight += 1
	wal.batchStatus = batchStatusPending
}

//----------------------------------------
// NOTE: Not goroutine safe

func (wal *WAL) OpenBatch() {
	if wal.batchStatus != batchStatusPending {
		panic(Fmt("Expected batchStatusPending, got %v", wal.batchStatus))
	}
	height := wal.batchHeight
	wal.batchStatus = batchStatusOpened
	wal.write(walMsgOpenBatch{height})
}

func (wal *WAL) AddNode(key []byte, value []byte) {
	if wal.batchStatus != batchStatusOpened {
		panic(Fmt("Expected batchStatusOpened, got %v", wal.batchStatus))
	}
	msg := walMsgAddNode{key, value}
	wal.batchMsgs = append(wal.batchMsgs, msg)
	wal.write(msg)
}

func (wal *WAL) DelNode(key []byte) {
	if wal.batchStatus != batchStatusOpened {
		panic(Fmt("Expected batchStatusOpened, got %v", wal.batchStatus))
	}
	msg := walMsgDelNode{key}
	wal.batchMsgs = append(wal.batchMsgs, msg)
	wal.write(msg)
}

func (wal *WAL) CloseBatchSync() {
	if wal.batchStatus != batchStatusOpened {
		panic(Fmt("Expected batchStatusOpened, got %v", wal.batchStatus))
	}
	height := wal.batchHeight
	wal.write(walMsgCloseBatch{height})
	wal.group.Head.Sync()
	wal.batchStatus = batchStatusClosed
}

func (wal *WAL) ProcessBatchSync() {
	if wal.batchStatus != batchStatusClosed {
		panic(Fmt("Expected batchStatusClosed, got %v", wal.batchStatus))
	}
	wal.processWALMessages(wal.batchHeight, wal.batchMsgs)
	wal.batchMsgs = nil
}
