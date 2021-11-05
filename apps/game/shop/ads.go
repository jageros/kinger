package shop

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"time"
	"kinger/gamedata"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/common/glog"
	"kinger/common/consts"
	"kinger/common/config"
	"math"
	"kinger/gopuppy/common/timer"
)

// 商城赞助
type freeAds struct {
	attr *attribute.MapAttr
}

func newFreeAdsByAttr(attr *attribute.MapAttr) *freeAds {
	fa := &freeAds{
		attr: attr,
	}
	if fa.getTimeout() <= 0 {
		attr.SetInt64("timeout", time.Now().Unix())
	}
	return fa
}

func newFreeAds(player types.IPlayer, type_ pb.ShopFreeAdsType, isNewbie bool) *freeAds {
	team := player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()
	var adata gamedata.IFreeShopAds = nil
	if type_ == pb.ShopFreeAdsType_GoldAds {
		adata = gamedata.GetGameData(consts.FreeGoldAds).(*gamedata.FreeGoldAdsGameData).RandomAdsByTeam(team)
	} else if type_ == pb.ShopFreeAdsType_JadeAds {
		adata = gamedata.GetGameData(consts.FreeGoodTreasureAds).(*gamedata.FreeGoodTreasureAdsGameData).RandomAdsByTeam(team)
	} else {
		adata = gamedata.GetGameData(consts.FreeTreasureAds).(*gamedata.FreeTreasureAdsGameData).RandomAdsByTeam(team)
	}
	if adata == nil {
		return nil
	}

	attr := attribute.NewMapAttr()
	attr.SetInt32("type", int32(type_))
	attr.SetInt("id", adata.GetID())
	timeout := time.Now().Unix()
	if !isNewbie {
		timeout += int64(adata.GetTime())
	}
	attr.SetInt64("timeout", timeout)
	return &freeAds{
		attr: attr,
	}
}

func (fa *freeAds) packMsg() *pb.ShopFreeAds {
	msg := &pb.ShopFreeAds{
		Type: fa.getType(),
		ID: int32(fa.getID()),
		CanGet: fa.isWxCanReward(),
	}
	remainTime := int32(fa.getTimeout() - time.Now().Unix())
	if remainTime < 0 {
		remainTime = 0
	}
	msg.RemainTime = remainTime
	return msg
}

func (fa *freeAds) String() string {
	return fmt.Sprintf("[shop ads type=%s, id=%d, timeout=%d]", fa.getType(), fa.getID(), fa.getTimeout())
}

func (fa *freeAds) getType() pb.ShopFreeAdsType {
	return pb.ShopFreeAdsType(fa.attr.GetInt32("type"))
}

func (fa *freeAds) getID() int {
	return fa.attr.GetInt("id")
}

func (fa *freeAds) getTimeout() int64 {
	return fa.attr.GetInt64("timeout")
}

func (fa *freeAds) isTimeToReward() bool {
	timeout := fa.getTimeout()
	if timeout < 0 || time.Now().Unix() < timeout {
		return false
	}
	return true
}

func (fa *freeAds) isWxCanReward() bool {
	return fa.attr.GetBool("canReward")
}

func (fa *freeAds) setWxCanReward(can bool) {
	fa.attr.SetBool("canReward", can)
}

func (fa *freeAds) getReward(player types.IPlayer, isConsumeJade bool) ([]byte, error) {
	if !fa.isTimeToReward() {
		return nil, gamedata.GameError(1)
	}

	var needJade int
	if isConsumeJade {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		needJade = player.GetSkipAdsNeedJade()
		if !resCpt.HasResource(consts.Jade, needJade) {
			return nil, gamedata.GameError(2)
		}
		resCpt.ModifyResource(consts.Jade, -needJade, consts.RmrUnknownConsume)
	}

	type_ := fa.getType()
	id := fa.getID()
	glog.Infof("shop ads get reward, uid=%d, type=%s, id=%d, isConsumeJade=%v, needJade=%d", player.GetUid(),
		type_, id, isConsumeJade, needJade)
	fa.attr.SetInt64("timeout", 0)
	fa.setWxCanReward(false)

	var payload []byte
	if type_ == pb.ShopFreeAdsType_GoldAds {
		data := gamedata.GetGameData(consts.FreeGoldAds).(*gamedata.FreeGoldAdsGameData).ID2Ads[id]
		player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).ModifyResource(consts.Gold, data.Gold, consts.RmrShopAds)
		payload, _ = (&pb.WatchShopFreeAdsReply_GoldReward{
			GoldAmount: int32(data.Gold),
		}).Marshal()
	//} else if type_ == pb.ShopFreeAdsType_JadeAds {
	//	data := gamedata.GetGameData(consts.FreeJddeAds).(*gamedata.FreeJadeAdsGameData).ID2Ads[id]
	//	player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).ModifyResource(consts.Jade, data.Jade, true)
	//	payload, _ = (&pb.WatchShopFreeAdsReply_JadeReward{
	//		JadeAmount: int32(data.Jade),
	//	}).Marshal()
	} else {
		var data *gamedata.FreeTreasureAds
		if type_ == pb.ShopFreeAdsType_JadeAds {
			data = gamedata.GetGameData(consts.FreeGoodTreasureAds).(*gamedata.FreeGoodTreasureAdsGameData).ID2Ads[id]
		} else {
			data = gamedata.GetGameData(consts.FreeTreasureAds).(*gamedata.FreeTreasureAdsGameData).ID2Ads[id]
		}
		reward := player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(data.TreasureModelID, false)
		payload, _ = reward.Marshal()
	}

	return payload, nil
}

type iAdsMgr interface {
	onLogin()
	onLogout()
	onMaxPvpLevelUpdate()
	watchAds(type_ pb.ShopFreeAdsType, id int, isConsumeJade bool) (*pb.WatchShopFreeAdsReply, error)
	onWxBeHelp(type_ pb.ShopFreeAdsType, id int) error
	caclHint(isLogin bool)
	forEachAds(callback func(a *freeAds))
}

type nilAdsMgrSt struct {
}
func (am *nilAdsMgrSt) onLogin()                             {}
func (am *nilAdsMgrSt) onLogout()                            {}
func (am *nilAdsMgrSt) onMaxPvpLevelUpdate()                 {}
func (am *nilAdsMgrSt) caclHint(isLogin bool)                {}
func (am *nilAdsMgrSt) forEachAds(callback func(a *freeAds)) {}

func (am *nilAdsMgrSt) watchAds(type_ pb.ShopFreeAdsType, id int, isConsumeJade bool) (*pb.WatchShopFreeAdsReply, error) {
	return nil, gamedata.InternalErr
}

func (am *nilAdsMgrSt) onWxBeHelp(type_ pb.ShopFreeAdsType, id int) error {
	return gamedata.InternalErr
}


type adsMgrSt struct {
	attr *attribute.ListAttr
	player types.IPlayer
	adses []*freeAds
	notifyTimer *timer.Timer
}

var nilAdsMgr = &nilAdsMgrSt{}
func newAdsMgr(cptAttr *attribute.MapAttr, player types.IPlayer) iAdsMgr {
	if config.GetConfig().IsXfServer() {
		// 仙峰
		return nilAdsMgr
	}

	attr := cptAttr.GetListAttr("ads")
	if attr == nil {
		attr = attribute.NewListAttr()
		cptAttr.SetListAttr("ads", attr)
	}

	mgr := &adsMgrSt{attr: attr, player: player}
	attr.ForEachIndex(func(index int) bool {
		mgr.adses = append(mgr.adses, newFreeAdsByAttr(attr.GetMapAttr(index)))
		return true
	})
	return mgr
}

func (am *adsMgrSt) onLogin() {
	am.caclHint(true)
	am.addNotifyTimer()
}

func (am *adsMgrSt) onLogout() {
	if am.notifyTimer != nil {
		am.notifyTimer.Cancel()
		am.notifyTimer = nil
	}
}

func (am *adsMgrSt) addNotifyTimer() {
	nextTime := int64(math.MaxInt64)
	for _, a := range am.adses {
		t := a.getTimeout()
		if t > 0 && t < nextTime {
			nextTime = t
		}
	}
	if nextTime == int64(math.MaxInt64) {
		return
	}

	now := time.Now().Unix()
	remainTime := nextTime - now
	if remainTime <= 0 {
		return
	}

	if am.notifyTimer != nil {
		am.notifyTimer.Cancel()
	}

	am.notifyTimer = timer.AfterFunc(time.Duration(remainTime + 1) * time.Second, func() {
		am.notifyTimer = nil
		am.caclHint(false)
		am.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_FreeAds)
	})
}

func (am *adsMgrSt) onMaxPvpLevelUpdate() {
	if len(am.adses) > 0 {
		return
	}

	a := newFreeAds(am.player, pb.ShopFreeAdsType_GoldAds, true)
	if a != nil {
		am.adses = append(am.adses, a)
		am.attr.AppendMapAttr(a.attr)
	}

	a = newFreeAds(am.player, pb.ShopFreeAdsType_JadeAds, true)
	if a != nil {
		am.adses = append(am.adses, a)
		am.attr.AppendMapAttr(a.attr)
	}

	a = newFreeAds(am.player, pb.ShopFreeAdsType_TreasureAds, true)
	if a != nil {
		am.adses = append(am.adses, a)
		am.attr.AppendMapAttr(a.attr)
	}

	am.caclHint(false)

	glog.Infof("add shop ads uid=%d, ads=%v", am.player.GetUid(), am.adses)
}

func (am *adsMgrSt) watchAds(type_ pb.ShopFreeAdsType, id int, isConsumeJade bool) (*pb.WatchShopFreeAdsReply, error) {
	var a *freeAds
	var idx int
	for i, a2 := range am.adses {
		if a2.getType() == type_ && a2.getID() == id {
			a = a2
			idx = i
			break
		}
	}
	if a == nil {
		return nil, gamedata.GameError(2)
	}

	rewardPayload, err := a.getReward(am.player, isConsumeJade)
	if err != nil {
		return nil, err
	}

	newa := newFreeAds(am.player, type_, false)
	if newa == nil {
		return nil, gamedata.GameError(3)
	}
	am.adses[idx] = newa
	am.attr.SetMapAttr(idx, newa.attr)
	am.addNotifyTimer()
	am.caclHint(false)

	return &pb.WatchShopFreeAdsReply{
		Type: type_,
		RewardPayload: rewardPayload,
		NextAds: newa.packMsg(),
	}, nil
}

func (am *adsMgrSt) onWxBeHelp(type_ pb.ShopFreeAdsType, id int) error {
	var a *freeAds
	for _, a2 := range am.adses {
		if a2.getType() == type_ && a2.getID() == id {
			a = a2
			break
		}
	}
	if a == nil {
		return gamedata.GameError(1)
	}

	if !a.isTimeToReward() || a.isWxCanReward() {
		return gamedata.GameError(2)
	}

	a.setWxCanReward(true)
	return nil
}

func (am *adsMgrSt) forEachAds(callback func(a *freeAds)) {
	for _, a := range am.adses {
		callback(a)
	}
}

func (am *adsMgrSt) caclHint(isLogin bool) {
	var count int
	for _, ads := range am.adses {
		if ads.isTimeToReward() {
			count++
		}
	}

	if count > 0 {
		if isLogin {
			am.player.AddHint(pb.HintType_HtFreeAds, count)
		} else {
			am.player.UpdateHint(pb.HintType_HtFreeAds, count)
		}
	} else {
		am.player.DelHint(pb.HintType_HtFreeAds)
	}
}
