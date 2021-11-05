// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/proto/pb"
	"kinger/gopuppy/network/protoc"
)

//@ C2S_FETCH_SURVEY_INFO    resp: SurveyInfo
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_SURVEY_INFO_Meta struct {
}

func (m *C2S_FETCH_SURVEY_INFO_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_SURVEY_INFO
}

func (m *C2S_FETCH_SURVEY_INFO_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_FETCH_SURVEY_INFO_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_FETCH_SURVEY_INFO_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SurveyInfo)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_SURVEY_INFO_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_SURVEY_INFO_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SurveyInfo{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_SURVEY_INFO END ----------------------------------------

//@ C2S_COMPLETE_SURVEY    req: SurveyAnswer
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_COMPLETE_SURVEY_Meta struct {
}

func (m *C2S_COMPLETE_SURVEY_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_COMPLETE_SURVEY
}

func (m *C2S_COMPLETE_SURVEY_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.SurveyAnswer)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_COMPLETE_SURVEY_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_COMPLETE_SURVEY_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.SurveyAnswer{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_COMPLETE_SURVEY_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_COMPLETE_SURVEY_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_COMPLETE_SURVEY END ----------------------------------------

//@ C2S_GET_SURVEY_REWARD    resp: OpenTreasureReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_SURVEY_REWARD_Meta struct {
}

func (m *C2S_GET_SURVEY_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_SURVEY_REWARD
}

func (m *C2S_GET_SURVEY_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_GET_SURVEY_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_GET_SURVEY_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.OpenTreasureReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_SURVEY_REWARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_SURVEY_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.OpenTreasureReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_SURVEY_REWARD END ----------------------------------------
