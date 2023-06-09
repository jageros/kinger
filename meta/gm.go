// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ G2G_GET_ONLINE_INFO    resp: OnlineInfo
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_GET_ONLINE_INFO_Meta struct {
}

func (m *G2G_GET_ONLINE_INFO_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_GET_ONLINE_INFO
}

func (m *G2G_GET_ONLINE_INFO_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_GET_ONLINE_INFO_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *G2G_GET_ONLINE_INFO_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.OnlineInfo)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_ONLINE_INFO_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *G2G_GET_ONLINE_INFO_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.OnlineInfo{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ G2G_GET_ONLINE_INFO END ----------------------------------------

//@ G2G_GET_CARD_AMOUNT_LOG    req: GetCardAmountLogArg    resp: GetCardAmountLogReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_GET_CARD_AMOUNT_LOG_Meta struct {
}

func (m *G2G_GET_CARD_AMOUNT_LOG_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_GET_CARD_AMOUNT_LOG
}

func (m *G2G_GET_CARD_AMOUNT_LOG_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetCardAmountLogArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_AMOUNT_LOG_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_GET_CARD_AMOUNT_LOG_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetCardAmountLogArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_GET_CARD_AMOUNT_LOG_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.GetCardAmountLogReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_AMOUNT_LOG_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *G2G_GET_CARD_AMOUNT_LOG_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.GetCardAmountLogReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ G2G_GET_CARD_AMOUNT_LOG END ----------------------------------------

//@ G2G_GET_CARD_LEVEL_LOG    req: GetCardLevelLogArg    resp: GetCardLevelLogReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_GET_CARD_LEVEL_LOG_Meta struct {
}

func (m *G2G_GET_CARD_LEVEL_LOG_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_GET_CARD_LEVEL_LOG
}

func (m *G2G_GET_CARD_LEVEL_LOG_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetCardLevelLogArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_LEVEL_LOG_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_GET_CARD_LEVEL_LOG_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetCardLevelLogArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_GET_CARD_LEVEL_LOG_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.GetCardLevelLogReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_LEVEL_LOG_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *G2G_GET_CARD_LEVEL_LOG_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.GetCardLevelLogReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ G2G_GET_CARD_LEVEL_LOG END ----------------------------------------

//@ G2G_GET_CARD_POOL_LOG    req: GetCardPoolLogArg    resp: GetCardPoolLogReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_GET_CARD_POOL_LOG_Meta struct {
}

func (m *G2G_GET_CARD_POOL_LOG_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_GET_CARD_POOL_LOG
}

func (m *G2G_GET_CARD_POOL_LOG_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetCardPoolLogArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_POOL_LOG_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_GET_CARD_POOL_LOG_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetCardPoolLogArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_GET_CARD_POOL_LOG_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.GetCardPoolLogReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_GET_CARD_POOL_LOG_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *G2G_GET_CARD_POOL_LOG_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.GetCardPoolLogReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ G2G_GET_CARD_POOL_LOG END ----------------------------------------

//@ G2G_ON_SERVER_STATUS_UPDATE    req: ServerStatus
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_ON_SERVER_STATUS_UPDATE_Meta struct {
}

func (m *G2G_ON_SERVER_STATUS_UPDATE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_ON_SERVER_STATUS_UPDATE
}

func (m *G2G_ON_SERVER_STATUS_UPDATE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ServerStatus)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_ON_SERVER_STATUS_UPDATE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_ON_SERVER_STATUS_UPDATE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ServerStatus{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_ON_SERVER_STATUS_UPDATE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_ON_SERVER_STATUS_UPDATE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ G2G_ON_SERVER_STATUS_UPDATE END ----------------------------------------

//@ G2G_ON_LOGIN_NOTICE_UPDATE    req: GmLoginNotice
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_ON_LOGIN_NOTICE_UPDATE_Meta struct {
}

func (m *G2G_ON_LOGIN_NOTICE_UPDATE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_ON_LOGIN_NOTICE_UPDATE
}

func (m *G2G_ON_LOGIN_NOTICE_UPDATE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GmLoginNotice)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_ON_LOGIN_NOTICE_UPDATE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_ON_LOGIN_NOTICE_UPDATE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GmLoginNotice{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_ON_LOGIN_NOTICE_UPDATE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_ON_LOGIN_NOTICE_UPDATE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ G2G_ON_LOGIN_NOTICE_UPDATE END ----------------------------------------
