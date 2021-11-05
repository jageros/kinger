package spring

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	atypes "kinger/apps/game/activitys/types"
	"kinger/gamedata"
	"kinger/apps/game/module"
	"fmt"
	"strconv"
)

func rpc_C2S_HuodongExchange(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.HuodongExchangeArg)
	cpt := atypes.IMod.NewPCM(player)
	aid := int(arg2.ActivityID)

	if !cpt.ConformTime(aid) {
		return nil, gamedata.GameError(atypes.GetTimeConditionError)
	}

	pdata := newComponent(player)
	pdata.checkVersion(aid)
	goods := getGoods(pb.HuodongTypeEnum_HSpringExchange, int(arg2.GoodsID))
	if goods == nil {
		glog.Infof("spring exchange no goods %s %d", pb.HuodongTypeEnum_HSeasonPvp, arg2.GoodsID)
		return nil, gamedata.InternalErr
	}

	goodsData := goods.getGameData()
	if goodsData.ExchangeCnt > 0 && pdata.getExchangeCnt(aid, goodsData.ID) >= goodsData.ExchangeCnt {
		return nil, gamedata.GameError(1)
	}

	if !goods.canExchange(player) {
		return nil, gamedata.GameError(2)
	}

	resType := GetEventItemType()
	if !module.Player.HasResource(player, resType, goodsData.Price) {
		return nil, gamedata.GameError(3)
	}

	module.Player.ModifyResource(player, resType, - goodsData.Price)
	itemID, itemName, treasure := goods.exchange(player)
	pdata.onExchangeGoods(aid, goodsData)

	if itemID != "" || itemName != "" {
		module.Shop.LogShopBuyItem(player, itemID, itemName, 1,
			fmt.Sprintf("%s_%d", pb.HuodongTypeEnum_HSeasonPvp, pdata.getActivityVersion(aid)),
			strconv.Itoa(resType), module.Player.GetResourceName(resType), goodsData.Price, "")
	}
	return &pb.HuodongExchangeReply{Treasure: treasure}, nil
}

func RegisterRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_HUODONG_EXCHANGE, rpc_C2S_HuodongExchange)
}
