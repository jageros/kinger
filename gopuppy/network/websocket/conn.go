package websocket

import (
	"bytes"
	"time"

	"github.com/gorilla/websocket"
	//"kinger/gopuppy/common/glog"
)

type wsConn struct {
	*websocket.Conn
	readBuff *bytes.Buffer
}

func NewWsConn(conn *websocket.Conn) *wsConn {
	return &wsConn{
		Conn:     conn,
		readBuff: bytes.NewBuffer([]byte{}),
	}
}

func (wc *wsConn) SetDeadline(t time.Time) error {
	err := wc.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = wc.SetWriteDeadline(t)
	if err != nil {
		return err
	}
	return nil
}

func (wc *wsConn) Read(b []byte) (int, error) {
	blen := len(b)
	rlen := 0
	//glog.Debugf("wsConn Read 111111111 blen %d", blen)
	if wc.readBuff.Len() > 0 {
		//glog.Debugf("wsConn Read 222222222 wc.readBuff.Len() %d", wc.readBuff.Len())
		n, err := wc.readBuff.Read(b)
		if n == blen {
			return n, nil
		} else if err != nil {
			return n, err
		}
		rlen += n
	}

	for rlen < blen {
		_, p, err := wc.ReadMessage()
		if err != nil {
			return rlen, err
		}

		plen := len(p)
		//glog.Debugf("wsConn Read 3333333333 plen %d", plen)
		n := copy(b[rlen:], p)
		rlen += n
		if n < plen {
			wc.readBuff.Write(p[n:])
			return rlen, nil
		}
	}

	return rlen, nil
}

func (wc *wsConn) Write(b []byte) (n int, err error) {
	if err := wc.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	} else {
		return len(b), nil
	}
}
