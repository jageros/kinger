package protoc

import (
	"encoding/binary"
	"errors"
	"io"

	_ "bufio"
	//"kinger/gopuppy/common/glog"
	"bufio"
	"kinger/gopuppy/common/utils"
)

var (
	packetTooBigErr   = errors.New("packetTooBigErr")
	packetTooSmallErr = errors.New("packetTooSmallErr")
	unknowMsgTypeErr  = errors.New("unknowMsgType")
)

const (
	SeqSize     = 4
	MsgIDSize   = 4
	MsgTypeSize = 1
	ErrcodeSize = 4
	PktSizeSize = 4
)

type (
	IProto interface {
		WritePacket(*Packet) error
		ReadPacket() (*Packet, error)
		Flush() error
	}
	IProtoFunc func(rw io.ReadWriter, maxPacketSize uint32, readbufSize, writeBufSize int) IProto

	IAuthor interface {
		DoAuth(io.ReadWriter) error
	}
	IAuthorFunc func() IAuthor

	IEncryptor interface {
		Init(io.ReadWriter) error
		Encrypt([]byte) ([]byte, error)
		Decrypt([]byte) ([]byte, error)
	}
	IEncryptorFunc func() IEncryptor

	ICompressor interface {
		Compress([]byte) ([]byte, error)
		Decompress([]byte) ([]byte, error)
	}
	ICompressorFunc func() ICompressor

	ProtocFactory struct {
		ProtoFunc      IProtoFunc
		AuthorFunc     IAuthorFunc
		EncryptorFunc  IEncryptorFunc
		CompressorFunc ICompressorFunc
	}
)

type DefaultProto struct {
	maxPacketSize uint32
	writeHeadBuf  []byte
	r             io.Reader
	w             io.Writer
	bw            *bufio.Writer
}

func NewDefaultProto(rw io.ReadWriter, maxPacketSize uint32, readbufSize, writeBufSize int) IProto {
	if maxPacketSize <= 0 {
		maxPacketSize = 1024 * 1024
	}
	dp := &DefaultProto{
		maxPacketSize: maxPacketSize,
		writeHeadBuf:  make([]byte, PktSizeSize+MsgTypeSize+MsgIDSize+SeqSize),
	}

	if readbufSize > 0 {
		dp.r = bufio.NewReaderSize(rw, readbufSize)
	} else {
		dp.r = rw
	}

	if writeBufSize > 0 {
		dp.bw = bufio.NewWriterSize(rw, writeBufSize)
		dp.w = dp.bw
	} else {
		dp.w = rw
	}

	return dp
}

func (dp *DefaultProto) WritePacket(p *Packet) error {
	var size uint32 = MsgTypeSize
	var err error
	dp.writeHeadBuf[PktSizeSize] = p.msgType
	if p.msgType == MsgPing || p.msgType == MsgPong {

	} else if p.msgType == MsgErr {
		binary.BigEndian.PutUint32(dp.writeHeadBuf[PktSizeSize+size:], uint32(p.errcode))
		size += ErrcodeSize
		binary.BigEndian.PutUint32(dp.writeHeadBuf[PktSizeSize+size:], uint32(p.seq))
		size += SeqSize
	} else if p.msgType == MsgReq || p.msgType == MsgPush {
		binary.BigEndian.PutUint32(dp.writeHeadBuf[PktSizeSize+size:], uint32(p.msgID))
		size += MsgIDSize
		if p.msgType == MsgReq {
			binary.BigEndian.PutUint32(dp.writeHeadBuf[PktSizeSize+size:], uint32(p.seq))
			size += SeqSize
		}
	} else if p.msgType == MsgReply {
		binary.BigEndian.PutUint32(dp.writeHeadBuf[PktSizeSize+size:], uint32(p.seq))
		size += SeqSize
	} else {
		return unknowMsgTypeErr
	}

	//glog.Debugf("WritePacket msgType=%d, msgID=%d, payload=%s", p.msgType, p.msgID, p.payload)

	headerSize := size
	var payload []byte
	if (p.msgType == MsgReq || p.msgType == MsgPush || p.msgType == MsgReply) && p.payload != nil {
		payload = p.payload
		size += uint32(len(p.payload))
	}

	if size > dp.maxPacketSize {
		return packetTooBigErr
	}

	fullPacketSize := size + PktSizeSize
	bb := utils.AcquireByteBuffer()
	bb.ChangeLen(int(fullPacketSize))
	binary.BigEndian.PutUint32(dp.writeHeadBuf, size)
	copy(bb.B, dp.writeHeadBuf[:headerSize+PktSizeSize])
	if payload != nil {
		copy(bb.B[headerSize+PktSizeSize:], payload)
	}

	// write header
	// binary.BigEndian.PutUint32(dp.writeHeadBuf, size)
	//if err = utils.WriteAll(dp.w, dp.writeHeadBuf[:headerSize+PktSizeSize]); err != nil {
	//	return err
	//}

	//glog.Debugf("write header %d", headerSize+PktSizeSize)
	//glog.Debugf("packet size %d", size)
	//glog.Debugf("packet headerSize %d", headerSize)

	// write body
	//if payload != nil {
	//glog.Debugf("write buff %d", len(buff))
	if err = utils.WriteAll(dp.w, bb.B); err != nil {
		utils.ReleaseByteBuffer(bb)
		return err
	}
	utils.ReleaseByteBuffer(bb)
	//}

	return nil
}

func (dp *DefaultProto) ReadPacket() (*Packet, error) {
	var size uint32
	err := binary.Read(dp.r, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}
	if size > dp.maxPacketSize {
		return nil, packetTooBigErr
	} else if size < MsgTypeSize {
		return nil, packetTooSmallErr
	}
	//glog.Debugf("ReadPacket size %d", size)

	p := GetPacket()
	err = binary.Read(dp.r, binary.BigEndian, &p.msgType)
	if err != nil {
		return nil, err
	}
	size -= MsgTypeSize

	//glog.Debugf("ReadPacket msgType %d", p.msgType)

	// read header
	if p.msgType == MsgPing || p.msgType == MsgPong {
		return p, nil
	} else if p.msgType == MsgErr {
		if size < ErrcodeSize+SeqSize {
			return nil, packetTooSmallErr
		}
		err = binary.Read(dp.r, binary.BigEndian, &p.errcode)
		if err != nil {
			return nil, err
		}
		err = binary.Read(dp.r, binary.BigEndian, &p.seq)
		if err != nil {
			return nil, err
		}
		return p, nil

	} else if p.msgType == MsgReq || p.msgType == MsgPush {
		if size < MsgIDSize {
			return nil, packetTooSmallErr
		}
		err = binary.Read(dp.r, binary.BigEndian, &p.msgID)
		if err != nil {
			return nil, err
		}
		size -= MsgIDSize

		if p.msgType == MsgReq {
			if size < SeqSize {
				return nil, packetTooSmallErr
			}
			err = binary.Read(dp.r, binary.BigEndian, &p.seq)
			if err != nil {
				return nil, err
			}
			size -= SeqSize
		}

	} else if p.msgType == MsgReply {
		if size < SeqSize {
			return nil, packetTooSmallErr
		}
		err = binary.Read(dp.r, binary.BigEndian, &p.seq)
		if err != nil {
			return nil, err
		}
		size -= SeqSize

	} else {
		return nil, unknowMsgTypeErr
	}

	// read body
	if size > 0 {
		p.payload = make([]byte, size)
		_, err = io.ReadFull(dp.r, p.payload)
		if err != nil {
			return nil, err
		}
	}
	return p, nil

}

func (dp *DefaultProto) Flush() error {
	if dp.bw != nil {
		return dp.bw.Flush()
	}
	return nil
}
