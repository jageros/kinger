package network

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/network/snet"
	"kinger/gopuppy/network/websocket"
	"net/url"
)

type RawRpcPacketHandler func(ses *Session, msgID int32, msgType protoc.MessageType, pktSeq uint32, payload []byte)

type Peer struct {
	sesMgr           *sessionMgr
	cfg              *PeerConfig
	listeners        []net.Listener
	rpcHandlers      map[int32]RpcHandler
	rawPacketHandler RawRpcPacketHandler
	running          bool
	waitStop         sync.WaitGroup
	closeOnce        sync.Once
}

func NewPeer(cfg *PeerConfig) *Peer {
	return &Peer{
		rpcHandlers: make(map[int32]RpcHandler),
		sesMgr:      newSessionMgr(),
		cfg:         cfg,
		running:     true,
	}
}

func (p *Peer) ListenTcp(ip string, port int, factory *protoc.ProtocFactory, snetConfig *snet.Config) {
	var ln net.Listener
	var err error
	if snetConfig != nil {
		ln, err = snet.Listen(*snetConfig, func() (net.Listener, error) {
			return net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
		})
	} else {
		ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	}

	if err != nil {
		panic(err)
	}
	p.listeners = append(p.listeners, ln)
	p.listen(ln, factory)
}

func (p *Peer) ListenWs(ip string, port int, certfile, keyfile string, factory *protoc.ProtocFactory, snetConfig *snet.Config) {
	var ln net.Listener
	if snetConfig != nil {
		ln, _ = snet.Listen(*snetConfig, func() (net.Listener, error) {
			return websocket.NewWsListener(ip, port, certfile, keyfile), nil
		})
	} else {
		ln = websocket.NewWsListener(ip, port, certfile, keyfile)
	}
	p.listeners = append(p.listeners, ln)
	p.listen(ln, factory)
}

func (p *Peer) SetupHttp(ip string, port int, certfile, keyfile string) {
	go func() {
		var err error
		address := fmt.Sprintf("%s:%d", ip, port)
		glog.Infof("Http listen %s", address)
		if certfile != "" && keyfile != "" {
			err = http.ListenAndServeTLS(address, certfile, keyfile, nil)
		} else {
			err = http.ListenAndServe(address, nil)
		}

		if err != nil {
			glog.Errorf("Http listen %s err %s", address, err)
		}

	}()
}

func (p *Peer) DialTcp(ip string, port int, factory *protoc.ProtocFactory, snetConfig *snet.Config) (*Session, error) {
	var conn net.Conn
	var err error
	if snetConfig != nil {
		conn, err = snet.Dial(*snetConfig, func() (net.Conn, error) {
			return net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		})
	} else {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	}
	if err != nil {
		return nil, err
	}

	return p.serveConn(conn, factory)
}

func (p *Peer) DialWs(ip string, port int, factory *protoc.ProtocFactory, snetConfig *snet.Config) (*Session, error) {
	var conn net.Conn
	var err error
	wsUrl := (&url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", ip, port), Path: "/"}).String()

	if snetConfig != nil {
		conn, err = snet.Dial(*snetConfig, func() (net.Conn, error) {
			c, _, err := ws.DefaultDialer.Dial(wsUrl, nil)
			return websocket.NewWsConn(c), err
		})
	} else {
		c, _, err2 := ws.DefaultDialer.Dial(wsUrl, nil)
		conn = websocket.NewWsConn(c)
		err = err2
	}
	if err != nil {
		return nil, err
	}

	return p.serveConn(conn, factory)
}

func (p *Peer) listen(ln net.Listener, factory *protoc.ProtocFactory) {
	p.waitStop.Add(1)
	go func() {
		for p.running {
			conn, err := ln.Accept()
			if err != nil {
				if p.running {
					glog.Errorf("Accept %s error %s", ln.Addr(), err)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				break
			}

			go func() {
				ses, err := p.serveConn(conn, factory)
				if err == nil {
					evq.PostEvent(evq.NewCommonEvent(consts.SESSION_ON_ACCEPT_EVENT, ses))

					if c, ok := conn.(*snet.Conn); ok {
						c.OnDisconnect = func() {
							evq.PostEvent(evq.NewCommonEvent(consts.SNET_ON_DISCONNECT, ses))
						}

						c.OnReconnect = func() {
							evq.PostEvent(evq.NewCommonEvent(consts.SNET_ON_RECONNECT, ses))
						}
					}
				}
			}()
		}

		p.waitStop.Done()
	}()
}

func (p *Peer) serveConn(conn net.Conn, factory *protoc.ProtocFactory) (*Session, error) {
	ses := newSession(p, conn, factory)
	p.sesMgr.addSession(ses)
	ses.onClose = func() {
		p.sesMgr.removeSession(ses)
		glog.Debugf("ses %d on close", ses.GetSesID())
		evq.PostEvent(evq.NewCommonEvent(consts.SESSION_ON_CLOSE_EVENT, ses))
	}
	err := ses.startReadAndWrite()
	return ses, err
}

func (p *Peer) GetSession(sesID int64) *Session {
	return p.sesMgr.getSession(sesID)
}

func (p *Peer) RegisterRpcHandler(msgID protoc.IMessageID, handler RpcHandler) {
	p.rpcHandlers[msgID.ID()] = handler
}

func (p *Peer) SetRawPacketHandler(handler RawRpcPacketHandler) {
	p.rawPacketHandler = handler
}

func (p *Peer) getRpcHandler(msgID protoc.IMessageID) RpcHandler {
	return p.rpcHandlers[msgID.ID()]
}

func (p *Peer) getRawPacketHandler() RawRpcPacketHandler {
	return p.rawPacketHandler
}

func (p *Peer) Close() {
	p.closeOnce.Do(func() {
		p.running = false
		for _, ln := range p.listeners {
			ln.Close()
		}
		p.waitStop.Wait()
		p.sesMgr.closeAllSession()
	})
}
