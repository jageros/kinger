package protoc

import (
	"sync"
)

type IMessageID interface {
	ID() int32
	String() string
}

type MessageType = byte

const (
	MsgReq MessageType = 1 + iota // need reply
	MsgReply
	MsgPush // no reply
	MsgErr  // error reply
	MsgPing
	MsgPong
)

var (
	PingPacket = &Packet{
		msgType: MsgPing,
	}
	PongPacket = &Packet{
		msgType: MsgPong,
	}
)

type Packet struct {
	seq     uint32
	msgID   int32
	msgType MessageType
	errcode int32
	body    interface{}
	payload []byte
	next    *Packet
}

var packetStack = new(struct {
	freePacket *Packet
	mu         sync.Mutex
})

func GetPacket() *Packet {
	packetStack.mu.Lock()
	p := packetStack.freePacket
	if p == nil {
		p = &Packet{
			payload: nil,
		}
	} else {
		packetStack.freePacket = p.next
	}
	packetStack.mu.Unlock()
	return p
}

func PutPacket(p *Packet) {
	if p == PingPacket || p == PongPacket {
		return
	}
	packetStack.mu.Lock()
	p.Reset()
	p.next = packetStack.freePacket
	packetStack.freePacket = p
	packetStack.mu.Unlock()
}

func GetReqPacket(seq uint32, msgID IMessageID, body interface{}) *Packet {
	p := GetPacket()
	p.seq = seq
	p.msgID = msgID.ID()
	p.msgType = MsgReq
	p.body = body
	return p
}

func GetPushPacket(msgID IMessageID, body interface{}) *Packet {
	p := GetPacket()
	p.msgID = msgID.ID()
	p.msgType = MsgPush
	p.body = body
	return p
}

func GetReplyPacket(msgID int32, seq uint32, body interface{}) *Packet {
	p := GetPacket()
	p.seq = seq
	p.msgID = msgID
	p.msgType = MsgReply
	p.body = body
	p.payload = nil
	return p
}

func GetErrReplyPacket(seq uint32, errcode int32) *Packet {
	p := GetPacket()
	p.seq = seq
	p.msgType = MsgErr
	p.errcode = errcode
	return p
}

func (p *Packet) Reset() {
	p.seq = 0
	p.msgID = 0
	p.msgType = 0
	p.errcode = 0
	p.body = nil
	p.payload = nil
	p.next = nil
}

func (p *Packet) GetSeq() uint32 {
	return p.seq
}

func (p *Packet) GetMsgID() int32 {
	return p.msgID
}

func (p *Packet) GetMsgType() MessageType {
	return p.msgType
}

func (p *Packet) GetErrcode() int32 {
	return p.errcode
}

func (p *Packet) GetPayload() []byte {
	return p.payload
}

func (p *Packet) SetPayload(payload []byte) {
	p.payload = payload
}

func (p *Packet) GetBody() interface{} {
	return p.body
}

func (p *Packet) SetBody(body interface{}) {
	p.body = body
}
