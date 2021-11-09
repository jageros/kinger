package network

import (
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network/protoc"

	"fmt"
	"github.com/pkg/errors"
	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"gopkg.in/eapache/queue.v1"
	"kinger/gopuppy/common/utils"
)

type Session struct {
	id              int64
	peer            *Peer
	conn            net.Conn
	onClose         func()
	alive           int32          // 0.false 1.true
	needNotifyWrite bool           // 是否需要通知写线程关闭
	endSync         sync.WaitGroup // 等待读写线程完成
	closeNotify     chan struct{}
	prop            map[string]interface{}
	ip              string

	writeQueueGuard sync.Mutex
	writeQueue      *queue.Queue
	syncWriteQueue  *xnsyncutil.SyncQueue

	rpcGuard   sync.Mutex
	rpcSeq     uint32
	rpcPending map[uint32]*rpcCaller

	proto      protoc.IProto
	author     protoc.IAuthor
	encryptor  protoc.IEncryptor
	compressor protoc.ICompressor
}

func newSession(peer *Peer, conn net.Conn, factory *protoc.ProtocFactory) *Session {
	ses := &Session{
		peer:            peer,
		conn:            conn,
		prop:            make(map[string]interface{}),
		rpcPending:      make(map[uint32]*rpcCaller),
		needNotifyWrite: true,
		closeNotify:     make(chan struct{}),
	}

	if peer.cfg.WriteBufSize > 0 {
		ses.writeQueue = queue.New()
	} else {
		ses.syncWriteQueue = xnsyncutil.NewSyncQueue()
	}

	if factory == nil || factory.ProtoFunc == nil {
		ses.proto = protoc.NewDefaultProto(conn, peer.cfg.MaxPacketSize, peer.cfg.ReadBufSize, peer.cfg.WriteBufSize)
	} else if factory != nil && factory.ProtoFunc != nil {
		ses.proto = factory.ProtoFunc(conn, peer.cfg.MaxPacketSize, peer.cfg.ReadBufSize, peer.cfg.WriteBufSize)
	}
	if factory != nil {
		if factory.AuthorFunc != nil {
			ses.author = factory.AuthorFunc()
		}
		if factory.EncryptorFunc != nil {
			ses.encryptor = factory.EncryptorFunc()
		}
		if factory.CompressorFunc != nil {
			ses.compressor = factory.CompressorFunc()
		}
	}

	return ses
}

func (s *Session) String() string {
	return fmt.Sprintf("[ses %d]", s.GetSesID())
}

func (s *Session) GetSesID() int64 {
	return s.id
}

func (s *Session) setSesID(id int64) {
	s.id = id
}

func (s *Session) FromPeer() *Peer {
	return s.peer
}

func (s *Session) IsAlive() bool {
	return atomic.LoadInt32(&s.alive) == 1
}

func (s *Session) GetProp(key string) interface{} {
	if v, ok := s.prop[key]; ok {
		return v
	} else {
		return nil
	}
}

func (s *Session) SetProp(key string, value interface{}) {
	s.prop[key] = value
}

func (s *Session) GetIP() string {
	if s.ip != "" {
		return s.ip
	}
	addr := strings.Split(s.conn.RemoteAddr().String(), ":")
	if len(addr) > 0 {
		s.ip = addr[0]
	}
	return s.ip
}

func (s *Session) Ping() {
	s.Write(protoc.PingPacket)
}

func (s *Session) getPeer() *Peer {
	return s.peer
}

func (s *Session) doClose() {
	if atomic.SwapInt32(&s.alive, 0) == 0 {
		return
	}

	s.rpcGuard.Lock()
	for _, caller := range s.rpcPending {
		caller.Done(nil, InternalErr.Errcode())
	}
	s.rpcGuard.Unlock()

	close(s.closeNotify)
	if s.onClose != nil {
		s.onClose()
	}
}

func (s *Session) Close() chan struct{} {
	if s.syncWriteQueue != nil {
		s.syncWriteQueue.Close()
	} else {
		s.writeQueueGuard.Lock()
		s.writeQueue.Add(nil)
		s.writeQueueGuard.Unlock()
	}
	return s.closeNotify
}

func (s *Session) pickRpcCaller(seq uint32) *rpcCaller {
	s.rpcGuard.Lock()
	caller, ok := s.rpcPending[seq]
	if ok {
		delete(s.rpcPending, seq)
	}
	s.rpcGuard.Unlock()
	return caller
}

func (s *Session) Write(p *protoc.Packet) {
	if p == nil {
		return
	}

	if s.syncWriteQueue != nil {
		s.syncWriteQueue.Push(p)
	} else {
		s.writeQueueGuard.Lock()
		s.writeQueue.Add(p)
		s.writeQueueGuard.Unlock()
	}
}

func (s *Session) startReadAndWrite() error {

	if s.author != nil {
		if err := s.author.DoAuth(s.conn); err != nil {
			glog.Errorf("session DoAuth error, ses=%s, err=%s", s.id, err)
			s.conn.Close()
			s.doClose()
			return err
		}
	}

	if s.encryptor != nil {
		if err := s.encryptor.Init(s.conn); err != nil {
			glog.Errorf("session encryptor Init error, ses=%s, err=%s", s.id, err)
			s.conn.Close()
			s.doClose()
			return err
		}
	}

	atomic.StoreInt32(&s.alive, 1)
	s.endSync.Add(2)

	go s.readThread()
	go s.writeThread()

	// 等待接收和发送结束
	go func() {
		s.endSync.Wait()
		s.doClose()
	}()

	return nil
}

func (s *Session) readThread() {
	readTimeout := s.peer.cfg.ReadTimeout
	for {
		if readTimeout > 0 {
			s.conn.SetReadDeadline(time.Now().Add(readTimeout))
		}

		pkt, err := s.proto.ReadPacket()
		if err != nil {
			glog.Debugf("session %d ReadPacket err %s", s.id, err)
			break
		}

		payload := pkt.GetPayload()
		msgType := pkt.GetMsgType()
		if msgType == protoc.MsgPong {
			continue
		} else if msgType == protoc.MsgPing {
			s.Write(protoc.PongPacket)
			continue
		}

		if msgType == protoc.MsgReq || msgType == protoc.MsgPush || msgType == protoc.MsgReply {
			if s.encryptor != nil {
				payload, err = s.encryptor.Decrypt(payload)
				if err != nil {
					glog.Errorf("session %d Decrypt err %s", s.id, err)
					break
				}
			}

			if s.compressor != nil {
				payload, err = s.compressor.Decompress(payload)
				if err != nil {
					glog.Errorf("session %s Decompress error %s", s.id, err)
					break
				}
			}
		}

		if msgType == protoc.MsgReq || msgType == protoc.MsgPush {
			meta := protoc.GetMeta(pkt.GetMsgID())
			if meta == nil {
				h := s.peer.getRawPacketHandler()
				if h == nil {
					glog.Errorf("%d no meta", pkt.GetMsgID())
					protoc.PutPacket(pkt)
					break
				} else {
					msgID := pkt.GetMsgID()
					msgType := pkt.GetMsgType()
					seq := pkt.GetSeq()
					payload := pkt.GetPayload()
					utils.CatchPanic(func() {
						h(s, msgID, msgType, seq, payload)
					})
					protoc.PutPacket(pkt)
					continue
				}
			}
			body, err := meta.DecodeArg(pkt.GetPayload())
			if err != nil {
				glog.Errorf("session %d %s DecodeArg err %s", s.id, meta.GetMessageID(), err)
				break
			}
			pkt.SetBody(body)
			evq.PostEvent(getNetMsgEvent(s, meta.GetMessageID(), pkt))
			protoc.PutPacket(pkt)
		} else {
			caller := s.pickRpcCaller(pkt.GetSeq())
			if caller == nil {
				glog.Warnf("session %d late rpc reply %d", s.id, pkt.GetSeq())
				continue
			}

			if msgType == protoc.MsgErr {
				caller.Done(nil, pkt.GetErrcode())
				continue
			}

			meta := protoc.GetMeta(caller.msgID.ID())
			if meta == nil {
				glog.Errorf("%d no meta", caller.msgID.ID())
				break
			}
			body, err := meta.DecodeReply(pkt.GetPayload())
			if err != nil {
				glog.Errorf("session %d %s DecodeReply err %s", s.id, meta.GetMessageID(), err)
				break
			}
			caller.Done(body, pkt.GetErrcode())
		}

	}

	if s.needNotifyWrite {
		// 通知发送线程停止
		s.Close()
	}

	// 通知接收线程ok
	s.endSync.Done()
}

func (s *Session) writePacket(item interface{}) error {
	if item == nil {
		return errors.Errorf("ses %d writeThread closing", s.id)
	}

	p, ok := item.(*protoc.Packet)
	if !ok {
		glog.Errorf("writeThread get %s not a protoc.Packet", item)
		return nil
	}

	var err error
	msgType := p.GetMsgType()
	if msgType == protoc.MsgPing || msgType == protoc.MsgPong || msgType == protoc.MsgErr {
		if err = s.proto.WritePacket(p); err != nil {
			protoc.PutPacket(p)
			return errors.Errorf("session %d WritePacket err %s", s.id, err)
		} else {
			protoc.PutPacket(p)
			return nil
		}
	}

	meta := protoc.GetMeta(p.GetMsgID())
	var payload []byte
	if meta == nil {
		if p.GetPayload() != nil {
			payload = p.GetPayload()
		} else {
			ok := false
			payload, ok = p.GetBody().([]byte)
			if !ok {
				protoc.PutPacket(p)
				return errors.Errorf("%d no meta", p.GetMsgID())
			}
		}
	} else {
		if msgType == protoc.MsgReq || msgType == protoc.MsgPush {
			payload, err = meta.EncodeArg(p.GetBody())
		} else {
			payload, err = meta.EncodeReply(p.GetBody())
		}
	}

	if err != nil {
		protoc.PutPacket(p)
		return errors.Errorf("session %d %s MarshalBody err %d", s.id, meta.GetMessageID(), err)
	}

	if s.compressor != nil {
		payload, err = s.compressor.Compress(payload)
		if err != nil {
			protoc.PutPacket(p)
			return errors.Errorf("session %d Compress err %d", s.id, err)
		}
	}

	if s.encryptor != nil {
		payload, err = s.encryptor.Encrypt(payload)
		if err != nil {
			protoc.PutPacket(p)
			return errors.Errorf("session %d Encrypt err %d", s.id, err)
		}
	}

	p.SetPayload(payload)

	if s.peer.cfg.WriteTimeout > 0 {
		s.conn.SetWriteDeadline(time.Now().Add(s.peer.cfg.WriteTimeout))
	}
	if err = s.proto.WritePacket(p); err != nil {
		protoc.PutPacket(p)
		return errors.Errorf("session %s WritePacket err %s", s.id, err)
	}

	protoc.PutPacket(p)
	return nil
}

func (s *Session) writeThread() {
	var err error
L1:
	for true {

		if s.syncWriteQueue != nil {

			item := s.syncWriteQueue.Pop()
			if err = s.writePacket(item); err != nil {
				glog.Errorf("%s", err)
				break
			}

		} else {

			time.Sleep(10 * time.Millisecond)
			s.writeQueueGuard.Lock()
			qlen := s.writeQueue.Length()
			if qlen == 0 {
				s.writeQueueGuard.Unlock()
				if err := s.proto.Flush(); err != nil {
					glog.Errorf("session %s Flush err %s", s.id, err)
					break
				}
				continue
			}

			packets := make([]interface{}, qlen)
			for i := 0; i < qlen; i++ {
				packets[i] = s.writeQueue.Remove()
			}
			s.writeQueueGuard.Unlock()

			for _, item := range packets {
				if err = s.writePacket(item); err != nil {
					break L1
				}
			}

			if err := s.proto.Flush(); err != nil {
				glog.Errorf("session %s Flush err %s", s.id, err)
				break
			}

		}
	}

	// 不需要读线程再次通知写线程
	s.needNotifyWrite = false
	s.proto.Flush()
	// 关闭socket,触发读错误, 结束读循环
	s.conn.Close()
	// 通知发送线程ok
	s.endSync.Done()
}

func (s *Session) CallAsync(msgID protoc.IMessageID, arg interface{}) chan *RpcResult {
	caller := newRpcCaller(msgID)
	s.rpcGuard.Lock()
	if !s.IsAlive() {
		caller.Done(nil, InternalErr.Errcode())
		s.rpcGuard.Unlock()
	} else {
		s.rpcSeq++
		seq := s.rpcSeq
		s.rpcPending[seq] = caller
		s.rpcGuard.Unlock()

		s.Write(protoc.GetReqPacket(seq, msgID, arg))
	}
	return caller.c
}

func (s *Session) Call(msgID protoc.IMessageID, arg interface{}) (reply interface{}, err error) {
	c := s.CallAsync(msgID, arg)
	var result *RpcResult
	evq.Await(func() {
		result = <-c
	})
	reply = result.Reply
	err = result.Err
	return
}

func (s *Session) Push(msgID protoc.IMessageID, arg interface{}) {
	if !s.IsAlive() {
		return
	}
	s.Write(protoc.GetPushPacket(msgID, arg))
}
