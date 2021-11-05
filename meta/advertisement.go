// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/proto/pb"
	"kinger/gopuppy/network/protoc"
)

//@ C2S_DAILY_TREASURE_READ_ADS    req: DailyTreasureReadAdsArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_DAILY_TREASURE_READ_ADS_Meta struct {
}

func (m *C2S_DAILY_TREASURE_READ_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_DAILY_TREASURE_READ_ADS
}

func (m *C2S_DAILY_TREASURE_READ_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.DailyTreasureReadAdsArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_DAILY_TREASURE_READ_ADS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_DAILY_TREASURE_READ_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.DailyTreasureReadAdsArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_DAILY_TREASURE_READ_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_DAILY_TREASURE_READ_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_DAILY_TREASURE_READ_ADS END ----------------------------------------

//@ C2S_TREASURE_READ_ADS    req: TreasureReadAdsArg    resp: TreasureReadAdsReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_TREASURE_READ_ADS_Meta struct {
}

func (m *C2S_TREASURE_READ_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_TREASURE_READ_ADS
}

func (m *C2S_TREASURE_READ_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TreasureReadAdsArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_TREASURE_READ_ADS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_TREASURE_READ_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TreasureReadAdsArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_TREASURE_READ_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.TreasureReadAdsReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_TREASURE_READ_ADS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_TREASURE_READ_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.TreasureReadAdsReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_TREASURE_READ_ADS END ----------------------------------------

//@ C2S_BATTLE_LOSE_READ_ADS    resp: BattleLoseReadAdsReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BATTLE_LOSE_READ_ADS_Meta struct {
}

func (m *C2S_BATTLE_LOSE_READ_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BATTLE_LOSE_READ_ADS
}

func (m *C2S_BATTLE_LOSE_READ_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_BATTLE_LOSE_READ_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_BATTLE_LOSE_READ_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BattleLoseReadAdsReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BATTLE_LOSE_READ_ADS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BATTLE_LOSE_READ_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BattleLoseReadAdsReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BATTLE_LOSE_READ_ADS END ----------------------------------------

//@ C2S_WATCH_UP_TREASURE_RARE_ADS    req: WatchUpTreasureRareAdsArg    resp: Treasure
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_UP_TREASURE_RARE_ADS_Meta struct {
}

func (m *C2S_WATCH_UP_TREASURE_RARE_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_UP_TREASURE_RARE_ADS
}

func (m *C2S_WATCH_UP_TREASURE_RARE_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WatchUpTreasureRareAdsArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_UP_TREASURE_RARE_ADS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WATCH_UP_TREASURE_RARE_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WatchUpTreasureRareAdsArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WATCH_UP_TREASURE_RARE_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.Treasure)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_UP_TREASURE_RARE_ADS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_WATCH_UP_TREASURE_RARE_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.Treasure{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_WATCH_UP_TREASURE_RARE_ADS END ----------------------------------------

//@ C2S_WATCH_SHOP_FREE_ADS    req: WatchShopFreeAdsArg   resp: WatchShopFreeAdsReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_SHOP_FREE_ADS_Meta struct {
}

func (m *C2S_WATCH_SHOP_FREE_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_SHOP_FREE_ADS
}

func (m *C2S_WATCH_SHOP_FREE_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WatchShopFreeAdsArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_SHOP_FREE_ADS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WATCH_SHOP_FREE_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WatchShopFreeAdsArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WATCH_SHOP_FREE_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.WatchShopFreeAdsReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_SHOP_FREE_ADS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_WATCH_SHOP_FREE_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.WatchShopFreeAdsReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_WATCH_SHOP_FREE_ADS END ----------------------------------------

//@ C2S_WATCH_TREASURE_ADD_CARD_ADS    req: WatchTreasureAddCardAdsArg    resp: WatchTreasureAddCardAdsReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta struct {
}

func (m *C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_TREASURE_ADD_CARD_ADS
}

func (m *C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.WatchTreasureAddCardAdsArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.WatchTreasureAddCardAdsArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.WatchTreasureAddCardAdsReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_WATCH_TREASURE_ADD_CARD_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.WatchTreasureAddCardAdsReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_WATCH_TREASURE_ADD_CARD_ADS END ----------------------------------------

//@ C2S_WATCH_ADS_BEGIN
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_ADS_BEGIN_Meta struct {
}

func (m *C2S_WATCH_ADS_BEGIN_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_ADS_BEGIN
}

func (m *C2S_WATCH_ADS_BEGIN_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_BEGIN_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_BEGIN_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_BEGIN_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WATCH_ADS_BEGIN END ----------------------------------------

//@ C2S_WATCH_ADS_END
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_WATCH_ADS_END_Meta struct {
}

func (m *C2S_WATCH_ADS_END_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_WATCH_ADS_END
}

func (m *C2S_WATCH_ADS_END_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_END_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_END_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_WATCH_ADS_END_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_WATCH_ADS_END END ----------------------------------------

//@ S2C_TRIGGER_SHOP_ADD_GOLD_ADS
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta struct {
}

func (m *S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_TRIGGER_SHOP_ADD_GOLD_ADS
}

func (m *S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_TRIGGER_SHOP_ADD_GOLD_ADS_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_TRIGGER_SHOP_ADD_GOLD_ADS END ----------------------------------------
