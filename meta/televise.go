// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/proto/pb"
	"kinger/gopuppy/network/protoc"
)

//@ S2C_PUSH_TELEVISE   req: Televise
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_PUSH_TELEVISE_Meta struct {
}

func (m *S2C_PUSH_TELEVISE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_PUSH_TELEVISE
}

func (m *S2C_PUSH_TELEVISE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.Televise)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_PUSH_TELEVISE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_PUSH_TELEVISE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.Televise{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_PUSH_TELEVISE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_PUSH_TELEVISE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_PUSH_TELEVISE END ----------------------------------------

