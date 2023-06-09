// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ S2C_ADD_OUT_STATUS    req: OutStatus
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_ADD_OUT_STATUS_Meta struct {
}

func (m *S2C_ADD_OUT_STATUS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_ADD_OUT_STATUS
}

func (m *S2C_ADD_OUT_STATUS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.OutStatus)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_ADD_OUT_STATUS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_ADD_OUT_STATUS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.OutStatus{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_ADD_OUT_STATUS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_ADD_OUT_STATUS_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_ADD_OUT_STATUS END ----------------------------------------

//@ S2C_DEL_OUT_STATUS    req: TargetOutStatus
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_DEL_OUT_STATUS_Meta struct {
}

func (m *S2C_DEL_OUT_STATUS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_DEL_OUT_STATUS
}

func (m *S2C_DEL_OUT_STATUS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TargetOutStatus)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_DEL_OUT_STATUS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_DEL_OUT_STATUS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TargetOutStatus{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_DEL_OUT_STATUS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_DEL_OUT_STATUS_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_DEL_OUT_STATUS END ----------------------------------------
