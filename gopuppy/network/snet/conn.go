package snet

import (
	//"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
	//"github.com/funny/crypto/dh64/go"
	//"kinger/gopuppy/common/glog"
	//"hash"
	"math"
)

var _ net.Conn = &Conn{}

type Config struct {
	EnableCrypt        bool
	HandshakeTimeout   time.Duration
	RewriterBufferSize int
	ReconnWaitTimeout  time.Duration
}

type Dialer func() (net.Conn, error)

type Conn struct {
	base     net.Conn
	id       uint32
	listener *Listener
	dialer   Dialer

	key         [8]byte
	enableCrypt bool

	closed    bool
	closeChan chan struct{}
	closeOnce sync.Once

	writeMutex  sync.Mutex
	writeCipher *rc4.Cipher

	readMutex  sync.Mutex
	readCipher *rc4.Cipher

	reconnMutex       sync.RWMutex
	reconnOpMutex     sync.Mutex
	readWaiting       bool
	writeWaiting      bool
	readWaitChan      chan struct{}
	writeWaitChan     chan struct{}
	reconnWaitTimeout time.Duration

	rewriter     rewriter
	rereader     rereader
	readCount    uint32
	writeCount   uint32
	OnDisconnect func()
	OnReconnect func()
}

func Dial(config Config, dialer Dialer) (net.Conn, error) {
	conn, err := dialer()
	if err != nil {
		return nil, err
	}

	var (
		buf    [12]byte
		field1 = buf[0:4]
		//field2 = buf[4:8]
	)

	//privKey, pubKey := dh64.KeyPair()
	//binary.BigEndian.PutUint64(field2, pubKey)
	if _, err := conn.Write(buf[:]); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return nil, err
	}

	//srvPubKey := binary.BigEndian.Uint64(field1)
	//secret := dh64.Secret(privKey, srvPubKey)

	sconn, err := newConn(conn, 0, 0, config)
	if err != nil {
		return nil, err
	}

	//sconn.readCipher.XORKeyStream(field2, field2)
	sconn.id = binary.BigEndian.Uint32(field1)
	sconn.dialer = dialer
	return sconn, nil
}

func newConn(base net.Conn, id uint32, secret uint64, config Config) (conn *Conn, err error) {
	conn = &Conn{
		base:              base,
		id:                id,
		enableCrypt:       config.EnableCrypt,
		reconnWaitTimeout: config.ReconnWaitTimeout,
		closeChan:         make(chan struct{}),
		readWaitChan:      make(chan struct{}),
		writeWaitChan:     make(chan struct{}),
		rewriter: rewriter{
			data: make([]byte, config.RewriterBufferSize),
		},
	}

	//binary.BigEndian.PutUint64(conn.key[:], secret)

	//conn.writeCipher, err = rc4.NewCipher(conn.key[:])
	//if err != nil {
	//	return nil, err
	//}

	//conn.readCipher, err = rc4.NewCipher(conn.key[:])
	//if err != nil {
	//	return nil, err
	//}

	return conn, nil
}

func (c *Conn) WrapBaseForTest(wrap func(net.Conn) net.Conn) {
	c.base = wrap(c.base)
}

func (c *Conn) RemoteAddr() net.Addr {
	c.reconnMutex.RLock()
	defer c.reconnMutex.RUnlock()
	return c.base.RemoteAddr()
}

func (c *Conn) LocalAddr() net.Addr {
	c.reconnMutex.RLock()
	defer c.reconnMutex.RUnlock()
	return c.base.LocalAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	c.reconnMutex.RLock()
	defer c.reconnMutex.RUnlock()
	return c.base.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	c.reconnMutex.RLock()
	defer c.reconnMutex.RUnlock()
	return c.base.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.reconnMutex.RLock()
	defer c.reconnMutex.RUnlock()
	return c.base.SetWriteDeadline(t)
}

func (c *Conn) SetReconnWaitTimeout(d time.Duration) {
	c.reconnWaitTimeout = d
}

func (c *Conn) Close() error {
	//glog.Debugf("Close()")
	c.closeOnce.Do(func() {
		c.closed = true
		if c.listener != nil {
			c.listener.delConn(c.id)
		}
		close(c.closeChan)
	})
	return c.base.Close()
}

func (c *Conn) TryReconn() {
	if c.listener == nil {
		c.reconnMutex.RLock()
		base := c.base
		c.reconnMutex.RUnlock()
		go c.tryReconn(base)
	}
}

func (c *Conn) Read(b []byte) (n int, err error) {
	//glog.Debugf("Read(%d)", len(b))
	if len(b) == 0 {
		return
	}

	//glog.Debugf("Read() wait write")
	c.readMutex.Lock()
	//glog.Debugf("Read() wait reconn")
	c.reconnMutex.RLock()
	c.readWaiting = true

	defer func() {
		c.readWaiting = false
		c.reconnMutex.RUnlock()
		c.readMutex.Unlock()
	}()

	for {
		n, err = c.rereader.Pull(b), nil
		//glog.Debugf("read from queue, n = %d", n)
		if n > 0 {
			break
		}

		base := c.base
		n, err = base.Read(b[n:])
		if err == nil {
			//glog.Debugf("read from conn, n = %d", n)
			break
		}
		base.Close()

		if c.listener == nil {
			go c.tryReconn(base)
		}

		if c.OnDisconnect != nil {
			c.OnDisconnect()
		}

		if !c.waitReconn('r', c.readWaitChan) {
			break
		} else if c.OnReconnect != nil {
			c.OnReconnect()
		}
	}

	if err == nil {
		if c.enableCrypt {
			c.readCipher.XORKeyStream(b[:n], b[:n])
		}
		c.readCount += uint32(n)
	}

	//glog.Debugf("Read(), n = %d, err = %v", n, err)
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	//glog.Debugf("Write(%d)", len(b))
	if len(b) == 0 {
		return
	}

	//glog.Debugf("Write() wait write")
	c.writeMutex.Lock()
	//glog.Debugf("Write() wait reconn")
	c.reconnMutex.RLock()
	c.writeWaiting = true

	defer func() {
		c.writeWaiting = false
		c.reconnMutex.RUnlock()
		c.writeMutex.Unlock()
	}()

	if c.enableCrypt {
		c.writeCipher.XORKeyStream(b, b)
	}

	c.rewriter.Push(b)
	c.writeCount += uint32(len(b))

	base := c.base
	n, err = base.Write(b)
	if err == nil {
		return
	}
	base.Close()

	if c.listener == nil {
		go c.tryReconn(base)
	}

	if c.OnDisconnect != nil {
		c.OnDisconnect()
	}

	if c.waitReconn('w', c.writeWaitChan) {
		n, err = len(b), nil
		if c.OnReconnect != nil {
			c.OnReconnect()
		}
	}
	return
}

func (c *Conn) waitReconn(who byte, waitChan chan struct{}) (done bool) {
	//glog.Debugf("waitReconn('%c', \"%s\")", who, c.reconnWaitTimeout)

	timeout := time.NewTimer(c.reconnWaitTimeout)
	defer timeout.Stop()

	c.reconnMutex.RUnlock()
	defer func() {
		c.reconnMutex.RLock()
		if done {
			<-waitChan
			//glog.Debugf("waitReconn('%c', \"%s\") done", who, c.reconnWaitTimeout)
		}
	}()

	var lsnCloseChan chan struct{}
	if c.listener == nil {
		lsnCloseChan = make(chan struct{})
	} else {
		lsnCloseChan = c.listener.closeChan
	}

	select {
	case <-waitChan:
		done = true
		//glog.Debugf("waitReconn('%c', \"%s\") wake up", who, c.reconnWaitTimeout)
		return
	case <-c.closeChan:
		//glog.Debugf("waitReconn('%c', \"%s\") closed", who, c.reconnWaitTimeout)
		return
	case <-timeout.C:
		//glog.Debugf("waitReconn('%c', \"%s\") timeout", who, c.reconnWaitTimeout)
		c.Close()
		return
	case <-lsnCloseChan:
		//glog.Debugf("waitReconn('%c', \"%s\") listener closed", who, c.reconnWaitTimeout)
		return
	}
}

func (c *Conn) handleReconn(conn net.Conn, writeCount, readCount uint32) {
	var done bool

	//glog.Debugf("handleReconn() wait handleReconn()")
	c.reconnOpMutex.Lock()
	defer c.reconnOpMutex.Unlock()

	c.base.Close()
	//glog.Debugf("handleReconn() wait Read() or Write()")
	c.reconnMutex.Lock()
	readWaiting := c.readWaiting
	writeWaiting := c.writeWaiting
	defer func() {
		c.reconnMutex.Unlock()
		if done {
			c.wakeUp(readWaiting, writeWaiting)
		} else {
			conn.Close()
		}
	}()
	//glog.Debugf("handleReconn() begin")

	var buf [8]byte
	binary.BigEndian.PutUint32(buf[0:4], c.writeCount)
	binary.BigEndian.PutUint32(buf[4:8], c.readCount+c.rereader.count)
	if _, err := conn.Write(buf[:8]); err != nil {
		//glog.Debugf("response failed")
		return
	}

	isDone, alive := c.doReconn(conn, writeCount, readCount)
	if isDone && alive {
		done = true
	} else if !alive {
		return
	}
}

func (c *Conn) tryReconn(badConn net.Conn) {
	var done bool

	//glog.Debugf("tryReconn() wait tryReconn()")
	c.reconnOpMutex.Lock()
	defer c.reconnOpMutex.Unlock()

	//glog.Debugf("tryReconn() wait Read() or Write()")
	badConn.Close()
	c.reconnMutex.Lock()
	readWaiting := c.readWaiting
	writeWaiting := c.writeWaiting
	defer func() {
		c.reconnMutex.Unlock()
		if done {
			c.wakeUp(readWaiting, writeWaiting)
		}
	}()
	//glog.Debugf("tryReconn() begin")

	if badConn != c.base {
		//glog.Debugf("badConn != c.base")
		return
	}

	//var buf [24 + md5.Size]byte
	var buf [12]byte
	binary.BigEndian.PutUint32(buf[0:4], c.id)
	binary.BigEndian.PutUint32(buf[4:8], c.writeCount)
	binary.BigEndian.PutUint32(buf[8:12], c.readCount+c.rereader.count)
	//hash := md5.New()
	//hash.Write(buf[0:24])
	//hash.Write(c.key[:])
	//copy(buf[24:], hash.Sum(nil))

	for i := 0; !c.closed; i++ {
		if i > 0 {
			time.Sleep(time.Second * 3)
		}

		//glog.Debugf("reconn dial")
		conn, err := c.dialer()
		if err != nil {
			//glog.Debugf("dial fialed: %v", err)
			continue
		}

		//glog.Debugf("send reconn request")
		if _, err := conn.Write(buf[:]); err != nil {
			//glog.Debugf("write fialed: %v", err)
			conn.Close()
			continue
		}

		//glog.Debugf("wait reconn response")
		var buf2 [8]byte
		if _, err := io.ReadFull(conn, buf2[:]); err != nil {
			//glog.Debugf("read fialed", err.Error())
			conn.Close()
			continue
		}

		writeCount := binary.BigEndian.Uint32(buf2[0:4])
		readCount := binary.BigEndian.Uint32(buf2[4:8])
		if reconnDone, alive := c.doReconn(conn, writeCount, readCount); reconnDone && alive {
			done = true
			break
		} else if !alive {
			return
		}
		conn.Close()
	}
}

func (c *Conn) doReconn(conn net.Conn, writeCount, readCount uint32) (reconnDone bool, alive bool) {
	//glog.Debugf(
	//	"doReconn(\"%s\", %d, %d), c.writeCount = %d, c.readCount = %d",
	//	conn.RemoteAddr(), writeCount, readCount, c.writeCount, c.readCount,
	//)
	if writeCount == math.MaxUint32 && readCount == math.MaxUint32 {
		c.Close()
		return
	}

	if writeCount < c.readCount {
		//glog.Debugf("writeCount < c.readCount")
		c.Close()
		return
	}

	if c.writeCount < readCount {
		//glog.Debugf("c.writeCount < readCount")
		c.Close()
		return
	}

	if int(c.writeCount-readCount) > len(c.rewriter.data) {
		//glog.Debugf("c.writeCount - readCount > len(c.rewriter.data)")
		c.Close()
		return
	}

	if writeCount != c.readCount {
		rereadWaitChan := make(chan bool)

		defer func() {
			//glog.Debugf("reread wait")
			if !<-rereadWaitChan {
				reconnDone = false
				//glog.Debugf("reread failed")
				return
			}
			//glog.Debugf("reread done")
		}()

		go func() {
			n := int(writeCount) - int(c.readCount)
			//glog.Debugf(
			//	"reread, writeCount = %d, c.readCount = %d, n = %d",
			//	writeCount, c.readCount, n,
			//)
			rereadWaitChan <- c.rereader.Reread(conn, n)
		}()
	}

	if c.writeCount != readCount {
		//glog.Debugf(
		//	"rewrite, c.writeCount = %d, readCount = %d, n = %d",
		//	c.writeCount, readCount, c.writeCount-readCount,
		//)
		if !c.rewriter.Rewrite(conn, c.writeCount, readCount) {
			//glog.Debugf("rewrite failed")
			alive = true
			return
		}
		//glog.Debugf("rewrite done")
	}

	c.base = conn
	reconnDone = true
	alive = true
	return
}

func (c *Conn) wakeUp(readWaiting, writeWaiting bool) {
	if readWaiting {
		//glog.Debugf("continue read")
		// make sure reader take over reconnMutex
		for i := 0; i < 2; i++ {
			select {
			case c.readWaitChan <- struct{}{}:
			case <-c.closeChan:
				//glog.Debugf("continue read closed")
				return
			}
		}
		//glog.Debugf("continue read done")
	}

	if writeWaiting {
		//glog.Debugf("continue write")
		// make sure writer take over reconnMutex
		for i := 0; i < 2; i++ {
			select {
			case c.writeWaitChan <- struct{}{}:
			case <-c.closeChan:
				//glog.Debugf("continue write closed")
				return
			}
		}
		//glog.Debugf("continue write done")
	}
}
