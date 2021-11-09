// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ C2S_GET_CAMP_ID    req: GetCampIDArg    resp: GetCampIDReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_CAMP_ID_Meta struct {
}

func (m *C2S_GET_CAMP_ID_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_CAMP_ID
}

func (m *C2S_GET_CAMP_ID_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetCampIDArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_CAMP_ID_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_GET_CAMP_ID_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetCampIDArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_GET_CAMP_ID_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.GetCampIDReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_CAMP_ID_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_CAMP_ID_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.GetCampIDReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_CAMP_ID END ----------------------------------------

//@ C2S_SET_CAMP_ID    req: SetCampIDArg    resp: SetCampIDReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_SET_CAMP_ID_Meta struct {
}

func (m *C2S_SET_CAMP_ID_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_SET_CAMP_ID
}

func (m *C2S_SET_CAMP_ID_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.SetCampIDArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SET_CAMP_ID_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_SET_CAMP_ID_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.SetCampIDArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_SET_CAMP_ID_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SetCampIDReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SET_CAMP_ID_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_SET_CAMP_ID_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SetCampIDReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_SET_CAMP_ID END ----------------------------------------

//@ C2S_START_TUTORIAL_BATTLE    req: StartTutorialBattleArg    resp: FightDesk
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_START_TUTORIAL_BATTLE_Meta struct {
}

func (m *C2S_START_TUTORIAL_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_START_TUTORIAL_BATTLE
}

func (m *C2S_START_TUTORIAL_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.StartTutorialBattleArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_START_TUTORIAL_BATTLE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_START_TUTORIAL_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.StartTutorialBattleArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_START_TUTORIAL_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.FightDesk)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_START_TUTORIAL_BATTLE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_START_TUTORIAL_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.FightDesk{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_START_TUTORIAL_BATTLE END ----------------------------------------

//@ S2C_TUTORIAL_FIGHT_END    req: TutorialFightEnd
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_TUTORIAL_FIGHT_END_Meta struct {
}

func (m *S2C_TUTORIAL_FIGHT_END_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_TUTORIAL_FIGHT_END
}

func (m *S2C_TUTORIAL_FIGHT_END_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TutorialFightEnd)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_TUTORIAL_FIGHT_END_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_TUTORIAL_FIGHT_END_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TutorialFightEnd{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_TUTORIAL_FIGHT_END_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_TUTORIAL_FIGHT_END_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_TUTORIAL_FIGHT_END END ----------------------------------------
