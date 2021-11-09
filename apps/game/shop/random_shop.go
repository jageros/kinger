package shop

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type iRandomShop interface {
	packMsg() *pb.VisitRandomShopData
	pushToClient()
	onCrossDay()
	onLogin()
	buyRefreshCnt(buyCnt int) (*pb.BuyRandomShopRefreshCntReply, error)
	buy(arg *pb.BuyRandomShopArg) (interface{}, error)
}

type nilRandomShopSt struct {
}

func (nrs *nilRandomShopSt) pushToClient() {}
func (nrs *nilRandomShopSt) onCrossDay()   {}
func (nrs *nilRandomShopSt) onLogin()      {}

func (nrs *nilRandomShopSt) packMsg() *pb.VisitRandomShopData {
	return nil
}

func (nrs *nilRandomShopSt) buyRefreshCnt(buyCnt int) (*pb.BuyRandomShopRefreshCntReply, error) {
	return nil, gamedata.InternalErr
}

func (nrs *nilRandomShopSt) buy(arg *pb.BuyRandomShopArg) (interface{}, error) {
	return nil, gamedata.InternalErr
}

type randomShop struct {
	player       types.IPlayer
	attr         *attribute.MapAttr
	cptAttr      *attribute.MapAttr
	goodsList    []*randShopGoods
	newGoodsList []*randShopGoods
}

type randShopGoods struct {
	attr *attribute.MapAttr
}

type card struct {
	Key   uint32
	Value float32
}

func (rsg randShopGoods) getGoodsID() string {
	return rsg.attr.GetStr("goodsID")
}

func (rsg randShopGoods) getAmount() int {
	return rsg.attr.GetInt("amount")
}

func (rsg randShopGoods) getIsBuy() bool {
	return rsg.attr.GetBool("isBuy")
}

func (rsg randShopGoods) getIsNeedGold() bool {
	return rsg.attr.GetBool("isNeedGold")
}

func (rsg randShopGoods) setGoodsID(value string) {
	rsg.attr.SetStr("goodsID", value)
}

func (rsg randShopGoods) setAmount(value int) {
	rsg.attr.SetInt("amount", value)
}

func (rsg randShopGoods) setIsBuy(value bool) {
	rsg.attr.SetBool("isBuy", value)
}

func (rsg randShopGoods) setIsNeedGold(value bool) {
	rsg.attr.SetBool("isNeedGold", value)
}

func (rsg randShopGoods) setData(goodsID string, amount int, isBuy, isNeedGold bool) {
	rsg.setGoodsID(goodsID)
	rsg.setAmount(amount)
	rsg.setIsBuy(isBuy)
	rsg.setIsNeedGold(isNeedGold)
}

var nilRandomShop = &nilRandomShopSt{}

func newRandomShop(attr *attribute.MapAttr, player types.IPlayer) iRandomShop {
	var goodsList []*randShopGoods
	randomShopAttr := attr.GetMapAttr("randomShop")
	if randomShopAttr == nil {
		randomShopAttr = attribute.NewMapAttr()
		randomShopAttr.SetInt("buyCnt", 0)
		randomShopAttr.SetInt("refreshTime", 0)
		attr.SetMapAttr("randomShop", randomShopAttr)
		goodsListAttr := attribute.NewListAttr()
		randomShopAttr.SetListAttr("goodsList", goodsListAttr)
	} else {
		goodsListAttr := randomShopAttr.GetListAttr("goodsList")
		goodsListAttr.ForEachIndex(func(index int) bool {
			goodsAttr := goodsListAttr.GetMapAttr(index)
			goodsList = append(goodsList, &randShopGoods{goodsAttr})
			return true
		})
	}

	r := &randomShop{
		player:    player,
		attr:      attr,
		cptAttr:   randomShopAttr,
		goodsList: goodsList,
	}
	return r
}

func (rs *randomShop) packMsg() *pb.VisitRandomShopData {
	refreshTime := rs.getRefreshTime()
	ver := rs.getVer()
	ok := false
	if refreshTime != 0 {
		unix8 := int64(8 * 60 * 60)
		timeNow := time.Now().Unix()
		subTime := timeNow - int64(refreshTime)
		if subTime > unix8 {
			ok = true
			rs.setRefreshTime(0)
		}
	}

	if len(rs.goodsList) > 0 && !ok && ver != 0 {
		msg := &pb.VisitRandomShopData{}
		randomTeam := gamedata.GetGameData(consts.RandomShop).(*gamedata.RandomShopGameData)
		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		for _, goods := range rs.goodsList {
			goodsID := goods.getGoodsID()
			shopName := strings.Split(goodsID, ":")
			goodsName := shopName[0]
			var needJade, amount, needGold int32
			if goodsName == "card" {
				n, _ := strconv.Atoi(shopName[1])
				card := poolGameData.GetCard(uint32(n), 1)
				if goods.getIsNeedGold() {
					onePrice := randomTeam.CardStarPriceGold[card.Rare][1]
					amount = int32(goods.getAmount())
					needGold = int32(onePrice) * amount
				} else {
					onePrice := randomTeam.CardStarPrice[card.Rare][1]
					amount = int32(goods.getAmount())
					needJade = int32(onePrice) * amount
				}
			} else if goodsName == "gold" {
				onePrice := randomTeam.GoldPrice[0] / randomTeam.GoldPrice[1]
				amount = int32(goods.getAmount())
				needJade = amount / int32(onePrice)
			} else if goodsName == "freeGold" {
				amount = int32(goods.getAmount())
			} else if goodsName == "freeJade" {
				amount = int32(goods.getAmount())
			}

			msg.RandomShops = append(msg.RandomShops, rs.getPackMsg(goods.getGoodsID(), amount, needJade, needGold, goods.getIsBuy()))
		}
		bugCnt := rs.getBuyCnt()
		msg.BuyTimes = int32(bugCnt)
		msg.Discount = int32(rs.getDiscount() * 100)
		msg.MaxBuyTimes = int32(randomTeam.RefreshCnt)
		msg.NeedJade = int32(randomTeam.RefreshPrice)
		msg.NexRemainTime = int32(rs.getNextRemainTime().Seconds())

		return msg
	} else {
		//isReset := rs.getIsReset()
		//if ok{
		//	isReset = true
		//}
		msg := rs.newPackMsg(true)
		rs.setAfterTime()
		return msg
	}
}

func (rs *randomShop) getRefreshTime() int {
	return rs.cptAttr.GetInt("refreshTime")
}

func (rs *randomShop) setRefreshTime(value int) {
	rs.cptAttr.SetInt("refreshTime", value)
}

func (rs *randomShop) getBuyCnt() int {
	return rs.cptAttr.GetInt("buyCnt")
}

func (rs *randomShop) setBuyCnt(cnt int) {
	rs.cptAttr.SetInt("buyCnt", cnt)
}

func (rs *randomShop) setVer(ver int) {
	rs.cptAttr.SetInt("ver", ver)
}

func (rs *randomShop) getVer() int {
	return rs.cptAttr.GetInt("ver")
}

func (rs *randomShop) getPackMsg(goodsID string, amount, needJade, needGold int32, isBuy bool) *pb.RandomShop {
	return &pb.RandomShop{
		GoodsId:  goodsID,
		Amount:   amount,
		NeedJade: needJade,
		NeedGold: needGold,
		IsBuy:    isBuy,
	}
}

func (rs *randomShop) setGoodsList() {
	var newList []*randShopGoods
	var i int
	for i = 0; i < len(rs.newGoodsList); i++ {
		newMap := rs.newGoodsList[i]
		if len(rs.goodsList) > i && len(rs.goodsList) > 0 {
			rs.goodsList[i].attr.SetStr("goodsID", newMap.getGoodsID())
			rs.goodsList[i].attr.SetInt("amount", newMap.getAmount())
			rs.goodsList[i].attr.SetBool("isBuy", newMap.getIsBuy())
			rs.goodsList[i].attr.SetBool("isNeedGold", newMap.getIsNeedGold())
		} else {
			newGoods := &randShopGoods{
				attr: attribute.NewMapAttr(),
			}
			newGoods.setData(newMap.getGoodsID(), newMap.getAmount(), newMap.getIsBuy(), newMap.getIsNeedGold())
			rs.cptAttr.GetListAttr("goodsList").AppendMapAttr(newGoods.attr)
			newList = append(newList, newGoods)
		}
	}

	subGoodsList := len(rs.goodsList) - len(rs.newGoodsList)
	if subGoodsList > 0 {
		for _, goods := range rs.goodsList[i:] {
			rs.cptAttr.GetListAttr("goodsList").DelMapAttr(goods.attr)
		}
		rs.goodsList = rs.goodsList[:subGoodsList+1]
	}

	if len(newList) > 0 {
		rs.goodsList = append(rs.goodsList, newList[:]...)
	}
}

func (rs *randomShop) setAfterTime() {
	subTime8, pTime8 := timer.TimePreDelta(8, 0, 0)
	subTime16, pTime16 := timer.TimePreDelta(16, 0, 0)
	subTime24, pTime24 := timer.TimePreDelta(24, 0, 0)

	if subTime8 < subTime16 {
		rs.setRefreshTime(int(pTime8))
	} else if subTime16 < subTime24 {
		rs.setRefreshTime(int(pTime16))
	} else {
		rs.setRefreshTime(int(pTime24))
	}
}

func (rs *randomShop) getNextRemainTime() time.Duration {
	nTime8 := timer.TimeDelta(8, 0, 0)
	nTime16 := timer.TimeDelta(16, 0, 0)
	nTime24 := timer.TimeDelta(24, 0, 0)

	if nTime8 < nTime16 && nTime8 < nTime24 {
		return nTime8
	} else if nTime16 < nTime24 {
		return nTime16
	} else {
		return nTime24
	}
}

func (rs *randomShop) getIsReset() (isReset bool) {
	buyCnt := rs.getBuyCnt()
	if buyCnt > 3 {
		isReset = false
	} else {
		isReset = true
	}
	return
}

func (rs *randomShop) buyRefreshCnt(_ int) (*pb.BuyRandomShopRefreshCntReply, error) {
	buyCnt := rs.getBuyCnt()
	randomTeam := gamedata.GetGameData(consts.RandomShop).(*gamedata.RandomShopGameData)
	if buyCnt >= randomTeam.RefreshCnt {
		return nil, gamedata.GameError(1)
	}

	discount := rs.getDiscount()

	var price int
	price = int(float64(randomTeam.RefreshPrice) * discount)

	if price > 0 {
		resCpt := rs.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, price) {
			return nil, gamedata.GameError(2)
		}
		resCpt.ModifyResource(consts.Jade, -price, consts.RmrRefreshRandomShop)
	}

	buyCnt += 1
	rs.setBuyCnt(buyCnt)
	//isReset := rs.getIsReset()
	msg := &pb.BuyRandomShopRefreshCntReply{
		RandomShopData: rs.newPackMsg(false),
	}

	msg.RandomShopData.NexRemainTime = int32(rs.getNextRemainTime().Seconds())
	rs.caclHint(false)

	mod.LogShopBuyItem(rs.player, "ReFreshCnt", "探访刷新", 1, "refresh_shop",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), price,
		fmt.Sprintf("discount=%f, buyCnt=%d", discount, buyCnt))

	return msg, nil
}

func (rs *randomShop) buy(arg *pb.BuyRandomShopArg) (interface{}, error) {
	goodsID := arg.GoodsID
	goodsAmount := arg.Amount
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	randomTeam := gamedata.GetGameData(consts.RandomShop).(*gamedata.RandomShopGameData)
	shopName := strings.Split(goodsID, ":")

	var needJade, needGold int
	curGoodsAttr := &randShopGoods{
		attr: attribute.NewMapAttr(),
	}
	n, _ := strconv.Atoi(shopName[1])

	hasRes := false
	for _, goods := range rs.goodsList {
		amount := goods.getAmount()
		if goods.getGoodsID() == goodsID && goods.getIsBuy() && goodsAmount == int32(amount) {
			if strings.HasPrefix(goodsID, "card") {
				card := poolGameData.GetCard(uint32(n), 1)
				if goods.getIsNeedGold() {
					oneCardPrice := randomTeam.CardStarPriceGold[card.Rare][1]
					needGold = oneCardPrice * int(amount)
				} else {
					oneCardPrice := randomTeam.CardStarPrice[card.Rare][1]
					needJade = oneCardPrice * int(amount)
				}

			} else if strings.HasPrefix(goodsID, "gold") {
				oneGoldPrice := randomTeam.GoldPrice[0] / randomTeam.GoldPrice[1]
				needJade = int(amount) / oneGoldPrice

			} else if strings.HasPrefix(goodsID, "freeGold") || strings.HasPrefix(goodsID, "freeJade") {
				needJade = 0
			} else {
				return nil, gamedata.GameError(1)
			}
			curGoodsAttr = goods
			hasRes = true
		}
	}

	if !hasRes {
		return nil, gamedata.GameError(1)
	}

	resCpt := rs.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)

	resType := consts.Jade
	needRes := needJade

	if needJade > 0 {
		if !resCpt.HasResource(consts.Jade, needJade) {
			return nil, gamedata.GameError(2)
		}
		resCpt.ModifyResource(consts.Jade, -needJade, consts.RmrRandomShop)
	}

	if needGold > 0 {
		if !resCpt.HasResource(consts.Gold, needGold) {
			return nil, gamedata.GameError(2)
		}
		resCpt.ModifyResource(consts.Gold, -needGold, consts.RmrRandomShop)
		resType = consts.Gold
		needRes = needGold
	}

	if strings.HasPrefix(goodsID, "card") {
		cardCpt := rs.player.GetComponent(consts.CardCpt).(types.ICardComponent)
		cardMap := make(map[uint32]*pb.CardInfo)
		cardMap[uint32(n)] = &pb.CardInfo{
			Amount: goodsAmount,
		}
		cardCpt.ModifyCollectCards(cardMap)
	} else if strings.HasPrefix(goodsID, "gold") || strings.HasPrefix(goodsID, "freeGold") {
		resCpt.ModifyResource(consts.Gold, int(goodsAmount), consts.RmrRandomShop)
	} else if strings.HasPrefix(goodsID, "freeJade") {
		resCpt.ModifyResource(consts.Jade, int(goodsAmount), consts.RmrRandomShop)
	} else {
		return nil, gamedata.GameError(1)
	}

	curGoodsAttr.setIsBuy(false)
	rs.caclHint(false)

	mod.LogShopBuyItem(rs.player, goodsID, fmt.Sprintf("%s_%s", shopName[0], shopName[1]), 1,
		"refresh_shop", strconv.Itoa(resType), module.Player.GetResourceName(resType), needRes,
		fmt.Sprintf("goodsAmount=%d", goodsAmount))

	return nil, nil
}

func (rs *randomShop) onCrossDay() {
	rs.setBuyCnt(0)
	rs.syncToClient()
	rs.setAfterTime()
	rs.caclHint(false)
}

func (rs *randomShop) onLogin() {
	rs.caclHint(true)
}

func (rs *randomShop) caclHint(isLogin bool) {
	if rs.hasFreeRes() {
		if isLogin {
			rs.player.AddHint(pb.HintType_HtRandomFree, 1)
		} else {
			rs.player.UpdateHint(pb.HintType_HtRandomFree, 1)
		}
	} else {
		rs.player.DelHint(pb.HintType_HtRandomFree)
	}
}

func (rs *randomShop) hasFreeRes() (hasRes bool) {
	for _, goods := range rs.goodsList {
		if strings.HasPrefix(goods.getGoodsID(), "free") && goods.getIsBuy() {
			hasRes = true
		}
	}
	return
}

func (rs *randomShop) pushToClient() {
	rs.syncToClient()
	rs.setAfterTime()
	rs.caclHint(false)
}

func (rs *randomShop) syncToClient() {
	agent := rs.player.GetAgent()
	if agent == nil {
		return
	}
	msg := rs.newPackMsg(true)
	if msg == nil {
		return
	}

	if config.GetConfig().IsXfServer() {
		rs.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_RandomShop)
	} else {
		agent.PushClient(pb.MessageID_S2C_UPDATE_RANDOM_SHOP, msg)
	}
}

func (rs *randomShop) newPackMsg(isReset bool) *pb.VisitRandomShopData {
	msg := &pb.VisitRandomShopData{}
	rs.newGoodsList = rs.newGoodsList[:0:0]

	cardRankUp := make(map[uint32]float32)
	cardRankDown := make(map[uint32]float32)
	cardIDsMap := make(map[uint32]int)
	allCardIDMap := make(map[uint32]int)

	randomTeam := gamedata.GetGameData(consts.RandomShop).(*gamedata.RandomShopGameData)
	allCards := module.Card.GetAllCollectCards(rs.player)
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	cardIDs := module.Card.GetUnlockCards(rs.player, 0)

	for _, cm := range cardIDs {
		cardIDsMap[cm] = 1
	}

	oldID2Type := make(map[uint32]string)
	for _, goods := range rs.goodsList {
		goodsName := strings.Split(goods.getGoodsID(), ":")
		if len(goodsName) < 2 {
			continue
		}
		n, _ := strconv.Atoi(goodsName[1])
		oldID2Type[uint32(n)] = goodsName[0]
	}

	for _, card := range allCards {
		allCardIDMap[card.GetCardID()] = 1

		if card.IsMaxCanUpLevel() {
			continue
		}

		cardData := card.GetCardGameData()
		if cardData.IsSpCard() {
			continue
		}

		if n, ok := oldID2Type[cardData.GetCardID()]; ok && n == "card" {
			continue
		}

		if _, ok := cardIDsMap[cardData.GetCardID()]; ok {
			delete(cardIDsMap, cardData.GetCardID())
		}

		levelupAmount := cardData.LevelupNum
		curAmount := card.GetAmount()
		rateLevelup := float32(curAmount) / float32(levelupAmount)

		if rateLevelup < float32(1) {
			cardRankUp[card.GetCardID()] = rateLevelup
		}

		cardAmount := 0
		for lv := 1; lv < card.GetLevel(); lv++ {
			cardData = poolGameData.GetCard(card.GetCardID(), lv)
			cardAmount += cardData.LevelupNum
		}
		cardAmount += card.GetAmount()

		allAmount := 0
		for lv := 1; lv < 5; lv++ {
			cardData = poolGameData.GetCard(card.GetCardID(), lv)
			allAmount += cardData.LevelupNum
		}

		if cardAmount < allAmount {
			cardRankDown[card.GetCardID()] = float32(cardAmount)
		}
	}

	for cardID, _ := range cardIDsMap {
		cardData := poolGameData.GetCard(cardID, 1)
		if cardData.IsSpCard() {
			continue
		}

		if n, ok := oldID2Type[cardID]; ok && n == "card" {
			continue
		}

		if _, ok := allCardIDMap[cardID]; ok {
			continue
		}

		if _, ok := cardRankDown[cardID]; !ok {
			cardRankDown[cardID] = 0
		}
		if _, ok := cardRankUp[cardID]; !ok {
			cardRankUp[cardID] = 0
		}
	}

	team := rs.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpTeam()

	goodsNum := randomTeam.GoodsNum
	oneGoldPrice := randomTeam.GoldPrice[0] / randomTeam.GoldPrice[1]

	if isReset {
		randName := rs.randWeights(randomTeam.FreeData, randomTeam.FreeRandomPara)
		goods := &randShopGoods{
			attr: attribute.NewMapAttr(),
		}
		var goodsID string
		var goodsAmount int
		if strings.HasSuffix(randName, "gold") {
			teamRes_0 := randomTeam.TeamFreeGold[team][0]
			teamRes_1 := randomTeam.TeamFreeGold[team][1]
			goodsAmount = rs.randInt(teamRes_0, teamRes_1+1)
			goodsID = fmt.Sprintf("freeGold:%d", goodsAmount)

		} else if strings.HasSuffix(randName, "jade") {
			teamRes_0 := randomTeam.TeamFreeJade[team][0]
			teamRes_1 := randomTeam.TeamFreeJade[team][1]
			goodsAmount = rs.randInt(teamRes_0, teamRes_1+1)
			goodsID = fmt.Sprintf("freeJade:%d", goodsAmount)

		}
		goods.setData(goodsID, int(goodsAmount), true, false)

		rs.newGoodsList = append(rs.newGoodsList, goods)
		msg.RandomShops = append(msg.RandomShops, rs.getPackMsg(goodsID, int32(goodsAmount), 0, 0, true))

		goodsNum -= 1

	} else if len(cardRankUp) > 0 {
		randList := rs.rankCard(cardRankUp, true)

		var num int
		if len(cardRankUp) >= randomTeam.CardLevelupNum {
			num = randomTeam.CardLevelupNum - 1
		} else {
			num = len(cardRankUp)
		}
		randNum := rand.Intn(num)
		chooseCardUp := randList[randNum]
		msg.RandomShops = append(msg.RandomShops, rs.getRandCard(chooseCardUp.Key, team))
		goodsNum -= 1
	}

	if len(cardRankDown) > 0 {
		randList := rs.rankCard(cardRankDown, false)

		var num int
		if len(cardRankDown) >= randomTeam.CardLessNum {
			num = randomTeam.CardLessNum - 1
		} else {
			num = len(cardRankDown)
		}
		randNum := rand.Intn(num)
		chooseCardDowm := randList[randNum]
		msg.RandomShops = append(msg.RandomShops, rs.getRandCard(chooseCardDowm.Key, team))
		goodsNum -= 1
	}

	for n := 0; n < goodsNum; n++ {
		randName := rs.randWeights(randomTeam.RandomData, randomTeam.RandomParaValue)
		if len(cardIDs) <= 0 {
			return nil
		}

		if strings.HasSuffix(randName, "card") {
			Index := rand.Intn(len(cardIDs) - 1)
			cardId := cardIDs[Index]
			msg.RandomShops = append(msg.RandomShops, rs.getRandCard(cardId, team))
		} else if strings.HasSuffix(randName, "gold") {
			teamRes_0 := randomTeam.TeamGold[team][0]
			teamRes_1 := randomTeam.TeamGold[team][1]
			goldNum := rs.getGoldNum(teamRes_0, teamRes_1, team, randomTeam.GoldPrice[0])
			goldPrice := goldNum / oneGoldPrice

			goods := &randShopGoods{
				attr: attribute.NewMapAttr(),
			}
			goods.setData(fmt.Sprintf("gold:%d", goldNum), int(goldNum), true, false)

			rs.newGoodsList = append(rs.newGoodsList, goods)
			msg.RandomShops = append(msg.RandomShops,
				rs.getPackMsg(fmt.Sprintf("gold:%d", goldNum), int32(goldNum), int32(goldPrice), 0, true))
		}
	}

	rs.setGoodsList()
	rs.setVer(1)

	bugCnt := rs.getBuyCnt()
	msg.BuyTimes = int32(bugCnt)
	msg.Discount = int32(rs.getDiscount() * 100)
	msg.MaxBuyTimes = int32(randomTeam.RefreshCnt)
	msg.NeedJade = int32(randomTeam.RefreshPrice)
	msg.NexRemainTime = int32(rs.getNextRemainTime().Seconds())

	return msg
}

func (rs *randomShop) getGoldNum(teamRes_0, teamRes_1, team, goldSpace int) int {
	var goldList []int
	for j := teamRes_0; j <= teamRes_1; j += goldSpace {
		goldList = append(goldList, j)
	}
	goldNum := goldList[rand.Intn(len(goldList))]
	return goldNum
}

func (rs *randomShop) getDiscount() float64 {
	buyCnt := rs.getBuyCnt()
	if buyCnt <= 0 {
		return 0
	}
	return 1
}

func (rs *randomShop) getRandCard(cardId uint32, team int) *pb.RandomShop {
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	randomTeam := gamedata.GetGameData(consts.RandomShop).(*gamedata.RandomShopGameData)
	goods := &randShopGoods{
		attr: attribute.NewMapAttr(),
	}

	card := poolGameData.GetCard(cardId, 1)
	cardNum := int32(rs.randInt(randomTeam.TeamCard[team][0], randomTeam.TeamCard[team][1]+1))

	var isNeedGold bool
	var cardPrice, cardPriceGold int32
	randName := rs.randWeights(randomTeam.CardPro, randomTeam.CardProValue)
	if strings.HasPrefix(randName, "card_jade") {
		oneCardPrice := randomTeam.CardStarPrice[card.Rare][1]
		cardPrice = cardNum * int32(oneCardPrice)
		isNeedGold = false
	} else {
		oneCardPrice := randomTeam.CardStarPriceGold[card.Rare][1]
		cardPriceGold = cardNum * int32(oneCardPrice)
		isNeedGold = true
	}

	goods.setData(fmt.Sprintf("card:%d", cardId), int(cardNum), true, isNeedGold)
	rs.newGoodsList = append(rs.newGoodsList, goods)

	return rs.getPackMsg(fmt.Sprintf("card:%d", cardId), cardNum, cardPrice, cardPriceGold, true)
}

func (rs *randomShop) rankCard(cardRankUp map[uint32]float32, isUp bool) []card {
	var cards []card
	cards = cards[:0]
	for cardID, Rate := range cardRankUp {
		cards = append(cards, card{cardID, Rate})
	}
	sort.Slice(cards, func(i, j int) bool {
		if isUp {
			return cards[i].Value > cards[j].Value
		} else {
			return cards[i].Value < cards[j].Value
		}
	})
	return cards
}

func (rs *randomShop) randInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

func (rs *randomShop) randWeights(m map[string][]int, maxParaValue int) string {
	randValue := rs.randInt(1, maxParaValue)
	var randName string
	for k, v := range m {
		if randValue >= v[0] && randValue <= v[1] {
			randName = k
			break
		}
	}
	return randName
}

func timeToRefreshRandomShop() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		player.GetComponent(consts.ShopCpt).(*shopComponent).getRandomShop().pushToClient()
	})
}

func initRandShop() {
	timer.RunEveryDay(8, 0, 0, timeToRefreshRandomShop)
	timer.RunEveryDay(16, 0, 0, timeToRefreshRandomShop)
}
