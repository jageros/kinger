package outstatus

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
)

var mod *outStatusModule

type outStatusModule struct {
}

func (m *outStatusModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("outstatus")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("outstatus", attr)
	}
	return &outstatusComponent{attr: attr}
}

func (m *outStatusModule) GetStatus(player types.IPlayer, statusID string) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).getStatus(statusID)
}

func (m *outStatusModule) GetBuff(player types.IPlayer, buffID int) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).getStatus(
		fmt.Sprintf("%s%d", consts.OtBuffPrefix, buffID))
}

func (m *outStatusModule) GetForbidStatus(player types.IPlayer, forbidID int) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).getStatus(
		fmt.Sprintf("%s%d", consts.OtForbid, forbidID))
}

func (m *outStatusModule) AddStatus(player types.IPlayer, statusID string, leftTime int, args ...interface{}) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).addStatus(statusID, leftTime, args...)
}

func (m *outStatusModule) AddBuff(player types.IPlayer, buffID int, leftTime int, args ...interface{}) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).addStatus(
		fmt.Sprintf("%s%d", consts.OtBuffPrefix, buffID), leftTime, args...)
}

func (m *outStatusModule) AddForbidStatus(player types.IPlayer, forbidID int, leftTime int, args ...interface{}) types.IOutStatus {
	return player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).addStatus(
		fmt.Sprintf("%s%d", consts.OtForbid, forbidID), leftTime, args...)
}

func (m *outStatusModule) DelForbidStatus(player types.IPlayer, forbidID int){
	player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).delStatus(fmt.Sprintf("%s%d", consts.OtForbid, forbidID))
}

func (m *outStatusModule) ForEachClientStatus(player types.IPlayer, callback func(st types.IOutStatus)) {
	player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).forEachClientStatus(func(st iClientOutStatus) {
		callback(st)
	})
}

func (m *outStatusModule) DelStatus(player types.IPlayer, statusID string) {
	player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).delStatus(statusID)
}

func (m *outStatusModule) BuffTreasureCard(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffTreasureCard(amount2, pcm.getBuffEffect(consts.PrivTreasureCard))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffTreasureGold(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffTreasureGold(amount2, pcm.getBuffEffect(consts.PrivTreasureGold))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffTreasureCnt(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffTreasureCnt(amount2, pcm.getBuffEffect(consts.PrivTreasureCnt))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffAccTreasureCnt(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffAccTreasureCnt(amount2, pcm.getBuffEffect(consts.PrivAccTreasureCnt))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffAccTreasureCntByActivity(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffAccTreasureCntByActivity(amount2, pcm.getBuffEffect(consts.PrivAccTreasureCnt))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffDayTreasureCard(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffDayTreasureCard(amount2, pcm.getBuffEffect(consts.PrivDayTreasureCard))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffTreasureTime(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffTreasureTime(amount2, pcm.getBuffEffect(consts.PrivTreasureTime))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffAddPvpGold(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffAddPvpGold(amount2, pcm.getBuffEffect(consts.PrivAddPvpGold))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffPvpAddStar(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffPvpAddStar(amount2, pcm.getBuffEffect(consts.PrivPvpAddStar))
	})
	return int(amount2)
}

func (m *outStatusModule) BuffPvpNoSubStar(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffPvpNoSubStar(amount2, pcm.getBuffEffect(consts.PrivPvpNoSubStar))
	})
	return int(amount2)
}

//pvp加一个宝箱
func (m *outStatusModule)BuffPvpAddTreasure(player types.IPlayer, amount int) int{
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffPvpAddTreasure(amount2, pcm.getBuffEffect(consts.PrivPvpAddTreasure))
	})
	return int(amount2)
}
//每日宝箱加金币
func (m *outStatusModule)BuffDayTreasureGlod(player types.IPlayer, amount int) int{
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffDayTreasureGlod(amount2, pcm.getBuffEffect(consts.PrivDayTreasureGold))
	})
	return int(amount2)
}

//VIP每日宝箱翻倍
func (m *outStatusModule)BuffDoubleRewardOfVip(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffDoubleRewardOfVip(amount2, pcm.getBuffEffect(consts.PrivDoubleRewardOfVip))
	})
	return int(amount2)
}

//VIP对战宝箱卡牌+2
func (m *outStatusModule)BuffAddCardOfVip(player types.IPlayer, amount int) int {
	amount2 := float64(amount)
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	pcm.forEachBuff(func(b iBuff) {
		amount2 = b.buffAddCardOfVip(amount2, pcm.getBuffEffect(consts.PrivAddCardOfVip))
	})
	return int(amount2)
}

func (m *outStatusModule) GetBuffLevel(player types.IPlayer) int {
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	return pcm.getBuffLevel()
}

func (m *outStatusModule) HasAllPriv(player types.IPlayer) bool {
	pcm := player.GetComponent(consts.OutStatusCpt).(*outstatusComponent)
	return pcm.hasAllPriv()
}

func (m *outStatusModule) AddVipBuff(p types.IPlayer) {
	if !p.IsVip() {
		return
	}

	st := module.OutStatus.GetStatus(p, consts.OtVipCard)
	if st == nil {
		st = module.OutStatus.GetStatus(p, consts.OtMinVipCard)
	}
	if st == nil {
		return
	}
	leftTime := st.GetRemainTime()
	deadline := st.GetLeftTime()

	vipBuff12 := m.AddBuff(p, consts.PrivDoubleRewardOfVip, leftTime)
	vipBuff12.SetLeftTime(deadline)

	vipBuff13 := m.AddBuff(p, consts.PrivAddCardOfVip, leftTime)
	vipBuff13.SetLeftTime(deadline)

	vipBuff14 := m.AddBuff(p, consts.PrivAutoOpenTreasureOfVip, leftTime)
	vipBuff14.SetLeftTime(deadline)

}

func Initialize() {
	mod = &outStatusModule{}
	module.OutStatus = mod
}
