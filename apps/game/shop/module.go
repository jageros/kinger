package shop

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"time"
)

var (
	mod                    *shopModule
	googlePlayPublicKey    *rsa.PublicKey
	strGooglePlayPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgN8kCwQMXe59knh6SzdL" +
		"56gl2hMYH9BudVqQmvzqFkKFNYD2Fe7MLJ6Zj4+2WlJsLAs/8kDvHduYsxlPy8dr" +
		"dE8og6PicoYF3LZcwOfE1FiidrW2cWbtvaznO5MX9mCyEdsnqDy699uD7rYPyut7" +
		"HnMps8DMhSAucBDJ1eNFLg93/m35Tev6u3EzsXlnmJGTC29L723Tbznw1vKd+3r1" +
		"k8FyWa8RlmpnEhvsuURur7c2AB7JfBTXOynzBR6Qq8I04Bpcxf6qZPHUk5N9ifHw" +
		"QMHuNZpyPW5YO3/3tBSGjTgfImGb+CaIGKoKcMP+CIHF64NFqClJQge0mNvBgWsYLwIDAQAB"
)

type shopModule struct {
	type2Shop map[string]*shopSt
	payment   *basePayment
}

func newShopModule() *shopModule {
	m := &shopModule{
		type2Shop: map[string]*shopSt{},
		payment:   &basePayment{},
	}

	s := newCardPieceShop()
	m.type2Shop[s.type_] = s
	s = newSkinPieceShop()
	m.type2Shop[s.type_] = s
	return m
}

func (sm *shopModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("shop")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("shop", attr)
	}
	return &shopComponent{attr: attr}
}

func (sm *shopModule) getShop(shopName string) *shopSt {
	return sm.type2Shop[shopName]
}

func (sm *shopModule) GetVipRemainTime(player types.IPlayer) int {
	if player.IsForeverVip() {
		return -1
	}
	st := module.OutStatus.GetStatus(player, consts.OtVipCard)
	if st != nil {
		return st.GetRemainTime()
	} else {
		return 0
	}
}

func (sm *shopModule) LogShopBuyItem(player types.IPlayer, itemID, itemName string, amount int, shopName, resType,
	resName string, resAmount int, msg string) {
	glog.JsonInfo("shop", glog.Uint64("uid", uint64(player.GetUid())), glog.String("itemID", itemID),
		glog.String("itemName", itemName), glog.Int("amount", amount), glog.String("shopName", shopName),
		glog.String("logMsg", msg), glog.String("accountType", player.GetLogAccountType().String()), glog.String(
			"channel", player.GetChannel()), glog.String("resType", resType), glog.String("resName", resName),
		glog.Int("resAmount", resAmount), glog.Int("area", player.GetArea()), glog.String("subChannel", player.GetSubChannel()))
}

func (sm *shopModule) GetLimitGiftPrice(goodsID string, player types.IPlayer) int {
	gdata := player.GetComponent(consts.ShopCpt).(*shopComponent).getLimitGiftGameData()
	if gdata == nil {
		glog.Errorf("shoModule GetLimitGiftPrice error1, uid=%d, goosID=%s", player.GetUid(), goodsID)
		return 0
	}
	gift := gdata.GetGiftByID(goodsID)
	if gift == nil {
		glog.Errorf("shoModule GetLimitGiftPrice error2, uid=%d, goosID=%s", player.GetUid(), goodsID)
		return 0
	}
	return gift.Price
}

func (sm *shopModule) GetRecruitCurIDs(player types.IPlayer) []int32 {
	p, ok := player.GetComponent(consts.ShopCpt).(*shopComponent)
	if !ok {
		return nil
	}
	tbName := p.getRecruitTreasure().getTreasureTblName()
	_, ids, _ := crd.getRecruitIDs(player.GetArea(), tbName)
	return ids
}

func (s *shopModule) GM_setRecruitVer(cmd string, player types.IPlayer) {
	if cmd == "reset" {
		for _, cycRew := range crd.area2CycleReward {
			cycRew.setCurCardVer(0)
			cycRew.setCurSkinVer(0)
			cycRew.setMaxExchangeCardVer(0)
			cycRew.setMaxExchangeSkinVer(0)
			cycRew.setSwitchVer(0)
			cycRew.setCurWeeks(0)
			cycRew.save()
			cycRew.pushRecruitIdsToClientOnceArea()
		}
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			p := player.GetComponent(consts.ShopCpt).(*shopComponent)
			p.recruitTreasure.syncToClient()
		})
		glog.Infof("player uid=%d reset recruit all ver!", player.GetUid())
		return
	}

	if cmd == "crossweek" {
		glog.Infof("player uid=%d set recruit cross week!", player.GetUid())
		crd.crossWeek()
	}
}

func onMaxPvpLevelUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	maxPvpLevel := args[2].(int)
	player.GetComponent(consts.ShopCpt).(*shopComponent).onMaxPvpLevelUpdate(maxPvpLevel)
}

func onPvpLevelUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	player.GetComponent(consts.ShopCpt).(*shopComponent).getSoldTreasure().onPvpLevelUpdate()
}

func heartBeat() {
	now := time.Now()
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		player.GetComponent(consts.ShopCpt).(*shopComponent).heartBeat(now)
	})
}

func onReborn(args ...interface{}) {
	player := args[0].(types.IPlayer)
	player.GetComponent(consts.ShopCpt).(*shopComponent).onReborn()
}

func genShopDataUpdater(type_ pb.UpdateShopDataArg_DataType) func(gamedata.IGameData) {
	return func(_ gamedata.IGameData) {
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(type_)
		})
	}
}

func genRecruitDataUpdate(type_ pb.UpdateShopDataArg_DataType) func(gamedata.IGameData) {
	return func(_ gamedata.IGameData) {
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			p := player.GetComponent(consts.ShopCpt).(*shopComponent)
			p.onShopDataUpdate(type_)
			p.pushExchangeIdToClient()
		})
	}
}

func genAccountTypeShopDataUpdater(types_ []pb.UpdateShopDataArg_DataType, accountTypes []pb.AccountTypeEnum) func(gamedata.IGameData) {
	return func(_ gamedata.IGameData) {
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			accountType := player.GetAccountType()
			isRightAccountType := false
			for _, at := range accountTypes {
				if at == accountType {
					isRightAccountType = true
					break
				}
			}

			if !isRightAccountType {
				return
			}

			cpt := player.GetComponent(consts.ShopCpt).(*shopComponent)
			for _, t := range types_ {
				cpt.onShopDataUpdate(t)
			}
		})
	}
}

func Initialize() {
	mod = newShopModule()
	module.Shop = mod
	registerRpc()
	eventhub.Subscribe(consts.EvMaxPvpLevelUpdate, onMaxPvpLevelUpdate)
	eventhub.Subscribe(consts.EvPvpLevelUpdate, onPvpLevelUpdate)
	eventhub.Subscribe(consts.EvReborn, onReborn)
	timer.AddTicker(5*time.Minute, heartBeat)
	crdInitialized()
	initRandShop()

	gamedata.GetGameData(consts.IosLimitGift).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_LimitGift, pb.UpdateShopDataArg_Vip},
		[]pb.AccountTypeEnum{pb.AccountTypeEnum_Ios}))
	gamedata.GetGameData(consts.AndroidLimitGift).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_LimitGift, pb.UpdateShopDataArg_Vip},
		[]pb.AccountTypeEnum{pb.AccountTypeEnum_Android}))
	gamedata.GetGameData(consts.WxLimitGift).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_LimitGift, pb.UpdateShopDataArg_Vip},
		[]pb.AccountTypeEnum{pb.AccountTypeEnum_Wxgame, pb.AccountTypeEnum_WxgameIos}))
	gamedata.GetGameData(consts.IosRecharge).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_Jade}, []pb.AccountTypeEnum{pb.AccountTypeEnum_Ios}))
	gamedata.GetGameData(consts.AndroidRecharge).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_Jade}, []pb.AccountTypeEnum{pb.AccountTypeEnum_Android}))
	gamedata.GetGameData(consts.WxRecharge).AddReloadCallback(genAccountTypeShopDataUpdater(
		[]pb.UpdateShopDataArg_DataType{pb.UpdateShopDataArg_Jade},
		[]pb.AccountTypeEnum{pb.AccountTypeEnum_Wxgame, pb.AccountTypeEnum_WxgameIos}))

	gamedata.GetSoldTreasureGameData().AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_SoldTreasure))
	gamedata.GetGameData(consts.RecruitTreausre).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_RecruitTreasure))
	gamedata.GetSoldGoldGameData().AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_Gold))
	gamedata.GetGameData(consts.FreeGoldAds).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_FreeAds))
	gamedata.GetGameData(consts.FreeGoodTreasureAds).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_FreeAds))
	gamedata.GetGameData(consts.FreeTreasureAds).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_FreeAds))
	gamedata.GetGameData(consts.RandomShop).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_RandomShop))
	gamedata.GetGameData(consts.SoldGoldGift).AddReloadCallback(genShopDataUpdater(pb.UpdateShopDataArg_GoldGift))
	gamedata.GetGameData(consts.RecruitTreausreCardRewardTbl).AddReloadCallback(genRecruitDataUpdate(pb.UpdateShopDataArg_RecruitTreasure))
	gamedata.GetGameData(consts.RecruitTreausreSkinRewardTbl).AddReloadCallback(genRecruitDataUpdate(pb.UpdateShopDataArg_RecruitTreasure))

	decodePublic, err := base64.StdEncoding.DecodeString(strGooglePlayPublicKey)
	if err != nil {
		glog.Errorf("GooglePlayPublicKey DecodeString err %s", err)
		return
	}
	pubInterface, err := x509.ParsePKIXPublicKey(decodePublic)
	if err != nil {
		glog.Errorf("GooglePlayPublicKey ParsePKIXPublicKey err %s", err)
		return
	}
	googlePlayPublicKey = pubInterface.(*rsa.PublicKey)
}
