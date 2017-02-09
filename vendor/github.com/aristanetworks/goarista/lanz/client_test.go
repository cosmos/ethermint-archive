// Copyright (C) 2016  Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package lanz_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/aristanetworks/goarista/lanz"
	pb "github.com/aristanetworks/goarista/lanz/proto"
	"github.com/aristanetworks/goarista/test"

	"github.com/aristanetworks/glog"
	"github.com/golang/protobuf/proto"
)

var testProtoBuf = &pb.LanzRecord{
	ErrorRecord: &pb.ErrorRecord{
		Timestamp:    proto.Uint64(146591697107549),
		ErrorMessage: proto.String("Error"),
	},
}

type testConnector struct {
	reader   *bytes.Reader
	open     bool
	refusing bool
	connect  chan bool
	block    chan bool
}

func (c *testConnector) Read(p []byte) (int, error) {
	if !c.open {
		return 0, errors.New("closed")
	}

	if c.reader.Len() == 0 {
		<-c.block
		return 0, io.EOF
	}

	return c.reader.Read(p)
}

func (c *testConnector) Close() error {
	if !c.open {
		return nil
	}

	c.open = false
	close(c.block)
	return nil
}

func (c *testConnector) Connect() error {
	var err error

	if c.refusing {
		err = errors.New("refused")
	} else {
		c.block = make(chan bool)
		c.open = true
	}
	if c.connect != nil {
		c.connect <- true
	}
	return err
}

func launchClient(ch chan<- *pb.LanzRecord, conn lanz.ConnectReadCloser) (lanz.Client, chan bool) {
	done := make(chan bool)

	c := lanz.New(lanz.WithConnector(conn), lanz.WithBackoff(1*time.Millisecond))
	go func() {
		c.Run(ch)
		done <- true
	}()

	return c, done
}

func pbsToStream(pbs []*pb.LanzRecord) []byte {
	var r []byte

	for _, p := range pbs {
		b, err := proto.Marshal(p)
		if err != nil {
			glog.Fatalf("Can't marshal pb: %v", err)
		}

		bLen := uint64(len(b))
		s := make([]byte, binary.MaxVarintLen64)
		sLen := binary.PutUvarint(s, bLen)

		r = append(r, s[:sLen]...)
		r = append(r, b...)
	}

	return r
}

func testStream() []byte {
	return pbsToStream([]*pb.LanzRecord{testProtoBuf})
}

// This function tests that basic workflow works.
func TestSuccessPath(t *testing.T) {
	pbs := []*pb.LanzRecord{
		{
			ConfigRecord: &pb.ConfigRecord{
				Timestamp:    proto.Uint64(146591697107544),
				LanzVersion:  proto.Uint32(1),
				NumOfPorts:   proto.Uint32(146),
				SegmentSize:  proto.Uint32(512),
				MaxQueueSize: proto.Uint32(524288000),
				PortConfigRecord: []*pb.ConfigRecord_PortConfigRecord{
					{
						IntfName:      proto.String("Cpu"),
						SwitchId:      proto.Uint32(2048),
						PortId:        proto.Uint32(4096),
						InternalPort:  proto.Bool(false),
						HighThreshold: proto.Uint32(50000),
						LowThreshold:  proto.Uint32(25000),
					},
				},
				GlobalUsageReportingEnabled: proto.Bool(true),
			},
		},
		{
			CongestionRecord: &pb.CongestionRecord{
				Timestamp: proto.Uint64(146591697107546),
				IntfName:  proto.String("Cpu"),
				SwitchId:  proto.Uint32(2048),
				PortId:    proto.Uint32(4096),
				QueueSize: proto.Uint32(30000),
			},
		},
		{
			ErrorRecord: &pb.ErrorRecord{
				Timestamp:    proto.Uint64(146591697107549),
				ErrorMessage: proto.String("Error"),
			},
		},
	}

	conn := &testConnector{reader: bytes.NewReader(pbsToStream(pbs))}
	ch := make(chan *pb.LanzRecord)
	c, done := launchClient(ch, conn)
	for i, p := range pbs {
		r, ok := <-ch
		if !ok {
			t.Fatalf("Unexpected closed channel")
		}
		if !test.DeepEqual(p, r) {
			t.Fatalf("Test case %d: expected %v, but got %v", i, p, r)
		}
	}
	c.Stop()
	<-done
	if conn.open {
		t.Fatalf("Connection still open after stopping")
	}
}

// This function tests that the client keeps retrying on connection error.
func TestRetryOnConnectionError(t *testing.T) {
	conn := &testConnector{
		refusing: true,
		connect:  make(chan bool),
	}

	ch := make(chan *pb.LanzRecord)
	c, done := launchClient(ch, conn)

	connects := 3
	stopped := false
	for !stopped {
		select {
		case <-conn.connect:
			connects--
			if connects == 0 {
				c.Stop()
			}
		case <-done:
			stopped = true
		}
	}
}

// This function tests that the client will reconnect if the connection gets closed.
func TestRetryOnClose(t *testing.T) {
	conn := &testConnector{
		reader:  bytes.NewReader(testStream()),
		connect: make(chan bool),
	}

	ch := make(chan *pb.LanzRecord)
	c, done := launchClient(ch, conn)
	<-conn.connect
	<-ch
	conn.Close()
	<-conn.connect
	conn.Close()
	<-conn.connect
	c.Stop()
	<-done

	if conn.open {
		t.Fatalf("Connection still open after stopping")
	}
}

// This function tests that the client will reconnect if it receives truncated input.
func TestRetryOnTrucatedInput(t *testing.T) {
	conn := &testConnector{
		reader:  bytes.NewReader(testStream()[:5]),
		connect: make(chan bool),
	}

	ch := make(chan *pb.LanzRecord)
	c, done := launchClient(ch, conn)
	<-conn.connect
	conn.block <- false
	<-conn.connect
	c.Stop()
	<-done
	if conn.open {
		t.Fatalf("Connection still open after stopping")
	}
}

// This function tests that the client will reconnect if it receives malformed input.
func TestRetryOnMalformedInput(t *testing.T) {
	stream := testStream()
	uLen := binary.PutUvarint(stream, 3)
	conn := &testConnector{
		reader:  bytes.NewReader(stream[:uLen+3]),
		connect: make(chan bool),
	}

	ch := make(chan *pb.LanzRecord)
	c, done := launchClient(ch, conn)
	<-conn.connect
	<-conn.connect
	c.Stop()
	<-done
	if conn.open {
		t.Fatalf("Connection still open after stopping")
	}
}

// This function tests the default connector.
func TestDefaultConnector(t *testing.T) {
	stream := testStream()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Can't listen: %v", err)
	}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Fatalf("Can't accept: %v", err)
		}
		conn.Write(stream)
		conn.Close()
	}()

	ch := make(chan *pb.LanzRecord)
	done := make(chan bool)
	c := lanz.New(lanz.WithAddr(l.Addr().String()), lanz.WithBackoff(1*time.Millisecond))
	go func() {
		c.Run(ch)
		done <- true
	}()
	p := <-ch
	c.Stop()
	<-done

	if !test.DeepEqual(p, testProtoBuf) {
		t.Fatalf("Expected protobuf %v, but got %v", testProtoBuf, p)
	}
}
