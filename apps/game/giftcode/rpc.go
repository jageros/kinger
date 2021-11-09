package giftcode

import (
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
)

func rpc_C2S_ExchangeGiftCode(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.GiftCodeCpt).(*giftCodeComponent).exchange(arg.(*pb.ExchangeCodeArg).Code)
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_EXCHANGE_GIFT_CODE, rpc_C2S_ExchangeGiftCode)
}
