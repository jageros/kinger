package shop

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"strconv"
	"time"
	//"kinger/gopuppy/common/evq"
)

var _ types.IShopComponent = &shopComponent{}

type shopComponent struct {
	attr         *attribute.MapAttr
	player       types.IPlayer
	limitGiftMgr *limitGiftMgrSt
	// 探访
	randomShop    iRandomShop
	orderID2Order map[string]*orderSt
	// 军备宝箱
	soldTreasure iSoldTreasure
	// 招募宝箱
	recruitTreasure iRecruitTreasure
	// 赞助
	adsMgr        iAdsMgr
	goldGift      *goldGiftSt
	recommendGift *recommendGiftSt
	midasPayment  *midasPaymentSt
	isClientInit  bool
	buyGoldTimer  *timer.Timer
}

func (sc *shopComponent) ComponentID() string {
	return consts.ShopCpt
}

func (sc *shopComponent) GetPlayer() types.IPlayer {
	return sc.player
}

func (sc *shopComponent) OnInit(player types.IPlayer) {
	sc.player = player
	sc.orderID2Order = map[string]*orderSt{}

	// soldTreasureAttr := sc.attr.GetMapAttr("soldTreasure")
	// if soldTreasureAttr != nil {
	// 	sc.recruitTreasure = newSoldTreasureByAttr(soldTreasureAttr, player)
	// }

	sc.limitGiftMgr = newLimitGiftMgr(player, sc.attr)
	sc.soldTreasure = newSoldTreasure(sc.player, sc.attr)
	sc.recruitTreasure = newRecruitTreasure(sc.attr, player)
	sc.randomShop = newRandomShop(sc.attr, player)
	sc.adsMgr = newAdsMgr(sc.attr, player)
	sc.goldGift = newGoldGiftSt(player, sc.attr)
	sc.midasPayment = newMidasPayment(player, sc.attr)
	sc.recommendGift = newRecommendGiftSt(player, sc.attr)

	sc.onMaxPvpLevelUpdate(player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel())
}

func (sc *shopComponent) OnLogin(isRelogin, isRestore bool) {
	sc.limitGiftMgr.caclHint(true)
	sc.adsMgr.onLogin()
	sc.recruitTreasure.onLogin()
	sc.randomShop.onLogin()
	sc.limitGiftMgr.forEachCanShowLimitGifts(func(gift iLimitGift) {
		gift.onLogin()
	})

	if sc.midasPayment != nil {
		sc.midasPayment.onLogin(sc.player)
	}

	sc.OnCrossDay(timer.GetDayNo())
	if !isRelogin {
		sc.beginBuyGoldTimer()
	}
}

func (sc *shopComponent) OnLogout() {
	sc.adsMgr.onLogout()
	sc.stopBuyGoldTimer()
	sc.soldTreasure.onLogout()
	if sc.midasPayment != nil {
		sc.midasPayment.onLogout()
	}
}

func (sc *shopComponent) OnCrossDay(dayno int) {
	if sc.player.GetDataDayNo() == dayno {
		return
	}

	sc.recruitTreasure.onCrossDay()
	sc.randomShop.onCrossDay()
	sc.limitGiftMgr.forEachCanShowLimitGifts(func(gift iLimitGift) {
		gift.onCrossDay()
	})
	if sc.goldGift != nil {
		sc.goldGift.onCrossDay()
	}
	sc.soldTreasure.onCrossDay()
	if sc.recommendGift != nil {
		sc.recommendGift.onCrossDay()
	}
}

func (sc *shopComponent) getBuySoldTreasureDay() int {
	return sc.attr.GetInt("buyStDay")
}

func (sc *shopComponent) setBuySoldTreasureDay(dayno int) {
	sc.attr.SetInt("buyStDay", dayno)
}

func (sc *shopComponent) getRecruitTreasure() iRecruitTreasure {
	return sc.recruitTreasure
}

func (sc *shopComponent) getSoldTreasure() iSoldTreasure {
	return sc.soldTreasure
}

func (sc *shopComponent) getRandomShop() iRandomShop {
	return sc.randomShop
}

func (sc *shopComponent) getLimitGiftGameData() gamedata.ILimitGiftGameData {
	accountType := sc.player.GetAccountType()
	channel := sc.player.GetChannel()
	if channel == "lzd_handjoy" {

		if accountType == pb.AccountTypeEnum_Ios {
			return gamedata.GetGameData(consts.IosHandjoyLimitGift).(*gamedata.IosHandJoyLimitGiftGameData)
		} else {
			return gamedata.GetGameData(consts.AndroidHandjoyLimitGift).(*gamedata.AndroidHandJoyLimitGiftGameData)
		}

	} else {
		if accountType == pb.AccountTypeEnum_Ios {
			return gamedata.GetGameData(consts.IosLimitGift).(*gamedata.LimitGiftGameData)
		} else if sc.player.IsWxgameAccount() {
			return gamedata.GetGameData(consts.WxLimitGift).(*gamedata.WxLimitGiftGameData)
		} else {
			return gamedata.GetGameData(consts.AndroidLimitGift).(*gamedata.AndroidLimitGiftGameData)
		}
	}
	return nil
}

func (sc *shopComponent) getJadeGoodsGameData() gamedata.IRechargeGameData {
	accountType := sc.player.GetAccountType()
	if accountType == pb.AccountTypeEnum_Ios {
		return gamedata.GetGameData(consts.IosRecharge).(*gamedata.RechargeGameData)
	} else if sc.player.IsWxgameAccount() {
		return gamedata.GetGameData(consts.WxRecharge).(*gamedata.WxRechargeGameData)
	} else {
		return gamedata.GetGameData(consts.AndroidRecharge).(*gamedata.AndroidRechargeGameData)
	}
	return nil
}

func (sc *shopComponent) onMaxPvpLevelUpdate(maxPvpLevel int) {
	sc.adsMgr.onMaxPvpLevelUpdate()
	sc.recruitTreasure.onMaxPvpLevelUpdate()
	// 最高等级提升，展示新的礼包
	sc.limitGiftMgr.onMaxPvpLevelUp(sc.getLimitGiftGameData(), maxPvpLevel)
}

func (sc *shopComponent) heartBeat(now time.Time) {
	sc.limitGiftMgr.heartBeat(now)
}

func (sc *shopComponent) getLimitGift(treasureModelID string) iLimitGift {
	return sc.limitGiftMgr.getLimitGiftByTreasure(treasureModelID)
}

func (sc *shopComponent) getLimitGiftByGoodsID(goodsID string) iLimitGift {
	return sc.limitGiftMgr.getLimitGiftByGoodsID(goodsID)
}

func (sc *shopComponent) getSoldTreasureRemainTime() int {
	remainTime := sc.attr.GetInt("soldTreasureTimeout") - int(time.Now().Unix())
	if remainTime < 0 {
		return 0
	}
	return remainTime
}

func (sc *shopComponent) getBuyGoldRemainTime() int {
	remainTime := sc.attr.GetInt("goldTimeout") - int(time.Now().Unix())
	if remainTime < 0 {
		return 0
	}
	return remainTime
}

func (sc *shopComponent) beginBuyGoldTimer() {
	sc.stopBuyGoldTimer()
	remainTime := sc.getBuyGoldRemainTime()
	if remainTime > 0 {
		sc.buyGoldTimer = timer.AfterFunc(time.Duration(remainTime)*time.Second, func() {
			sc.onShopDataUpdate(pb.UpdateShopDataArg_Gold)
		})
	}
}

func (sc *shopComponent) stopBuyGoldTimer() {
	if sc.buyGoldTimer != nil {
		sc.buyGoldTimer.Cancel()
		sc.buyGoldTimer = nil
	}
}

func (sc *shopComponent) buyGold(goodsID string) (*pb.BuyGoldReply, error) {
	remainTime := sc.getBuyGoldRemainTime()
	if remainTime > 0 {
		return nil, gamedata.GameError(3)
	}

	goodsList := gamedata.GetSoldGoldGameData().GetGoodsList(sc.player.GetArea())
	for _, goldGoods := range goodsList {
		if goldGoods.GoodsID == goodsID {

			resCpt := sc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
			var resType, resAmount int
			if goldGoods.BowlderPrice > 0 {
				if !sc.player.HasBowlder(goldGoods.BowlderPrice) {
					return nil, gamedata.GameError(1)
				}
				resType = consts.Bowlder
				resAmount = goldGoods.BowlderPrice
				sc.player.SubBowlder(goldGoods.BowlderPrice, consts.RmrShopBuyGold)
			} else {

				if !resCpt.HasResource(consts.Jade, goldGoods.JadePrice) {
					return nil, gamedata.GameError(1)
				}
				resType = consts.Jade
				resAmount = goldGoods.JadePrice
				resCpt.ModifyResource(consts.Jade, -goldGoods.JadePrice, consts.RmrShopBuyGold)
			}

			now := time.Now()
			funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
			nextRemainTime := funcPrice.ShopGoldCD
			nextTime := now.Add(nextRemainTime)
			sc.attr.SetInt("goldTimeout", int(nextTime.Unix()))

			resCpt.ModifyResource(consts.Gold, goldGoods.Gold, consts.RmrShopBuyGold)

			mod.LogShopBuyItem(sc.player, fmt.Sprintf("soldGold_%d", goldGoods.Gold),
				fmt.Sprintf("金币_%d", goldGoods.Gold), 1, "shop",
				strconv.Itoa(resType), module.Player.GetResourceName(resType), resAmount,
				fmt.Sprintf("goodsID=%s, needjade=%d, needBowlder=%d, nextRemainTime=%f",
					goodsID, goldGoods.JadePrice, goldGoods.BowlderPrice, nextRemainTime.Seconds()))

			sc.beginBuyGoldTimer()

			return &pb.BuyGoldReply{
				Gold:           int32(goldGoods.Gold),
				NextRemainTime: int32(nextRemainTime.Seconds()),
			}, nil
		}
	}
	return nil, gamedata.GameError(2)
}

func (sc *shopComponent) getAddGoldRemainTime() int64 {
	nextAddGoldTime := sc.attr.GetInt64("nextAddGoldTime")
	now := time.Now().Unix()
	t := now - nextAddGoldTime
	if t < 0 {
		t = 0
	}
	return t
}

func (sc *shopComponent) WatchShopFreeAds(type_ pb.ShopFreeAdsType, id int, isConsumeJade bool) (*pb.WatchShopFreeAdsReply, error) {
	return sc.adsMgr.watchAds(type_, id, isConsumeJade)
}

func (sc *shopComponent) OnShopAdsBeHelp(type_ pb.ShopFreeAdsType, id int) error {
	return sc.adsMgr.onWxBeHelp(type_, id)
}

func (sc *shopComponent) addCumulativePay(amount int) {
	sc.attr.SetInt("cumulativePay", sc.attr.GetInt("cumulativePay")+amount)
}

func (sc *shopComponent) GetCumulativePay() int {
	return sc.attr.GetInt("cumulativePay")
}

func (sc *shopComponent) loadOrder(orderID string) *orderSt {
	order, ok := sc.orderID2Order[orderID]
	if ok {
		return order
	}
	orderAttr := attribute.NewAttrMgr("order", orderID)
	err := orderAttr.Load()

	order, ok = sc.orderID2Order[orderID]
	if ok {
		return order
	}

	if err != nil {
		return nil
	}

	order = newOrderByAttr(orderAttr)
	sc.orderID2Order[orderID] = order
	return order
}

func (sc *shopComponent) OnSdkRecharge(channelUid, cpOrderID, channelOrderID string, paymentAmount int, needCheckMoney bool) {
	order := sc.loadOrder(cpOrderID)
	reply := &pb.SdkRechargeResult{Errcode: pb.SdkRechargeResult_Fail}
	if order == nil {
		glog.Errorf("OnSdkRecharge no order, uid=%d, accountType=%s, channelUid=%s, cpOrderID=%s, channelOrderID=%s, "+
			"paymentAmount=%d", sc.player.GetUid(), sc.player.GetAccountType(), channelUid, cpOrderID, channelOrderID, paymentAmount)
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT, reply)
		return
	}

	if order.isComplete() {
		glog.Errorf("OnSdkRecharge order isComplete, uid=%d, accountType=%s, channelUid=%s, cpOrderID=%s, channelOrderID=%s, "+
			"paymentAmount=%d", sc.player.GetUid(), sc.player.GetAccountType(), channelUid, cpOrderID, channelOrderID, paymentAmount)
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT, reply)
		return
	}

	order.setChannelOrderID(channelOrderID)
	reply = mod.payment.onRecharge(sc.player, order, sc.player.GetChannel(), float64(paymentAmount), needCheckMoney)
	sc.player.GetAgent().PushClient(pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT, reply)
}

func (sc *shopComponent) CompensateRecharge(cpOrderID, channelOrderID, goodsID string) {
	if cpOrderID == "" {
		orderID := fmt.Sprintf("%d%d", sc.player.GetUid(), time.Now().UnixNano()/1000)
		order := newOrder(orderID, goodsID, "", sc.player)
		order.setIsTest()
		reply := mod.payment.onRecharge(sc.player, order, sc.player.GetChannel(), 0, false)
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_NOTIFY_SDK_RECHARGE_RESULT, reply)
		return
	}

	if channelOrderID == "" {
		channelOrderID = cpOrderID
	}
	sc.OnSdkRecharge(sc.player.GetChannelUid(), cpOrderID, channelOrderID, 0, false)
}

func (sc *shopComponent) isGooglePlayOrderComplete(orderID string) bool {
	googlePlayOrderAttr := sc.attr.GetMapAttr("googlePlayOrder")
	if googlePlayOrderAttr == nil {
		return false
	} else {
		return googlePlayOrderAttr.HasKey(orderID)
	}
}

func (sc *shopComponent) googlePlayOrderComplete(orderID string) {
	googlePlayOrderAttr := sc.attr.GetMapAttr("googlePlayOrder")
	if googlePlayOrderAttr == nil {
		googlePlayOrderAttr = attribute.NewMapAttr()
		sc.attr.SetMapAttr("googlePlayOrder", googlePlayOrderAttr)
	}
	googlePlayOrderAttr.SetBool(orderID, true)
}

func (sc *shopComponent) onReborn() {
	sc.limitGiftMgr.onReborn()
}

func (sc *shopComponent) getAdsMgr() iAdsMgr {
	return sc.adsMgr
}

func (sc *shopComponent) hasEverBuyVip() bool {
	return false
	//return sc.attr.GetStr("everBuyVip") != ""
}
func (sc *shopComponent) onBuyVip(goodsID string) {
	sc.attr.SetStr("everBuyVip", goodsID)
	//evq.CallLater(func() {
	//	sc.onShopDataUpdate(pb.UpdateShopDataArg_Vip)
	//})
}

func (sc *shopComponent) setJadeData(recharge *gamedata.Recharge) (isDouble bool) {
	var hasGoods bool
	buyJadeAttr := sc.getBuyJadeListAttr()
	buyJadeAttr.ForEachIndex(func(index int) bool {
		goodsAttr := buyJadeAttr.GetMapAttr(index)
		if goodsAttr.GetStr("goodsID") == recharge.GoodsID {
			if goodsAttr.GetInt("ver") != recharge.FirstJadrVer {
				buyJadeAttr.DelMapAttr(goodsAttr)
			} else {
				isDouble = goodsAttr.GetBool("isDouble")
				hasGoods = true
			}
		}
		return true
	})

	if !hasGoods {
		newGoods := attribute.NewMapAttr()
		newGoods.SetStr("goodsID", recharge.GoodsID)
		newGoods.SetBool("isDouble", true)
		newGoods.SetInt("ver", recharge.FirstJadrVer)
		buyJadeAttr.AppendMapAttr(newGoods)
		isDouble = true
	}
	return
}

func (sc *shopComponent) setJadeIsDouble(goodsID string) {
	buyJadeAttr := sc.getBuyJadeListAttr()
	buyJadeAttr.ForEachIndex(func(index int) bool {
		goodsAttr := buyJadeAttr.GetMapAttr(index)
		if goodsAttr.GetStr("goodsID") == goodsID {
			goodsAttr.SetBool("isDouble", false)
		}
		return true
	})
}

func (sc *shopComponent) getJadeIsDouble(goodsID string) (isDouble bool) {
	buyJadeAttr := sc.getBuyJadeListAttr()
	buyJadeAttr.ForEachIndex(func(index int) bool {
		goodsAttr := buyJadeAttr.GetMapAttr(index)
		if goodsAttr.GetStr("goodsID") == goodsID {
			isDouble = goodsAttr.GetBool("isDouble")
		}
		return true
	})
	return
}

func (sc *shopComponent) getBuyJadeListAttr() *attribute.ListAttr {
	buyJadeAttr := sc.attr.GetListAttr("buyJadeList")
	if buyJadeAttr == nil {
		buyJadeAttr = attribute.NewListAttr()
		sc.attr.SetListAttr("buyJadeList", buyJadeAttr)
	}
	return buyJadeAttr
}

func (sc *shopComponent) packJadeGoodsMsg() []*pb.JadeGoods {
	rechargeGameData := sc.getJadeGoodsGameData()
	var goods []*pb.JadeGoods
	if rechargeGameData == nil {
		return goods
	}

	for _, recharge := range rechargeGameData.GetAllGoods(sc.player.GetArea()) {
		isDouble := sc.setJadeData(recharge)
		goods = append(goods, &pb.JadeGoods{
			GoodsID:  recharge.GoodsID,
			Price:    int32(recharge.Price),
			Jade:     int32(recharge.JadeCnt),
			IsDouble: isDouble,
		})
	}

	return goods
}

func (sc *shopComponent) packVipGoodsMsg() *pb.VipCardGoods {
	if sc.player.IsForeverVip() {
		return nil
	}
	limitGiftGameData := sc.getLimitGiftGameData()
	if limitGiftGameData != nil {
		vipCardData := limitGiftGameData.GetVipCard(sc.player.GetArea(), !sc.hasEverBuyVip())
		if vipCardData != nil {
			return &pb.VipCardGoods{
				GoodsID:   vipCardData.GiftID,
				JadePrice: int32(vipCardData.JadePrice),
				Price:     int32(vipCardData.Price),
			}
		}
	}

	return nil
}

func (sc *shopComponent) packGoldGoodsMsg() ([]*pb.GoldGoods, int32) {
	var goldGoodsList []*pb.GoldGoods
	goodsDataList := gamedata.GetSoldGoldGameData().GetGoodsList(sc.player.GetArea())
	for _, goldGoods := range goodsDataList {
		goldGoodsList = append(goldGoodsList, &pb.GoldGoods{
			GoodsID:  goldGoods.GoodsID,
			Gold:     int32(goldGoods.Gold),
			NeedJade: int32(goldGoods.JadePrice),
		})
	}

	var buyGoldRemainTime int32
	if len(goldGoodsList) > 0 {
		buyGoldRemainTime = int32(sc.getBuyGoldRemainTime())
	}

	return goldGoodsList, buyGoldRemainTime
}

func (sc *shopComponent) packFreeAdsMsg() []*pb.ShopFreeAds {
	var adsList []*pb.ShopFreeAds
	sc.getAdsMgr().forEachAds(func(a *freeAds) {
		if len(adsList) < 3 {
			adsList = append(adsList, a.packMsg())
		}
	})
	return adsList
}

func (sc *shopComponent) packMsg() *pb.ShopData {
	reply := &pb.ShopData{
		Gift:            sc.limitGiftMgr.packMsg(),
		SoldTreasures:   sc.getSoldTreasure().packMsg(),
		RecruitTreasure: sc.getRecruitTreasure().packMsg(),
		RandomShopData:  sc.getRandomShop().packMsg(),
		JadeGoodsList:   sc.packJadeGoodsMsg(),
		VipCard:         sc.packVipGoodsMsg(),
		Adses:           sc.packFreeAdsMsg(),
	}

	if sc.goldGift != nil {
		reply.GoldGift = sc.goldGift.packMsg()
	}

	if sc.recommendGift != nil {
		reply.RecommendGift = sc.recommendGift.packMsg()
	}

	reply.GoldGoodsList, reply.BuyGoldRemainTime = sc.packGoldGoodsMsg()
	sc.isClientInit = true
	return reply
}

func (sc *shopComponent) onShopDataUpdate(type_ pb.UpdateShopDataArg_DataType) {
	if !sc.isClientInit {
		return
	}
	if !config.GetConfig().IsXfServer() {
		return
	}

	agent := sc.player.GetAgent()
	if agent == nil {
		return
	}

	var data proto.Marshaler = nil
	switch type_ {
	case pb.UpdateShopDataArg_LimitGift:
		data = &pb.UpdateLimitGiftArg{Gift: sc.limitGiftMgr.packMsg()}

	case pb.UpdateShopDataArg_SoldTreasure:
		msg := sc.getSoldTreasure().packMsg()
		if msg != nil {
			data = msg
		}

	case pb.UpdateShopDataArg_RecruitTreasure:
		msg := sc.getRecruitTreasure().packMsg()
		if msg != nil {
			data = msg
		}

	case pb.UpdateShopDataArg_Jade:
		data = &pb.UpdateJadeGoodsArg{JadeGoodsList: sc.packJadeGoodsMsg()}

	case pb.UpdateShopDataArg_Gold:
		msg := &pb.UpdateGoldGoodsArg{}
		msg.GoldGoodsList, msg.BuyGoldRemainTime = sc.packGoldGoodsMsg()
		data = msg

	case pb.UpdateShopDataArg_FreeAds:
		data = &pb.UpdateFreeAdsArg{
			Adses: sc.packFreeAdsMsg(),
		}

	case pb.UpdateShopDataArg_Vip:
		msg := sc.packVipGoodsMsg()
		if msg != nil {
			data = msg
		}

	case pb.UpdateShopDataArg_RandomShop:
		msg := sc.getRandomShop().packMsg()
		if msg != nil {
			data = msg
		}

	case pb.UpdateShopDataArg_GoldGift:
		if sc.goldGift != nil {
			data = sc.goldGift.packMsg()
		}
	case pb.UpdateShopDataArg_RecommendGift:
		if sc.recommendGift != nil {
			data = sc.recommendGift.packMsg()
		}

	default:
		return
	}

	msg := &pb.UpdateShopDataArg{Type: type_}
	if data != nil {
		msg.Data, _ = data.Marshal()
	}
	agent.PushClient(pb.MessageID_S2C_UPDATE_SHOP_DATA, msg)
}

func (sc *shopComponent) onMidasRecharge(arg *pb.MidasRechargeArg) {
	if sc.midasPayment != nil {
		sc.midasPayment.onRecharge(arg.OrderID, arg.MidasOpenkey, arg.Pf, arg.Pfkey)
	}
}

func (sc *shopComponent) pushExchangeIdToClient() {
	crd.pushRecruitIdsToClient(sc.player)
}
