// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ C2S_WX_INVITE_BATTLE    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WX_INVITE_BATTLE_Meta struct {
}

func (m *C2S_WX_INVITE_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WX_INVITE_BATTLE
}

func (m *C2S_WX_INVITE_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WX_INVITE_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_WX_INVITE_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WX_INVITE_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WX_INVITE_BATTLE END ----------------------------------------

//@ C2S_WX_CANCEL_INVITE_BATTLE    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WX_CANCEL_INVITE_BATTLE_Meta struct {
}

func (m *C2S_WX_CANCEL_INVITE_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WX_CANCEL_INVITE_BATTLE
}

func (m *C2S_WX_CANCEL_INVITE_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WX_CANCEL_INVITE_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_WX_CANCEL_INVITE_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WX_CANCEL_INVITE_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WX_CANCEL_INVITE_BATTLE END ----------------------------------------

//@ C2S_WX_REPLY_INVITE_BATTLE    req: ReplyWxInviteBattleArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WX_REPLY_INVITE_BATTLE_Meta struct {
}

func (m *C2S_WX_REPLY_INVITE_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WX_REPLY_INVITE_BATTLE
}

func (m *C2S_WX_REPLY_INVITE_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ReplyWxInviteBattleArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WX_REPLY_INVITE_BATTLE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WX_REPLY_INVITE_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ReplyWxInviteBattleArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WX_REPLY_INVITE_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WX_REPLY_INVITE_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WX_REPLY_INVITE_BATTLE END ----------------------------------------

//@ C2S_GET_SHARE_TREASURE_REWARD    req: GetShareTreasureArg    resp: OpenTreasureReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_SHARE_TREASURE_REWARD_Meta struct {
}

func (m *C2S_GET_SHARE_TREASURE_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_SHARE_TREASURE_REWARD
}

func (m *C2S_GET_SHARE_TREASURE_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GetShareTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_SHARE_TREASURE_REWARD_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_GET_SHARE_TREASURE_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GetShareTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_GET_SHARE_TREASURE_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.OpenTreasureReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_SHARE_TREASURE_REWARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_SHARE_TREASURE_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.OpenTreasureReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_SHARE_TREASURE_REWARD END ----------------------------------------

//@ C2S_WXGAME_SHARE    req: WxgameShareArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WXGAME_SHARE_Meta struct {
}

func (m *C2S_WXGAME_SHARE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WXGAME_SHARE
}

func (m *C2S_WXGAME_SHARE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WxgameShareArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WXGAME_SHARE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WXGAME_SHARE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WxgameShareArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WXGAME_SHARE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WXGAME_SHARE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WXGAME_SHARE END ----------------------------------------

//@ C2S_CLICK_WXGAME_SHARE    req: ClickWxgameShareArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_CLICK_WXGAME_SHARE_Meta struct {
}

func (m *C2S_CLICK_WXGAME_SHARE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_CLICK_WXGAME_SHARE
}

func (m *C2S_CLICK_WXGAME_SHARE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ClickWxgameShareArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_CLICK_WXGAME_SHARE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_CLICK_WXGAME_SHARE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ClickWxgameShareArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_CLICK_WXGAME_SHARE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_CLICK_WXGAME_SHARE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_CLICK_WXGAME_SHARE END ----------------------------------------

//@ C2S_FETCH_DAILY_SHARE_INFO    resp: DailyShareInfo
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_DAILY_SHARE_INFO_Meta struct {
}

func (m *C2S_FETCH_DAILY_SHARE_INFO_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_DAILY_SHARE_INFO
}

func (m *C2S_FETCH_DAILY_SHARE_INFO_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_FETCH_DAILY_SHARE_INFO_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_FETCH_DAILY_SHARE_INFO_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.DailyShareInfo)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_DAILY_SHARE_INFO_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_DAILY_SHARE_INFO_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.DailyShareInfo{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_DAILY_SHARE_INFO END ----------------------------------------

//@ C2S_GET_DAILY_SHARE_REWARD    resp: DailyShareReward
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GET_DAILY_SHARE_REWARD_Meta struct {
}

func (m *C2S_GET_DAILY_SHARE_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GET_DAILY_SHARE_REWARD
}

func (m *C2S_GET_DAILY_SHARE_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_GET_DAILY_SHARE_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_GET_DAILY_SHARE_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.DailyShareReward)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GET_DAILY_SHARE_REWARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GET_DAILY_SHARE_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.DailyShareReward{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GET_DAILY_SHARE_REWARD END ----------------------------------------

//@ C2S_SHARE_TREASURE    req: ShareTreasureArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_SHARE_TREASURE_Meta struct {
}

func (m *C2S_SHARE_TREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_SHARE_TREASURE
}

func (m *C2S_SHARE_TREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ShareTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SHARE_TREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_SHARE_TREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ShareTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_SHARE_TREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_SHARE_TREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_SHARE_TREASURE END ----------------------------------------

//@ C2S_HELP_SHARE_TREASURE    req: HelpShareTreasureArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_HELP_SHARE_TREASURE_Meta struct {
}

func (m *C2S_HELP_SHARE_TREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_HELP_SHARE_TREASURE
}

func (m *C2S_HELP_SHARE_TREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.HelpShareTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_HELP_SHARE_TREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_HELP_SHARE_TREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.HelpShareTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_HELP_SHARE_TREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_HELP_SHARE_TREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_HELP_SHARE_TREASURE END ----------------------------------------

//@ C2S_SHARE_BATTLE_LOSE    req: ShareBattleLoseArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_SHARE_BATTLE_LOSE_Meta struct {
}

func (m *C2S_SHARE_BATTLE_LOSE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_SHARE_BATTLE_LOSE
}

func (m *C2S_SHARE_BATTLE_LOSE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.ShareBattleLoseArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SHARE_BATTLE_LOSE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_SHARE_BATTLE_LOSE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.ShareBattleLoseArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_SHARE_BATTLE_LOSE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_SHARE_BATTLE_LOSE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_SHARE_BATTLE_LOSE END ----------------------------------------

//@ C2S_END_SHARE_BATTLE_LOSE
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_END_SHARE_BATTLE_LOSE_Meta struct {
}

func (m *C2S_END_SHARE_BATTLE_LOSE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_END_SHARE_BATTLE_LOSE
}

func (m *C2S_END_SHARE_BATTLE_LOSE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_END_SHARE_BATTLE_LOSE_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_END_SHARE_BATTLE_LOSE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_END_SHARE_BATTLE_LOSE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_END_SHARE_BATTLE_LOSE END ----------------------------------------

//@ C2S_HELP_SHARE_BATTLE_LOSE    req: HelpShareBattleLoseArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_HELP_SHARE_BATTLE_LOSE_Meta struct {
}

func (m *C2S_HELP_SHARE_BATTLE_LOSE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_HELP_SHARE_BATTLE_LOSE
}

func (m *C2S_HELP_SHARE_BATTLE_LOSE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.HelpShareBattleLoseArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_HELP_SHARE_BATTLE_LOSE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_HELP_SHARE_BATTLE_LOSE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.HelpShareBattleLoseArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_HELP_SHARE_BATTLE_LOSE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_HELP_SHARE_BATTLE_LOSE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_HELP_SHARE_BATTLE_LOSE END ----------------------------------------

//@ S2C_WX_INVITE_BATTLE_RESULT    req: WxInviteBattleResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_WX_INVITE_BATTLE_RESULT_Meta struct {
}

func (m *S2C_WX_INVITE_BATTLE_RESULT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_WX_INVITE_BATTLE_RESULT
}

func (m *S2C_WX_INVITE_BATTLE_RESULT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WxInviteBattleResult)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_WX_INVITE_BATTLE_RESULT_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_WX_INVITE_BATTLE_RESULT_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WxInviteBattleResult{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_WX_INVITE_BATTLE_RESULT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_WX_INVITE_BATTLE_RESULT_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_WX_INVITE_BATTLE_RESULT END ----------------------------------------

//@ S2C_DAILY_TREASURE_BE_HELP
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_DAILY_TREASURE_BE_HELP_Meta struct {
}

func (m *S2C_DAILY_TREASURE_BE_HELP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_DAILY_TREASURE_BE_HELP
}

func (m *S2C_DAILY_TREASURE_BE_HELP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_DAILY_TREASURE_BE_HELP_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *S2C_DAILY_TREASURE_BE_HELP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_DAILY_TREASURE_BE_HELP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_DAILY_TREASURE_BE_HELP END ----------------------------------------

//@ S2C_TREASURE_BE_HELP    req: TreasureBeHelp
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_TREASURE_BE_HELP_Meta struct {
}

func (m *S2C_TREASURE_BE_HELP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_TREASURE_BE_HELP
}

func (m *S2C_TREASURE_BE_HELP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TreasureBeHelp)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_TREASURE_BE_HELP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_TREASURE_BE_HELP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TreasureBeHelp{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_TREASURE_BE_HELP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_TREASURE_BE_HELP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_TREASURE_BE_HELP END ----------------------------------------

//@ S2C_BATTLE_LOSE_BE_HELP    req: BattleLoseBeHelp
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_BATTLE_LOSE_BE_HELP_Meta struct {
}

func (m *S2C_BATTLE_LOSE_BE_HELP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_BATTLE_LOSE_BE_HELP
}

func (m *S2C_BATTLE_LOSE_BE_HELP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BattleLoseBeHelp)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_BATTLE_LOSE_BE_HELP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_BATTLE_LOSE_BE_HELP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BattleLoseBeHelp{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_BATTLE_LOSE_BE_HELP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_BATTLE_LOSE_BE_HELP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_BATTLE_LOSE_BE_HELP END ----------------------------------------

//@ S2C_UPDATE_WX_EXAMINE_STATE    req: WxExamineState
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_UPDATE_WX_EXAMINE_STATE_Meta struct {
}

func (m *S2C_UPDATE_WX_EXAMINE_STATE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_UPDATE_WX_EXAMINE_STATE
}

func (m *S2C_UPDATE_WX_EXAMINE_STATE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WxExamineState)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_UPDATE_WX_EXAMINE_STATE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_UPDATE_WX_EXAMINE_STATE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WxExamineState{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_UPDATE_WX_EXAMINE_STATE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_UPDATE_WX_EXAMINE_STATE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_UPDATE_WX_EXAMINE_STATE END ----------------------------------------

//@ S2C_DAILY_SHARE_RETURN_REWARD    req: DailyShareReturnReward
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_DAILY_SHARE_RETURN_REWARD_Meta struct {
}

func (m *S2C_DAILY_SHARE_RETURN_REWARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_DAILY_SHARE_RETURN_REWARD
}

func (m *S2C_DAILY_SHARE_RETURN_REWARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.DailyShareReturnReward)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_DAILY_SHARE_RETURN_REWARD_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_DAILY_SHARE_RETURN_REWARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.DailyShareReturnReward{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_DAILY_SHARE_RETURN_REWARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_DAILY_SHARE_RETURN_REWARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_DAILY_SHARE_RETURN_REWARD END ----------------------------------------

//@ G2G_WX_REPLY_INVITE_BATTLE    req: G2GReplyWxInviteBattleArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type G2G_WX_REPLY_INVITE_BATTLE_Meta struct {
}

func (m *G2G_WX_REPLY_INVITE_BATTLE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_G2G_WX_REPLY_INVITE_BATTLE
}

func (m *G2G_WX_REPLY_INVITE_BATTLE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.G2GReplyWxInviteBattleArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("G2G_WX_REPLY_INVITE_BATTLE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *G2G_WX_REPLY_INVITE_BATTLE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.G2GReplyWxInviteBattleArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *G2G_WX_REPLY_INVITE_BATTLE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *G2G_WX_REPLY_INVITE_BATTLE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ G2G_WX_REPLY_INVITE_BATTLE END ----------------------------------------

//@ C2S_CANCEL_WX_SHARE    req: CancelWxShareArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_CANCEL_WX_SHARE_Meta struct {
}

func (m *C2S_CANCEL_WX_SHARE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_CANCEL_WX_SHARE
}

func (m *C2S_CANCEL_WX_SHARE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.CancelWxShareArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_CANCEL_WX_SHARE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_CANCEL_WX_SHARE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.CancelWxShareArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_CANCEL_WX_SHARE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_CANCEL_WX_SHARE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_CANCEL_WX_SHARE END ----------------------------------------

//@ S2C_WX_SHARE_BE_HELP    req: WxShareBeHelpArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_WX_SHARE_BE_HELP_Meta struct {
}

func (m *S2C_WX_SHARE_BE_HELP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_WX_SHARE_BE_HELP
}

func (m *S2C_WX_SHARE_BE_HELP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WxShareBeHelpArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_WX_SHARE_BE_HELP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_WX_SHARE_BE_HELP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WxShareBeHelpArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_WX_SHARE_BE_HELP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_WX_SHARE_BE_HELP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_WX_SHARE_BE_HELP END ----------------------------------------

//@ C2S_IOS_SHARE    resp: DailyShareReward
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_IOS_SHARE_Meta struct {
}

func (m *C2S_IOS_SHARE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_IOS_SHARE
}

func (m *C2S_IOS_SHARE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_IOS_SHARE_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_IOS_SHARE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.DailyShareReward)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_IOS_SHARE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_IOS_SHARE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.DailyShareReward{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_IOS_SHARE END ----------------------------------------
