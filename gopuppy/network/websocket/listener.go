package websocket

import (
	"fmt"
	"net"
	"net/http"

	"kinger/gopuppy/common/glog"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type addr struct {
	network string
	address string
}

func (a *addr) Network() string {
	return a.network
}

func (a *addr) String() string {
	return fmt.Sprintf("%s://%s", a.network, a.address)
}

type WsListener struct {
	a *addr
	c chan net.Conn
}

func (wl *WsListener) Accept() (net.Conn, error) {
	conn := <-wl.c
	if conn == nil {
		return nil, errors.Errorf("%s close", wl.a)
	} else {
		return conn, nil
	}
}

func (wl *WsListener) Close() error {
	wl.c <- nil
	return nil
}

func (wl *WsListener) Addr() net.Addr {
	return wl.a
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wl *WsListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Debugf("ws get http req %s", r.RemoteAddr)

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("websocket Upgrade err %s", err)
		return
	}

	wl.c <- NewWsConn(conn)
}

func NewWsListener(ip string, port int, certfile, keyfile string) *WsListener {

	address := fmt.Sprintf("%s:%d", ip, port)
	network := "ws"
	if certfile != "" && keyfile != "" {
		network = "wss"
	}

	ln := &WsListener{
		a: &addr{
			network: network,
			address: address,
		},
		c: make(chan net.Conn),
	}

	go func() {
		var err error
		glog.Infof("websocket.listen %s", address)
		if certfile != "" && keyfile != "" {
			err = http.ListenAndServeTLS(address, certfile, keyfile, ln)
		} else {
			err = http.ListenAndServe(address, ln)
		}

		if err != nil {
			glog.Errorf("websocket.listen %s err %s", address, err)
		}

	}()

	return ln
}
