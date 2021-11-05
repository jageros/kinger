package shop

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	//"kinger/gopuppy/common/glog"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
)


// 招募宝箱
type iRecruitTreasure interface {
	packMsg() *pb.RecruitTreasureData
	buy(buyCnt int) (*pb.BuyRecruitTreasureReply, error)
	onLogin()
	onCrossDay()
	onMaxPvpLevelUpdate()
	syncToClient()
	getDiscount() float64
	getTreasureTblName() string
}

func newRecruitTreasure(cptAttr *attribute.MapAttr, player types.IPlayer) iRecruitTreasure {
	attrKey := "soldTreasure"
	isXfServer := config.GetConfig().IsXfServer()
	if isXfServer {
		attrKey = "recruitTreasure"
	}

	attr := cptAttr.GetMapAttr(attrKey)
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr(attrKey, attr)
	}

	if isXfServer {

		t := &xfRecruitTreasureSt{}
		t.player = player
		t.attr = attr
		t.i = t
		return t

	} else {

		t := &oldRecruitTreasureSt{
			cptAttr: cptAttr,
			team: player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam(),
		}
		t.player = player
		t.attr = attr
		t.i = t
		return t

	}
}

type baseRecruitTreasureSt struct {
	attr *attribute.MapAttr
	player types.IPlayer
	i iRecruitTreasure
}

func (rt *baseRecruitTreasureSt) getBuyCnt() int {
	return rt.attr.GetInt("buyCnt")
}

func (rt *baseRecruitTreasureSt) setBuyCnt(cnt int) {
	rt.attr.SetInt("buyCnt", cnt)
}
//func (rt *baseRecruitTreasureSt) caclRecruitHintByVersion(isFetch bool) {
//	key := "recruitRewardVer"
//	verAttr := rt.attr.GetMapAttr(key)
//	if verAttr == nil {
//		verAttr = attribute.NewMapAttr()
//		rt.attr.SetMapAttr(key, verAttr)
//	}
//	rewardTbl := gamedata.GetGameData(consts.RecruitTreausreRewardTbl).(*gamedata.RewardTblGameData)
//	for _, rw := range rewardTbl.Team2Rewards {
//		for _, r := range rw.Rewards {
//			if !r.AreaLimit.IsEffective(rt.player.GetArea()) {
//				continue
//			}
//
//			ridStr := strconv.Itoa(r.ID)
//			version := verAttr.GetInt(ridStr)
//			if version != r.Version {
//				if isFetch {
//					verAttr.SetInt(ridStr, r.Version)
//				}else {
//					rt.player.UpdateHint(pb.HintType_HtRecruitReward+pb.HintType(r.ID), 1)
//				}
//			}else {
//				rt.player.DelHint(pb.HintType_HtRecruitReward+pb.HintType(r.ID))
//			}
//		}
//	}
//}

func (rt *baseRecruitTreasureSt) getTreasureData() *gamedata.RecruitTreasure {
	team := rt.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()
	area := rt.player.GetArea()
	gdata := gamedata.GetGameData(consts.RecruitTreausre)
	if gdata == nil {
		return nil
	}
	sw := crd.getRecruiteSwitchVer(area)
	if config.GetConfig().IsOldServer() {
		sw = 1
	}

	team2Treasure := gdata.(*gamedata.RecruitTreasureGameData).GetTeam2Treausre(area)
	treasures, ok := team2Treasure[team]
	if !ok {
		treasures, ok = team2Treasure[0]
		if !ok {
			return nil
		}
	}

	if treasure, ok :=  treasures[sw]; ok {
		//glog.Infof("baseRecruitTreasureSt getTreasureData %v", treasure)
		return treasure
	}
	return nil
}

func (rt *baseRecruitTreasureSt) syncToClient() {
	agent := rt.player.GetAgent()
	if agent == nil {
		return
	}

	msg := rt.i.packMsg()
	if msg == nil {
		return
	}

	agent.PushClient(pb.MessageID_S2C_UPDATE_RECRUIT_TREASURE, msg)
}

func (rt *baseRecruitTreasureSt) caclHint(isLogin bool) {
	if rt.i.getDiscount() >= 1 {
		rt.player.DelHint(pb.HintType_HtShopTreasure)
	} else if rt.getTreasureData() != nil {
		if isLogin {
			rt.player.AddHint(pb.HintType_HtShopTreasure, 1)
		} else {
			rt.player.UpdateHint(pb.HintType_HtShopTreasure, 1)
		}
	}
}

func (rt *baseRecruitTreasureSt) getTreasureTblName() string {
	treasure := rt.getTreasureData()
	if treasure == nil {
		return ""
	}
	treData, ok := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[treasure.TreasureID]
	if !ok {
		return ""
	}
	tbName := treData.RewardTbl
	return tbName
}

func (rt *baseRecruitTreasureSt) onLogin() {
	rt.caclHint(true)
	rt.caclRecruitHintByVersion(rt.getTreasureTblName())
}

func (rt *baseRecruitTreasureSt) onCrossDay() {
	rt.setBuyCnt(0)
	rt.caclHint(false)
	rt.i.syncToClient()
}

func (rt *baseRecruitTreasureSt) caclRecruitHintByVersion(tbName string) {
	var rwTypeKey string
	if tbName == consts.RecruitTreausreCardRewardTbl {
		rwTypeKey = "card"
	}else {
		rwTypeKey = "skin"
	}
	key := fmt.Sprintf("recruit_sp_sk_ver_%s", rwTypeKey)
	ver := rt.attr.GetInt(key)
	if ver == 0 {
		ver = 1
	}
	area := rt.player.GetArea()
	rwType, curIds, _ := crd.getRecruitIDs(area, tbName)
	curVer := crd.getCurVerByArea(area, rwType)
	var mid int32
	if ver != curVer {
		for _, id := range curIds {
			if id > mid {
				mid = id
			}
		}
		rt.attr.SetInt(key, curVer)
		rt.player.UpdateHint(pb.HintType_HtRecruitReward+pb.HintType(mid), 1)
	}
}

// ------------------------- 1、2服 begin ----------------------------
type oldRecruitTreasureSt struct {
	baseRecruitTreasureSt
	cptAttr *attribute.MapAttr
	team int
}

func (rt *oldRecruitTreasureSt) getTeam() int {
	return rt.team
}

func (rt *oldRecruitTreasureSt) setTeam(team int) {
	rt.team = team
}

func (rt *oldRecruitTreasureSt) getDiscount() float64 {
	buyCnt := rt.getBuyCnt()
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if buyCnt > funcPrice.ShopTreasureMaxBuyCnt {
		return 1
	}

	if buyCnt <= 0 {
		return 0.5
	}

	if buyCnt >= funcPrice.ShopTreasureMaxBuyCnt {
		return 0
	}
	return 1
}

func (rt *oldRecruitTreasureSt) packMsg() *pb.RecruitTreasureData {
	treasure := rt.getTreasureData()
	if treasure == nil {
		return nil
	}

	buyCnt := rt.getBuyCnt()
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	msg := &pb.RecruitTreasureData{
		TreasureModelID: treasure.TreasureID,
		BuyTimes: int32(buyCnt),
		Discount: int32(rt.getDiscount() * 100),
		MaxBuyTimes: int32(funcPrice.ShopTreasureMaxBuyCnt),
		NeedJade: int32(treasure.JadePrice),
	}

	if buyCnt > funcPrice.ShopTreasureMaxBuyCnt {
		msg.NexRemainTime = int32(timer.TimeDelta(0, 0, 0).Seconds())
	}

	return msg
}

func (rt *oldRecruitTreasureSt) onMaxPvpLevelUpdate() {
	maxTeam := rt.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()
	if rt.getTeam() < maxTeam {
		rt.setTeam(maxTeam)
		rt.syncToClient()
	}
}

func (rt *oldRecruitTreasureSt) buy(_ int) (*pb.BuyRecruitTreasureReply, error) {
	buyCnt := rt.getBuyCnt()
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if buyCnt > funcPrice.ShopTreasureMaxBuyCnt {
		return nil, gamedata.GameError(1)
	}

	discount := rt.getDiscount()
	treasure := rt.getTreasureData()
	if treasure == nil {
		return nil, gamedata.GameError(2)
	}

	var resType, price int
	resType = consts.Jade
	price = int(float64(treasure.JadePrice) * discount)
	if price > 0 {
		resCpt := rt.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, price) {
			return nil, gamedata.GameError(3)
		}
		resCpt.ModifyResource(consts.Jade, - price, consts.RmrRecruitTreasure)
	}

	treasureReward := rt.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
		treasure.TreasureID, false)

	buyCnt += 1
	rt.setBuyCnt(buyCnt)
	mod.LogShopBuyItem(rt.player, fmt.Sprintf("recruitTreasure_%s", treasure.TreasureID),
		fmt.Sprintf("招募宝箱_%s", treasure.TreasureID), 1, "shop",
		strconv.Itoa(resType), module.Player.GetResourceName(resType), price,
		fmt.Sprintf("discount=%f, buyCnt=%d", discount, buyCnt + 1))

	var nextRemainTime int32
	if buyCnt > funcPrice.ShopTreasureMaxBuyCnt {
		nextRemainTime = int32(timer.TimeDelta(0, 0, 0).Seconds())
	}

	rt.caclHint(false)
	return &pb.BuyRecruitTreasureReply{
		TreasureReward: treasureReward,
		Discount: int32(rt.getDiscount() * 100),
		NextRemainTime: nextRemainTime,
	}, nil
}

// ------------------------- 1、2服 end ----------------------------

// ------------------------- xf begin ----------------------------
type xfRecruitTreasureSt struct {
	baseRecruitTreasureSt
}

func (rt *xfRecruitTreasureSt) packMsg() *pb.RecruitTreasureData {
	area := rt.player.GetArea()
	treasure := rt.getTreasureData()
	if treasure == nil {
		return nil
	}
	treData, ok := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[treasure.TreasureID]
	if !ok {
		return nil
	}
	tbName := treData.RewardTbl
	rt.caclRecruitHintByVersion(tbName)
	rwType, rids, tim := crd.getRecruitIDs(area, tbName)
	rd := &pb.RecruitTreasureData{
		TreasureModelID: treasure.TreasureID,
		NeedJade: int32(treasure.JadePrice),
		NexRemainTime: int32(treasure.GetNextOpenRemainTime()),
		Discount: int32(rt.getDiscount() * 100),
		Rid: rids,
		RwType: rwType,
		NextRefreshTime: tim,
	}
	//glog.Infof("xf packMsg RecruitTreasureData %v", rd)
	return rd
}

func (rt *xfRecruitTreasureSt) getDiscount() float64 {
	buyCnt := rt.getBuyCnt()
	if buyCnt <= 0 {
		return 0.5
	}
	return 1
}

func (rt *xfRecruitTreasureSt) buy(buyCnt int) (*pb.BuyRecruitTreasureReply, error) {
	/*
	if rt.player.GetArea() != 3 && time.Now().Unix() <= 1559232005 {
		timer.AfterFunc(time.Second, func() {
			rt.player.Tellme("招募正在维护中，请0点后再试！", 0)
		})
		return nil, gamedata.GameError(10)
	}
	*/

	if rt.player.IsMultiRpcForbid(consts.FmcRecruit) {
		return nil, gamedata.GameError(10)
	}

	realCnt := buyCnt
	if buyCnt >= 10 {
		realCnt++
	}

	treasure := rt.getTreasureData()
	if treasure == nil {
		return nil, gamedata.GameError(1)
	}

	if !treasure.IsOpen() {
		rt.syncToClient()
		return nil, gamedata.GameError(2)
	}

	resType := consts.Jade
	oldDiscount := rt.getDiscount()
	var price int
	if buyCnt == 1 {
		price = int(float64(treasure.JadePrice) * oldDiscount)
	} else {
		price = treasure.JadePrice * buyCnt
	}

	if price > 0 {
		resCpt := rt.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if !resCpt.HasResource(consts.Jade, price) {
			return nil, gamedata.GameError(3)
		}
		resCpt.ModifyResource(consts.Jade, - price, consts.RmrRecruitTreasure)
	}

	reply := &pb.BuyRecruitTreasureReply{}
	ignoreRewardTbl := false
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)

	for i := 0; i < realCnt; i++ {
		treasureReward := rt.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
			treasure.TreasureID, false, ignoreRewardTbl)

		if len(treasureReward.CardSkins) > 0 {
			ignoreRewardTbl = true
		}

		if !ignoreRewardTbl && realCnt > 1 {

			for _, cardID := range treasureReward.CardIDs {
				cardData := poolGameData.GetCard(cardID, 1)
				if cardData != nil && cardData.IsSpCard() {
					if len(treasureReward.ConvertResources) <= 0 {
						module.Televise.SendNotice(pb.TeleviseEnum_RecruitGetCard, rt.player.GetName(), cardID)
					}
					ignoreRewardTbl = true
					break
				}
			}

		}

		if reply.TreasureReward == nil {
			reply.TreasureReward = treasureReward
		} else {

			reply.TreasureReward.CardIDs = append(reply.TreasureReward.CardIDs, treasureReward.CardIDs...)
			reply.TreasureReward.CardSkins = append(reply.TreasureReward.CardSkins, treasureReward.CardSkins...)
			reply.TreasureReward.EmojiTeams = append(reply.TreasureReward.EmojiTeams, treasureReward.EmojiTeams...)
			reply.TreasureReward.Headframes = append(reply.TreasureReward.Headframes, treasureReward.Headframes...)
			reply.TreasureReward.UpLevelRewardCards = append(reply.TreasureReward.UpLevelRewardCards,
				treasureReward.UpLevelRewardCards...)

		L1:
			for _, res := range treasureReward.Resources {
				for _, res2 := range reply.TreasureReward.Resources {
					if res.Type == res2.Type {
						res2.Amount += res.Amount
						continue L1
					}
				}
				reply.TreasureReward.Resources = append(reply.TreasureReward.Resources, res)
			}

		L2:
			for _, res := range treasureReward.ConvertResources {
				for _, res2 := range reply.TreasureReward.ConvertResources {
					if res.Type == res2.Type {
						res2.Amount += res.Amount
						continue L2
					}
				}
				reply.TreasureReward.ConvertResources = append(reply.TreasureReward.ConvertResources, res)
			}
		}
	}

	if buyCnt == 1 {
		rt.setBuyCnt(rt.getBuyCnt() + buyCnt)
		rt.caclHint(false)
	}
	reply.Discount = int32(rt.getDiscount() * 100)

	mod.LogShopBuyItem(rt.player, fmt.Sprintf("recruitTreasure_%s", treasure.TreasureID),
		fmt.Sprintf("招募宝箱_%s", treasure.TreasureID), buyCnt, "shop",
		strconv.Itoa(resType), module.Player.GetResourceName(resType), price, "")
	return reply, nil
}

func (rt *xfRecruitTreasureSt) syncToClient() {
	rt.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_RecruitTreasure)
}

func (rt *xfRecruitTreasureSt) onMaxPvpLevelUpdate() {}
// ------------------------- xf end ----------------------------
