package cardpool

import (
	"kinger/gopuppy/common/eventhub"
	"math"
	"math/rand"
	"strconv"

	"encoding/json"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"sort"
	"kinger/apps/game/module"
	"kinger/common/config"
)

const compensateVersion = 1

var _ types.ICardComponent = &cardComponent{}

type collectCardList []*collectCard

func (cl collectCardList) Len() int {
	return len(cl)
}

func (cl collectCardList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

func (cl collectCardList) Less(i, j int) bool {
	useCount1 := cl[i].getUseCount()
	useCount2 := cl[j].getUseCount()
	if useCount1 > useCount2 {
		return true
	} else if useCount1 == useCount2 {
		return cl[i].getLastUseTime() > cl[j].getLastUseTime()
	} else {
		return false
	}
}

type cardComponent struct {
	poolGData *gamedata.PoolGameData
	diyGData  *gamedata.DiyGameData
	player    types.IPlayer
	attr      *attribute.MapAttr

	// 当前收集的卡
	collectCardMap map[uint32]*collectCard
	collectCards   collectCardList
	// 曾经有，但现在没有的卡
	onceCards map[uint32]*collectCard
	// 所有diy卡
	diycardMap map[uint32]*diyCard
	// map[poolID]*pvpCardPool
	pvpCardPools map[int]*pvpCardPool
	// map[camp]*pvpCardPool  各阵营出战卡组
	pvpCampFightPool map[int]*pvpCardPool
}

func (cc *cardComponent) ComponentID() string {
	return consts.CardCpt
}

func (cc *cardComponent) GetPlayer() types.IPlayer {
	return cc.player
}

func (cc *cardComponent) OnInit(player types.IPlayer) {
	cc.player = player
	cc.poolGData = gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	//cc.diyGData = gamedata.GetGameData(consts.Diy).(*gamedata.DiyGameData)
	cc.initPvpCardPool()
	cc.initCollectCard()
	cc.initDiyCard()
}

func (cc *cardComponent) OnLogin(isRelogin, isRestore bool) {
	cVersion := cc.attr.GetInt("compensateVersion")
	needCompensate := cVersion < compensateVersion
	cc.attr.SetInt("compensateVersion", compensateVersion)
	modifyCards := map[uint32]*pb.CardInfo{}
	for _, c := range cc.collectCardMap {
		if needCompensate && c.GetLevel() == 2 {
			modifyCards[c.GetCardID()] = &pb.CardInfo{
				Amount: 10,
			}
		}
		c.Reset(cc.player)
		if _, ok := cc.collectCardMap[c.GetCardID()]; !ok {
			continue
		}

		equipID := c.GetEquip()
		if equipID != "" {
			it, ok := module.Bag.GetItem(cc.player, consts.ItEquip, equipID).(types.IEquipItem)
			if ok {
				it.SetOwner(c.GetCardID())
			} else {
				c.DeEquip()
			}
		}
	}

	if len(modifyCards) > 0 {
		cc.player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(modifyCards)
	}
}

func (cc *cardComponent) OnLogout() {
}

func (cc *cardComponent) initPvpCardPool() {
	// 各阵营出战卡组
	poolsAttr := cc.attr.GetListAttr("pvpCardPool")
	cc.pvpCardPools = make(map[int]*pvpCardPool)
	cc.pvpCampFightPool = make(map[int]*pvpCardPool)

	if poolsAttr == nil {
		// newbie
		cc.attr.SetInt("compensateVersion", compensateVersion)
		poolsAttr = attribute.NewListAttr()
		cc.attr.SetListAttr("pvpCardPool", poolsAttr)
		poolID := 1
		for _, camp := range []int{consts.Shu, consts.Wei, consts.Wu} {
			for i := 0; i < 3; i++ {
				pool := newPvpCardPool(poolID, camp, i == 0, []uint32{0, 0, 0, 0, 0})
				poolsAttr.AppendMapAttr(pool.getAttr())
				poolID++
				cc.pvpCardPools[pool.getPoolID()] = pool
				if i == 0 {
					cc.pvpCampFightPool[pool.getCamp()] = pool
				}
			}
		}
	} else {
		poolsAttr.ForEachIndex(func(index int) bool {
			poolAttr := poolsAttr.GetMapAttr(index)
			pool := newPvpCardPoolByAttr(poolAttr)
			cc.pvpCardPools[pool.getPoolID()] = pool
			if pool.isFight() {
				cc.pvpCampFightPool[pool.getCamp()] = pool
			}
			return true
		})
	}
}

func (cc *cardComponent) initCollectCard() {
	cc.collectCardMap = make(map[uint32]*collectCard)
	collectCardsAttr := cc.attr.GetMapAttr("collectCards")
	if collectCardsAttr == nil {
		collectCardsAttr = attribute.NewMapAttr()
		cc.attr.SetMapAttr("collectCards", collectCardsAttr)
	}

	cc.onceCards = map[uint32]*collectCard{}
	collectCardsAttr.ForEachKey(func(key string) {
		card := newCollectCardByAttr(collectCardsAttr.GetMapAttr(key))
		if card.GetLevel() > 0 {
			cc.collectCardMap[card.GetCardID()] = card
			cc.collectCards = append(cc.collectCards, card)
		} else {
			cc.onceCards[card.GetCardID()] = card
		}
	})

	sort.Sort(cc.collectCards)
}

func (cc *cardComponent) initDiyCard() {
	cc.diycardMap = make(map[uint32]*diyCard)
	diyCardsAttr := cc.attr.GetMapAttr("diycard")
	if diyCardsAttr == nil {
		diyCardsAttr = attribute.NewMapAttr()
		cc.attr.SetMapAttr("diycard", diyCardsAttr)
	}

	diyCardsAttr.ForEachKey(func(key string) {
		card := newDiyCardByAttr(diyCardsAttr.GetMapAttr(key))
		cc.diycardMap[card.GetCardID()] = card
	})
}

func (cc *cardComponent) GetCollectCard(cardId uint32) types.ICollectCard {
	if card, ok := cc.collectCardMap[cardId]; ok {
		return card
	} else {
		return nil
	}
}

func (cc *cardComponent) GetCollectCardNumByLevel(lvl int) int {
	var counts int
	for _, card := range cc.collectCardMap {
		if lvl == 0 || lvl == card.GetLevel(){
			counts++
		}
	}
	return counts
}

func (cc *cardComponent) GetCollectCardNumByStar(star int) int {
	var counts int
	for _, card := range cc.collectCardMap {
		if star == 0 || star == card.getRare() {
			counts++
		}
	}
	return counts
}

func (cc *cardComponent) getCollectCardByGCardID(gcardID uint32) types.ICollectCard {
	cardData := cc.poolGData.GetCardByGid(gcardID)
	if cardData == nil {
		return nil
	}
	return cc.GetCollectCard(cardData.CardID)
}

func (cc *cardComponent) getDiyCard(cardId uint32) *diyCard {
	return cc.diycardMap[cardId]
}

func (cc *cardComponent) PackDiyCardMsg() []*pb.DiyCardData {
	var msg []*pb.DiyCardData
	for _, card := range cc.diycardMap {
		msg = append(msg, card.packDataMsg())
	}
	return msg
}

func (cc *cardComponent) GetPvpCardPoolByCamp(camp int) []types.ICollectCard {
	var cardIDs []uint32
	for _, pool := range cc.pvpCardPools {
		if pool.getCamp() == camp && pool.isFight() {
			cardIDs = pool.getCards()
			break
		}
	}

	var cards []types.ICollectCard
	for _, cardID := range cardIDs {
		card := cc.GetCollectCard(cardID)
		if card != nil {
			cards = append(cards, card)
		}
	}
	return cards
}

func (cc *cardComponent) onReborn() {
	for _, pool := range cc.pvpCardPools {
		pool.reset()
	}
	cc.collectCards = collectCardList{}
	for _, card := range cc.collectCardMap {
		cc.collectCards = append(cc.collectCards, card)
	}
	sort.Sort(cc.collectCards)
}

func (cc *cardComponent) GetAllCollectCards() []types.ICollectCard {
	var cards []types.ICollectCard
	for _, card := range cc.collectCardMap {
		cards = append(cards, card)
	}
	return cards
}

func (cc *cardComponent) GetCollectCardLevelInfo() (maxLevel, minLevel, avlLevel int) {
	amount := len(cc.collectCardMap)
	if amount <= 0 {
		return
	}

	totalLevel := 0
	minLevel = 10000
	for _, card := range cc.collectCardMap {
		lv := card.GetLevel()
		if lv > maxLevel {
			maxLevel = lv
		}

		if lv < minLevel {
			minLevel = lv
		}

		totalLevel += lv
	}

	avlLevel = totalLevel / amount

	return
}

func (cc *cardComponent) GetAllCollectCardDatas() []*gamedata.Card {
	var cardDatas []*gamedata.Card
	for _, card := range cc.collectCardMap {
		cardDatas = append(cardDatas, card.GetCardGameData())
	}
	return cardDatas
}

func (cc *cardComponent) getAllPvpCardPool() []*pvpCardPool {
	var poolIDs []int
	for id, _ := range cc.pvpCardPools {
		poolIDs = append(poolIDs, id)
	}
	sort.Ints(poolIDs)

	var pools []*pvpCardPool
	for _, id := range poolIDs {
		pools = append(pools, cc.pvpCardPools[id])
	}

	return pools
}

func (cc *cardComponent) poolUpdateCard(poolId int, cardIds []uint32) error {
	if len(cardIds) != consts.MaxHandCardAmount {
		return gamedata.InternalErr
	}

	pool, ok := cc.pvpCardPools[poolId]
	if !ok {
		return gamedata.InternalErr
	}

	pvpLevel := cc.player.GetPvpLevel()
	for _, cardID := range cardIds {
		card := cc.GetCollectCard(cardID)
		if card == nil || card.GetState() == pb.CardState_InCampaignMs {
			return gamedata.GameError(1)
		}

		cardData := card.GetCardGameData()
		if cardData == nil || (cardData.LevelLimit > 0 && cardData.LevelLimit > pvpLevel) {
			return gamedata.GameError(2)
		}
	}

	pool.setCards(cardIds)
	return nil
}

func (cc *cardComponent) NewbieInitPvpCardPool(camp int, cardIDs []uint32) {
	initCards := make(map[uint32]*pb.CardInfo)
	var gcardIDs []uint32
	for _, cardID := range cardIDs {
		initCards[cardID] = &pb.CardInfo{
			CardId: cardID,
			Amount: 1,
		}

		cardData := cc.poolGData.GetCard(cardID, 1)
		if cardData != nil {
			gcardIDs = append(gcardIDs, cardData.GCardID)
		}
	}
	cc.ModifyCollectCards(initCards)

	pool, ok := cc.pvpCampFightPool[camp]
	if !ok {
		return
	}
	pool.setCards(cardIDs)
	cc.attr.SetInt("fightCamp", camp)
	cc.OnPvpBattleEnd(gcardIDs)
}

func (cc *cardComponent) CreatePvpHandCards(camp int) []*pb.SkinGCard {
	cardPool := cc.GetPvpCardPoolByCamp(camp)
	var handCards []*pb.SkinGCard
	for _, card := range cardPool {
		handCards = append(handCards, &pb.SkinGCard{
			GCardID: card.GetCardGameData().GetGCardID(),
			Skin: card.GetSkin(),
			Equip: card.GetEquip(),
		})
	}
	return handCards
}

func (cc *cardComponent) GetFightCamp() int {
	camp := cc.attr.GetInt("fightCamp")
	if camp == 0 {
		return consts.Wei
	} else {
		return camp
	}
}

func (cc *cardComponent) poolAddCard(cardId uint32, poolId, idx int) error {
	if idx >= 5 || idx < 0 {
		return gamedata.InternalErr
	}

	var cardData types.IFightCardData
	if !mod.IsDiyCard(cardId) {
		card := cc.GetCollectCard(cardId)
		if card == nil {
			return gamedata.InternalErr
		}

		cardData2 := card.GetCardGameData()
		if cardData2 == nil || (cardData2.LevelLimit > 0 && cardData2.LevelLimit > cc.player.GetPvpLevel()) {
			return gamedata.GameError(2)
		}
		cardData = cardData2
	} else {
		cardData = cc.getDiyCard(cardId)
		if cardData == nil {
			return gamedata.InternalErr
		}
	}

	pool, ok := cc.pvpCardPools[poolId]
	if !ok {
		return gamedata.InternalErr
	}

	if cardData.GetCamp() != pool.getCamp() && cardData.GetCamp() != consts.Heroes {
		return gamedata.InternalErr
	}

	pool.updateCard(cardId, idx)
	eventhub.Publish(consts.EvCardUpdate, cc.player)
	return nil
}

func (cc *cardComponent) poolDelCard(cardId uint32) {
	for _, p := range cc.pvpCardPools {
		p.delCard(cardId)
	}
}

func (cc *cardComponent) updatePvpFightCardPool(pools []*pb.FightPool, fightCamp int) {
	for _, campFightPool := range pools {
		for poolID, pool := range cc.pvpCardPools {
			if int(campFightPool.Camp) == pool.getCamp() {
				pool.setFight(poolID == int(campFightPool.PoolId))
			}
		}
	}


	cc.attr.SetInt("fightCamp", fightCamp)
}

func (cc *cardComponent) getOneCardJadePrice(rare int) (int, error) {
	if config.GetConfig().IsXfServer() {
		switch rare {
		case 1:
			return 4, nil
		case 2:
			return 5, nil
		case 3:
			return 6, nil
		case 4:
			return 7, nil
		case 5:
			return 8, nil
		default:
			return 0, gamedata.GameError(6)
		}
	} else {

		switch rare {
		case 1:
			return 3, nil
		case 2:
			return 4, nil
		case 3:
			return 5, nil
		case 4:
			return 6, nil
		case 5:
			return 7, nil
		default:
			return 0, gamedata.GameError(6)
		}
	}
}

// 卡升级
func (cc *cardComponent) uplevelCard(cardId uint32, isConsumeJade, isNeedJade bool) (err error) {
	_card := cc.GetCollectCard(cardId)
	err = gamedata.InternalErr
	if _card == nil {
		err = gamedata.GameError(1)
		return
	}

	card := _card.(*collectCard)

	cardData := card.GetCardGameData()
	if cardData == nil {
		err = gamedata.GameError(2)
		return
	}

	curLevel := card.GetLevel()
	nextLevelCardData := cc.poolGData.GetCard(cardId, curLevel + 1)
	if nextLevelCardData == nil {
		err = gamedata.GameError(3)
		return
	}

	if cardData.ConsumeBook > 0 && card.GetMaxUnlockLevel() < nextLevelCardData.Level {
		err = gamedata.GameError(4)
		return
	}

	funcData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	rare := nextLevelCardData.Rare
	resComponent := cc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	needRes := make(map[int]int)
	var newLevel int
	var newAmount int
	if isConsumeJade {
		newLevel = 3
		newAmount := card.GetAmount()
		var gold int
		if curLevel >= newLevel {
			err = gamedata.GameError(5)
			return
		}

		for level := curLevel; level < newLevel; level++ {
			cardData := cc.poolGData.GetCard(cardId, level)
			gold += cardData.LevelupGold
			newAmount -= cardData.LevelupNum
		}
		jade := gold / 80
		
		if newAmount < 0 {
			amount := - newAmount
			newAmount = 0

			jadePrice, err := cc.getOneCardJadePrice(rare)
			if err != nil {
				return err
			}

			jade += jadePrice * amount
		}

		if !resComponent.HasResource(consts.Jade, jade) {
			err = gamedata.GameError(7)
			return
		}
		needRes[consts.Jade] = - jade

		module.Shop.LogShopBuyItem(cc.player, "upLevelCard", "宝玉升级卡", 1, "gameplay",
			strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), jade, "")

	} else {
		if cardData.LevelupGold > 0 {
			levelupGold := cardData.LevelupGold
			if !resComponent.HasResource(consts.Gold, cardData.LevelupGold) {
				if isNeedJade {
					var needJade int
					hasGood := resComponent.GetResource(consts.Gold)
					subGold :=  levelupGold - hasGood
					needJade = int(math.Ceil(float64(subGold) / float64(funcData.JadeToGold)))
					if !resComponent.HasResource(consts.Jade, needJade){
						err = gamedata.GameError(8)
						return
					}
					needRes[consts.Jade] = - needJade
					needRes[consts.Gold] = - hasGood
				}else {
					err = gamedata.GameError(8)
					return
				}
			}else {
				needRes[consts.Gold] = -levelupGold
			}
		}
		if cardData.LevelupHor > 0 {
			if !resComponent.HasResource(consts.Horse, cardData.LevelupHor) {
				err = gamedata.GameError(8)
				return
			}
			needRes[consts.Heroes] = -cardData.LevelupHor
		}
		if cardData.LevelupMat > 0 {
			if !resComponent.HasResource(consts.Mat, cardData.LevelupMat) {
				err = gamedata.GameError(8)
				return
			}
			needRes[consts.Mat] = -cardData.LevelupMat
		}
		if cardData.LevelupWeap > 0 {
			if !resComponent.HasResource(consts.Weap, cardData.LevelupWeap) {
				err = gamedata.GameError(8)
				return
			}
			needRes[consts.Weap] = -cardData.LevelupWeap
		}

		if card.GetAmount() < cardData.LevelupNum {
			err = gamedata.GameError(9)
			return
		}

		newAmount = card.GetAmount() - cardData.LevelupNum
		newLevel = card.GetLevel() + 1
	}

	resComponent.BatchModifyResource(needRes, consts.RmrUpLevelCard)
	card.SetAmount(newAmount)
	card.setLevel(newLevel)
	glog.Infof("uplevelCard, uid=%d, card=%s, needRes=%v", cc.player.GetUid(), card, needRes)

	if card.GetLevel() == 4 {
		module.Televise.SendNotice(pb.TeleviseEnum_CardLevelupPurple, cc.player.GetName(), cardId)
	}else if card.GetLevel() == 5 {
		module.Televise.SendNotice(pb.TeleviseEnum_CardLevelupOrange, cc.player.GetName(), cardId)
	}

	cc.player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
		Cards: []*pb.CardInfo{card.PackMsg()},
	})

	log := getLog(cc.player.GetLogAccountType(), cc.player.GetArea())
	if log != nil {
		log.modifyCardLevel(cardId, curLevel, newLevel)
	}
	err = nil
	eventhub.Publish(consts.EvCardUpdate, cc.player)
	return
}

// 卡复活
func (cc *cardComponent) cardRelive(cardId uint32) (info *pb.CardInfo, err error) {
	_card := cc.GetCollectCard(cardId)
	err = gamedata.InternalErr
	if _card == nil {
		return
	}
	card := _card.(*collectCard)

	if !card.IsDead() {
		return
	}

	cardData := card.GetCardGameData()
	if cardData == nil {
		return
	}

	resComponent := cc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resComponent.HasResource(consts.Med, 1) {
		return
	}

	resComponent.ModifyResource(consts.Med, -1)
	if card.GetAmount() <= 0 {
		card.SetAmount(1)
	}
	if card.GetLevel() <= 0 {
		card.setLevel(1)
	}
	card.setEnergy(cardData.Energy)

	info = &pb.CardInfo{
		CardId: cardId,
		Amount: int32(card.GetAmount()),
		Level:  int32(card.GetLevel()),
		Energy: card.GetEnergy(),
	}
	cc.player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
		Cards: []*pb.CardInfo{info},
	})

	err = nil
	return
}

// 卡治疗
func (cc *cardComponent) cardTreat(cardId uint32) (info *pb.CardInfo, err error) {
	_card := cc.GetCollectCard(cardId)
	err = gamedata.InternalErr
	if _card == nil {
		return
	}
	card := _card.(*collectCard)

	if card.IsDead() {
		return
	}

	cardData := card.GetCardGameData()
	if cardData == nil {
		return
	}

	resComponent := cc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resComponent.HasResource(consts.Ban, 1) {
		return
	}

	resComponent.ModifyResource(consts.Ban, -1)
	energy := card.GetEnergy() + 1
	if energy > cardData.Energy {
		energy = cardData.Energy
	}
	card.setEnergy(energy)

	info = &pb.CardInfo{
		CardId: cardId,
		Amount: int32(card.GetAmount()),
		Level:  int32(card.GetLevel()),
		Energy: card.GetEnergy(),
	}
	cc.player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
		Cards: []*pb.CardInfo{info},
	})
	err = nil
	return
}

func (cc *cardComponent) addNewCollectCard(cardData *gamedata.Card, amount int) *collectCard {
	cardID := cardData.CardID
	var card *collectCard
	if oldCc, ok := cc.onceCards[cardID]; ok {
		card = newCollectCardByOldData(oldCc, cardData, amount)
		delete(cc.onceCards, cardID)
	} else {
		card = newCollectCard(cardData, amount)
		cc.attr.GetMapAttr("collectCards").SetMapAttr(strconv.Itoa(int(card.GetCardID())), card.getAttr())
	}

	cc.collectCardMap[cardID] = card
	cc.collectCards = append(cc.collectCards, card)

	log := getLog(cc.player.GetLogAccountType(), cc.player.GetArea())
	if log != nil {
		log.modifyCardAmount(cardID, amount)
		log.modifyCardLevel(cardID, 0, cardData.Level)
	}
	return card
}

func (cc *cardComponent) ModifyCollectCards(cardsChange map[uint32]*pb.CardInfo) []*pb.ChangeCardInfo {

	var changeCards []*pb.ChangeCardInfo
	var newCards []*pb.CardInfo
	//campaignComponent := cc.player.GetComponent(consts.CampaignCpt).(types.ICampaignComponent)
	log := getLog(cc.player.GetLogAccountType(), cc.player.GetArea())

	for cardId, cardInfo := range cardsChange {

		cardData := cc.poolGData.GetCard(cardId, 1)
		if cardData == nil {
			continue
		}

		card, ok := cc.collectCardMap[cardId]
		change := &pb.ChangeCardInfo{}

		if ok {

			if card.IsDead() {
				// 死亡
				continue
			}
			cardData := cc.poolGData.GetCard(cardId, card.GetLevel())
			if cardData == nil {
				continue
			}
			change.Old = card.PackMsg()

			modifyAmount := int(cardInfo.Amount)
			oldAmount := card.GetAmount()
			card.SetAmount(oldAmount + modifyAmount)

			oldLevel := card.GetLevel()
			level := oldLevel + int(cardInfo.Level)
			if log != nil {
				log.modifyCardLevel(cardId, oldLevel, level)
			}

			if level <= 0 {
				// 删掉
				card.del(cc.player)
				delete(cc.collectCardMap, cardId)
				//cc.attr.GetMapAttr("collectCards").Del(strconv.Itoa(int(cardId)))
				cc.onceCards[cardId] = card
				card = nil

				if log != nil {
					log.modifyCardAmount(cardId, - oldAmount)
				}
			} else {
				card.setLevel(level)
				if log != nil {
					log.modifyCardAmount(cardId, modifyAmount)
				}
			}

		} else if cardInfo.Amount > 0 || cardInfo.Level > 0 {

			if cardInfo.Level > 0 {
				cardData = cc.poolGData.GetCard(cardId, int(cardInfo.Level))
				if cardData == nil {
					continue
				}
			}

			card = cc.addNewCollectCard(cardData, int(cardInfo.Amount))

		} else {
			continue
		}

		if card != nil {
			energy := card.GetEnergy() + cardInfo.Energy
			card.setEnergy(energy)

			if energy <= 0 {
				card.setEnergy(0)
				cc.poolDelCard(card.GetCardID())
			} else if energy > cardData.Energy {
				card.setEnergy(cardData.Energy)
			}

			change.New = card.PackMsg()
			newCards = append(newCards, change.New)
		} else {
			newCards = append(newCards, &pb.CardInfo{
				CardId: cardId,
			})
		}

		changeCards = append(changeCards, change)
	}

	cc.player.GetAgent().PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
		Cards: newCards,
	})

	logInfo, _ := json.Marshal(changeCards)
	glog.Infof("player %d ModifyCollectCards %s", cc.player.GetUid(), logInfo)

	return changeCards
}

func (cc *cardComponent) genDiyCardNum(singleMin, singleMax, avMin, avMax int) (min int, max int) {
	minValue := int(math.Max(float64(singleMin), float64(avMin)))
	maxValue := int(math.Min(float64(singleMax), float64(avMax))) - 1
	if minValue > maxValue {
		minValue = maxValue
	}
	if minValue <= 0 {
		minValue = 0
	}
	if maxValue <= 0 {
		maxValue = 0
	}

	min = rand.Intn(maxValue-minValue+1) + minValue
	max = min + 1
	return
}

func (cc *cardComponent) makeDiyCard(name string, diySkillId1, diySkillId2 int, weapon, img string) (*diyCard, error) {
	if diySkillId1 == diySkillId2 {
		return nil, gamedata.InternalErr
	}
	diyData1 := cc.diyGData.GetDiyData(diySkillId1)
	diyData2 := cc.diyGData.GetDiyData(diySkillId2)
	if diyData1 == nil || diyData2 == nil {
		return nil, gamedata.InternalErr
	}

	pvpComponent := cc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	pvpLevel := pvpComponent.GetPvpLevel()
	if pvpLevel < diyData1.Level || pvpLevel < diyData2.Level {
		return nil, gamedata.InternalErr
	}

	for _, banId := range diyData1.Ban {
		if banId == diySkillId2 {
			return nil, gamedata.InternalErr
		}
	}
	for _, banId := range diyData2.Ban {
		if banId == diySkillId1 {
			return nil, gamedata.InternalErr
		}
	}
	if diyData1.MinSingle > diyData2.MaxSingle || diyData1.MinTotal > diyData2.MaxTotal ||
		diyData2.MinSingle > diyData1.MaxSingle || diyData2.MinTotal > diyData1.MaxTotal {
		return nil, gamedata.InternalErr
	}

	needWine := diyData1.Wine + diyData2.Wine
	needBook := diyData1.Book + diyData2.Book
	resComponent := cc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resComponent.HasResource(consts.Wine, needWine) || !resComponent.HasResource(consts.Book, needBook) {
		return nil, gamedata.InternalErr
	}

	cardId, err := common.Gen32UUid("diycard")
	if err != nil {
		glog.Errorf("diycard GenUuid err %s", err)
		return nil, err
	}

	resComponent.BatchModifyResource(map[int]int{
		consts.Wine: -needWine,
		consts.Book: -needBook,
	})

	minSingle := int(math.Max(float64(diyData1.MinSingle), float64(diyData2.MinSingle)))
	maxSingle := int(math.Min(float64(diyData1.MaxSingle), float64(diyData2.MaxSingle)))
	if minSingle > maxSingle {
		minSingle = maxSingle
	}
	minTotal := int(math.Max(float64(diyData1.MinTotal), float64(diyData2.MinTotal)))
	maxTotal := int(math.Min(float64(diyData1.MaxTotal), float64(diyData2.MaxTotal)))
	if minTotal > maxTotal {
		minTotal = maxTotal
	}

	attr := attribute.NewMapAttr()
	card := newDiyCardByAttr(attr)
	card.setCardID(cardId)
	card.setDiySkillId(diySkillId1, diySkillId2)
	card.setResource(needWine, needBook)
	card.setName(name)
	card.setWeapon(weapon)

	dirs := []int{consts.UP, consts.DOWN, consts.LEFT, consts.RIGHT}
	utils.ShuffleInt(dirs)
	for i, dir := range dirs {
		avMin := minTotal/(len(dirs)-i) + (minTotal % (len(dirs) - i))
		avMax := maxTotal/(len(dirs)-i) + (maxTotal % (len(dirs) - i))
		min, max := cc.genDiyCardNum(minSingle, maxSingle, avMin, avMax)
		minTotal -= min
		maxTotal -= max
		switch dir {
		case consts.UP:
			card.setUpNum(min, max)
			break
		case consts.DOWN:
			card.setDownNum(min, max)
			break
		case consts.LEFT:
			card.setLeftNum(min, max)
			break
		case consts.RIGHT:
			card.setRightNum(min, max)
			break
		default:
			break
		}
	}

	cc.diycardMap[card.GetCardID()] = card
	cc.attr.GetMapAttr("diycard").SetMapAttr(strconv.Itoa(int(cardId)), attr)
	pa := cc.player.GetAgent()
	if pa != nil {
		pa.PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
			DiyCards: []*pb.DiyCardData{
				card.packDataMsg(),
			},
		})
	}
	return card, nil
}

// 分解diy卡
func (cc *cardComponent) decomposeDiyCard(card *diyCard) {
	cardID := card.GetCardID()
	cc.attr.GetMapAttr("diycard").Del(strconv.Itoa(int(cardID)))
	delete(cc.diycardMap, cardID)

	wine, book := card.getResource()
	resComponent := cc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	resComponent.BatchModifyResource(map[int]int{
		consts.Wine: wine / 2,
		consts.Book: book / 2,
	})

	pa := cc.player.GetAgent()
	if pa != nil {
		pa.PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
			DiyCards: []*pb.DiyCardData{
				&pb.DiyCardData{
					CardId: card.GetCardID(),
				},
			},
		})
	}
}

// 重制diy卡
func (cc *cardComponent) remakeDiyCard(oldCard *diyCard) (*diyCard, error) {
	cc.decomposeDiyCard(oldCard)
	// TODO img
	return cc.makeDiyCard(oldCard.GetName(), oldCard.getDiySkillId1(), oldCard.getDiySkillId2(), oldCard.getWeapon(), "")
}

func (cc *cardComponent) OnPvpBattleEnd(fightCards []uint32) {
	fightCardsAttr := attribute.NewListAttr()
	for _, gcardID := range fightCards {
		c := cc.poolGData.GetCardByGid(gcardID)
		if c == nil {
			continue
		}
		fightCardsAttr.AppendUInt32(c.GetGCardID())
		cl := cc.GetCollectCard(c.GetCardID())
		if cl != nil {
			cl.(*collectCard).use()
		}
	}

	cc.attr.SetListAttr("lastFightCards", fightCardsAttr)
	sort.Sort(cc.collectCards)
}

// []gcardID
func (cc *cardComponent) GetLastFightCards() []*pb.SkinGCard {
	fightCardsAttr := cc.attr.GetListAttr("lastFightCards")
	var cards []*pb.SkinGCard
	if fightCardsAttr != nil {
		fightCardsAttr.ForEachIndex(func(index int) bool {
			gcardID := fightCardsAttr.GetUInt32(index)
			c := cc.getCollectCardByGCardID(gcardID)
			if c != nil {
				cards = append(cards, &pb.SkinGCard{
					GCardID: gcardID,
					Skin: c.GetSkin(),
					Equip: c.GetEquip(),
				})
			}
			return true
		})
	}
	return cards
}

func (cc *cardComponent) GetFavoriteCards() []*pb.SkinGCard {
	n := 3
	if cc.collectCards.Len() < 3 {
		n = cc.collectCards.Len()
	}

	var cards []*pb.SkinGCard
	for i := 0; i < n; i++ {
		c := cc.collectCards[i]
		cardData := c.GetCardGameData()
		if cardData == nil {
			continue
		}
		cards = append(cards, &pb.SkinGCard{
			GCardID: cardData.GetGCardID(),
			Skin: c.GetSkin(),
			Equip: c.GetEquip(),
		})
	}
	return cards
}

func (cc *cardComponent) updateCardSkin(cardID uint32, skin string) error {
	c := cc.GetCollectCard(cardID)
	if c == nil {
		return gamedata.GameError(1)
	}
	if skin != "" && !module.Bag.HasItem(cc.player, consts.ItCardSkin, skin) {
		return gamedata.GameError(2)
	}
	c.(*collectCard).setSkin(skin)
	return nil
}

func (cc *cardComponent) getOnceCards() []types.ICollectCard {
	var cards []types.ICollectCard
	if cc.onceCards == nil {
		return cards
	}

	for _, card := range cc.onceCards {
		cards = append(cards, card)
	}
	return cards
}

func (cc *cardComponent) onEquipDel(equipID string) {
	for _, card := range cc.collectCards {
		if card.GetEquip() == equipID {
			card.DeEquip()
			agent := cc.player.GetAgent()
			if agent == nil {
				return
			}

			agent.PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
				Cards: []*pb.CardInfo{card.PackMsg()},
			})
			return
		}
	}
}

func (cc *cardComponent) GetPvpCampFightPool(camp int){

}