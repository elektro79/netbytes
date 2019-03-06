package netBytes

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

const maxMsg uint64 = 1024 * 8

type NetBytesConn struct {
	Conn     net.Conn
	MaxMsg   uint64
	queueOut chan []byte
}

type NetBytesProcessor interface {
	Connected(*NetBytesConn)
	Msg(*NetBytesConn, []byte)
	Disconected(*NetBytesConn, error)
}

type NetBytes struct {
	MaxMsg uint64
	nbp    NetBytesProcessor
}

func NewNetBytes(nbp NetBytesProcessor) *NetBytes {
	ns := NetBytes{maxMsg, nbp}
	return &ns
}

func (n *NetBytes) Connect(conn net.Conn) {
	n.process(conn)
}

func (n *NetBytes) Listen(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go n.process(conn)
	}
}

func (n *NetBytes) process(conn net.Conn) {
	nbc := &NetBytesConn{conn, n.MaxMsg, make(chan []byte, 10)}
	go nbc.process()
	n.nbp.Connected(nbc)
	reader := bufio.NewReader(conn)
	for {
		longs, err := reader.ReadString(':')
		if err != nil {
			n.nbp.Disconected(nbc, err)
			break
		}
		long, err := strconv.ParseUint(longs[0:len(longs)-1], 10, 64)
		if err != nil {
			n.nbp.Disconected(nbc, fmt.Errorf("error in convert len data: %s\n", err.Error()))
			break
		}
		if long > nbc.MaxMsg {
			n.nbp.Disconected(nbc, fmt.Errorf("Data len (%d) is bigger than max msg len (%d)\n", long, nbc.MaxMsg))
			break
		}
		b := make([]byte, long)
		if nread, err := reader.Read(b); err != nil {
			n.nbp.Disconected(nbc, fmt.Errorf("error in receive data: %s\n", err.Error()))
			break
		} else if nread != len(b) {
			n.nbp.Disconected(nbc, fmt.Errorf("error, receive %d data and expected %d\n", nread, long))
			break
		}
		if by, err := reader.ReadByte(); err != nil {
			n.nbp.Disconected(nbc, fmt.Errorf("error in last byte, must be ',': %s", err.Error()))
			break
		} else if by != ',' {
			n.nbp.Disconected(nbc, fmt.Errorf("error, last byte must be ',' and is '%v'", by))
			break
		}
		n.nbp.Msg(nbc, b)
	}
	close(nbc.queueOut)
}

func (n *NetBytesConn) process() {
	w := bufio.NewWriter(n.Conn)
	for s := range n.queueOut {
		w.Write([]byte(strconv.Itoa(len(s))))
		w.WriteByte(':')
		w.Write(s)
		w.WriteByte(',')
		w.Flush()
	}
}

func (n *NetBytesConn) Send(s []byte) {
	n.queueOut <- s
}
