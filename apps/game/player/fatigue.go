package player

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"time"
)

// 防沉迷

const (
	//fsHealthy = iota    // 今天在线没到7小时
	//fsUnHealthy         // 今天在线已到7小时
	//fsFatigue           // 不能登录（不能战斗）

	unHealthyOnlineTime = 3 * 60 * 60
	//unHealthyOnlineTime2 = 60 * 60
	//unHealthyDayOnlineTime = 7 * 60 * 60
)

type fatigueComponent struct {
	p    *Player
	attr *attribute.MapAttr
}

func newFatigueComponent(playerAttr *attribute.AttrMgr, isAdult bool) types.IPlayerComponent {
	if !config.GetConfig().IsBanShu || isAdult {
		return nil
	} else {
		attr := playerAttr.GetMapAttr("fatigue")
		if attr == nil {
			attr = attribute.NewMapAttr()
			playerAttr.SetMapAttr("fatigue", attr)
		}
		return &fatigueComponent{
			attr: attr,
		}
	}
}

func (fc *fatigueComponent) ComponentID() string {
	return consts.FatigueCpt
}

func (fc *fatigueComponent) GetPlayer() types.IPlayer {
	return fc.p
}

func (fc *fatigueComponent) OnInit(player types.IPlayer) {
	fc.p = player.(*Player)
}

func (fc *fatigueComponent) OnLogout() {
	fc.setTodayOnlineTime(fc.getTodayOnlineTime() + int(time.Now().Unix()-fc.getLoginTime()))
}

func (fc *fatigueComponent) OnLogin(isRelogin, isRestore bool) {
	if isRestore {
		return
	}

	if isRelogin {
		fc.setTodayOnlineTime(fc.getTodayOnlineTime() + int(time.Now().Unix()-fc.getLoginTime()))
	}

	fc.setLoginTime(time.Now().Unix())
	fc.OnCrossDay(timer.GetDayNo())
}

func (fc *fatigueComponent) OnCrossDay(dayno int) {
	if dayno == fc.p.GetDataDayNo() {
		return
	}

	fc.setLoginTime(time.Now().Unix())
	fc.setTodayOnlineTime(0)
	module.OutStatus.DelStatus(fc.p, consts.OtFatigue)
}

func (fc *fatigueComponent) getTodayOnlineTime() int {
	return fc.attr.GetInt("todayOnlineTime")
}

func (fc *fatigueComponent) setTodayOnlineTime(t int) {
	fc.attr.SetInt("todayOnlineTime", t)
}

func (fc *fatigueComponent) getLoginTime() int64 {
	return fc.attr.GetInt64("loginTime")
}

func (fc *fatigueComponent) setLoginTime(t int64) {
	fc.attr.SetInt64("loginTime", t)
}

func (fc *fatigueComponent) OnHeartBeat() {
	if module.OutStatus.GetStatus(fc.p, consts.OtFatigue) != nil {
		return
	}

	onlineTime := int(time.Now().Unix() - fc.getLoginTime())
	todayOnlineTime := fc.getTodayOnlineTime() + onlineTime
	fc.setTodayOnlineTime(todayOnlineTime)

	if fc.isTodayUnHealthy() {
		module.OutStatus.AddStatus(fc.p, consts.OtFatigue, -1)
	}
}

func (fc *fatigueComponent) isTodayUnHealthy() bool {
	return fc.getTodayOnlineTime() >= unHealthyOnlineTime
}
