package main

import (
	"kinger/proto/pb"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/eventhub"
	"time"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"fmt"
	"kinger/gopuppy/common/evq"
)

// TODO
// 国战期间是否可以接任务，运输任务怎么办
// 怎样表现驱逐剩余时间
// 驱逐官大的
// 国战期间是否可以驱逐
// 国战期间是否可以辞官
// 战斗界面怎样表现城防技能
// 国战期间独立了 ？？
// 国战结束、撤退，玩家身上的粮草是否返回城市
// C2S_ACC_DEF_CITY_LOSE_LOADING  是耗城市的粮草吗

var (
	warMgr = &warMgrSt{}
	nextReadyWarTime time.Time
	nextBeignWarTime time.Time
	nextEndWarTime time.Time
)

type warMgrSt struct {
	attr *attribute.AttrMgr
	yourMajestyRankAttr *attribute.ListAttr
	beAttackCity common.IntSet
	normalTicker *timer.Timer
	readyTicker *timer.Timer
	warTicker *timer.Timer
	unifiedInfo []byte
}

func (wm *warMgrSt) initialize() {
	wm.beAttackCity = common.IntSet{}
	wm.attr = attribute.NewAttrMgr("campaign_war", version)
	err := wm.attr.Load()
	if err != nil {
		if err == attribute.NotExistsErr {
			wm.attr.SetDirty(true)
			wm.attr.Save(false)
		} else {
			panic(err)
		}
	}

	wm.yourMajestyRankAttr = wm.attr.GetListAttr("yourMajestyRank")
	if wm.yourMajestyRankAttr == nil {
		wm.yourMajestyRankAttr = attribute.NewListAttr()
		wm.attr.SetListAttr("yourMajestyRank", wm.yourMajestyRankAttr)
	}

	unifiedInfo := wm.attr.GetStr("unifiedInfo")
	if unifiedInfo != "" {
		wm.unifiedInfo = []byte(unifiedInfo)
	}

	wm.setState(pb.CampaignState_Pause)
	timer.AddTicker(5 * time.Minute, func() {
		wm.save(false)
	})
}

func (wm *warMgrSt) calcWarTime() {
	now := time.Now()
	todayEndTime := time.Date(now.Year(), now.Month(), now.Day(), coutryWarEndHour, coutryWarBeginMin, 0, 0, now.Location())
	if !now.Before(todayEndTime) {
		now = now.Add(24 * time.Hour)
	}

	nextDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var diffDay int
	switch nextDate.Weekday() {
	case time.Monday:
		diffDay = 1
	case time.Tuesday:
	case time.Wednesday:
		diffDay = 1
	case time.Thursday:
	case time.Friday:
		diffDay = 1
	case time.Saturday:
	case time.Sunday:
		diffDay = 2
	}

	if diffDay > 0 {
		nextDate = nextDate.Add(24 * time.Hour * time.Duration(diffDay))
	}

	nextReadyWarTime = nextDate.Add(coutryWarReadyBeginHour * time.Hour + coutryWarReadyBeginMin * time.Minute)
	nextBeignWarTime = nextDate.Add(coutryWarBeginHour * time.Hour + coutryWarBeginMin * time.Minute)
	nextEndWarTime = nextDate.Add(coutryWarEndHour * time.Hour + coutryWarEndMin * time.Minute)

	now = time.Now()
	if wm.normalTicker != nil {
		wm.normalTicker.Cancel()
	}
	if nextReadyWarTime.After(now) {
		wm.normalTicker = timer.AfterFunc(nextReadyWarTime.Sub(now), wm.onWarReady)
	}

	if wm.readyTicker != nil {
		wm.readyTicker.Cancel()
	}
	if nextBeignWarTime.After(now) {
		wm.readyTicker = timer.AfterFunc(nextBeignWarTime.Sub(now), wm.onWarBegin)
	}

	if wm.warTicker != nil {
		wm.warTicker.Cancel()
	}
	if nextEndWarTime.After(now) {
		wm.warTicker = timer.AfterFunc(nextEndWarTime.Sub(now), wm.onWarEnd)
	}

	glog.Infof("calcWarTime nextReadyWarTime=%s, nextBeignWarTime=%s, nextEndWarTime=%s", nextReadyWarTime,
		nextBeignWarTime, nextEndWarTime)
}

func (wm *warMgrSt) initializeTimer()  {
	state := wm.getState()
	glog.Infof("war initializeTimer state=%s", state)
	if state == pb.CampaignState_Unified {
		return
	}

	/*
	now := time.Now()
	switch state {
	case pb.CampaignState_Normal:
		beginTime := time.Date(now.Year(), now.Month(), now.Day(), coutryWarReadyBeginHour, coutryWarReadyBeginMin, 0, 0,
			now.Location())
		if now.Before(beginTime) {
			break
		}
		wm.onWarReady()
		fallthrough

	case pb.CampaignState_ReadyWar:
		beginTime := time.Date(now.Year(), now.Month(), now.Day(), coutryWarBeginHour, coutryWarBeginMin, 0, 0,
			now.Location())
		if now.Before(beginTime) {
			break
		}
		wm.onWarBegin()
		fallthrough

	case pb.CampaignState_InWar:
		beginTime := time.Date(now.Year(), now.Month(), now.Day(), coutryWarEndHour, coutryWarEndMin, 0, 0,
			now.Location())
		if now.Before(beginTime) {
			break
		}
		wm.onWarEnd()

	case pb.CampaignState_Unified:
		return
	}
	*/

	wm.calcWarTime()
	//wm.normalTicker = timer.RunEveryDay(coutryWarReadyBeginHour, coutryWarReadyBeginMin, 0, wm.onWarReady)
	//wm.readyTicker = timer.RunEveryDay(coutryWarBeginHour, coutryWarBeginMin, 0, wm.onWarBegin)
	//wm.warTicker = timer.RunEveryDay(coutryWarEndHour, coutryWarEndMin, 0, wm.onWarEnd)
}

func (wm *warMgrSt) onWarReady() {
	if wm.getState() != pb.CampaignState_Normal {
		return
	}

	glog.Infof("onWarReady")
	wm.setState(pb.CampaignState_ReadyWar)
	eventhub.Publish(evWarReady)

	msg := &pb.CampaignState{
		State: pb.CampaignState_ReadyWar,
	}
	msg.Arg, _ = (&pb.CaStateWarArg{RemainTime: int32(wm.getStateRemainTime())}).Marshal()
	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_STATE, msg)
}

func (wm *warMgrSt) onWarBegin() {
	if wm.getState() != pb.CampaignState_ReadyWar {
		return
	}

	glog.Infof("onWarBegin")
	wm.setState(pb.CampaignState_InWar)
	eventhub.Publish(evWarBegin)

	msg := &pb.CampaignState{
		State: pb.CampaignState_InWar,
	}
	msg.Arg, _ = (&pb.CaStateWarArg{RemainTime: int32(wm.getStateRemainTime())}).Marshal()
	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_STATE, msg)
}

func (wm *warMgrSt) onWarEnd() {
	if wm.getState() != pb.CampaignState_InWar {
		return
	}

	glog.Infof("onWarEnd")
	version := wm.getVersion() + 1
	wm.attr.SetInt("version", version)
	wm.setState(pb.CampaignState_Normal)
	wm.calcWarTime()
	playerMgr.onWarEnd()
	eventhub.Publish(evWarEnd)
}

func (wm *warMgrSt) onUnified(cry *country) {
	wm.setState(pb.CampaignState_Unified)
	if wm.normalTicker != nil {
		wm.normalTicker.Cancel()
	}
	if wm.readyTicker != nil {
		wm.readyTicker.Cancel()
	}
	if wm.warTicker != nil {
		wm.warTicker.Cancel()
	}

	evq.CallLater(func() {
		unifiedMsg := &pb.CaStateUnifiedArg{
			CountryID: cry.getID(),
			CountryPlayerAmount: int32(cry.getPlayerAmount()),
		}
		p := cry.getYourMajesty()
		if p != nil {
			unifiedMsg.YourMajestyName = p.getName()
		}
		wm.unifiedInfo, _ = unifiedMsg.Marshal()
		wm.attr.SetStr("unifiedInfo", string(wm.unifiedInfo))
		wm.attr.Save(false)

		uid2Rank := map[common.UUid]int{}
		uid2Rank[p.getUid()] = 1
		rankLen := wm.yourMajestyRankAttr.Size()
		wm.yourMajestyRankAttr.ForEachIndex(func(index int) bool {
			uid := common.UUid(wm.yourMajestyRankAttr.GetUInt64(index))
			uid2Rank[uid] = rankLen + 1
			rankLen --
			return true
		})
		playerMgr.onUnified(uid2Rank, unifiedMsg.YourMajestyName, cry.getName())
		eventhub.Publish(evUnified)
	})
}

func (wm *warMgrSt) onCountryDestory(cry *country) {
	yourMajesty := cry.getYourMajesty()
	if yourMajesty != nil {
		wm.yourMajestyRankAttr.AppendUInt64(uint64(yourMajesty.getUid()))
		if wm.yourMajestyRankAttr.Size() > 2 {
			wm.yourMajestyRankAttr.DelByIndex(0)
		}
	}
}

func (wm *warMgrSt) addBeAttackCity(cityID int) {
	if wm.getState() == pb.CampaignState_InWar {
		wm.beAttackCity.Add(cityID)
	}
}

func (wm *warMgrSt) syncCityDefense() {
	if wm.beAttackCity.Size() <= 0 {
		return
	}

	msg := &pb.SyncCityDefenseArg{}
	wm.beAttackCity.ForEach(func(cityID int) {
		cty := cityMgr.getCity(cityID)
		if cty != nil {
			msg.CityDefenses = append(msg.CityDefenses, &pb.CityDefense{
				CityID: int32(cityID),
				Defense: int32(cty.getResource(resDefense)),
			})
		}
	})
	campaignMgr.broadcastClient(pb.MessageID_S2C_SYNC_CITY_DEFENSE, msg)
}

func (wm *warMgrSt) setState(state pb.CampaignState_StateEnum) {
	wm.attr.SetInt("state", int(state))
	wm.attr.Save(false)
}

func (wm *warMgrSt) getState() pb.CampaignState_StateEnum {
	return pb.CampaignState_StateEnum(wm.attr.GetInt("state"))
}

func (wm *warMgrSt) getStateRemainTime() int {
	now := time.Now()
	switch wm.getState() {
	case pb.CampaignState_Normal:
		return int(nextReadyWarTime.Sub(now).Seconds())
	case pb.CampaignState_ReadyWar:
		return int(nextBeignWarTime.Sub(now).Seconds())
	case pb.CampaignState_InWar:
		return int(nextEndWarTime.Sub(now).Seconds())
	}
	return 0
}

func (wm *warMgrSt) isNormalState() bool {
	return wm.getState() == pb.CampaignState_Normal
}

func (wm *warMgrSt) isReadyWar() bool {
	return wm.getState() == pb.CampaignState_ReadyWar
}

func (wm *warMgrSt) isInWar() bool {
	return wm.getState() == pb.CampaignState_InWar
}

func (wm *warMgrSt) isUnified() bool {
	return wm.getState() == pb.CampaignState_Unified
}

func (wm *warMgrSt) genTeamID() int {
	id := wm.attr.GetInt("teamID") + 1
	wm.attr.SetInt("teamID", id)
	return id
}

func (wm *warMgrSt) save(isStopServer bool) {
	wm.attr.Save(isStopServer)
}

func (wm *warMgrSt) getVersion() int {
	return wm.attr.GetInt("version")
}

func (wm *warMgrSt) getTeamAttrName() string {
	return fmt.Sprintf("%s_%d", campaignMgr.genAttrName("campaignTeam"), wm.getVersion())
}

func (wm *warMgrSt) getUnifiedInfo() []byte {
	return wm.unifiedInfo
}

func (wm *warMgrSt) nextState() {
	state := wm.getState()
	switch state {
	case pb.CampaignState_Normal:
		wm.onWarReady()
	case pb.CampaignState_ReadyWar:
		wm.onWarBegin()
	case pb.CampaignState_InWar:
		wm.onWarEnd()
	}
}

func (wm *warMgrSt) isPause() bool {
	return wm.getState() == pb.CampaignState_Pause
}
