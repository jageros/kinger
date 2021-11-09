package treasure

import (
	"kinger/gopuppy/common/eventhub"
	"math/rand"
	"time"

	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"strconv"
)

var _ types.ITreasureComponent = &treasureComponent{}

type treasureComponent struct {
	player          types.IPlayer
	gdata           *gamedata.TreasureGameData
	gdataRewardFake *gamedata.TreasureRewardFakeGameData
	gdataDailyFake  *gamedata.TreasureDailyFakeGameData

	attr          *attribute.MapAttr
	treasuresAttr *attribute.ListAttr

	treasures     []*treasureSt
	dailyTreasure *dailyTreasureSt

	addCardEvent *addCardEventSt
	upRareEvent  *upRareEventSt
}

func (tc *treasureComponent) ComponentID() string {
	return consts.TreasureCpt
}

func (tc *treasureComponent) GetPlayer() types.IPlayer {
	return tc.player
}

func (tc *treasureComponent) OnInit(player types.IPlayer) {
	tc.player = player
	tc.gdata = gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
	tc.gdataRewardFake = gamedata.GetGameData(consts.TreasureRewardFake).(*gamedata.TreasureRewardFakeGameData)
	tc.gdataDailyFake = gamedata.GetGameData(consts.TreasureDailyFake).(*gamedata.TreasureDailyFakeGameData)

	if tc.getTotalDailyTreasureCount() <= 0 {
		tc.incrTotalDailyTreasureCount()
	}
	if tc.getTotalRewardTreasureCount() <= 0 {
		tc.incrTotalRewardTreasureCount()
	}

	tc.treasuresAttr = tc.attr.GetListAttr("treasures")
	if tc.treasuresAttr == nil {
		tc.treasuresAttr = attribute.NewListAttr()
		tc.attr.SetListAttr("treasures", tc.treasuresAttr)
	}

	dailyTreasureAttr := tc.attr.GetMapAttr("dailyTreasure")
	if dailyTreasureAttr != nil {
		tc.dailyTreasure = newDailyTreasureByAttr(dailyTreasureAttr)
	}

	tc.treasuresAttr.ForEachIndex(func(index int) bool {
		tc.treasures = append(tc.treasures, newTreasureByAttr(tc.treasuresAttr.GetMapAttr(index)))
		return true
	})

	tc.addCardEvent = newAddCardEvent(player, tc.attr)
	tc.upRareEvent = newUpRareEventSt(tc, tc.attr)

	tc.OnCrossDay(timer.GetDayNo())
}

func (tc *treasureComponent) OnLogin(isRelogin, isRestore bool) {
	if !isRelogin {
		tc.AddDailyTreasure(false)
	}

	//tc.checkCrossDay()
	if !isRestore {
		if tc.player.IsVip() {
			tc.checkTreasuresOfflineTime()
		} else if st := module.OutStatus.GetStatus(tc.player, consts.OtMinVipCard); st != nil {
			tc.checkTreasuresOfflineTime()
		}
	}
}

func (tc *treasureComponent) OnLogout() {
}

func (tc *treasureComponent) getUpRareTreasureID() uint32 {
	return tc.attr.GetUInt32("canUpRareTreasureID")
}

func (tc *treasureComponent) setUpRareTreasureID(treasureID uint32) {
	tc.attr.SetUInt32("canUpRareTreasureID", treasureID)
}

func (tc *treasureComponent) getUpRareTreasureModelID() string {
	return tc.attr.GetStr("upRareTreasureModelID")
}

func (tc *treasureComponent) setUpRareTreasureModelID(modelID string) {
	tc.attr.SetStr("upRareTreasureModelID", modelID)
}

func (tc *treasureComponent) getNextTreasureID() int32 {
	id := tc.attr.GetInt32("nextTreasureID")
	tc.attr.SetInt32("nextTreasureID", id+1)
	return id
}

func (tc *treasureComponent) getTotalRewardTreasureCount() int32 {
	return tc.attr.GetInt32("totalRewardTreasureCount")
}

func (tc *treasureComponent) incrTotalRewardTreasureCount() int32 {
	cnt := tc.getTotalRewardTreasureCount() + 1
	tc.attr.SetInt32("totalRewardTreasureCount", cnt)
	return cnt
}

func (tc *treasureComponent) getTotalDailyTreasureCount() int32 {
	return tc.attr.GetInt32("totalDailyTreasureCount")
}

func (tc *treasureComponent) incrTotalDailyTreasureCount() int32 {
	cnt := tc.getTotalDailyTreasureCount() + 1
	tc.attr.SetInt32("totalDailyTreasureCount", cnt)
	return cnt
}

// 哪一天的数据
func (tc *treasureComponent) getDayno() int {
	return tc.attr.GetInt("dayno")
}

func (tc *treasureComponent) setDayno(dayno int) {
	tc.attr.SetInt("dayno", dayno)
}

// 今天触发了升级品质的次数
func (tc *treasureComponent) getUpRareCnt() int {
	return tc.attr.GetInt("canUpRareCnt")
}

func (tc *treasureComponent) setUpRareCnt(cnt int) {
	tc.attr.SetInt("canUpRareCnt", cnt)
}

// 下次触发升级品质的时间
func (tc *treasureComponent) getNextUpRareTime() int64 {
	return tc.attr.GetInt64("nextUpRareTime")
}

func (tc *treasureComponent) setNextUpRareTime(t int64) {
	tc.attr.SetInt64("nextUpRareTime", t)
}

// 今天触发了加卡广告的次数
func (tc *treasureComponent) getAddCardAdsCnt() int {
	return tc.attr.GetInt("addCardAdsCnt")
}

func (tc *treasureComponent) setAddCardAdsCnt(cnt int) {
	tc.attr.SetInt("addCardAdsCnt", cnt)
}

// 下次触发加卡广告的时间
func (tc *treasureComponent) getNextAddCardAdsTime() int64 {
	return tc.attr.GetInt64("nextAddCardAdsTime")
}

func (tc *treasureComponent) setNextAddCardAdsTime(t int64) {
	tc.attr.SetInt64("nextAddCardAdsTime", t)
}

// 离线自动开箱子
func (tc *treasureComponent) checkTreasuresOfflineTime() {
	lastOnlineTime := tc.player.GetLastOnlineTime()
	offlineTime := int(time.Now().Unix()) - lastOnlineTime
	//glog.Infof("checkTreasuresOfflineTime uid=%d, offlineTime=%d", tc.player.GetUid(), offlineTime)
	if offlineTime < 120 {
		return
	}

	now := int32(time.Now().Unix())
	var treasures []*treasureSt
	var activatedTreasures []*treasureSt
	for _, t := range tc.treasures {
		if t.isActivated() {
			activatedTreasures = append(activatedTreasures, t)
		} else {
			//if t.getRare() < 4 {
			treasures = append(treasures, t)
			//}
		}
	}

	tcnt := len(treasures)
	if tcnt <= 0 {
		// 没有未激活的宝箱
		return
	}

	for _, t := range activatedTreasures {
		openTime := t.getOpenTime()
		if openTime > now {
			// 正在激活的宝箱，还有没够时间开的，后面不用再算了
			return
		}
		if lastOnlineTime >= int(openTime) {
			continue
		}

		offlineTime = offlineTime - (int(openTime) - lastOnlineTime)
	}
	if offlineTime <= 0 {
		return
	}

	// 按品质排序
	for i := tcnt - 1; i > 0; i-- {
		for j := 0; j < i; j++ {
			if treasures[j].getOpenNeedTime() > treasures[j+1].getOpenNeedTime() {
				treasures[j], treasures[j+1] = treasures[j+1], treasures[j]
			}
		}
	}

	for _, t := range treasures {
		if offlineTime <= 0 {
			return
		}

		data := t.getGameData()
		if data == nil {
			continue
		}

		var remainTime int
		unlockTime := module.OutStatus.BuffTreasureTime(tc.player, data.RewardUnlockTime)
		if unlockTime >= offlineTime {
			remainTime = unlockTime - offlineTime
			offlineTime = 0
		} else {
			offlineTime -= unlockTime
		}

		t.setOpenTime(now + int32(remainTime))
		tc.CancelTreasureAddCardAds()
	}
}

func (tc *treasureComponent) getTreasures() ([]*pb.Treasure, *pb.DailyTreasure) {
	var ts []*pb.Treasure
	var dailyMsg *pb.DailyTreasure
	if tc.dailyTreasure != nil {
		dailyMsg = tc.dailyTreasure.packMsg(tc.player).(*pb.DailyTreasure)
	}

	for _, t := range tc.treasures {
		ts = append(ts, t.packMsg(tc.player).(*pb.Treasure))
	}

	return ts, dailyMsg
}

func (tc *treasureComponent) AddRewardTreasureByID(treasureID string, canTriggerUpRare bool) (bool, string) {
	// 找宝箱位置
	poses := [consts.TreasureMaxReward]bool{}
	for _, t := range tc.treasures {
		poses[t.getPos()] = true
	}

	pos := -1
	for i, _ := range poses {
		if !poses[i] {
			pos = i
			break
		}
	}

	if pos < 0 {
		return false, ""
	}

	tdata := tc.gdata.Treasures[treasureID]
	if tdata == nil {
		return false, ""
	}

	tc.incrTotalRewardTreasureCount()
	t := newTreasure(tc.getNextTreasureID(), int32(pos), treasureID)
	tc.treasures = append(tc.treasures, t)
	tc.treasuresAttr.AppendMapAttr(t.attr)

	// 升级品质
	var upRareTreasureModelID string
	if canTriggerUpRare {
		upRareTreasureModelID = tc.upRareEvent.trigger(t)
	}

	glog.Infof("AddRewardTreasure uid=%d, modelID=%s, upRareTreasureModelID=%s", tc.player.GetUid(),
		treasureID, upRareTreasureModelID)

	agnet := tc.player.GetAgent()
	if agnet != nil {
		agnet.PushClient(pb.MessageID_S2C_GAIN_TREASURE, &pb.GainTreasure{Treasure: t.packMsg(tc.player).(*pb.Treasure)})
	}
	return true, upRareTreasureModelID
}

func (tc *treasureComponent) OnCrossDay(dayno int) {
	oldDayno := tc.getDayno()
	if dayno == oldDayno {
		return
	}

	tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).SetResource(consts.PvpTreasureCnt, 0)
	tc.setDayno(dayno)
	tc.addAccTicketEveryDay(oldDayno, dayno)
}

func (tc *treasureComponent) addAccTicketEveryDay(oldDayno, dayno int) {
	if oldDayno <= 0 {
		return
	}

	accTicket := mod.GetDayAccTicketCanAdd(tc.player) * (dayno - oldDayno)
	if accTicket > 0 {
		module.Player.ModifyResource(tc.player, consts.AccTreasureCnt, accTicket)
	}
}

func (tc *treasureComponent) onReborn() {
	resCpt := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	resCpt.SetResource(consts.PvpTreasureCnt, 0)
	resCpt.SetResource(consts.PvpGoldCnt, 0)
	ticket := mod.GetDayAccTicketCanAdd(tc.player)
	if resCpt.GetResource(consts.AccTreasureCnt) < ticket {
		resCpt.SetResource(consts.AccTreasureCnt, ticket)
	}
	tc.AddDailyTreasure(true)
}

func (tc *treasureComponent) AddRewardTreasure(ignoreLimit, canTriggerUpRare bool) (string, pb.NoTreasureReasonEnum, string) {
	//tc.checkCrossDay()
	resCpt := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	addTreasureAmount := resCpt.GetResource(consts.PvpTreasureCnt)
	if !ignoreLimit {
		if addTreasureAmount >= module.OutStatus.BuffTreasureCnt(tc.player, consts.TreasureDayAmountLimit) {
			return "", pb.NoTreasureReasonEnum_AmountLimit, ""
		}
	}

	pvpComponent := tc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	pvpLevel := pvpComponent.GetPvpLevel()
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	r := rgd.Ranks[pvpLevel]
	ts := tc.gdata.TreasuresOfTeam[r.Team]

	if len(ts) == 0 {
		return "", pb.NoTreasureReasonEnum_Unknow, ""
	}

	totalRewardTreasureCount := tc.getTotalRewardTreasureCount()
	var t *gamedata.Treasure = nil

	if rare, ok := tc.gdataRewardFake.Cnt2TreaureRare[int(totalRewardTreasureCount)]; ok {
		for _, t2 := range ts {
			if t2.Rare == rare {
				t = t2
				break
			}
		}
	}

	if t == nil {
		weights := []int{}

		for _, t2 := range ts {
			weights = append(weights, t2.RewardProb)
		}

		i := utils.RandIndexOfWeights(weights)

		if i < 0 {
			return "", pb.NoTreasureReasonEnum_NoPos, ""
		}

		t = ts[i]
	}

	if ok, upRareTreasureModelID := tc.AddRewardTreasureByID(t.ID, canTriggerUpRare); ok {
		if !ignoreLimit {
			resCpt.ModifyResource(consts.PvpTreasureCnt, 1)
		}
		return t.ID, pb.NoTreasureReasonEnum_Unknow, upRareTreasureModelID
	} else {
		return "", pb.NoTreasureReasonEnum_Unknow, ""
	}
}

func (tc *treasureComponent) AddDailyTreasureByID(treasureID string, dayIdx int) bool {
	tdata := tc.gdata.Treasures[treasureID]
	if tdata == nil {
		return false
	}

	tc.incrTotalDailyTreasureCount()
	t := newDailyTreasure(tc.getNextTreasureID(), tdata)
	tc.dailyTreasure = t
	tc.attr.SetMapAttr("dailyTreasure", t.attr)

	glog.Infof("AddDailyTreasure uid=%d, modelID=%s", tc.player.GetUid(), tdata.ID)

	tc.dailyTreasure.setDayIdx(dayIdx)
	tc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_DAILY_TREASURE, t.packMsg(tc.player))
	return true
}

func (tc *treasureComponent) onOldAddDailyTreasure(isReset bool) bool {
	curDayno := timer.GetDayNo()
	if tc.dailyTreasure != nil && !isReset {
		dayno := tc.dailyTreasure.getDayno()
		if !tc.dailyTreasure.isOpen() {
			if dayno != curDayno {
				tc.dailyTreasure.setDayno(curDayno)
			}
			return false
		}

		if dayno == curDayno {
			return false
		}
	}

	pvpComponent := tc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	pvpLevel := pvpComponent.GetPvpLevel()
	if pvpLevel <= 1 {
		pvpLevel = 2
	}
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	r := rgd.Ranks[pvpLevel]
	ts := tc.gdata.TreasuresOfTeam[r.Team]

	if len(ts) == 0 {
		return false
	}

	totalDailyTreasureCount := tc.getTotalDailyTreasureCount()
	var t *gamedata.Treasure = nil

	if rare, ok := tc.gdataDailyFake.Cnt2TreaureRare[int(totalDailyTreasureCount)]; ok {
		for _, t2 := range ts {
			if t2.Rare == rare {
				t = t2
				break
			}
		}
	}

	if t == nil {
		weights := []int{}

		for _, t2 := range ts {
			weights = append(weights, t2.DailyProb)
		}

		i := utils.RandIndexOfWeights(weights)

		if i < 0 {
			return false
		}

		t = ts[i]
	}

	return tc.AddDailyTreasureByID(t.ID, 0)
}

func (tc *treasureComponent) onXfAddDailyTreasure(isReset bool) bool {
	var idx int
	if tc.dailyTreasure != nil && !isReset {
		curDayno := timer.GetDayNo()
		isCrossDay := curDayno != tc.dailyTreasure.getDayno()
		if !tc.dailyTreasure.isOpen() {
			if !isCrossDay {
				return false
			}

			if tc.dailyTreasure.getDayIdx() <= 0 {
				tc.dailyTreasure.setDayno(curDayno)
				return false
			}
		} else if !isCrossDay {
			idx = tc.dailyTreasure.getDayIdx() + 1
		}
	}

	pvpComponent := tc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	pvpLevel := pvpComponent.GetPvpLevel()
	if pvpLevel <= 1 {
		pvpLevel = 2
	}
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	r := rgd.Ranks[pvpLevel]
	ts := tc.gdata.Team2DailyTreasure[r.Team]

	if len(ts) == 0 {
		return false
	}

	/*
		totalDailyTreasureCount := tc.getTotalDailyTreasureCount()
		var t *gamedata.Treasure = nil

		if rare, ok := tc.gdataDailyFake.Cnt2TreaureRare[int(totalDailyTreasureCount)]; ok {
			for _, t2 := range ts {
				if t2.Rare == rare {
					t = t2
					break
				}
			}
		}
	*/

	if idx >= 3 {
		// 超过3个
		return false
	}

	if len(ts) < idx+1 {
		return false
	}
	t := ts[idx]

	if tc.AddDailyTreasureByID(t.ID, idx) {
		return true
	} else {
		return false
	}
}

func (tc *treasureComponent) AddDailyTreasure(isReset bool) bool {
	if config.GetConfig().IsXfServer() {
		return tc.onXfAddDailyTreasure(isReset)
	} else {
		return tc.onOldAddDailyTreasure(isReset)
	}
}

func (tc *treasureComponent) AddDailyTreasureStar(star int) {
	if tc.dailyTreasure == nil {
		return
	}

	starCount := tc.dailyTreasure.getOpenStarCount()
	if starCount >= 1 {
		newStar := starCount - int32(star)
		if newStar < 0 {
			newStar = 0
		}
		tc.dailyTreasure.setOpenStartCount(newStar)
		tc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_DAILY_TREASURE, tc.dailyTreasure.packMsg(tc.player))
	}
}

func (tc *treasureComponent) OpenTreasureByModelID(modelID string, isDobule bool, ignoreRewardTbl ...bool) *pb.OpenTreasureReply {
	return tc.openTreasureWithType(modelID, isDobule, ttUnknowTreasure, len(ignoreRewardTbl) > 0 && ignoreRewardTbl[0])
}

func (tc *treasureComponent) treasureGetUnlockCards(data *gamedata.Treasure) (unlockCards []uint32,
	rare2UnlockCardSet map[int]common.UInt32Set) {

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	team := data.Team
	if team <= 0 {
		team = tc.player.GetPvpTeam()
	}
	if team <= 1 {
		team = 2
	}

	rs := rgd.RanksOfTeam[team]
	unlockCardsSet := common.UInt32Set{}
	rare2UnlockCardSet = map[int]common.UInt32Set{}

	unlockCards = append(unlockCards, gamedata.GetGameData(
		consts.EventFirstRechargeReward).(*gamedata.ActivityFirstRechargeRewardGameData).UnlockCards...)

	var _getUnlockCard = func(cardIDs []uint32) {
		for _, cardID := range cardIDs {
			cardData := poolGameData.GetCard(uint32(cardID), 1)
			if cardData == nil {
				continue
			}

			if data.Camp > 0 && cardData.Camp != data.Camp {
				continue
			}

			unlockCardsSet.Add(cardID)
			if cardIDSet, ok := rare2UnlockCardSet[cardData.Rare]; ok {
				cardIDSet.Add(cardID)
			} else {
				cardIDSet = common.UInt32Set{}
				cardIDSet.Add(cardID)
				rare2UnlockCardSet[cardData.Rare] = cardIDSet
			}
		}
	}

	for _, r := range rs {
		_getUnlockCard(r.Unlock)
	}
	_getUnlockCard(module.Card.GetFirstRechargeUnlockCards(tc.player))

	levelCpt := tc.player.GetComponent(consts.LevelCpt).(types.ILevelComponent)
	for _, cardID := range levelCpt.GetUnlockCards() {
		cardData := poolGameData.GetCard(uint32(cardID), 1)
		if cardData == nil {
			continue
		}
		if data.Camp > 0 && cardData.Camp != data.Camp {
			continue
		}

		unlockCardsSet.Add(cardID)

		if cardIDSet, ok := rare2UnlockCardSet[cardData.Rare]; ok {
			cardIDSet.Add(cardID)
		} else {
			cardIDSet = common.UInt32Set{}
			cardIDSet.Add(cardID)
			rare2UnlockCardSet[cardData.Rare] = cardIDSet
		}
	}

	unlockCards = unlockCardsSet.ToList()
	utils.ShuffleUInt32(unlockCards)
	return
}

func (tc *treasureComponent) treasureGetCollectCards(data *gamedata.Treasure) (rare2CollectCardIDs map[int][]uint32) {
	cardComponent := tc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	allCollectCards := cardComponent.GetAllCollectCards()
	var collectCardIDs []uint32
	rare2CollectCardIDs = map[int][]uint32{}
	for _, c := range allCollectCards {
		cardData := c.GetCardGameData()
		if cardData == nil || cardData.IsSpCard() || (data.Camp > 0 && data.Camp != cardData.Camp) {
			continue
		}

		cardID := cardData.GetCardID()
		collectCardIDs = append(collectCardIDs, cardID)
		if cardIDs, ok := rare2CollectCardIDs[cardData.Rare]; ok {
			rare2CollectCardIDs[cardData.Rare] = append(cardIDs, cardID)
		} else {
			rare2CollectCardIDs[cardData.Rare] = []uint32{cardID}
		}
	}
	return
}

// 品质卡牌，优先从FakeRandomCard取
func (tc *treasureComponent) treasureRandomRareCards(data *gamedata.Treasure, rare2CollectCardIDs map[int][]uint32,
	rare2UnlockCardSet map[int]common.UInt32Set, cardCnt int) (uniqueCardIDs []uint32, newCardCnt, fakeRandomCard, randomCard int) {

	fakeRandomCard = data.FakeRandomCard
	randomCard = data.RandomCard
	for i, cardStar := range []int{data.CardStar5, data.CardStar4, data.CardStar3, data.CardStar2, data.CardStar1} {
		if cardStar <= 0 {
			continue
		}

		if fakeRandomCard <= 0 && randomCard <= 0 {
			break
		}

		curRare := 5 - i
		cardStar2 := cardStar
		if fakeRandomCard > 0 {
			amount := cardStar
			if amount > fakeRandomCard {
				amount = fakeRandomCard
			}
			if amount > cardCnt {
				amount = cardCnt
			}

			rareCollectCards := rare2CollectCardIDs[curRare]
			rareCollectCardsAmount := len(rareCollectCards)
			if amount > rareCollectCardsAmount {
				amount = rareCollectCardsAmount
			}

			fakeRandomCard -= amount
			cardStar2 -= amount
			cardCnt -= amount

			if amount > 0 {
				utils.ShuffleUInt32(rareCollectCards)
				uniqueCardIDs = append(uniqueCardIDs, utils.RandUInt32Sample(rareCollectCards, amount, false)...)
			}
		}

		if randomCard > 0 && cardStar2 > 0 {
			amount := cardStar2
			if amount > randomCard {
				amount = randomCard
			}
			if amount > cardCnt {
				amount = cardCnt
			}

			curRareUnlockCardSet := rare2UnlockCardSet[curRare]
			curRareUnlockCard := curRareUnlockCardSet.ToList()
			curRareUnlockCardAmount := len(curRareUnlockCard)
			if amount > curRareUnlockCardAmount {
				amount = curRareUnlockCardAmount
			}

			randomCard -= amount
			cardStar2 -= amount
			cardCnt -= amount

			if amount > 0 {
				utils.ShuffleUInt32(curRareUnlockCard)
				uniqueCardIDs = append(uniqueCardIDs, utils.RandUInt32Sample(curRareUnlockCard, amount, false)...)
			}
		}
	}

	newCardCnt = cardCnt
	return
}

func (tc *treasureComponent) treasureRandomNewCards(data *gamedata.Treasure, randomCard, cardCnt int, unlockCards []uint32) (
	newCardIDs []uint32, newRandomCard, newCardCnt int) {

	newCardNum := data.GetNewCardNum()
	cardComponent := tc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	if newCardNum > 0 && randomCard > 0 && cardCnt > 0 {
		for i := 0; i < len(unlockCards); {
			cardID := unlockCards[i]
			card := cardComponent.GetCollectCard(uint32(cardID))
			if card == nil {
				newCardNum--
				randomCard--
				cardCnt--
				newCardIDs = append(newCardIDs, cardID)
				unlockCards = append(unlockCards[:i], unlockCards[i+1:]...)
				if newCardNum <= 0 || randomCard <= 0 || cardCnt <= 0 {
					break
				}
			} else {
				i++
			}
		}
	}

	newRandomCard = randomCard
	newCardCnt = cardCnt
	return
}

func (tc *treasureComponent) treasureRandomCards(data *gamedata.Treasure, isDobule bool, type_ int) []uint32 {
	// 共能获得多少张卡
	cardCnt := data.CardCnt

	if type_ == ttDailyTreasure {
		cardCnt = module.OutStatus.BuffDayTreasureCard(tc.player, cardCnt)
	}

	if isDobule {
		cardCnt *= 2
	}

	if type_ == ttRewardTreasure {
		cardCnt = module.OutStatus.BuffTreasureCard(tc.player, cardCnt)
		cardCnt = module.OutStatus.BuffAddCardOfVip(tc.player, cardCnt)
	}

	// 获得了哪些卡
	var cardIDs []uint32
	if len(data.Reward) > 0 {
		// 指定卡牌
		cardCnt -= len(data.Reward)
		cardIDs = append(cardIDs, data.Reward...)
	}

	if cardCnt > 0 {
		unlockCards, rare2UnlockCardSet := tc.treasureGetUnlockCards(data)
		//glog.Debugf("unlockCards = %s", unlockCards)
		rare2CollectCardIDs := tc.treasureGetCollectCards(data)

		uniqueCardIDs, newCardCnt, fakeRandomCard, randomCard := tc.treasureRandomRareCards(data, rare2CollectCardIDs,
			rare2UnlockCardSet, cardCnt)
		cardCnt = newCardCnt

		var newCardIDs []uint32
		newCardIDs, randomCard, cardCnt = tc.treasureRandomNewCards(data, randomCard, cardCnt, unlockCards)
		uniqueCardIDs = append(uniqueCardIDs, newCardIDs...)

		if randomCard > 0 {
			// 从解锁牌池随机卡
			amount := randomCard
			if randomCard > cardCnt {
				amount = cardCnt
			}
			cardCnt -= amount
			uniqueCardIDs = append(uniqueCardIDs, utils.RandUInt32Sample(unlockCards, amount, false)...)
		}

		if fakeRandomCard > 0 && cardCnt > 0 {
			// 从已有的牌随机卡
			cardComponent := tc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
			allCollectCards := cardComponent.GetAllCollectCards()
			var uniqueFakeCardIDs []uint32
			for _, c := range allCollectCards {
				cardData := c.GetCardGameData()
				if cardData == nil || cardData.IsSpCard() || (data.Camp > 0 && data.Camp != cardData.Camp) {
					continue
				}
				uniqueFakeCardIDs = append(uniqueFakeCardIDs, cardData.GetCardID())
			}

			if len(uniqueFakeCardIDs) > 0 {
				amount := fakeRandomCard
				if amount > cardCnt {
					amount = cardCnt
				}
				if amount > len(uniqueFakeCardIDs) {
					amount = len(uniqueFakeCardIDs)
				}
				cardCnt -= amount
				uniqueCardIDs = append(uniqueCardIDs, utils.RandUInt32Sample(uniqueFakeCardIDs, amount, false)...)
			}
		}

		if cardCnt > 0 && len(uniqueCardIDs) > 0 {
			uniqueCardIDs = append(uniqueCardIDs, utils.RandUInt32Sample(uniqueCardIDs, cardCnt, true)...)
		}

		cardIDs = append(cardIDs, uniqueCardIDs...)
	}

	cardsChange := map[uint32]*pb.CardInfo{}
	for _, cardID := range cardIDs {
		if ci, ok := cardsChange[cardID]; ok {
			ci.Amount += 1
		} else {
			cardsChange[cardID] = &pb.CardInfo{CardId: cardID, Amount: 1}
		}
	}

	cardComponent := tc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardComponent.ModifyCollectCards(cardsChange)
	return cardIDs
}

func (tc *treasureComponent) caclTreasureResource(data *gamedata.Treasure, isDobule bool, type_ int) map[int]int {
	gold := data.GoldMin
	goldRand := data.GoldMax - data.GoldMin + 1
	if goldRand > 0 {
		gold += rand.Intn(goldRand)
	}

	jade := data.JadeMin
	jadeRand := data.JadeMax - data.JadeMin + 1
	if jadeRand > 0 {
		jade += rand.Intn(jadeRand)
	}

	bowlder := data.BowlderMin
	bowlderRand := data.BowlderMax - data.BowlderMin + 1
	if bowlderRand > 0 {
		bowlder += rand.Intn(bowlderRand)
	}

	var eventItemAmount, eventItemType int
	if data.EventItemCnt > 0 && data.EventProp > 0 {
		if rand.Intn(10000) < data.EventProp {
			eventItemAmount = data.EventItemCnt
			eventItemType = module.Activitys.OnGetSpringHuodongItemType(tc.player, eventItemAmount)
		}
	}

	if type_ == ttRewardTreasure {
		gold = module.OutStatus.BuffTreasureGold(tc.player, gold)
	}
	if type_ == ttDailyTreasure {
		gold = module.OutStatus.BuffDayTreasureGlod(tc.player, gold)
	}
	if isDobule {
		gold *= 2
		jade *= 2
	}

	resources := map[int]int{}
	if gold > 0 {
		resources[consts.Gold] = gold
	}
	if jade > 0 {
		resources[consts.Jade] = jade
	}
	if bowlder > 0 {
		resources[consts.Bowlder] = bowlder
	}
	if data.Star > 0 {
		resources[consts.Score] = data.Star
	}
	if eventItemAmount > 0 && eventItemType > 0 {
		resources[eventItemType] = eventItemAmount
	}
	return resources
}

func (tc *treasureComponent) caclTreasureRewardTbl(data *gamedata.Treasure, msg *pb.OpenTreasureReply) {
	if data.RewardTbl == "" {
		return
	}

	rr := module.Reward.GiveReward(tc.player, data.RewardTbl)
	if rr == nil {
		return
	}

	msg.CardSkins = append(msg.CardSkins, rr.GetCardSkins()...)
	msg.Headframes = append(msg.Headframes, rr.GetHeadFrames()...)

	resources := rr.GetResources()
L1:
	for resType, amount := range resources {
		resType2 := int32(resType)
		for _, resMsg := range msg.Resources {
			if resMsg.Type == resType2 {
				resMsg.Amount += int32(amount)
				continue L1
			}
		}

		msg.Resources = append(msg.Resources, &pb.Resource{
			Type:   resType2,
			Amount: int32(amount),
		})
	}

	cards := rr.GetCards()
	for cardID, amount := range cards {
		for i := 0; i < amount; i++ {
			msg.CardIDs = append(msg.CardIDs, cardID)
		}
	}

	emojis := rr.GetEmojis()
	for _, emojiTeam := range emojis {
		msg.EmojiTeams = append(msg.EmojiTeams, int32(emojiTeam))
	}

	resources = rr.GetConvertResources()
L2:
	for resType, amount := range resources {
		resType2 := int32(resType)
		for _, resMsg := range msg.ConvertResources {
			if resMsg.Type == resType2 {
				resMsg.Amount += int32(amount)
				continue L2
			}
		}

		msg.ConvertResources = append(msg.ConvertResources, &pb.Resource{
			Type:   resType2,
			Amount: int32(amount),
		})
	}

	cards = rr.GetUpLevelRewardCards()
	for cardID, amount := range cards {
		for i := 0; i < amount; i++ {
			msg.UpLevelRewardCards = append(msg.UpLevelRewardCards, cardID)
		}
	}
}

func (tc *treasureComponent) openTreasureWithType(modelID string, isDobule bool, type_ int, ignoreRewardTbl bool) *pb.OpenTreasureReply {
	reply := &pb.OpenTreasureReply{}
	data := tc.gdata.Treasures[modelID]
	if data == nil {
		return reply
	}

	cardIDs := tc.treasureRandomCards(data, isDobule, type_)

	resources := tc.caclTreasureResource(data, isDobule, type_)
	if len(resources) > 0 {
		resComponent := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		oldPvpLevel := tc.player.GetPvpLevel()
		resComponent.BatchModifyResource(resources, consts.RmrTreasure+modelID)
		newPvpLevel := tc.player.GetPvpLevel()

		if resources[consts.Score] > 0 {
			logic.PushBackend("", 0, pb.MessageID_G2R_UPDATE_PVP_SCORE, module.Player.PackUpdateRankMsg(
				tc.player, []*pb.SkinGCard{}, 0))
		}
		if newPvpLevel > oldPvpLevel {
			reply.UpLevelRewardCards = tc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).UplevelReward()
		}
	}

	glog.Infof("open treasure: uid=%d, modelID=%s, cardIDs=%v, resource=%v, cardSkins=%v", tc.player.GetUid(),
		data.ID, cardIDs, resources, data.CardSkins)

	module.Mission.OnOpenTreasure(tc.player)

	for _, cardSkinID := range data.CardSkins {
		module.Bag.AddCardSkin(tc.player, cardSkinID)
	}
	var emojiTeams []int32
	for _, emojiID := range data.Emojis {
		module.Bag.AddEmoji(tc.player, emojiID)
		emojiTeams = append(emojiTeams, int32(emojiID))
	}
	for _, headFrame := range data.HeadFrames {
		module.Bag.AddHeadFrame(tc.player, headFrame)
	}

	hdSkins, eventItem := module.Huodong.GetTreasureHuodongSkin(tc.player, data)
	reply.CardSkins = append(reply.CardSkins, hdSkins...)
	if eventItem > 0 {
		reply.ConvertResources = append(reply.ConvertResources, &pb.Resource{
			Type:   int32(consts.EventItem1),
			Amount: int32(eventItem),
		})
	}

	for resType, amount := range resources {
		reply.Resources = append(reply.Resources, &pb.Resource{
			Type:   int32(resType),
			Amount: int32(amount),
		})
	}

	reply.OK = true
	reply.CardIDs = cardIDs
	reply.CardSkins = data.CardSkins
	reply.EmojiTeams = emojiTeams
	reply.Headframes = data.HeadFrames
	reply.TreasureID = modelID
	if !ignoreRewardTbl {
		tc.caclTreasureRewardTbl(data, reply)
	}
	return reply
}

func (tc *treasureComponent) triggerAddCardAds(treasureID uint32, treasureModelID string, cardIDs []uint32) bool {
	return tc.addCardEvent.trigger(treasureID, treasureModelID, cardIDs)
}

func (tc *treasureComponent) openTreasure(treasureID int) (*pb.OpenTreasureReply, bool, string) {
	isDaily := tc.IsDailyTreasure(uint32(treasureID))
	var t iTreasure = nil
	if isDaily {
		if tc.dailyTreasure != nil {
			t = tc.dailyTreasure
		}
	} else {
		t2 := tc.getTreasureByID(uint32(treasureID))
		if t2 != nil {
			t = t2
		}
	}

	if t == nil || t.isOpen() || !t.isCanOpen() {
		return &pb.OpenTreasureReply{}, isDaily, ""
	}

	if !isDaily {
		tc.treasuresAttr.DelMapAttr(t.getAttr())
		for i, t2 := range tc.treasures {
			if t2.getID() == uint32(treasureID) {
				tc.treasures = append(tc.treasures[:i], tc.treasures[i+1:]...)
			}
		}
	} else {
		t.opened()
		tc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_DAILY_TREASURE, t.packMsg(tc.player))
	}

	modelID := t.getModelID()
	treasureType := ttRewardTreasure
	if isDaily {
		treasureType = ttDailyTreasure
	}
	treasureReward := tc.openTreasureWithType(modelID, isDaily && t.isDobule(tc.player), treasureType, false)

	if isDaily && treasureReward != nil && treasureReward.OK && config.GetConfig().IsXfServer() {
		tc.AddDailyTreasure(false)
	}

	eventhub.Publish(consts.EvOpenTreasure, tc.player, isDaily)

	return treasureReward, isDaily, modelID
}

func (tc *treasureComponent) activateRewardTreasure(treasureID int) (int32, bool) {
	t := tc.getTreasureByID(uint32(treasureID))
	if t == nil || t.isActivated() {
		return -1, false
	}

	tdata := t.getGameData()
	if tdata == nil {
		return -1, false
	}

	unlockTime := module.OutStatus.BuffTreasureTime(tc.player, tdata.RewardUnlockTime)
	openTimeout := int32(unlockTime)
	now := int32(time.Now().Unix())
	openTime := now + openTimeout
	t.setOpenTime(openTime)
	t.setActivateTime(now)
	tc.CancelTreasureAddCardAds()
	return openTimeout, true
}

func (tc *treasureComponent) IsDailyTreasure(treasureID uint32) bool {
	return tc.dailyTreasure != nil && tc.dailyTreasure.getID() == treasureID
}

func (tc *treasureComponent) getOldDailyTreasureDobulePrice() int {
	return gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).DailyDouble
}

func (tc *treasureComponent) getXfDailyTreasureDobulePrice() int {
	if tc.dailyTreasure == nil {
		return 0
	}

	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(
		tc.dailyTreasure.getRare())
	if eventConfig == nil {
		return 0
	}
	return eventConfig.Double
}

func (tc *treasureComponent) DailyTreasureBeDobule(isConsumeJade bool) bool {
	if tc.dailyTreasure == nil || tc.dailyTreasure.isOpen() || tc.dailyTreasure.isDobule(tc.player) {
		return false
	}

	var needJade int
	if isConsumeJade {
		if config.GetConfig().IsXfServer() {
			needJade = tc.getXfDailyTreasureDobulePrice()
		} else {
			needJade = tc.getOldDailyTreasureDobulePrice()
		}

		if needJade <= 0 {
			return false
		}
		if !tc.player.HasBowlder(needJade) {
			return false
		}
		tc.player.SubBowlder(needJade, consts.RmrDailyTreasureDouble)

		module.Shop.LogShopBuyItem(tc.player, "dailyTreasureDouble", "每日宝箱翻倍", 1, "gameplay",
			strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), needJade, "")
	}

	glog.Infof("DailyTreasureBeDobule uid=%d, treasureID=%d, isConsumeJade=%v. needJade=%d", tc.player.GetUid(),
		tc.dailyTreasure.getID(), isConsumeJade, needJade)
	tc.dailyTreasure.beDouble()
	return true
}

func (tc *treasureComponent) AccTreasure(treasureID uint32, isConsumeJade bool) (int32, error) {
	t := tc.getTreasureByID(treasureID)
	if t == nil {
		return 0, gamedata.GameError(1)
	}

	now := int32(time.Now().Unix())
	openTime := t.getOpenTime()
	if now >= openTime {
		return 0, nil
	}

	var needJade int
	/*
		if isConsumeJade {
			if config.GetConfig().IsMultiLan && tc.player.IsVip() {

			} else {
				needJade = tc.player.GetSkipAdsNeedJade()
				if !tc.player.HasBowlder(needJade) {
					return 0, gamedata.GameError(2)
				}
				tc.player.SubBowlder(needJade)
			}
		}
	*/

	glog.Infof("AccTreasure uid=%d, treasureID=%d, isConsumeJade=%v, needJade=%d", tc.player.GetUid(), treasureID,
		isConsumeJade, needJade)
	openTime -= int32(config.GetConfig().Wxgame.AccTreasureTime)
	t.setOpenTime(openTime)
	tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).ModifyResource(consts.AccTreasureCnt, -1,
		consts.RmrUnknownConsume)
	openTimeout := openTime - now
	if openTimeout < 0 {
		openTimeout = 0
	}
	return openTimeout, nil
}

func (tc *treasureComponent) HelpOpenTreasure(treasureID uint32) {
	if tc.dailyTreasure != nil && tc.dailyTreasure.getID() == treasureID {
		if tc.dailyTreasure.isOpen() || tc.dailyTreasure.isDobule(tc.player) {
			return
		}
		tc.DailyTreasureBeDobule(false)
		//tc.player.GetAgent().PushClient(pb.MessageID_S2C_DAILY_TREASURE_BE_HELP, nil)
	} else {

		openTimeout, err := tc.AccTreasure(treasureID, false)
		if err != nil {
			return
		}
		tc.player.GetAgent().PushClient(pb.MessageID_S2C_TREASURE_BE_HELP, &pb.TreasureBeHelp{
			TreasureID:  treasureID,
			OpenTimeout: openTimeout,
		})
	}
}

func (tc *treasureComponent) getTreasureByID(treasureID uint32) *treasureSt {
	for _, t := range tc.treasures {
		if t.getID() == treasureID {
			return t
		}
	}
	return nil
}

func (tc *treasureComponent) jadeAccTreasure(treasureID uint32) error {
	t := tc.getTreasureByID(treasureID)
	if t == nil {
		return gamedata.InternalErr
	}

	now := float64(time.Now().Unix())
	openTime := float64(t.getOpenTime())
	remainTime := openTime - now
	if remainTime <= 0 {
		return nil
	}

	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	count := remainTime / funcPrice.AccTreasure
	var needJade int
	if count > float64(int(count)) {
		needJade = int(count) + 1
	} else {
		needJade = int(count)
	}

	if !tc.player.HasBowlder(needJade) {
		return gamedata.GameError(1)
	}

	tc.player.SubBowlder(needJade, consts.RmrAccTreasure)
	t.setOpenTime(int32(now))
	module.Mission.OnAccTreasure(tc.player)

	module.Shop.LogShopBuyItem(tc.player, "accTreasure", "加速宝箱", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), needJade, "")

	return nil
}

func (tc *treasureComponent) UpTreasureRare(isConsumeJade bool) (*pb.Treasure, error) {
	return tc.upRareEvent.doAction()
}

func (tc *treasureComponent) WatchTreasureAddCardAds(treasureID uint32, isConsumeJade bool) (
	*pb.WatchTreasureAddCardAdsReply, error) {
	return tc.addCardEvent.doAction(treasureID)
}

func (tc *treasureComponent) CancelTreasureAddCardAds() {
	tc.addCardEvent.cancel()
}

func (tc *treasureComponent) wxHelpDoubleDailyTreasure(treasureID uint32, helperUid common.UUid, helperHeadImg,
	helperHeadFrame, helperName string) bool {

	if tc.dailyTreasure == nil || tc.dailyTreasure.getID() != treasureID {
		return false
	}

	ok := tc.DailyTreasureBeDobule(false)
	if !ok {
		return false
	}

	tc.dailyTreasure.setHelper(helperUid, helperHeadImg, helperHeadFrame, helperName)
	return true
}
