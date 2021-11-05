package spring

/*
import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	htypes "kinger/apps/game/huodong/types"
	"kinger/gamedata"
	"kinger/apps/game/module"
	"kinger/common/consts"
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
	hd := htypes.Mod.GetHuodong(player.GetArea(), arg2.Type)
	if hd == nil || !hd.IsOpen() {
		return nil, gamedata.InternalErr
	}

	hd2, ok := hd.(*springHd)
	if !ok {
		return nil, gamedata.InternalErr
	}

	pdata := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(arg2.Type)
	if pdata == nil {
		return nil, gamedata.InternalErr
	}

	pdata2, ok := pdata.(*springHdPlayerData)
	if !ok {
		return nil, gamedata.InternalErr
	}

	goods := getGoods(arg2.Type, int(arg2.GoodsID))
	if goods == nil {
		glog.Infof("huodong exchange no goods %s %d", arg2.Type, arg2.GoodsID)
		return nil, gamedata.InternalErr
	}

	goodsData := goods.getGameData()
	if goodsData.ExchangeCnt > 0 && pdata2.getExchangeCnt(goodsData.ID) >= goodsData.ExchangeCnt {
		return nil, gamedata.GameError(1)
	}

	if !goods.canExchange(player) {
		return nil, gamedata.GameError(2)
	}

	if !hd2.HasEventItem(player, goodsData.Price) {
		return nil, gamedata.GameError(3)
	}

	hd2.SubEventItem(player, goodsData.Price)
	itemID, itemName, treasure := goods.exchange(player)
	pdata2.onExchangeGoods(goodsData)

	if itemID != "" || itemName != "" {
		eventResType := hd2.GetEventItemType()
		module.Shop.LogShopBuyItem(player, itemID, itemName, 1, fmt.Sprintf("%s_%d", arg2.Type, hd.GetVersion()),
			strconv.Itoa(eventResType), module.Player.GetResourceName(eventResType), goodsData.Price, "")
	}
	return &pb.HuodongExchangeReply{Treasure: treasure}, nil
}
*/

func registerRpc() {
	//logic.RegisterAgentRpcHandler(pb.MessageID_C2S_HUODONG_EXCHANGE, rpc_C2S_HuodongExchange)
}
