package shop

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
	"time"
)

func rpc_C2S_FetchShopData(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.ShopCpt).(*shopComponent).packMsg(), nil
}

func rpc_C2S_BuySoldTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.ShopCpt).(*shopComponent).getSoldTreasure().buy(arg.(*pb.BuySoldTreasureArg).TreasureModelID)
}

func rpc_C2S_BuyGold(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.ShopCpt).(*shopComponent).buyGold(arg.(*pb.BuyGoldArg).GoodsID)
}

// 已废弃
func rpc_C2S_BuyJade(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.BuyJadeArg)
	glog.Infof("rpc_C2S_BuyJade begin uid=%d, goodsID=%s", player.GetUid(), arg2.GoodsID)
	order, err := newIosOrder(player, arg2.GoodsID, arg2.Receipt)
	if err != nil {
		glog.Errorf("rpc_C2S_BuyJade appStoreVerifyReceipt err=%s, uid=%d, goods=%s", err, player.GetUid(), arg2.GoodsID)
		return nil, gamedata.GameError(1)
	}

	reply := mod.payment.onRecharge(player, order, player.GetChannel(), 0, false)
	if reply.Errcode == pb.SdkRechargeResult_Success && reply.Type == pb.SdkRechargeResult_Jade {
		return reply.Data, nil
	} else {
		return nil, gamedata.GameError(2)
	}
}

// 已废弃
func rpc_C2S_BuyLimitGift(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	arg2 := arg.(*pb.BuyLimitGiftArg)
	gift := shopCpt.getLimitGift(arg2.GiftID)
	if gift == nil || !gift.canBuy() {
		glog.Errorf("rpc_C2S_BuyLimitGift no %s, uid=%d", arg2.GiftID, player.GetUid())
		return nil, gamedata.GameError(1)
	}
	goodsID := gift.getGiftID()

	order, err := newIosOrder(player, goodsID, arg2.Receipt)
	if err != nil {
		glog.Errorf("rpc_C2S_BuyLimitGift appStoreVerifyReceipt err=%s, uid=%d, goods=%s", err, player.GetUid(), goodsID)
		return nil, gamedata.GameError(1)
	}

	reply := mod.payment.onRecharge(player, order, player.GetChannel(), 0, false)
	if reply.Errcode == pb.SdkRechargeResult_Success && reply.Type == pb.SdkRechargeResult_LimitGift {
		return reply.Data, nil
	} else {
		return nil, gamedata.GameError(2)
	}
}

func rpc_C2S_WatchShopAddGoldAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	arg2 := arg.(*pb.WatchShopFreeAdsArg)
	return shopCpt.WatchShopFreeAds(arg2.Type, int(arg2.ID), arg2.IsConsumeJade)
}

func rpc_C2S_SdkCreateOrder(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.SdkCreateOrderArg)
	var price int
	jadeGoods := mod.payment.getJadeGoods(arg2.GoodsID, player)
	if jadeGoods != nil {
		price = jadeGoods.Price
	} else {

		giftGoods := mod.payment.getGiftGoods(arg2.GoodsID, player)
		if giftGoods != nil {
			price = giftGoods.getGameData().Price
		} else {

			vipGoods := mod.payment.getVipGoods(arg2.GoodsID, player)
			if vipGoods != nil {
				price = vipGoods.Price
			} else {
				return nil, gamedata.GameError(1)
			}
		}
	}

	orderID := fmt.Sprintf("%d%d", uid, time.Now().UnixNano()/1000)
	order := newOrder(orderID, arg2.GoodsID, "", player)
	order.setPrice(price)
	return &pb.SdkCreateOrderReply{OrderID: orderID}, order.save(true)
}

func rpc_C2S_IosPrePay(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.IosPrePayArg)
	glog.Infof("rpc_C2S_IosPrePay uid=%d, goodsID=%s", uid, arg2.GoodsID)
	return nil, nil
}

func rpc_C2S_BuyLimitGiftByJade(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	limitGiftGameData := shopCpt.getLimitGiftGameData()
	if limitGiftGameData == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.BuyLimitGiftByJadeArg)
	gift := shopCpt.getLimitGiftByGoodsID(arg2.GiftID)
	if gift == nil || !gift.canBuy() {
		glog.Errorf("rpc_C2S_BuyLimitGiftByJade no %s, uid=%d", arg2.GiftID, player.GetUid())
		return nil, gamedata.GameError(1)
	}

	gdata := gift.getGameData()
	if gdata == nil {
		return nil, gamedata.GameError(4)
	}

	jade := gdata.JadePrice
	bowlder := gdata.BowlderPrice
	var resType, resAmount int

	if bowlder <= 0 {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, jade) {
			return nil, gamedata.GameError(2)
		}

		resType = consts.Jade
		resAmount = jade
		resCpt.ModifyResource(consts.Jade, -jade, consts.RmrLimitGift)
	} else {

		if !player.HasBowlder(bowlder) {
			return nil, gamedata.GameError(3)
		}

		resType = consts.Bowlder
		resAmount = bowlder
		player.SubBowlder(bowlder, consts.RmrLimitGift)
	}

	itemID := fmt.Sprintf("%s_%s", gdata.GiftID, gdata.Reward)
	mod.LogShopBuyItem(player, itemID, itemID, 1, "shop", strconv.Itoa(resType),
		module.Player.GetResourceName(resType), resAmount, "")

	reply, _ := gift.buy()
	return reply, nil
}

func rpc_C2S_BuyOneGaChaByJade(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	limitGiftGameData := shopCpt.getLimitGiftGameData()
	if limitGiftGameData == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.BuyLimitGiftByJadeArg)
	gift := shopCpt.getLimitGiftByGoodsID(arg2.GiftID)
	if gift == nil || !gift.canBuy() {
		glog.Errorf("rpc_C2S_BuyOneGaChaByJade no %s, uid=%d", arg2.GiftID, player.GetUid())
		return nil, gamedata.GameError(1)
	}

	gdata := gift.getGameData()
	if gdata == nil {
		return nil, gamedata.GameError(4)
	}

	jade := gdata.JadePrice
	bowlder := gdata.BowlderPrice
	if jade <= 0 {
		return nil, gamedata.GameError(5)
	}

	var resType, resAmount int
	if bowlder <= 0 {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, jade) {
			return nil, gamedata.GameError(2)
		}

		resType = consts.Jade
		resAmount = jade
		resCpt.ModifyResource(consts.Jade, -jade, consts.RmrLimitGift)
	} else {

		if !player.HasBowlder(bowlder) {
			return nil, gamedata.GameError(3)
		}

		resType = consts.Bowlder
		resAmount = bowlder
		player.SubBowlder(bowlder, consts.RmrLimitGift)
	}

	itemID := fmt.Sprintf("%s_%s", gdata.GiftID, gdata.Reward)
	mod.LogShopBuyItem(player, itemID, itemID, 1, "shop", strconv.Itoa(resType),
		module.Player.GetResourceName(resType), resAmount, "")

	reply, _ := gift.buy()
	return reply, nil
}

type googlePlayPurchaseData struct {
	OrderId          string `json:"orderId"`
	PurchaseState    int    `json:"purchaseState"`
	ProductId        string `json:"productId"`
	DeveloperPayload string `json:"developerPayload"`
}

func rpc_C2S_GooglePlayRecharge(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.SdkRechargeResult{Errcode: pb.SdkRechargeResult_Fail}
	arg2 := arg.(*pb.GooglePlayRechargeArg)
	decodeSign, err := base64.StdEncoding.DecodeString(arg2.InappDataSignature)
	if err != nil {
		glog.Errorf("rpc_C2S_GooglePlayRecharge decodeSign error uid=%d, InappPurchaseData=%s, err=%s",
			uid, arg2.InappPurchaseData, err)
		return reply, nil
	}

	sh1 := sha1.New()
	sh1.Write([]byte(arg2.InappPurchaseData))
	hashData := sh1.Sum(nil)

	result := rsa.VerifyPKCS1v15(googlePlayPublicKey, crypto.SHA1, hashData, decodeSign)
	if result != nil {
		glog.Errorf("rpc_C2S_GooglePlayRecharge VerifyPKCS1v15 error, uid=%d, InappPurchaseData=%s,  err=%s",
			uid, arg2.InappPurchaseData, err)
		return reply, nil
	}

	payData := &googlePlayPurchaseData{}
	err = json.Unmarshal([]byte(arg2.InappPurchaseData), payData)
	if err != nil {
		glog.Errorf("rpc_C2S_GooglePlayRecharge Unmarshal error, uid=%d, InappPurchaseData=%s, err=%s",
			uid, arg2.InappPurchaseData, err)
		return reply, nil
	}

	var isTest bool
	if payData.OrderId == "" {
		if !config.GetConfig().AppStoreCanTest {
			glog.Errorf("rpc_C2S_GooglePlayRecharge cant test, uid=%d, InappPurchaseData=%s", uid, arg2.InappPurchaseData)
			return reply, nil
		} else {
			isTest = true
		}
	}

	if payData.ProductId != arg2.GoodsID {
		glog.Errorf("rpc_C2S_GooglePlayRecharge ProductId error, uid=%d, InappPurchaseData=%s, isTest=%v", uid,
			arg2.InappPurchaseData, isTest)
		return reply, nil
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	if !isTest && shopCpt.isGooglePlayOrderComplete(payData.OrderId) {
		glog.Errorf("rpc_C2S_GooglePlayRecharge google order isComplete, uid=%d, orderID=%s, InappPurchaseData=%s, "+
			"isTest=%v", uid, payData.OrderId, arg2.InappPurchaseData, isTest)
		return reply, nil
	}

	order := shopCpt.loadOrder(payData.DeveloperPayload)
	if order == nil {
		glog.Errorf("rpc_C2S_GooglePlayRecharge no order, uid=%d, InappPurchaseData=%s, isTest=%v", uid,
			arg2.InappPurchaseData, isTest)
		return reply, nil
	}
	if order.isComplete() {
		glog.Errorf("rpc_C2S_GooglePlayRecharge cp order isComplete, uid=%d, cpOrderID=%s, InappPurchaseData=%s, "+
			"isTest=%v", uid, payData.DeveloperPayload, arg2.InappPurchaseData, isTest)
		return reply, nil
	}

	order.setChannelOrderID(payData.OrderId)
	order.setCurrency(arg2.Currency)
	reply = mod.payment.onRecharge(player, order, "play.google.com", float64(arg2.Money)/1000000.0, false)
	if !isTest && reply.Errcode == pb.SdkRechargeResult_Success {
		shopCpt.googlePlayOrderComplete(payData.OrderId)
	}
	return reply, nil
}

func rpc_C2S_BuyVipCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	if player.IsForeverVip() {
		return nil, gamedata.GameError(1)
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	limitGiftGameData := shopCpt.getLimitGiftGameData()
	if limitGiftGameData == nil {
		return nil, gamedata.GameError(2)
	}
	vipCardGameData := limitGiftGameData.GetVipCard(player.GetArea(), !shopCpt.hasEverBuyVip())

	jade := vipCardGameData.JadePrice
	bowlder := vipCardGameData.BowlderPrice
	var resType, resAmount int

	if bowlder <= 0 {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, jade) {
			return nil, gamedata.GameError(3)
		}

		resType, resAmount = consts.Jade, jade
		resCpt.ModifyResource(consts.Jade, -jade, consts.RmrBuyVip)
	} else {

		if !player.HasBowlder(bowlder) {
			return nil, gamedata.GameError(4)
		}

		resType, resAmount = consts.Bowlder, bowlder
		player.SubBowlder(bowlder, consts.RmrBuyVip)
	}

	mod.LogShopBuyItem(player, vipCardGameData.GiftID, vipCardGameData.GiftID, 1, "shop",
		strconv.Itoa(resType), module.Player.GetResourceName(resType), resAmount, "")

	vipSt := module.OutStatus.GetStatus(player, consts.OtVipCard)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if vipSt == nil {
		vipSt = module.OutStatus.AddStatus(player, consts.OtVipCard, funcPrice.VipContinuedTime)
	} else {
		vipSt.Over(funcPrice.VipContinuedTime)
	}

	headFrame := "" //vipCardGameData.HeadFrame
	//if headFrame != "" {
	//	if !module.Bag.HasItem(player, consts.ItHeadFrame, headFrame) {
	//		module.Bag.AddHeadFrame(player, vipCardGameData.HeadFrame)
	//	} else {
	//		headFrame = ""
	//	}
	//}

	reply := &pb.BuyVipCardReply{
		HeadFrame:  headFrame,
		RemainTime: -1,
	}

	if vipSt != nil {
		reply.RemainTime = int32(vipSt.GetRemainTime())
	}

	if vipCardGameData.IsFirstVip() {
		shopCpt.onBuyVip(vipCardGameData.GiftID)
	}

	return reply, nil
}

func rpc_C2S_PieceExchangeItem(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.PieceExchangeArg)
	var shopName string
	if arg2.PieceType == pb.PieceExchangeArg_Card {
		shopName = "cardPiece"
	} else {
		shopName = "skinPiece"
	}

	s := mod.getShop(shopName)
	if s == nil {
		return nil, gamedata.GameError(1)
	}

	return nil, s.buyGoods(player, int(arg2.GoodsID))
}

func rpc_C2S_LookOverLimitGift(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetLimitGift)
	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	gift := shopCpt.getLimitGiftByGoodsID(arg2.GiftID)
	if gift == nil || !gift.isNew() {
		return nil, gamedata.InternalErr
	}

	gift.setIsNew(false)
	shopCpt.limitGiftMgr.caclHint(false)
	return nil, nil
}

func rpc_C2S_BuyRecruitTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	return player.GetComponent(consts.ShopCpt).(*shopComponent).getRecruitTreasure().buy(int(arg.(*pb.BuyRecruitTreasureArg).BuyCnt))
}

func rpc_C2S_BuyRandomShopRefreshCnt(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	return player.GetComponent(consts.ShopCpt).(*shopComponent).getRandomShop().buyRefreshCnt(
		int(arg.(*pb.BuyRandomShopRefreshCntArg).BuyCnt))
}

func rpc_C2S_BuyRandomShop(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	return player.GetComponent(consts.ShopCpt).(*shopComponent).getRandomShop().buy(arg.(*pb.BuyRandomShopArg))
}

func rpc_C2S_IosRecharge(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.IosRechargeArg)
	glog.Infof("rpc_C2S_IosRecharge begin uid=%d, goodsID=%s", player.GetUid(), arg2.GoodsID)
	order, err := newIosOrder(player, arg2.GoodsID, arg2.Receipt)
	if err != nil {
		glog.Errorf("IosRechargeArg appStoreVerifyReceipt err=%s, uid=%d, goods=%s", err, player.GetUid(), arg2.GoodsID)
		return nil, gamedata.GameError(1)
	}

	return mod.payment.onRecharge(player, order, player.GetChannel(), 0, false), nil
}

func rpc_C2S_TestRecharge(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	if !config.GetConfig().Debug {
		return nil, gamedata.InternalErr
	}
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	arg2 := arg.(*pb.IosRechargeArg)
	orderID := fmt.Sprintf("%d%d", uid, time.Now().UnixNano()/1000)
	order := newOrder(orderID, arg2.GoodsID, "", player)
	order.setIsTest()
	return mod.payment.onRecharge(player, order, player.GetChannel(), 0, false), nil
}

func rpc_C2S_FetchPieceExchangeIds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	msg := &pb.PieceExchangeIds{}
	cids, sids := crd.getExchangeIdsByArea(player.GetArea())
	msg.ExchangeCardIds = cids
	msg.ExchangeSkinIds = sids
	return msg, nil
}

func rpc_C2S_BuySoldGoldGift(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	if shopCpt.goldGift == nil {
		return nil, gamedata.GameError(1)
	}
	return shopCpt.goldGift.buy()
}

func rpc_C2S_MidasRecharge(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	player.GetComponent(consts.ShopCpt).(*shopComponent).onMidasRecharge(arg.(*pb.MidasRechargeArg))
	return nil, nil
}

func rpc_C2S_BuyRecommendGift(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	if shopCpt.recommendGift == nil {
		return nil, gamedata.GameError(1)
	}
	return shopCpt.recommendGift.buy()
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SHOP_DATA, rpc_C2S_FetchShopData)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_SOLDTREASURE, rpc_C2S_BuySoldTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_GOLD, rpc_C2S_BuyGold)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_JADE, rpc_C2S_BuyJade)            // 已废弃
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_LIMIT_GITF, rpc_C2S_BuyLimitGift) // 已废弃
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_SHOP_FREE_ADS, rpc_C2S_WatchShopAddGoldAds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SDK_CREATE_ORDER, rpc_C2S_SdkCreateOrder)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_IOS_PRE_PAY, rpc_C2S_IosPrePay)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_LIMIT_GITF_BY_JADE, rpc_C2S_BuyLimitGiftByJade)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_ONE_GA_CHA_BY_JADE, rpc_C2S_BuyOneGaChaByJade)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GOOGLE_PLAY_RECHARGE, rpc_C2S_GooglePlayRecharge)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_VIP_CARD, rpc_C2S_BuyVipCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_PIECE_EXCHANGE_ITEM, rpc_C2S_PieceExchangeItem)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LOOK_OVER_LIMIT_GIFT, rpc_C2S_LookOverLimitGift)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_RECRUIT_TREASURE, rpc_C2S_BuyRecruitTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_IOS_RECHARGE, rpc_C2S_IosRecharge)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_RANDOM_SHOP_REFRESH_CNT, rpc_C2S_BuyRandomShopRefreshCnt)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_RANDOM_SHOP, rpc_C2S_BuyRandomShop)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_TEST_RECHARGE, rpc_C2S_TestRecharge)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_EXCHANGE_CARD_SKIN_IDS, rpc_C2S_FetchPieceExchangeIds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_SOLD_GOLD_GIFT, rpc_C2S_BuySoldGoldGift)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MIDAS_RECHARGE, rpc_C2S_MidasRecharge)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BUY_RECOMMEND_GIFT, rpc_C2S_BuyRecommendGift)
}
