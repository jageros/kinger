package reborn

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
)

func rpc_C2S_RefineCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.RefineCardArg)
	cards := make([]types.ICollectCard, len(arg2.CardIDs))
	modifyCards := map[uint32]*pb.CardInfo{}
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardCaculGameData := gamedata.GetGameData(consts.RebornCardCacul).(*gamedata.RebornCardCaculGameData)
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	hasSkyBook := resCpt.HasResource(consts.SkyBook, 1)
	var reputation float32
	for i, cardID := range arg2.CardIDs {
		if _, ok := modifyCards[cardID]; ok {
			return nil, gamedata.GameError(1)
		}

		c := cardCpt.GetCollectCard(cardID)
		if c == nil {
			return nil, gamedata.GameError(2)
		}

		if hasSkyBook {
			if !c.IsMaxLevel() {
				return nil, gamedata.GameError(4)
			}
		} else {
			if !c.IsMaxCanUpLevel() {
				return nil, gamedata.GameError(4)
			}
		}

		cardData := c.GetCardGameData()
		f, ok := cardCaculGameData.Start2Feats[cardData.Rare]
		if !ok {
			return nil, gamedata.GameError(3)
		}

		cards[i] = c
		amount := c.GetAmount()
		modifyCards[cardID] = &pb.CardInfo{Amount: -int32(amount)}
		reputation += f * float32(amount)
	}

	glog.Infof("rpc_C2S_RefineCard uid=%d, reputation=%d, cards=%v", uid, int(reputation), cards)
	cardCpt.ModifyCollectCards(modifyCards)
	module.Player.ModifyResource(player, consts.Reputation, int(reputation))
	return &pb.RefineCardReply{Reputation: int32(reputation)}, nil
}

func rpc_C2S_BuyRebornGoods(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.BuyRebornGoodsArg)
	rebornCpt := player.GetComponent(consts.RebornCpt).(*rebornComponent)
	var err error
	var itemID, itemName string
	var resType, resAmount int
	switch arg2.Type {
	case pb.BuyRebornGoodsArg_Card:
		itemID, itemName, resType, resAmount, err = rebornCpt.buyCard(int(arg2.GoodsID))
	case pb.BuyRebornGoodsArg_Privilege:
		itemID, itemName, resType, resAmount, err = rebornCpt.buyPrivilege(int(arg2.GoodsID))
	case pb.BuyRebornGoodsArg_CardSkin:
		itemID, itemName, resType, resAmount, err = rebornCpt.buyCardSkin(int(arg2.GoodsID))
	case pb.BuyRebornGoodsArg_Equip:
		itemID, itemName, resType, resAmount, err = rebornCpt.buyEquip(int(arg2.GoodsID))
	default:
		err = gamedata.GameError(1)
	}

	if err == nil {
		module.Shop.LogShopBuyItem(player, itemID, itemName, 1, "reborn",
			strconv.Itoa(resType), module.Player.GetResourceName(resType), resAmount, "")
	}

	return nil, err
}

func rpc_C2S_Reborn(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.RebornCpt).(*rebornComponent).reborn()
}

func rpc_C2S_FetchRebornData(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	rebornCpt := player.GetComponent(consts.RebornCpt).(*rebornComponent)
	return &pb.RebornData{
		Prestige:  int32(rebornPrestige),
		RemainDay: int32(rebornCpt.getRebornRemainDay()),
		Cnt:       int32(rebornCpt.getRebornCnt()),
	}, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REFINE_CARD, rpc_C2S_RefineCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_REBORN_GOODS, rpc_C2S_BuyRebornGoods)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REBORN, rpc_C2S_Reborn)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_REBORN_DATA, rpc_C2S_FetchRebornData)
}
