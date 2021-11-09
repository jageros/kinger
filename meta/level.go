// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ C2S_FETCH_LEVEL_INFO    resp: LevelInfo
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_LEVEL_INFO_Meta struct {
}

func (m *C2S_FETCH_LEVEL_INFO_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_LEVEL_INFO
}

func (m *C2S_FETCH_LEVEL_INFO_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_FETCH_LEVEL_INFO_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_FETCH_LEVEL_INFO_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.LevelInfo)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_LEVEL_INFO_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_LEVEL_INFO_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.LevelInfo{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_LEVEL_INFO END ----------------------------------------

//@ C2S_BEGIN_LEVEL_BATTLE    req: BeginLevelBattle    resp: LevelBattle
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BEGIN_LEVEL_BATTLE_Meta struct {
}

func (m *C2S_BEGIN_LEVEL_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BEGIN_LEVEL_BATTLE
}

func (m *C2S_BEGIN_LEVEL_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BeginLevelBattle)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BEGIN_LEVEL_BATTLE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BEGIN_LEVEL_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BeginLevelBattle{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BEGIN_LEVEL_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.LevelBattle)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BEGIN_LEVEL_BATTLE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BEGIN_LEVEL_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.LevelBattle{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BEGIN_LEVEL_BATTLE END ----------------------------------------

//@ C2S_LEVEL_READY_DONE    req: LevelChooseCard    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_LEVEL_READY_DONE_Meta struct {
}

func (m *C2S_LEVEL_READY_DONE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_LEVEL_READY_DONE
}

func (m *C2S_LEVEL_READY_DONE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.LevelChooseCard)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_LEVEL_READY_DONE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_LEVEL_READY_DONE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.LevelChooseCard{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_LEVEL_READY_DONE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_LEVEL_READY_DONE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_LEVEL_READY_DONE END ----------------------------------------

//@ C2S_OPEN_LEVEL_TREASURE    req: OpenLevelTreasureArg    resp: OpenTreasureReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_OPEN_LEVEL_TREASURE_Meta struct {
}

func (m *C2S_OPEN_LEVEL_TREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_OPEN_LEVEL_TREASURE
}

func (m *C2S_OPEN_LEVEL_TREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.OpenLevelTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_OPEN_LEVEL_TREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_OPEN_LEVEL_TREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.OpenLevelTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_OPEN_LEVEL_TREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.OpenTreasureReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_OPEN_LEVEL_TREASURE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_OPEN_LEVEL_TREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.OpenTreasureReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_OPEN_LEVEL_TREASURE END ----------------------------------------

//@ C2S_FETCH_LEVEL_HELP_RECORD    req: TargetLevel    resp: LevelHelpRecord
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_LEVEL_HELP_RECORD_Meta struct {
}

func (m *C2S_FETCH_LEVEL_HELP_RECORD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_LEVEL_HELP_RECORD
}

func (m *C2S_FETCH_LEVEL_HELP_RECORD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TargetLevel)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_LEVEL_HELP_RECORD_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_FETCH_LEVEL_HELP_RECORD_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TargetLevel{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_FETCH_LEVEL_HELP_RECORD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.LevelHelpRecord)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_LEVEL_HELP_RECORD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_LEVEL_HELP_RECORD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.LevelHelpRecord{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_LEVEL_HELP_RECORD END ----------------------------------------

//@ C2S_LEVEL_HELP_OTHER    req: LevelHelpArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_LEVEL_HELP_OTHER_Meta struct {
}

func (m *C2S_LEVEL_HELP_OTHER_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_LEVEL_HELP_OTHER
}

func (m *C2S_LEVEL_HELP_OTHER_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.LevelHelpArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_LEVEL_HELP_OTHER_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_LEVEL_HELP_OTHER_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.LevelHelpArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_LEVEL_HELP_OTHER_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_LEVEL_HELP_OTHER_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_LEVEL_HELP_OTHER END ----------------------------------------

//@ C2S_WATCH_HELP_VIDEO    req: WatchHelpVideoArg    resp: VideoBattleData
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_HELP_VIDEO_Meta struct {
}

func (m *C2S_WATCH_HELP_VIDEO_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_HELP_VIDEO
}

func (m *C2S_WATCH_HELP_VIDEO_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WatchHelpVideoArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_HELP_VIDEO_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WATCH_HELP_VIDEO_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WatchHelpVideoArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WATCH_HELP_VIDEO_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.VideoBattleData)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_HELP_VIDEO_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_WATCH_HELP_VIDEO_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.VideoBattleData{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_WATCH_HELP_VIDEO END ----------------------------------------

//@ C2S_FETCH_LEVEL_VIDEO_ID    req: FetchLevelVideoIDArg    resp: FetchLevelVideoIDRely
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_LEVEL_VIDEO_ID_Meta struct {
}

func (m *C2S_FETCH_LEVEL_VIDEO_ID_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_LEVEL_VIDEO_ID
}

func (m *C2S_FETCH_LEVEL_VIDEO_ID_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.FetchLevelVideoIDArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_LEVEL_VIDEO_ID_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_FETCH_LEVEL_VIDEO_ID_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.FetchLevelVideoIDArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_FETCH_LEVEL_VIDEO_ID_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.FetchLevelVideoIDRely)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_LEVEL_VIDEO_ID_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_LEVEL_VIDEO_ID_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.FetchLevelVideoIDRely{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_LEVEL_VIDEO_ID END ----------------------------------------

//@ C2S_CLEAR_LEVEL_CHAPTER    resp: LevelInfo
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_CLEAR_LEVEL_CHAPTER_Meta struct {
}

func (m *C2S_CLEAR_LEVEL_CHAPTER_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_CLEAR_LEVEL_CHAPTER
}

func (m *C2S_CLEAR_LEVEL_CHAPTER_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_CLEAR_LEVEL_CHAPTER_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_CLEAR_LEVEL_CHAPTER_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.LevelInfo)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_CLEAR_LEVEL_CHAPTER_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_CLEAR_LEVEL_CHAPTER_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.LevelInfo{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_CLEAR_LEVEL_CHAPTER END ----------------------------------------

//@ S2C_LEVEL_FIGHT_END    req: LevelFightResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_LEVEL_FIGHT_END_Meta struct {
}

func (m *S2C_LEVEL_FIGHT_END_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_LEVEL_FIGHT_END
}

func (m *S2C_LEVEL_FIGHT_END_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.LevelFightResult)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_LEVEL_FIGHT_END_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_LEVEL_FIGHT_END_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.LevelFightResult{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_LEVEL_FIGHT_END_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_LEVEL_FIGHT_END_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_LEVEL_FIGHT_END END ----------------------------------------

//@ S2C_CHAPTER_UNLOCK    req: ChapterUnlock
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_CHAPTER_UNLOCK_Meta struct {
}

func (m *S2C_CHAPTER_UNLOCK_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_CHAPTER_UNLOCK
}

func (m *S2C_CHAPTER_UNLOCK_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ChapterUnlock)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_CHAPTER_UNLOCK_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_CHAPTER_UNLOCK_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ChapterUnlock{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_CHAPTER_UNLOCK_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_CHAPTER_UNLOCK_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_CHAPTER_UNLOCK END ----------------------------------------

//@ S2C_LEVEL_BE_HELP    req: LevelBeHelpArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_LEVEL_BE_HELP_Meta struct {
}

func (m *S2C_LEVEL_BE_HELP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_LEVEL_BE_HELP
}

func (m *S2C_LEVEL_BE_HELP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.LevelBeHelpArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_LEVEL_BE_HELP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_LEVEL_BE_HELP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.LevelBeHelpArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_LEVEL_BE_HELP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_LEVEL_BE_HELP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_LEVEL_BE_HELP END ----------------------------------------

//@ S2C_LEVEL_ON_RECHARGE
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_LEVEL_ON_RECHARGE_Meta struct {
}

func (m *S2C_LEVEL_ON_RECHARGE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_LEVEL_ON_RECHARGE
}

func (m *S2C_LEVEL_ON_RECHARGE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_LEVEL_ON_RECHARGE_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *S2C_LEVEL_ON_RECHARGE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_LEVEL_ON_RECHARGE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_LEVEL_ON_RECHARGE END ----------------------------------------
