// Generated by gen_meta.py
// DO NOT EDIT!

package meta

import (
	"errors"

	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
)

//@ C2S_FETCH_SHOP_DATA    resp: ShopData
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_SHOP_DATA_Meta struct {
}

func (m *C2S_FETCH_SHOP_DATA_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_SHOP_DATA
}

func (m *C2S_FETCH_SHOP_DATA_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_FETCH_SHOP_DATA_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_FETCH_SHOP_DATA_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.ShopData)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_SHOP_DATA_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_SHOP_DATA_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.ShopData{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_SHOP_DATA END ----------------------------------------

//@ C2S_BUY_LIMIT_GITF    req: BuyLimitGiftArg    resp: BuyLimitGiftReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_LIMIT_GITF_Meta struct {
}

func (m *C2S_BUY_LIMIT_GITF_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_LIMIT_GITF
}

func (m *C2S_BUY_LIMIT_GITF_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyLimitGiftArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_LIMIT_GITF_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_LIMIT_GITF_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyLimitGiftArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_LIMIT_GITF_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyLimitGiftReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_LIMIT_GITF_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_LIMIT_GITF_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyLimitGiftReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_LIMIT_GITF END ----------------------------------------

//@ C2S_BUY_JADE    req: BuyJadeArg    resp: BuyJadeReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_JADE_Meta struct {
}

func (m *C2S_BUY_JADE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_JADE
}

func (m *C2S_BUY_JADE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyJadeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_JADE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_JADE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyJadeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_JADE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyJadeReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_JADE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_JADE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyJadeReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_JADE END ----------------------------------------

//@ C2S_BUY_SOLDTREASURE    req: BuySoldTreasureArg    resp: BuySoldTreasureReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_SOLDTREASURE_Meta struct {
}

func (m *C2S_BUY_SOLDTREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_SOLDTREASURE
}

func (m *C2S_BUY_SOLDTREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuySoldTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_SOLDTREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_SOLDTREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuySoldTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_SOLDTREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuySoldTreasureReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_SOLDTREASURE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_SOLDTREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuySoldTreasureReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_SOLDTREASURE END ----------------------------------------

//@ C2S_BUY_RANDOM_SHOP    req: BuyRandomShopArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_RANDOM_SHOP_Meta struct {
}

func (m *C2S_BUY_RANDOM_SHOP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_RANDOM_SHOP
}

func (m *C2S_BUY_RANDOM_SHOP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyRandomShopArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RANDOM_SHOP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_RANDOM_SHOP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyRandomShopArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_RANDOM_SHOP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_BUY_RANDOM_SHOP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_BUY_RANDOM_SHOP END ----------------------------------------

//@ C2S_BUY_GOLD    req: BuyGoldArg    resp: BuyGoldReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_GOLD_Meta struct {
}

func (m *C2S_BUY_GOLD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_GOLD
}

func (m *C2S_BUY_GOLD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyGoldArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_GOLD_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_GOLD_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyGoldArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_GOLD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyGoldReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_GOLD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_GOLD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyGoldReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_GOLD END ----------------------------------------

//@ C2S_SDK_CREATE_ORDER    req: SdkCreateOrderArg    resp: SdkCreateOrderReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_SDK_CREATE_ORDER_Meta struct {
}

func (m *C2S_SDK_CREATE_ORDER_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_SDK_CREATE_ORDER
}

func (m *C2S_SDK_CREATE_ORDER_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.SdkCreateOrderArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SDK_CREATE_ORDER_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_SDK_CREATE_ORDER_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.SdkCreateOrderArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_SDK_CREATE_ORDER_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SdkCreateOrderReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_SDK_CREATE_ORDER_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_SDK_CREATE_ORDER_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SdkCreateOrderReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_SDK_CREATE_ORDER END ----------------------------------------

//@ C2S_IOS_PRE_PAY    req: IosPrePayArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_IOS_PRE_PAY_Meta struct {
}

func (m *C2S_IOS_PRE_PAY_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_IOS_PRE_PAY
}

func (m *C2S_IOS_PRE_PAY_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.IosPrePayArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_IOS_PRE_PAY_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_IOS_PRE_PAY_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.IosPrePayArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_IOS_PRE_PAY_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_IOS_PRE_PAY_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_IOS_PRE_PAY END ----------------------------------------

//@ C2S_BUY_LIMIT_GITF_BY_JADE    req: BuyLimitGiftByJadeArg    resp: BuyLimitGiftReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_LIMIT_GITF_BY_JADE_Meta struct {
}

func (m *C2S_BUY_LIMIT_GITF_BY_JADE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_LIMIT_GITF_BY_JADE
}

func (m *C2S_BUY_LIMIT_GITF_BY_JADE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyLimitGiftByJadeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_LIMIT_GITF_BY_JADE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_LIMIT_GITF_BY_JADE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyLimitGiftByJadeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_LIMIT_GITF_BY_JADE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyLimitGiftReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_LIMIT_GITF_BY_JADE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_LIMIT_GITF_BY_JADE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyLimitGiftReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_LIMIT_GITF_BY_JADE END ----------------------------------------

//@ C2S_BUY_ONE_GA_CHA_BY_JADE    req: BuyLimitGiftByJadeArg    resp: RechargeLotteryReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_ONE_GA_CHA_BY_JADE_Meta struct {
}

func (m *C2S_BUY_ONE_GA_CHA_BY_JADE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_ONE_GA_CHA_BY_JADE
}

func (m *C2S_BUY_ONE_GA_CHA_BY_JADE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyLimitGiftByJadeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_ONE_GA_CHA_BY_JADE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_ONE_GA_CHA_BY_JADE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyLimitGiftByJadeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_ONE_GA_CHA_BY_JADE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.RechargeLotteryReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_ONE_GA_CHA_BY_JADE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_ONE_GA_CHA_BY_JADE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.RechargeLotteryReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_ONE_GA_CHA_BY_JADE END ----------------------------------------

//@ C2S_GOOGLE_PLAY_RECHARGE    req: GooglePlayRechargeArg    resp: SdkRechargeResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_GOOGLE_PLAY_RECHARGE_Meta struct {
}

func (m *C2S_GOOGLE_PLAY_RECHARGE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_GOOGLE_PLAY_RECHARGE
}

func (m *C2S_GOOGLE_PLAY_RECHARGE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.GooglePlayRechargeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GOOGLE_PLAY_RECHARGE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_GOOGLE_PLAY_RECHARGE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.GooglePlayRechargeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_GOOGLE_PLAY_RECHARGE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SdkRechargeResult)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_GOOGLE_PLAY_RECHARGE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_GOOGLE_PLAY_RECHARGE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SdkRechargeResult{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_GOOGLE_PLAY_RECHARGE END ----------------------------------------

//@ C2S_BUY_VIP_CARD    resp: BuyVipCardReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_VIP_CARD_Meta struct {
}

func (m *C2S_BUY_VIP_CARD_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_VIP_CARD
}

func (m *C2S_BUY_VIP_CARD_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_BUY_VIP_CARD_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_BUY_VIP_CARD_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyVipCardReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_VIP_CARD_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_VIP_CARD_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyVipCardReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_VIP_CARD END ----------------------------------------

//@ C2S_PIECE_EXCHANGE_ITEM    req: PieceExchangeArg    resp: ok or err
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_PIECE_EXCHANGE_ITEM_Meta struct {
}

func (m *C2S_PIECE_EXCHANGE_ITEM_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_PIECE_EXCHANGE_ITEM
}

func (m *C2S_PIECE_EXCHANGE_ITEM_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.PieceExchangeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_PIECE_EXCHANGE_ITEM_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_PIECE_EXCHANGE_ITEM_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.PieceExchangeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_PIECE_EXCHANGE_ITEM_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_PIECE_EXCHANGE_ITEM_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_PIECE_EXCHANGE_ITEM END ----------------------------------------

//@ C2S_LOOK_OVER_LIMIT_GIFT    req: TargetLimitGift
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_LOOK_OVER_LIMIT_GIFT_Meta struct {
}

func (m *C2S_LOOK_OVER_LIMIT_GIFT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_LOOK_OVER_LIMIT_GIFT
}

func (m *C2S_LOOK_OVER_LIMIT_GIFT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.TargetLimitGift)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_LOOK_OVER_LIMIT_GIFT_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_LOOK_OVER_LIMIT_GIFT_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.TargetLimitGift{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_LOOK_OVER_LIMIT_GIFT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_LOOK_OVER_LIMIT_GIFT_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_LOOK_OVER_LIMIT_GIFT END ----------------------------------------

//@ C2S_IOS_RECHARGE    req: IosRechargeArg    resp: SdkRechargeResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_IOS_RECHARGE_Meta struct {
}

func (m *C2S_IOS_RECHARGE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_IOS_RECHARGE
}

func (m *C2S_IOS_RECHARGE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.IosRechargeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_IOS_RECHARGE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_IOS_RECHARGE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.IosRechargeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_IOS_RECHARGE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SdkRechargeResult)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_IOS_RECHARGE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_IOS_RECHARGE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SdkRechargeResult{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_IOS_RECHARGE END ----------------------------------------

//@ C2S_BUY_RECRUIT_TREASURE    req: BuyRecruitTreasureArg    resp: BuyRecruitTreasureReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_RECRUIT_TREASURE_Meta struct {
}

func (m *C2S_BUY_RECRUIT_TREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_RECRUIT_TREASURE
}

func (m *C2S_BUY_RECRUIT_TREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyRecruitTreasureArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RECRUIT_TREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_RECRUIT_TREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyRecruitTreasureArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_RECRUIT_TREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyRecruitTreasureReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RECRUIT_TREASURE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_RECRUIT_TREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyRecruitTreasureReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_RECRUIT_TREASURE END ----------------------------------------

//@ C2S_BUY_RANDOM_SHOP_REFRESH_CNT    req: BuyRandomShopRefreshCntArg   resp: BuyRandomShopRefreshCntReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta struct {
}

func (m *C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_RANDOM_SHOP_REFRESH_CNT
}

func (m *C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.BuyRandomShopRefreshCntArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.BuyRandomShopRefreshCntArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyRandomShopRefreshCntReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_RANDOM_SHOP_REFRESH_CNT_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyRandomShopRefreshCntReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_RANDOM_SHOP_REFRESH_CNT END ----------------------------------------

//@ C2S_TEST_RECHARGE    req: IosRechargeArg    resp: SdkRechargeResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_TEST_RECHARGE_Meta struct {
}

func (m *C2S_TEST_RECHARGE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_TEST_RECHARGE
}

func (m *C2S_TEST_RECHARGE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.IosRechargeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_TEST_RECHARGE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_TEST_RECHARGE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.IosRechargeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_TEST_RECHARGE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.SdkRechargeResult)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_TEST_RECHARGE_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_TEST_RECHARGE_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.SdkRechargeResult{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_TEST_RECHARGE END ----------------------------------------

//@ C2S_FETCH_EXCHANGE_CARD_SKIN_IDS    resp: PieceExchangeIds
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta struct {
}

func (m *C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_FETCH_EXCHANGE_CARD_SKIN_IDS
}

func (m *C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.PieceExchangeIds)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_FETCH_EXCHANGE_CARD_SKIN_IDS_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.PieceExchangeIds{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_FETCH_EXCHANGE_CARD_SKIN_IDS END ----------------------------------------

//@ C2S_BUY_SOLD_GOLD_GIFT    resp: BuySoldGoldGiftReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_SOLD_GOLD_GIFT_Meta struct {
}

func (m *C2S_BUY_SOLD_GOLD_GIFT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_SOLD_GOLD_GIFT
}

func (m *C2S_BUY_SOLD_GOLD_GIFT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_BUY_SOLD_GOLD_GIFT_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_BUY_SOLD_GOLD_GIFT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuySoldGoldGiftReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_SOLD_GOLD_GIFT_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_SOLD_GOLD_GIFT_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuySoldGoldGiftReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_SOLD_GOLD_GIFT END ----------------------------------------

//@ C2S_MIDAS_RECHARGE    req: MidasRechargeArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_MIDAS_RECHARGE_Meta struct {
}

func (m *C2S_MIDAS_RECHARGE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_MIDAS_RECHARGE
}

func (m *C2S_MIDAS_RECHARGE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.MidasRechargeArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_MIDAS_RECHARGE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *C2S_MIDAS_RECHARGE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.MidasRechargeArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *C2S_MIDAS_RECHARGE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_MIDAS_RECHARGE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ C2S_MIDAS_RECHARGE END ----------------------------------------

//@ C2S_BUY_RECOMMEND_GIFT resp: BuyRecommendGiftReply
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type C2S_BUY_RECOMMEND_GIFT_Meta struct {
}

func (m *C2S_BUY_RECOMMEND_GIFT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_C2S_BUY_RECOMMEND_GIFT
}

func (m *C2S_BUY_RECOMMEND_GIFT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	return nil, nil
}

func (m *C2S_BUY_RECOMMEND_GIFT_Meta) DecodeArg(data []byte) (interface{}, error) {
	return nil, nil
}

func (m *C2S_BUY_RECOMMEND_GIFT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	_reply, ok := reply.(*pb.BuyRecommendGiftReply)
	if !ok {
		p, ok := reply.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("C2S_BUY_RECOMMEND_GIFT_Meta EncodeReply error type")
	}

	return _reply.Marshal()
}

func (m *C2S_BUY_RECOMMEND_GIFT_Meta) DecodeReply(data []byte) (interface{}, error) {
	reply := &pb.BuyRecommendGiftReply{}
	if err := reply.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return reply, nil
	}
}

//------------------------------------ C2S_BUY_RECOMMEND_GIFT END ----------------------------------------

//@ S2C_NOTIFY_SDK_RECHARGE_RESULT    req: SdkRechargeResult
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta struct {
}

func (m *S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT
}

func (m *S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.SdkRechargeResult)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.SdkRechargeResult{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_NOTIFY_SDK_RECHARGE_RESULT_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_NOTIFY_SDK_RECHARGE_RESULT END ----------------------------------------

//@ S2C_UPDATE_RECRUIT_TREASURE    req: RecruitTreasureData
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_UPDATE_RECRUIT_TREASURE_Meta struct {
}

func (m *S2C_UPDATE_RECRUIT_TREASURE_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_UPDATE_RECRUIT_TREASURE
}

func (m *S2C_UPDATE_RECRUIT_TREASURE_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.RecruitTreasureData)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_UPDATE_RECRUIT_TREASURE_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_UPDATE_RECRUIT_TREASURE_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.RecruitTreasureData{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_UPDATE_RECRUIT_TREASURE_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_UPDATE_RECRUIT_TREASURE_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_UPDATE_RECRUIT_TREASURE END ----------------------------------------

//@ S2C_UPDATE_RANDOM_SHOP    req: VisitRandomShopData
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_UPDATE_RANDOM_SHOP_Meta struct {
}

func (m *S2C_UPDATE_RANDOM_SHOP_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_UPDATE_RANDOM_SHOP
}

func (m *S2C_UPDATE_RANDOM_SHOP_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.VisitRandomShopData)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_UPDATE_RANDOM_SHOP_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_UPDATE_RANDOM_SHOP_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.VisitRandomShopData{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_UPDATE_RANDOM_SHOP_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_UPDATE_RANDOM_SHOP_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_UPDATE_RANDOM_SHOP END ----------------------------------------

//@ S2C_UPDATE_SHOP_DATA    req: UpdateShopDataArg
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_UPDATE_SHOP_DATA_Meta struct {
}

func (m *S2C_UPDATE_SHOP_DATA_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_UPDATE_SHOP_DATA
}

func (m *S2C_UPDATE_SHOP_DATA_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.UpdateShopDataArg)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_UPDATE_SHOP_DATA_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_UPDATE_SHOP_DATA_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.UpdateShopDataArg{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_UPDATE_SHOP_DATA_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_UPDATE_SHOP_DATA_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_UPDATE_SHOP_DATA END ----------------------------------------

//@ S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS    req: PieceExchangeIds
//------------------------------------------------------------------------------------------
// implement protoc.IMeta
type S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta struct {
}

func (m *S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta) GetMessageID() protoc.IMessageID {
	return pb.MessageID_S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS
}

func (m *S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta) EncodeArg(arg interface{}) ([]byte, error) {
	_arg, ok := arg.(*pb.PieceExchangeIds)
	if !ok {
		p, ok := arg.([]byte)
		if ok {
			return p, nil
		}

		return nil, errors.New("S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta EncodeArg error type")
	}

	return _arg.Marshal()
}

func (m *S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta) DecodeArg(data []byte) (interface{}, error) {
	arg := &pb.PieceExchangeIds{}
	if err := arg.Unmarshal(data); err != nil {
		return nil, err
	} else {
		return arg, nil
	}
}

func (m *S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta) EncodeReply(reply interface{}) ([]byte, error) {
	return nil, nil
}

func (m *S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS_Meta) DecodeReply(data []byte) (interface{}, error) {
	return nil, nil
}

//------------------------------------ S2C_UPDATE_EXCHANGE_CARD_SKIN_IDS END ----------------------------------------
