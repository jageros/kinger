package shop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/td"
	"io/ioutil"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"net/http"
	"strings"
	"time"
	"github.com/gogo/protobuf/proto"
)

// 花钱，怎样买东西

type orderSt struct {
	attr *attribute.AttrMgr
}

func newOrder(cpOrderID, goodsID, currency string, player types.IPlayer) *orderSt {
	attr := attribute.NewAttrMgr("order", cpOrderID)
	attr.SetStr("goodsID", goodsID)
	attr.SetUInt64("uid", uint64(player.GetUid()))
	attr.SetStr("accountType", player.GetLogAccountType().String())
	attr.SetStr("channel", player.GetChannel())
	attr.SetStr("time", time.Now().String())
	attr.SetStr("currency", currency)
	return &orderSt{attr: attr}
}

func newOrderByAttr(attr *attribute.AttrMgr) *orderSt {
	return &orderSt{
		attr: attr,
	}
}

func (o *orderSt) String() string {
	return fmt.Sprintf("[order uid=%d, goodsID=%s, orderID=%s]", o.attr.GetUInt64("uid"),
		o.getGoodsID(), o.getCpOrderID())
}

func (o *orderSt) save(needReply bool) error {
	return o.attr.Save(needReply)
}

func (o *orderSt) getGoodsID() string {
	return o.attr.GetStr("goodsID")
}

func (o *orderSt) isComplete() bool {
	return o.attr.GetBool("isComplete")
}

func (o *orderSt) onComplete() {
	o.attr.SetBool("isComplete", true)
}

func (o *orderSt) getCpOrderID() string {
	return o.attr.GetAttrID().(string)
}

func (o *orderSt) getChannelOrderID() string {
	return o.attr.GetStr("channelOrderID")
}

func (o *orderSt) setChannelOrderID(orderID string) {
	o.attr.SetStr("channelOrderID", orderID)
}

func (o *orderSt) getCurrency() string {
	currency := o.attr.GetStr("currency")
	if currency == "" {
		return "CNY"
	}
	return currency
}

func (o *orderSt) setCurrency(currency string) {
	o.attr.SetStr("currency", currency)
}

func (o *orderSt) isTest() bool {
	return o.attr.GetBool("isTest")
}

func (o *orderSt) setIsTest() {
	o.attr.SetBool("isTest", true)
}

func (o *orderSt) setPrice(price int) {
	o.attr.SetInt("price", price)
}

func (o *orderSt) getPrice() int {
	return o.attr.GetInt("price")
}

func newIosOrder(player types.IPlayer, goodsID, receipt string) (*orderSt, error) {
	url := "https://buy.itunes.apple.com/verifyReceipt"
	payload, err := json.Marshal(map[string]string{
		"receipt-data": receipt,
	})
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	var body []byte
	evq.Await(func() {
		resp, err = http.Post(url, "", bytes.NewReader(payload))
		if err != nil {
			return
		}

		if resp.StatusCode != 200 {
			err = errors.Errorf("appStoreVerifyReceipt http.Post status %d", resp.StatusCode)
			return
		}

		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return
		}
	})

	if err != nil {
		return nil, err
	}

	reply := map[string]interface{}{}
	err = json.Unmarshal(body, &reply)
	glog.Infof("appStoreVerifyReceipt body %v", reply)
	if err != nil {
		return nil, err
	}

	status, ok := reply["status"].(float64)
	if !ok {
		return nil, errors.Errorf("appStoreVerifyReceipt status %s", status)
	}
	status2 := int(status)
	if status2 != 0 && (!config.GetConfig().AppStoreCanTest || status2 != 21007) {
		return nil, errors.Errorf("appStoreVerifyReceipt status %d", status2)
	}

	isTest := status2 == 21007
	if isTest {
		order := newOrder(fmt.Sprintf("%d%d", uint64(player.GetUid()), time.Now().UnixNano()/1000), goodsID,
			"", player)
		order.setIsTest()
		order.save(false)
		return order, nil
	}

	replyReceipt2, ok := reply["receipt"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("appStoreVerifyReceipt no replyReceipt2")
	}
	replyReceipt3, ok := replyReceipt2["in_app"].([]interface{})
	if !ok || len(replyReceipt3) <= 0 {
		return nil, errors.Errorf("appStoreVerifyReceipt no replyReceipt3")
	}

	replyReceipt, ok := replyReceipt3[0].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("appStoreVerifyReceipt no replyReceipt")
	} else {
		goodsID2, ok := replyReceipt["product_id"].(string)
		if strings.Index(goodsID2, goodsID) < 0 {
			return nil, errors.Errorf("appStoreVerifyReceipt goodsID wrong %s, need %s", goodsID2, goodsID)
		}

		transactionID, ok := replyReceipt["transaction_id"].(string)
		if !ok {
			return nil, errors.Errorf("appStoreVerifyReceipt no transactionID")
		}

		orderID := fmt.Sprintf("%d%s", uint64(player.GetUid()), transactionID)
		isExists, err := attribute.NewAttrMgr("order", orderID).Exists()
		if isExists {
			return nil, errors.Errorf("appStoreVerifyReceipt transactionID %s isExists",
				transactionID)
		}

		if err != nil {
			return nil, errors.Errorf("appStoreVerifyReceipt check transactionID %s err %s",
				transactionID)
		}

		order := newOrder(orderID, goodsID, "", player)
		order.save(false)
		return order, nil
	}
}

type iPayment interface {
	onRecharge(player types.IPlayer, order *orderSt, channel string, paymentAmount float64, needCheckMoney bool) *pb.SdkRechargeResult
}

type basePayment struct {
}

func (p *basePayment) getJadeGoods(goodsID string, player types.IPlayer) *gamedata.Recharge {
	jadeGoodsGameData := player.GetComponent(consts.ShopCpt).(*shopComponent).getJadeGoodsGameData()
	if jadeGoodsGameData != nil {
		return jadeGoodsGameData.GetJadeGoods(goodsID)
	} else {
		return nil
	}
}

func (p *basePayment) getGiftGoods(goodsID string, player types.IPlayer) iLimitGift {
	gift := player.GetComponent(consts.ShopCpt).(*shopComponent).getLimitGiftByGoodsID(goodsID)
	if gift == nil || gift.isBuy() {
		return nil
	} else {
		return gift
	}
}

func (p *basePayment) getVipGoods(goodsID string, player types.IPlayer) *gamedata.LimitGift {
	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	limitGiftGameData := shopCpt.getLimitGiftGameData()
	if limitGiftGameData == nil {
		return nil
	}

	vipCardData := limitGiftGameData.GetVipCard(player.GetArea(), !shopCpt.hasEverBuyVip())
	if vipCardData == nil || vipCardData.GiftID != goodsID {
		return nil
	}

	if player.IsForeverVip() {
		return nil
	}
	return vipCardData
}

func (p *basePayment) logMoneyNotEnough(player types.IPlayer, goodsID, cpOrderID, channelOrderID, channel,
currency string, paymentAmount float64, needMoney int) {
	glog.Errorf("onRecharge paymentAmount not enough, uid=%d, accountType=%s, channelUid=%s, cpOrderID=%s, "+
		"channelOrderID=%s, paymentAmount=%f, goodsID=%s, needPayment=%d, channel=%s, currency=%s",
		player.GetUid(), player.GetAccountType(), player.GetChannelUid(), cpOrderID, channelOrderID, paymentAmount, goodsID,
		needMoney, channel, currency)
}

func (p *basePayment) buyJade(player types.IPlayer, goodsID, cpOrderID, channelOrderID, channel, currency string,
	paymentAmount float64, needCheckMoney bool) (replyData proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType,
	jade, price int, isRightGoodsType bool, errcode pb.SdkRechargeResult_RechargeErr) {

	errcode = pb.SdkRechargeResult_Fail
	jadeGoods := p.getJadeGoods(goodsID, player)
	if jadeGoods == nil {
		return
	}

	isRightGoodsType = true
	if needCheckMoney && paymentAmount < float64(jadeGoods.Price) {
		p.logMoneyNotEnough(player, goodsID, cpOrderID, channelOrderID, channel, currency, paymentAmount,
			jadeGoods.Price)
		return
	}

	jade = jadeGoods.JadeCnt
	rewardJade := module.Huodong.OnRecharge(player, jade, int(paymentAmount))

	shopCpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
	if shopCpt.getJadeIsDouble(goodsID) {
		jade = jadeGoods.FirstJadeCnt
	}

	player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).ModifyResource(consts.Jade,
		jade, consts.RmrRecharge)

	shopCpt.setJadeIsDouble(goodsID)

	replyData = &pb.BuyJadeReply{Jade: int32(jade), RewardJade: int32(rewardJade)}
	errcode = pb.SdkRechargeResult_Success
	goodsType = pb.SdkRechargeResult_Jade
	price = jadeGoods.Price

	shopCpt.onShopDataUpdate(pb.UpdateShopDataArg_Jade)
	return
}

func (p *basePayment) buyGift(player types.IPlayer, goodsID, cpOrderID, channelOrderID, channel, currency string,
	paymentAmount float64, needCheckMoney bool) (replyData proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType,
	isRightGoodsType bool, price int, errcode pb.SdkRechargeResult_RechargeErr) {

	errcode = pb.SdkRechargeResult_Fail
	gift := p.getGiftGoods(goodsID, player)
	if gift == nil {
		return
	}

	isRightGoodsType = true
	giftData := gift.getGameData()
	if giftData == nil {
		return
	}
	
	if needCheckMoney && paymentAmount < float64(giftData.Price) {
		p.logMoneyNotEnough(player, goodsID, cpOrderID, channelOrderID, channel, currency, paymentAmount,
			giftData.Price)
		return
	}

	replyData, goodsType = gift.buy()

	errcode = pb.SdkRechargeResult_Success
	price = giftData.Price

	itemID := fmt.Sprintf("%s_%s_money", giftData.GiftID, giftData.Reward)
	mod.LogShopBuyItem(player, itemID, itemID, 1, "shop", currency,
		currency, price, "")

	if giftData.GiftID == "gift12"{
		module.Televise.SendNotice(pb.TeleviseEnum_LimitGiftGetCard, player.GetName(), uint32(109))
	}

	return
}

func (p *basePayment) buyVip(player types.IPlayer, goodsID, cpOrderID, channelOrderID, channel, currency string,
	paymentAmount float64, needCheckMoney bool) (replyData proto.Marshaler, goodsType pb.SdkRechargeResult_GoodsType,
	isRightGoodsType bool, price int, errcode pb.SdkRechargeResult_RechargeErr) {

	errcode = pb.SdkRechargeResult_Fail
	vipCardData := p.getVipGoods(goodsID, player)
	if vipCardData == nil {
		return
	}

	isRightGoodsType = true
	if needCheckMoney && paymentAmount < float64(vipCardData.Price) {
		p.logMoneyNotEnough(player, goodsID, cpOrderID, channelOrderID, channel, currency, paymentAmount,
			vipCardData.Price)
		return
	}

	vipSt := module.OutStatus.GetStatus(player, consts.OtVipCard)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if vipSt == nil {
		vipSt = module.OutStatus.AddStatus(player, consts.OtVipCard, funcPrice.VipContinuedTime)
	} else {
		vipSt.Over(funcPrice.VipContinuedTime)
	}

	headFrame := vipCardData.HeadFrame
	if headFrame != "" {
		if !module.Bag.HasItem(player, consts.ItHeadFrame, headFrame) {
			module.Bag.AddHeadFrame(player, vipCardData.HeadFrame)
		} else {
			headFrame = ""
		}
	}

	var remainTime int32 = -1
	if vipSt != nil {
		remainTime = int32(vipSt.GetRemainTime())
	}
	replyData = &pb.BuyVipCardReply{
		HeadFrame:  headFrame,
		RemainTime: remainTime,
	}

	errcode = pb.SdkRechargeResult_Success
	price = vipCardData.Price

	goodsType = pb.SdkRechargeResult_Vip
	goodsName := vipCardData.GiftID + "_money"
	mod.LogShopBuyItem(player, goodsName, goodsName, 1, "shop",
		currency, currency, price, "")
	if vipCardData.IsFirstVip() {
		player.GetComponent(consts.ShopCpt).(*shopComponent).onBuyVip(vipCardData.GiftID)
	}

	module.Televise.SendNotice(pb.TeleviseEnum_BuyVip, player.GetName())

	return
}

func (p *basePayment) onRecharge(player types.IPlayer, order *orderSt, channel string, paymentAmount float64,
	needCheckMoney bool) *pb.SdkRechargeResult {

	channelUid := player.GetChannelUid()
	accountType := player.GetLogAccountType()
	reply := &pb.SdkRechargeResult{Errcode: pb.SdkRechargeResult_Fail}
	cpOrderID, channelOrderID := order.getCpOrderID(), order.getChannelOrderID()
	var isRightGoodsType bool
	var jade, price int
	var replyData proto.Marshaler
	var goodsType pb.SdkRechargeResult_GoodsType
	currency := order.getCurrency()
	goodsID := order.getGoodsID()

	replyData, goodsType, jade, price, isRightGoodsType, reply.Errcode = p.buyJade(player, goodsID, cpOrderID,
		channelOrderID, channel, currency, paymentAmount, needCheckMoney)
	if isRightGoodsType && reply.Errcode != pb.SdkRechargeResult_Success {
		return reply
	}

	if !isRightGoodsType {
		replyData, goodsType, isRightGoodsType, price, reply.Errcode = p.buyVip(player, goodsID, cpOrderID,
			channelOrderID, channel, currency, paymentAmount, needCheckMoney)
		if isRightGoodsType && reply.Errcode != pb.SdkRechargeResult_Success {
			return reply
		}
	}

	if !isRightGoodsType {
		replyData, goodsType, isRightGoodsType, price, reply.Errcode = p.buyGift(player, goodsID, cpOrderID,
			channelOrderID, channel, currency, paymentAmount, needCheckMoney)
		if isRightGoodsType && reply.Errcode != pb.SdkRechargeResult_Success {
			return reply
		}
	}

	if isRightGoodsType {
		order.onComplete()
		if paymentAmount <= 0 {
			paymentAmount = float64(price)
		}
		order.setPrice(int(paymentAmount))
		order.save(false)
		if !order.isTest() {
			os := "ios"
			if player.GetAccountType() != pb.AccountTypeEnum_Ios {
				os = "android"
			}
			td.OnPay(os, config.GetConfig().GetChannelTdkey(channel), player.GetServerID(), channel, player.GetUid(),
				player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel(), cpOrderID, paymentAmount,
				int(jade), currency)
		}

		player.GetComponent(consts.ShopCpt).(*shopComponent).addCumulativePay(int(paymentAmount))
		eventhub.Publish(consts.EvRecharge, player, price, goodsID)

		if order.isTest() {
			glog.Infof("onRecharge ok, uid=%d, accountType=%s, channelUid=%s, cpOrderID=%s, channelOrderID=%s, "+
				"paymentAmount=%f, goodsID=%s, jade=%d, isTest=true, channel=%s, currency=%s", player.GetUid(),
				accountType, channelUid, cpOrderID, channelOrderID, paymentAmount, goodsID, jade, channel, currency)
		} else {
			glog.JsonInfo("pay", glog.Uint64("uid", uint64(player.GetUid())), glog.String("accountType",
				accountType.String()), glog.String("channelID", channelUid), glog.String("cpOrderID", cpOrderID),
				glog.String("channelOrderID", channelOrderID), glog.Float64("money", paymentAmount), glog.String(
				"goodsID", goodsID), glog.Int("jade", jade), glog.String("channel", channel), glog.String(
				"currency", currency), glog.Int("pvpLevel", player.GetPvpLevel()), glog.Int("area", player.GetArea()),
				glog.String("subChannel", player.GetSubChannel()))
		}

	} else {
		glog.Errorf("onRecharge no goods, uid=%d, accountType=%s, channelUid=%s, cpOrderID=%s, channelOrderID=%s, "+
			"paymentAmount=%f, goodsID=%s, channel=%s, currency=%s", player.GetUid(), accountType, channelUid,
			cpOrderID, channelOrderID, paymentAmount, goodsID, channel, currency)
	}

	reply.Type = goodsType
	reply.Data, _ = replyData.Marshal()
	player.Save(false)
	return reply
}
