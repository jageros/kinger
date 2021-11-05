package shop

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"math/rand"
	"strconv"
)

type recommendGiftSt struct {
	player types.IPlayer
	attr   *attribute.MapAttr
}

func newRecommendGiftSt(player types.IPlayer, cptAttr *attribute.MapAttr) *recommendGiftSt {
	attr := cptAttr.GetMapAttr("recommendGift")
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr("recommendGift", attr)
	}
	return &recommendGiftSt{
		player: player,
		attr:   attr,
	}
}

func (gg *recommendGiftSt) getCurGiftId() string {
	return gg.attr.GetStr("GiftId")
}

func (gg *recommendGiftSt) setCurGiftId(giftId string) {
	gg.attr.SetStr("GiftId", giftId)
}

func (gg *recommendGiftSt) getGiftCardId() uint32 {
	return gg.attr.GetUInt32("GiftCardId")
}

func (gg *recommendGiftSt) setGiftCardId(cid uint32) {
	gg.attr.SetUInt32("GiftCardId", cid)
}

func (gg *recommendGiftSt) hasBuy() bool {
	return gg.attr.GetBool("hasBuy")
}

func (gg *recommendGiftSt) setBuy(hasBuy bool) {
	gg.attr.SetBool("hasBuy", hasBuy)
}

func (gg *recommendGiftSt) getGiftID() string {
	gid := gg.getCurGiftId()
	if gid == "" {
		gg.refreshGiftData()
	}
	return gg.getCurGiftId()
}

func (gg *recommendGiftSt) getGameData() *gamedata.LimitGift {
	goodsID := gg.getGiftID()
	gift := gg.player.GetComponent(consts.ShopCpt).(*shopComponent).getLimitGiftGameData().GetGiftByID(goodsID)
	if gift == nil {
		glog.Errorf("recommend_gift GetLimitGiftPrice error, uid=%d, goosID=%s", gg.player.GetUid(), goodsID)
		return nil
	}
	return gift
}

func (gg *recommendGiftSt) packMsg() *pb.SoldRecommendGift {
	data := gg.getGameData()
	var remainTime int32
	if data == nil {
		return &pb.SoldRecommendGift{}
	}
	if gg.hasBuy() {
		return &pb.SoldRecommendGift{}
	}
	treasureId := data.Reward
	trData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).GetTreasureByBXID(treasureId)
	remainTime = int32(timer.TimeDelta(0, 0, 0).Seconds())
	cards := map[uint32]int32{}
	gCid := gg.getGiftCardId()
	for _, cid := range trData.Reward {
		if cid == 0 {
			if _, ok := cards[gCid]; ok {
				cards[gCid] += 1
			} else {
				cards[gCid] = 1
			}
		} else {
			if _, ok := cards[cid]; ok {
				cards[cid] += 1
			} else {
				cards[cid] = 1
			}
		}
	}
	return &pb.SoldRecommendGift{
		GiftID:            data.GiftID,
		JadePrice:         int32(data.JadePrice),
		Cards:             cards,
		RefreshRemainTime: remainTime,
	}
}

func (gg *recommendGiftSt) refreshGiftData() {
	curCamp := gg.player.GetCurCamp()
	var iCards []types.ICollectCard
	for lv := 4; lv <=5; lv ++ {
		iCards = gg.player.GetPvpCardPoolsByCamp(curCamp)
		for i := 0; i < len(iCards); i++ {
			if iCards[i].IsSpCard() || iCards[i].GetLevel() >= lv{
				iCards = append(iCards[:i], iCards[i+1:] ... )
				i--
			}
		}
		if len(iCards) <= 0 {
			iCards = gg.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetAllCollectCards()
			for i := 0; i < len(iCards); i++ {
				if iCards[i].IsSpCard() || iCards[i].GetLevel() >= lv {
					iCards = append(iCards[:i], iCards[i+1:] ... )
					i--
				}
			}
		}

		if len(iCards) > 0 {
			break
		}
	}

	if len(iCards) <= 0 {
		return
	}
	idx := rand.Intn(len(iCards))
	cid := iCards[idx].GetCardID()
	clv := iCards[idx].GetLevel()
	giftIDs := "dailybox1"
	if clv >= 3 {
		giftIDs = "dailybox2"
	}
	glog.Infof("cid=%d, lvl=%d, card-len=%d", cid, clv, len(iCards))
	gg.setGiftCardId(cid)
	gg.setCurGiftId(giftIDs)
	gg.setBuy(false)
}

func (gg *recommendGiftSt) onCrossDay() {
	gg.refreshGiftData()
	gg.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_RecommendGift)
}

func (gg *recommendGiftSt) buy() (*pb.BuyRecommendGiftReply, error) {
	if gg.hasBuy() {
		return nil, gamedata.GameError(1)
	}
	data := gg.getGameData()
	if data == nil {
		return nil, gamedata.GameError(1)
	}

	if !module.Player.HasResource(gg.player, consts.Jade, data.JadePrice) {
		return nil, gamedata.GameError(2)
	}
	module.Player.ModifyResource(gg.player, consts.Jade, - data.JadePrice, consts.RmrBuyRecommendGift)

	module.Shop.LogShopBuyItem(gg.player, gg.getCurGiftId(), "推荐礼包", 1, "shop",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), data.JadePrice, "")

	reply := &pb.BuyRecommendGiftReply{}
	reply.TreasureReward = module.Treasure.OpenTreasureByModelID(gg.player, data.Reward, false)
	var cardNum int32
	curCid := gg.getGiftCardId()
	for k, cid := range reply.TreasureReward.CardIDs {
		if cid == 0 {
			cardNum++
			reply.TreasureReward.CardIDs[k] = curCid
		}
	}
	cardCpt := gg.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cif := &pb.CardInfo{Amount: cardNum}
	cardMap := map[uint32]*pb.CardInfo{curCid: cif}
	cardCpt.ModifyCollectCards(cardMap)
	gg.setBuy(true)
	glog.Infof("player uid=%d, buy recommend_gift modify collect card cardID=%d, Amount=%d", gg.player.GetUid(), curCid, cardNum)
	return reply, nil
}
