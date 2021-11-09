package snet

import (
	//"bytes"
	//"crypto/md5"
	"encoding/binary"
	//"github.com/funny/crypto/dh64/go"
	"io"
	"kinger/gopuppy/common/glog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
	//"hash"
	"math"
)

var _ net.Listener = &Listener{}

type Listener struct {
	base         net.Listener
	config       Config
	acceptChan   chan net.Conn
	closed       bool
	closeOnce    sync.Once
	closeChan    chan struct{}
	atomicConnID uint32
	connsMutex   sync.Mutex
	conns        map[uint32]*Conn
}

func Listen(config Config, listenFunc func() (net.Listener, error)) (*Listener, error) {
	listener, err := listenFunc()
	if err != nil {
		return nil, err
	}
	l := &Listener{
		base:       listener,
		config:     config,
		closeChan:  make(chan struct{}),
		acceptChan: make(chan net.Conn, 1000),
		conns:      make(map[uint32]*Conn),
	}
	go l.acceptLoop()
	return l, nil
}

func (l *Listener) Addr() net.Addr {
	return l.base.Addr()
}

func (l *Listener) Close() error {
	l.closeOnce.Do(func() {
		l.closed = true
		close(l.closeChan)
	})
	return l.base.Close()
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case <-l.closeChan:
	}
	return nil, os.ErrInvalid
}

func (l *Listener) acceptLoop() {
	for {
		conn, err := l.base.Accept()
		if err != nil {
			if !l.closed {
				glog.Errorf("accept failed: %v", err)
			}
			break
		}
		go l.handshake(conn)
	}
}

func (l *Listener) handshake(conn net.Conn) {
	if l.config.HandshakeTimeout > 0 {
		conn.SetDeadline(time.Now().Add(l.config.HandshakeTimeout))
		defer conn.SetDeadline(time.Time{})
	}

	var (
		//buf    [24 + md5.Size]byte
		buf    [12]byte
		field1 = buf[0:4]
		//field2 = buf[8:16]
		//field3 = buf[16:24]
		//field4 = buf[24 : 24+md5.Size]
		field2 = buf[4:8]
		field3 = buf[8:12]
	)
	if _, err := io.ReadFull(conn, buf[:]); err != nil {
		conn.Close()
		return
	}

	connID := binary.BigEndian.Uint32(field1)
	switch connID {
	case 0:
		//glog.Debugf("new conn")

		//connPubKey := binary.BigEndian.Uint64(field2)
		//if connPubKey == 0 {
		//	l.trace("zero public key")
		//	conn.Close()
		//	return
		//}

		//privKey, pubKey := dh64.KeyPair()
		//secret := dh64.Secret(privKey, connPubKey)

		connID = atomic.AddUint32(&l.atomicConnID, 1)
		//sconn, err := newConn(conn, connID, secret, l.config)
		sconn, err := newConn(conn, connID, 0, l.config)
		if err != nil {
			glog.Errorf("new conn failed: %s", err)
			conn.Close()
			return
		}

		//binary.BigEndian.PutUint64(field1, pubKey)
		//binary.BigEndian.PutUint64(field2, connID)
		//sconn.writeCipher.XORKeyStream(field2, field2)
		binary.BigEndian.PutUint32(field1, connID)
		//if _, err := conn.Write(buf[:16]); err != nil {
		if _, err := conn.Write(buf[:4]); err != nil {
			glog.Errorf("send handshake response failed: %s", err)
			conn.Close()
			return
		}

		sconn.listener = l
		l.putConn(connID, sconn)
		select {
		case l.acceptChan <- sconn:
		case <-l.closeChan:
		}
	default:
		//glog.Debugf("reconn")

		sconn, exists := l.getConn(connID)
		if !exists {
			glog.Errorf("conn %d not exists", connID)
			binary.BigEndian.PutUint32(buf[0:4], math.MaxUint32)
			binary.BigEndian.PutUint32(buf[4:8], math.MaxUint32)
			conn.Write(buf[:8])
			conn.Close()
			return
		}

		//hash := md5.New()
		//hash.Write(buf[:24])
		//hash.Write(sconn.key[:])
		//md5sum := hash.Sum(nil)
		//if !bytes.Equal(field4, md5sum) {
		//	glog.Errorf("not equals: %x, %x", field4, md5sum)
		//	conn.Close()
		//	return
		//}

		writeCount := binary.BigEndian.Uint32(field2)
		readCount := binary.BigEndian.Uint32(field3)
		sconn.handleReconn(conn, writeCount, readCount)
	}
}

func (l *Listener) getConn(id uint32) (*Conn, bool) {
	l.connsMutex.Lock()
	defer l.connsMutex.Unlock()
	conn, exists := l.conns[id]
	return conn, exists
}

func (l *Listener) putConn(id uint32, conn *Conn) {
	l.connsMutex.Lock()
	defer l.connsMutex.Unlock()
	l.conns[id] = conn
}

func (l *Listener) delConn(id uint32) {
	l.connsMutex.Lock()
	defer l.connsMutex.Unlock()
	if _, exists := l.conns[id]; exists {
		delete(l.conns, id)
	}
}
