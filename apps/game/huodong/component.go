package huodong

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
)

var _ types.IPlayerComponent = &huodongComponent{}
var _ htypes.IHuodongComponent = &huodongComponent{}

type huodongComponent struct {
	attr   *attribute.MapAttr
	player types.IPlayer
	id2Hd  map[pb.HuodongTypeEnum]htypes.IHdPlayerData
}

func (hc *huodongComponent) ComponentID() string {
	return consts.HuodongCpt
}

func (hc *huodongComponent) GetPlayer() types.IPlayer {
	return hc.player
}

func (hc *huodongComponent) OnInit(player types.IPlayer) {
	hc.player = player
}

func (hc *huodongComponent) OnLogin(isRelogin, isRestore bool) {
	if isRelogin {
		return
	}

	hc.id2Hd = map[pb.HuodongTypeEnum]htypes.IHdPlayerData{}
	hdDatasAttr := hc.attr.GetMapAttr("data")
	if hdDatasAttr == nil {
		hdDatasAttr = attribute.NewMapAttr()
		hc.attr.SetMapAttr("data", hdDatasAttr)
	}
	hdDatasAttr.ForEachKey(func(key string) {
		htype, _ := strconv.Atoi(key)
		htype2 := pb.HuodongTypeEnum(htype)
		hd := mod.GetHuodong(hc.player.GetArea(), htype2)
		if hd == nil {
			glog.Errorf("huodongComponent OnLogin newPlayerDataByAttr no huodong %s", htype2)
			return
		}
		hc.id2Hd[htype2] = hd.NewPlayerDataByAttr(hc.player, hdDatasAttr.GetMapAttr(key))
	})

	for id, _ := range pb.HuodongTypeEnum_name {
		htype := pb.HuodongTypeEnum(id)
		if htype == pb.HuodongTypeEnum_HUnknow {
			continue
		}

		hd := mod.GetHuodong(hc.player.GetArea(), htype)
		if hd == nil {
			glog.Errorf("huodongComponent OnLogin no huodong %d", htype)
			continue
		}
		hdData := hc.GetOrNewHdData(htype)
		if hdData != nil {
			hd.OnPlayerLogin(hc.player, hdData)
		}
	}
}

func (hc *huodongComponent) GetOrNewHdData(htype pb.HuodongTypeEnum) htypes.IHdPlayerData {
	hd := mod.GetHuodong(hc.player.GetArea(), htype)
	if hd == nil {
		return nil
	}

	hdData, ok := hc.id2Hd[htype]
	if !ok && hd.IsOpen() {
		hdData = hd.NewPlayerData(hc.player)
		hc.id2Hd[htype] = hdData
		hdDatasAttr := hc.attr.GetMapAttr("data")
		hdDatasAttr.SetMapAttr(strconv.Itoa(int(htype)), hdData.GetAttr())
	}
	return hdData
}

func (hc *huodongComponent) OnLogout() {

}

func (hc *huodongComponent) onChristmasRecharge(oldJade, money int) int {
	if money >= 30 {
		headFrame := "11"
		if !module.Bag.HasItem(hc.player, consts.ItHeadFrame, headFrame) {
			sender := module.Mail.NewMailSender(hc.player.GetUid())
			sender.SetTypeAndArgs(pb.MailTypeEnum_SpringRecharge)
			mailReward := sender.GetRewardObj()
			mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, headFrame, 1)
			sender.Send()
		}
	}

	if hc.attr.GetBool(htypes.RechargeHdKey) {
		return 0
	}

	hc.attr.SetBool(htypes.RechargeHdKey, true)
	glog.Infof("onChristmasRecharge %s uid=%d", htypes.RechargeHdKey, hc.player.GetUid())

	jade := int(float64(oldJade) * 0.5)
	sender := module.Mail.NewMailSender(hc.player.GetUid())
	sender.SetTypeAndArgs(pb.MailTypeEnum_YuandanReward)
	mailReward := sender.GetRewardObj()
	mailReward.AddAmountByType(pb.MailRewardType_MrtJade, jade)
	sender.Send()
	return jade
}
