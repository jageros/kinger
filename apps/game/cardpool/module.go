package cardpool

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/config"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"math"
)

var mod module.ICardModule = &cardModule{}

type cardModule struct {
}

type cardInfoLog struct {
	name string
	id   int32
}

func (ci *cardInfoLog) MarshalLogObject(encoder glog.ObjectEncoder) error {
	encoder.AddString("cardName", ci.name)
	encoder.AddInt32("cardID", ci.id)
	return nil
}

func (m *cardModule) NewCardComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	cardAttr := playerAttr.GetMapAttr("card")
	if cardAttr == nil {
		cardAttr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("card", cardAttr)
	}
	return &cardComponent{attr: cardAttr}
}

func (m *cardModule) IsDiyCard(cardId uint32) bool {
	return cardId >= minDiyCardID
}

func (m *cardModule) UpdateCardSkin(player types.IPlayer, cardID uint32, skin string) error {
	return player.GetComponent(consts.CardCpt).(*cardComponent).updateCardSkin(cardID, skin)
}

func (m *cardModule) GetCollectCard(player types.IPlayer, cardID uint32) types.ICollectCard {
	return player.GetComponent(consts.CardCpt).(*cardComponent).GetCollectCard(cardID)
}

func (m *cardModule) GetOnceCards(player types.IPlayer) []types.ICollectCard {
	return player.GetComponent(consts.CardCpt).(*cardComponent).getOnceCards()
}

func (m *cardModule) GetAllCollectCards(player types.IPlayer) []types.ICollectCard {
	return player.GetComponent(consts.CardCpt).(*cardComponent).GetAllCollectCards()
}

func (m *cardModule) GetAllCollectCardsByCamp(player types.IPlayer, camps []int) []types.ICollectCard {
	collectCardMap := player.GetComponent(consts.CardCpt).(*cardComponent).collectCardMap
	var cards []types.ICollectCard
	for _, card := range collectCardMap {
		camp := card.GetCardGameData().Camp
		for _, c := range camps {
			if c == camp {
				cards = append(cards, card)
				break
			}
		}
	}
	return cards
}

func (m *cardModule) OnCampaignMissionDone(player types.IPlayer, cardIDs []uint32) {
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
	allCards := cardCpt.GetAllCollectCards()
	arg := &pb.CardDatas{}
	for _, card := range allCards {
		if card.GetState() == pb.CardState_InCampaignMs {
			c := card.(*collectCard)
			c.setState(pb.CardState_NormalCState)
			arg.Cards = append(arg.Cards, c.PackMsg())
		}
	}
	player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, arg)
}

func (m *cardModule) OnAcceptCampaignMission(player types.IPlayer, cardIDs []uint32) {
	arg := &pb.CardDatas{}
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
	for _, cardID := range cardIDs {
		card := cardCpt.GetCollectCard(cardID)
		if card != nil {
			c := card.(*collectCard)
			c.setState(pb.CardState_InCampaignMs)
			arg.Cards = append(arg.Cards, c.PackMsg())
			cardCpt.poolDelCard(cardID)
		}
	}
	player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, arg)
}

func (m *cardModule) SetCardsState(player types.IPlayer, cardIDs []uint32, state pb.CardState) {
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
	arg := &pb.CardDatas{}
	for _, cardID := range cardIDs {
		card := cardCpt.GetCollectCard(cardID)
		if card != nil {
			c := card.(*collectCard)
			if c.GetState() != state {
				c.setState(state)
				arg.Cards = append(arg.Cards, c.PackMsg())
			}
		}
	}

	if len(arg.Cards) > 0 {
		player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, arg)
	}
}

func (m *cardModule) GetUnlockCards(player types.IPlayer, team int) []uint32 {
	if team <= 0 {
		team = player.GetPvpTeam()
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	rs := rgd.RanksOfTeam[team]
	var unlockCards []uint32

	if len(rs) > 0 {
		r := rs[len(rs)-1]
		for _, cardID := range r.Unlock {
			cardData := poolGameData.GetCard(uint32(cardID), 1)
			if cardData == nil {
				continue
			}
			unlockCards = append(unlockCards, cardID)
		}
	}

	levelCpt := player.GetComponent(consts.LevelCpt).(types.ILevelComponent)
	for _, cardID := range levelCpt.GetUnlockCards() {
		cardData := poolGameData.GetCard(uint32(cardID), 1)
		if cardData == nil {
			continue
		}

		unlockCards = append(unlockCards, cardID)
	}

	unlockCards = append(unlockCards, m.GetFirstRechargeUnlockCards(player)...)
	return unlockCards
}

func (m *cardModule) GetFirstRechargeUnlockCards(player types.IPlayer) []uint32 {
	var unlockCards []uint32
	firstRechargeUnlockCards := gamedata.GetGameData(
		consts.EventFirstRechargeReward).(*gamedata.ActivityFirstRechargeRewardGameData).UnlockCards
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)

	for _, cardID := range firstRechargeUnlockCards {
		if cardCpt.GetCollectCard(cardID) != nil {
			unlockCards = append(unlockCards, cardID)
		}
	}
	return unlockCards
}

func (m *cardModule) ModifyCollectCards(player types.IPlayer, cardsChange map[uint32]*pb.CardInfo) []*pb.ChangeCardInfo {
	return player.GetComponent(consts.CardCpt).(*cardComponent).ModifyCollectCards(cardsChange)
}

func (m *cardModule) GmAllCardUpLevel(player types.IPlayer) {
	allCards := m.GetAllCollectCards(player)
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
	for _, card := range allCards {
		for {
			err := cardCpt.uplevelCard(card.GetCardID(), false, true)
			if err != nil {
				break
			}
		}
	}
}

func (m *cardModule) GetCardAmountLog(accountType pb.AccountTypeEnum, area int) map[pb.AccountTypeEnum]map[uint32]int {
	appID := uint16(module.Service.GetAppID())
	allApps := config.GetConfig().GetLogicConfigsByName(gconsts.AppGame)
	reply := map[pb.AccountTypeEnum]map[uint32]int{}
	arg := &pb.GetCardAmountLogArg{AccountType: accountType, Area: int32(area)}

	for _, cfg := range allApps {
		if cfg.ID == appID {
			continue
		}

		result, err := logic.CallBackend(gconsts.AppGame, uint32(cfg.ID), pb.MessageID_G2G_GET_CARD_AMOUNT_LOG, arg)
		if err != nil {
			continue
		}

		result2 := result.(*pb.GetCardAmountLogReply)
		for _, cml := range result2.Logs {
			id2Amount, ok := reply[cml.AccountType]
			if !ok {
				id2Amount = map[uint32]int{}
				reply[cml.AccountType] = id2Amount
			}

			for _, l := range cml.Logs {
				id2Amount[l.CardID] = id2Amount[l.CardID] + int(l.Amount)
			}
		}
	}

	var packAtLog = func(at pb.AccountTypeEnum, log *logHub) {
		id2Amount, ok := reply[at]
		if !ok {
			id2Amount = map[uint32]int{}
			reply[at] = id2Amount
		}

		log.forEachCardLog(func(cl *cardLog) {
			cardID := cl.getCardID()
			id2Amount[cardID] = id2Amount[cardID] + cl.getAmount()
		})
	}

	if accountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(area, accountType, func(log *logHub) {
			packAtLog(accountType, log)
		})
	} else {
		forEachLog(area, packAtLog)
	}

	return reply
}

func (m *cardModule) GetCardLevelLog(accountType pb.AccountTypeEnum, cardID uint32, area int) map[pb.AccountTypeEnum]map[uint32]map[int]int {
	appID := uint16(module.Service.GetAppID())
	allApps := config.GetConfig().GetLogicConfigsByName(gconsts.AppGame)
	reply := map[pb.AccountTypeEnum]map[uint32]map[int]int{}
	arg := &pb.GetCardLevelLogArg{AccountType: accountType, CardID: cardID, Area: int32(area)}

	for _, cfg := range allApps {
		if cfg.ID == appID {
			continue
		}

		result, err := logic.CallBackend(gconsts.AppGame, uint32(cfg.ID), pb.MessageID_G2G_GET_CARD_LEVEL_LOG, arg)
		if err != nil {
			continue
		}

		result2 := result.(*pb.GetCardLevelLogReply)
		for _, cll := range result2.Logs {
			id2LevelAmount, ok := reply[cll.AccountType]
			if !ok {
				id2LevelAmount = map[uint32]map[int]int{}
				reply[cll.AccountType] = id2LevelAmount
			}

			for _, l := range cll.Logs {
				level2Amount, ok := id2LevelAmount[l.CardID]
				if !ok {
					level2Amount = map[int]int{}
					id2LevelAmount[l.CardID] = level2Amount
				}
				for _, lvLog := range l.Levels {
					level2Amount[int(lvLog.Level)] += int(lvLog.Amount)
				}
			}
		}
	}

	var packAtLog = func(at pb.AccountTypeEnum, log *logHub) {
		id2LevelAmount, ok := reply[at]
		if !ok {
			id2LevelAmount = map[uint32]map[int]int{}
			reply[at] = id2LevelAmount
		}

		var packCardLog = func(cl *cardLog) {
			cardID := cl.getCardID()
			level2Amount, ok := id2LevelAmount[cardID]
			if !ok {
				level2Amount = map[int]int{}
				id2LevelAmount[cardID] = level2Amount
			}

			cl.forEachLevel(func(level, amount int) {
				level2Amount[level] += amount
			})
		}

		cl := log.getCardLog(cardID)
		if cl != nil {
			packCardLog(cl)
		} else {
			log.forEachCardLog(packCardLog)
		}
	}

	if accountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(area, accountType, func(log *logHub) {
			packAtLog(accountType, log)
		})
	} else {
		forEachLog(area, packAtLog)
	}
	return reply
}

func (m *cardModule) LogBattleCards(player types.IPlayer, cards *pb.EndFighterData) {
	var cardInfos []glog.ObjectMarshaler
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range cards.InitHandCards {
		cardData := poolGameData.GetCardByGid(card.GCardID)
		if cardData == nil {
			continue
		}

		cardData2 := poolGameData.GetCard(cardData.CardID, 1)

		cardInfos = append(cardInfos, &cardInfoLog{
			name: cardData2.GetName(),
			id:   int32(cardData2.CardID),
		})
	}

	glog.JsonInfo("battlecard", glog.Uint64("uid", uint64(player.GetUid())), glog.String("accountType",
		player.GetLogAccountType().String()), glog.Int("area", player.GetArea()), glog.Int("pvpLevel",
		player.GetPvpLevel()), glog.Objects("battleCards", cardInfos))
}

/*
func (m *cardModule) GetCardPoolLog(accountType pb.AccountTypeEnum, pvpLevel, area int) (
	cardPoolLogs map[pb.AccountTypeEnum]map[int]map[uint32]int, battleAmountLogs map[pb.AccountTypeEnum]map[int]int) {

	cardPoolLogs = map[pb.AccountTypeEnum]map[int]map[uint32]int{}
	battleAmountLogs = map[pb.AccountTypeEnum]map[int]int{}
	appID := uint16(module.Service.GetAppID())
	allApps := config.GetConfig().GetLogicConfigsByName(gconsts.AppGame)
	arg := &pb.GetCardPoolLogArg{AccountType: accountType, PvpLevel: int32(pvpLevel), Area: int32(area)}

	for _, cfg := range allApps {
		if cfg.ID == appID {
			continue
		}

		result, err := logic.CallBackend(gconsts.AppGame, uint32(cfg.ID), pb.MessageID_G2G_GET_CARD_POOL_LOG, arg)
		if err != nil {
			continue
		}

		result2 := result.(*pb.GetCardPoolLogReply)
		for _, cpl := range result2.Logs {
			lv2CardAmount, ok := cardPoolLogs[cpl.AccountType]
			if !ok {
				lv2CardAmount = map[int]map[uint32]int{}
				cardPoolLogs[cpl.AccountType] = lv2CardAmount
			}

			lv2BattleAmount, ok := battleAmountLogs[cpl.AccountType]
			if !ok {
				lv2BattleAmount = map[int]int{}
				battleAmountLogs[cpl.AccountType] = lv2BattleAmount
			}

			for _, l := range cpl.Logs {
				lv2BattleAmount[int(l.PvpLevel)] += int(l.BattleAmount)
				id2Amount, ok := lv2CardAmount[int(l.PvpLevel)]
				if !ok {
					id2Amount = map[uint32]int{}
					lv2CardAmount[int(l.PvpLevel)] = id2Amount
				}
				for _, cLog := range l.CardLogs {
					id2Amount[cLog.CardID] += int(cLog.Amount)
				}
			}
		}
	}

	var packAtLog = func(at pb.AccountTypeEnum, log *logHub) {
		lv2CardAmount, ok := cardPoolLogs[at]
		if !ok {
			lv2CardAmount = map[int]map[uint32]int{}
			cardPoolLogs[at] = lv2CardAmount
		}

		lv2BattleAmount, ok := battleAmountLogs[at]
		if !ok {
			lv2BattleAmount = map[int]int{}
			battleAmountLogs[at] = lv2BattleAmount
		}

		var packLvLog = func(bl *battleLog) {
			lv2BattleAmount[bl.getPvpLevel()] += bl.getAmount()
			id2Amount, ok := lv2CardAmount[bl.getPvpLevel()]
			if !ok {
				id2Amount = map[uint32]int{}
				lv2CardAmount[bl.getPvpLevel()] = id2Amount
			}

			bl.forEachCard(func(cardID uint32, amount int) {
				id2Amount[cardID] += amount
			})
		}

		if pvpLevel > 0 {
			packLvLog(log.getBattleLog(pvpLevel))
		} else {
			log.forEachBattleLog(packLvLog)
		}
	}

	if accountType != pb.AccountTypeEnum_UnknowAccountType {
		forEachAccountTypeLog(area, accountType, func(log *logHub) {
			packAtLog(accountType, log)
		})
	} else {
		forEachLog(area, packAtLog)
	}

	return
}
*/

func onEquipDel(args ...interface{}) {
	player := args[0].(types.IPlayer)
	equipID := args[1].(string)
	player.GetComponent(consts.CardCpt).(*cardComponent).onEquipDel(equipID)
}

func onFixServer1Data(args ...interface{}) {
	player := args[0].(types.IPlayer)
	cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	rebornCnt := module.Reborn.GetRebornCnt(player)
	glog.Infof("cardpool begin fixServer1Data uid=%d, rebornCnt=%d", player.GetUid(), rebornCnt)

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	allCards := cardCpt.GetAllCollectCards()
	// 已有5级卡解锁5级
	var useSkyBook int
	for _, card := range allCards {
		level := card.GetLevel()
		if level >= 5 {
			card.(*collectCard).setMaxUnlockLevel(level)
			data := poolGameData.GetCard(card.GetCardID(), level-1)
			if data != nil {
				useSkyBook += data.ConsumeBook
			}
		}
	}

	var skyBook int
	if rebornCnt > 0 {
		// 曾经下过野
		rebornCntGameData := gamedata.GetGameData(consts.RebornCnt).(*gamedata.RebornCntGameData)
		for cnt := 1; cnt <= rebornCnt; cnt++ {
			skyBook += rebornCntGameData.Cnt2BookAmount[cnt]
		}
		skyBook -= useSkyBook

		if skyBook > 0 {
			resCpt.ModifyResource(consts.SkyBook, skyBook)
		}
	}

	// 回收重生商店的sp武将，返回功勋
	var cardPrice, returnFeats int
	modifyCards := map[uint32]*pb.CardInfo{}
	for _, card := range allCards {
		c := card.(*collectCard)
		if c.getRare() >= 99 && c.GetFrom() == consts.FromReborn {
			cardID := card.GetCardID()
			returnRes, _, _, _ := c.resetRebornSpCard(c.GetCardGameData(), player)
			if len(returnRes) <= 0 {
				continue
			}

			if cardPrice <= 0 {
				cardPrice = returnRes[consts.Feats]
			}

			returnFeats += returnRes[consts.Feats]
			modifyCards[cardID] = &pb.CardInfo{Level: int32(-card.GetLevel())}
			cardCpt.poolDelCard(cardID)
		}
	}

	if len(modifyCards) > 0 {
		cardCpt.ModifyCollectCards(modifyCards)
	}

	if cardPrice == 0 {
		cardPrice = 50000
	}
	returnFeats += int(math.Ceil(float64(resCpt.GetResource(consts.Feats))/float64(cardPrice))) * cardPrice
	if returnFeats > 0 {
		resCpt.SetResource(consts.Feats, returnFeats)
	}

	glog.Infof("cardpool fixServer1Data uid=%d, returnFeats=%d, skyBook=%d, sp=%v", player.GetUid(), returnFeats,
		skyBook, modifyCards)
}

func onReborn(args ...interface{}) {
	player := args[0].(types.IPlayer)
	player.GetComponent(consts.CardCpt).(*cardComponent).onReborn()
}

func Initialize() {
	registerRpc()
	initializeLog()
	module.Card = mod
	eventhub.Subscribe(consts.EvFixServer1Data, onFixServer1Data)
	eventhub.Subscribe(consts.EvEquipDel, onEquipDel)
	eventhub.Subscribe(consts.EvReborn, onReborn)
}
