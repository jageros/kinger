// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/proto/pb"
	"kinger/gopuppy/network/protoc"
)

//@ C2S_FETCH_MAIL_LIST    req: FetchMailListArg    resp: MailList
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_MAIL_LIST_Meta struct {
}

func (m *C2S_FETCH_MAIL_LIST_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_MAIL_LIST
}

func (m *C2S_FETCH_MAIL_LIST_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.FetchMailListArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_MAIL_LIST_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_FETCH_MAIL_LIST_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.FetchMailListArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_FETCH_MAIL_LIST_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.MailList)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_MAIL_LIST_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_MAIL_LIST_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.MailList{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_MAIL_LIST END ----------------------------------------

//@ C2S_GET_MAIL_REWARD    req: GetMailRewardArg    resp: MailRewardReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_MAIL_REWARD_Meta struct {
}

func (m *C2S_GET_MAIL_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_MAIL_REWARD
}

func (m *C2S_GET_MAIL_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetMailRewardArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_MAIL_REWARD_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_GET_MAIL_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetMailRewardArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_GET_MAIL_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.MailRewardReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_MAIL_REWARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_MAIL_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.MailRewardReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_MAIL_REWARD END ----------------------------------------

//@ C2S_READ_MAIL    req: ReadMailArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_READ_MAIL_Meta struct {
}

func (m *C2S_READ_MAIL_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_READ_MAIL
}

func (m *C2S_READ_MAIL_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ReadMailArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_READ_MAIL_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_READ_MAIL_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ReadMailArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_READ_MAIL_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_READ_MAIL_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_READ_MAIL END ----------------------------------------

//@ C2S_GET_ALL_MAIL_REWARD    resp: MailRewardReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_ALL_MAIL_REWARD_Meta struct {
}

func (m *C2S_GET_ALL_MAIL_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_ALL_MAIL_REWARD
}

func (m *C2S_GET_ALL_MAIL_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_GET_ALL_MAIL_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_GET_ALL_MAIL_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.MailRewardReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_ALL_MAIL_REWARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_ALL_MAIL_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.MailRewardReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_ALL_MAIL_REWARD END ----------------------------------------

//@ S2C_NOTIFY_NEW_MAIL
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_NOTIFY_NEW_MAIL_Meta struct {
}

func (m *S2C_NOTIFY_NEW_MAIL_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_NOTIFY_NEW_MAIL
}

func (m *S2C_NOTIFY_NEW_MAIL_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_NOTIFY_NEW_MAIL_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *S2C_NOTIFY_NEW_MAIL_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_NOTIFY_NEW_MAIL_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_NOTIFY_NEW_MAIL END ----------------------------------------

//@ G2G_ON_SEND_WHOLE_SERVER_MAIL    req: WholeServerMail
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta struct {
}

func (m *G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_ON_SEND_WHOLE_SERVER_MAIL
}

func (m *G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WholeServerMail)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WholeServerMail{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_ON_SEND_WHOLE_SERVER_MAIL_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ G2G_ON_SEND_WHOLE_SERVER_MAIL END ----------------------------------------

//@ G2G_ON_UPDATE_WHOLE_SERVER_MAIL    req: WholeServerMail
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta struct {
}

func (m *G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_ON_UPDATE_WHOLE_SERVER_MAIL
}

func (m *G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WholeServerMail)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WholeServerMail{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_ON_UPDATE_WHOLE_SERVER_MAIL_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ G2G_ON_UPDATE_WHOLE_SERVER_MAIL END ----------------------------------------
