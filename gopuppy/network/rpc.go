package network

import (
	"fmt"

	//"encoding/json"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network/protoc"
	"sync"
)

type (
	rpcCaller struct {
		msgID protoc.IMessageID
		c     chan *RpcResult
	}

	RpcResult struct {
		Reply interface{}
		Err   IRpcError
	}

	IRpcError interface {
		error
		Errcode() int32
	}

	RpcError int32

	RpcHandler func(ses *Session, arg interface{}) (reply interface{}, err error)

	netMsgEvent struct {
		msgID   protoc.IMessageID
		ses     *Session
		seq     uint32
		msgType protoc.MessageType
		body    interface{}
		next    *netMsgEvent
	}
)

const (
	Success     RpcError = 0
	InternalErr RpcError = -1
)

func (re RpcError) Error() string {
	return fmt.Sprintf("RpcError %d", re)
}

func (re RpcError) Errcode() int32 {
	return int32(re)
}

func (rc *rpcCaller) Done(reply interface{}, errcode int32) {
	if errcode != Success.Errcode() {
		rc.c <- &RpcResult{Err: RpcError(errcode)}
	} else {
		rc.c <- &RpcResult{Reply: reply, Err: nil}
	}
}

func newRpcCaller(msgID protoc.IMessageID) *rpcCaller {
	return &rpcCaller{
		msgID: msgID,
		c:     make(chan *RpcResult, 1),
	}
}

func (n *netMsgEvent) GetEventId() int {
	return consts.NET_MSG_EVENT
}

var netMsgEventStack = new(struct {
	freeEvent *netMsgEvent
	mu        sync.Mutex
})

func getNetMsgEvent(ses *Session, msgID protoc.IMessageID, packet *protoc.Packet) *netMsgEvent {
	netMsgEventStack.mu.Lock()
	evt := netMsgEventStack.freeEvent
	if evt == nil {
		evt = &netMsgEvent{}
	} else {
		netMsgEventStack.freeEvent = evt.next
	}
	netMsgEventStack.mu.Unlock()
	evt.msgID = msgID
	evt.ses = ses
	evt.seq = packet.GetSeq()
	evt.msgType = packet.GetMsgType()
	evt.body = packet.GetBody()
	return evt
}

func putNetMsgEvent(evt *netMsgEvent) {
	netMsgEventStack.mu.Lock()
	evt.ses = nil
	evt.body = nil
	evt.next = netMsgEventStack.freeEvent
	netMsgEventStack.freeEvent = evt
	netMsgEventStack.mu.Unlock()
}

func onRecvNetMsg(ev evq.IEvent) {
	netEv, ok := ev.(*netMsgEvent)
	if !ok {
		glog.Errorf("onRecvNetMsg %s not a netMsgEvent", ev)
		return
	}

	peer := netEv.ses.getPeer()
	handler := peer.getRpcHandler(netEv.msgID)
	if handler == nil {
		glog.Errorf("onRecvNetMsg %s no handler", netEv.msgID)
		if netEv.msgType == protoc.MsgReq && netEv.ses.IsAlive() {
			netEv.ses.Write(protoc.GetErrReplyPacket(netEv.seq, InternalErr.Errcode()))
		}
		putNetMsgEvent(netEv)
		return
	}

	glog.Debugf("rpc call %s", netEv.msgID)
	reply, err := handler(netEv.ses, netEv.body)
	if netEv.msgType == protoc.MsgReq && netEv.ses.IsAlive() {
		glog.Debugf("rpc reply %s %s", netEv.msgID, err)
		//replyJson, _ := json.Marshal(reply)
		//glog.Debugf("rpc reply %s %s %s", netEv.msgID, replyJson, err)
		if err != nil {
			rerr, ok := err.(IRpcError)
			if !ok {
				rerr = InternalErr
			}
			netEv.ses.Write(protoc.GetErrReplyPacket(netEv.seq, rerr.Errcode()))
		} else {
			netEv.ses.Write(protoc.GetReplyPacket(netEv.msgID.ID(), netEv.seq, reply))
		}
	}
	putNetMsgEvent(netEv)
}

func init() {
	evq.HandleEvent(consts.NET_MSG_EVENT, onRecvNetMsg)
}
