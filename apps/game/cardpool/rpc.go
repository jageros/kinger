package cardpool

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/network"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
)

func rpc_C2S_FetchCardData(agent *logic.PlayerAgent, _ interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	reply := &pb.CardPools{
		FightCamp: int32(cardComponent.GetFightCamp()),
	}

	pools := cardComponent.getAllPvpCardPool()
	for _, p := range pools {
		reply.Pools = append(reply.Pools, &pb.CardPool{
			PoolId:  int32(p.getPoolID()),
			Cards:   p.getCards(),
			Camp:    int32(p.getCamp()),
			IsFight: p.isFight(),
		})
	}

	return reply, nil
}

func rpc_C2S_PoolAddCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.PoolAddCard)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	err := cardComponent.poolAddCard(_arg.Card, int(_arg.PoolId), int(_arg.Idx))
	return nil, err
}

func rpc_C2S_PoolUpdateCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.PoolUpdateCard)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	err := cardComponent.poolUpdateCard(int(_arg.PoolId), _arg.Cards)
	return nil, err
}

func rpc_C2S_UpdateCardPool(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, nil
	}

	_arg := arg.(*pb.UpdateCardPools)
	cardComponent := player.GetComponent("card").(*cardComponent)
	cardComponent.updatePvpFightCardPool(_arg.Pools, int(_arg.FightCamp))
	return nil, nil
}

func rpc_C2S_UplevelCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.UpLevelCardArg)
	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	return nil, cardComponent.uplevelCard(arg2.CardId, arg2.IsConsumeJade, arg2.IsNeedJade)
}

func rpc_C2S_CardRelive(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.TargetCard)
	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	return cardComponent.cardRelive(_arg.CardId)
}

func rpc_C2S_CardTreat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.TargetCard)
	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	return cardComponent.cardTreat(_arg.CardId)
}

func rpc_C2S_DiyCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.DiyCardArg)
	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	card, err := cardComponent.makeDiyCard(_arg.Name, int(_arg.DiySkillId1), int(_arg.DiySkillId2), _arg.Weapon, _arg.Img)

	if err != nil {
		return nil, err
	} else {
		return card.packReplyMsg(), nil
	}
}

func rpc_C2S_FetchDiyCardImg(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	return nil, gamedata.InternalErr
}

func rpc_C2S_DiyCardAgain(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.TargetCard)
	cardComponent := player.GetComponent("card").(*cardComponent)
	card := cardComponent.getDiyCard(_arg.CardId)
	if card == nil {
		return nil, gamedata.InternalErr
	}

	newCard, err := cardComponent.remakeDiyCard(card)
	if err != nil {
		return nil, err
	} else {
		return newCard.packReplyMsg(), nil
	}
}

func rpc_C2S_DecomposeDiyCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.TargetCard)
	cardComponent := player.GetComponent(consts.CardCpt).(*cardComponent)
	card := cardComponent.getDiyCard(_arg.CardId)
	if card == nil {
		return nil, gamedata.InternalErr
	}

	cardComponent.decomposeDiyCard(card)
	return nil, nil
}

func rpc_C2S_UnlockCardLevel(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.UnlockCardLevelArg)
	card := module.Card.GetCollectCard(player, arg2.CardID)
	if card == nil {
		return nil, gamedata.GameError(2)
	}

	level := int(arg2.Level)
	cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(arg2.CardID, level - 1)
	if cardData == nil || cardData.ConsumeBook <= 0 {
		return nil, gamedata.GameError(1)
	}

	card2 := card.(*collectCard)
	if card2.GetMaxUnlockLevel() >= level {
		return nil, gamedata.GameError(3)
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.SkyBook, cardData.ConsumeBook) {
		return nil, gamedata.GameError(4)
	}

	resCpt.ModifyResource(consts.SkyBook, - cardData.ConsumeBook, consts.RmrUnlockCardLevel)
	card2.setMaxUnlockLevel(level)

	card2.pushClient(player)

	return nil, nil
}

func rpc_C2S_BackCardUnlock(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.BackCardUnlockArg)
	card := module.Card.GetCollectCard(player, arg2.CardID)
	if card == nil {
		return nil, gamedata.GameError(1)
	}

	level := int(arg2.Level)
	level = 4
	card2 := card.(*collectCard)

	if card2.GetMaxUnlockLevel() == 0 || card2.GetMaxUnlockLevel() <= level {
		return nil, gamedata.GameError(2)
	}

	cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(arg2.CardID, level)
	if cardData == nil || cardData.ConsumeBook <= 0 {
		return nil, gamedata.GameError(3)
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Jade, 100) {
		return nil, gamedata.GameError(4)
	}

	reply := &pb.BackCardUnlockReply{}

	needRes := make(map[int]int)
	needRes[consts.SkyBook] = cardData.ConsumeBook

	var levelupNum int32
	if card2.GetLevel() > level {
		needRes[consts.Gold] = cardData.LevelupGold
		newAmount := card2.GetAmount() + cardData.LevelupNum
		levelupNum = int32(cardData.LevelupNum)
		card2.SetAmount(newAmount)
		card2.setLevel(level)
	}

	for resType, resAmount := range needRes {
		reply.Resources = append(reply.Resources, &pb.Resource{
			Type: int32(resType),
			Amount: int32(resAmount),
		})
	}

	needRes[consts.Jade] = -100
	resCpt.BatchModifyResource(needRes, consts.RmrBackCardUnlock)
	module.Shop.LogShopBuyItem(player, "backCardUnlock", "回退卡等级解锁", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), -100, "")
	card2.setMaxUnlockLevel(0)

	reply.CardId = card2.GetCardID()
	reply.CardAmount += levelupNum
	card2.pushClient(player)

	return reply, nil
}

func rpc_G2G_GetCardAmountLog(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GetCardAmountLogArg)
	reply := &pb.GetCardAmountLogReply{}

	var packAtLog = func(accountType pb.AccountTypeEnum, log *logHub) {
		logsMsg := &pb.CardsAmountLog{AccountType: accountType}
		reply.Logs = append(reply.Logs, logsMsg)
		log.forEachCardLog(func(cl *cardLog) {
			msg := &pb.CardAmountLog{}
			msg.CardID = cl.getCardID()
			msg.Amount = int32(cl.getAmount())
			logsMsg.Logs = append(logsMsg.Logs, msg)
		})
	}

	if arg2.AccountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(int(arg2.Area), arg2.AccountType, func(log *logHub) {
			packAtLog(arg2.AccountType, log)
		})
	} else {
		forEachLog(int(arg2.Area), packAtLog)
	}

	return reply, nil
}

func rpc_G2G_GetCardLevelLog(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GetCardLevelLogArg)
	reply := &pb.GetCardLevelLogReply{}

	var packAtLog = func(accountType pb.AccountTypeEnum, log *logHub) {

		logsMsg := &pb.CardsLevelLog{AccountType: accountType}
		reply.Logs = append(reply.Logs, logsMsg)

		var packLevelLog = func(cl *cardLog) {
			msg := &pb.CardLevelLog{CardID: cl.getCardID()}
			cl.forEachLevel(func(level, amount int) {
				msg.Levels = append(msg.Levels, &pb.CardLevelLog_LevelAmount{
					Level:  int32(level),
					Amount: int32(amount),
				})
			})
			logsMsg.Logs = append(logsMsg.Logs, msg)
		}

		if arg2.CardID > 0 {
			packLevelLog( log.getCardLog(arg2.CardID) )
		} else {
			log.forEachCardLog(packLevelLog)
		}
	}

	if arg2.AccountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(int(arg2.Area), arg2.AccountType, func(log *logHub) {
			packAtLog(arg2.AccountType, log)
		})
	} else {
		forEachLog(int(arg2.Area), packAtLog)
	}

	return reply, nil
}

/*
func rpc_G2G_GetCardPoolLog(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GetCardPoolLogArg)
	reply := &pb.GetCardPoolLogReply{}

	var packAtLog = func(accountType pb.AccountTypeEnum, log *logHub) {

		logsMsg := &pb.CardPoolsLog{AccountType: accountType}
		reply.Logs = append(reply.Logs, logsMsg)

		var packPvpLevelLog = func(bl *battleLog) {
			msg := &pb.CardPoolLog{PvpLevel: int32(bl.getPvpLevel()), BattleAmount: int32(bl.getAmount())}
			bl.forEachCard(func(cardID uint32, amount int) {
				msg.CardLogs = append(msg.CardLogs, &pb.CardPoolLog_CardLog{
					CardID: cardID,
					Amount: int32(amount),
				})
			})
			logsMsg.Logs = append(logsMsg.Logs, msg)
		}

		if arg2.PvpLevel > 0 {
			packPvpLevelLog( log.getBattleLog(int(arg2.PvpLevel)) )
		} else {
			log.forEachBattleLog(packPvpLevelLog)
		}
	}

	if arg2.AccountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(int(arg2.Area), arg2.AccountType, func(log *logHub) {
			packAtLog(arg2.AccountType, log)
		})
	} else {
		forEachLog(int(arg2.Area), packAtLog)
	}

	return reply, nil
}
*/

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CARD_DATA, rpc_C2S_FetchCardData)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_POOL_ADD_CARD, rpc_C2S_PoolAddCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_POOL_UPDATE_CARD, rpc_C2S_PoolUpdateCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_CARD_POOL, rpc_C2S_UpdateCardPool)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPLEVEL_CARD, rpc_C2S_UplevelCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CARD_RELIVE, rpc_C2S_CardRelive)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CARD_TREAT, rpc_C2S_CardTreat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DIY_CARD, rpc_C2S_DiyCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_DIY_CARD_IMG, rpc_C2S_FetchDiyCardImg)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DIY_CARD_AGAIN, rpc_C2S_DiyCardAgain)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DECOMPOSE_DIY_CARD, rpc_C2S_DecomposeDiyCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UNLOCK_CARD_LEVEL, rpc_C2S_UnlockCardLevel)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BACK_CARD_UNLOCK, rpc_C2S_BackCardUnlock)

	logic.RegisterRpcHandler(pb.MessageID_G2G_GET_CARD_AMOUNT_LOG, rpc_G2G_GetCardAmountLog)
	logic.RegisterRpcHandler(pb.MessageID_G2G_GET_CARD_LEVEL_LOG, rpc_G2G_GetCardLevelLog)
	//logic.RegisterRpcHandler(pb.MessageID_G2G_GET_CARD_POOL_LOG, rpc_G2G_GetCardPoolLog)
}
