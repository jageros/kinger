package main

import (
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"kinger/gamedata"
	"kinger/common/consts"
	"time"
	"kinger/gopuppy/common/timer"
	"fmt"
	"kinger/gopuppy/common/glog"
	"math"
	"kinger/common/utils"
)

type baseMission struct {
	attr *attribute.MapAttr
}

func (m *baseMission) getType() pb.CampaignMsType {
	return pb.CampaignMsType(m.attr.GetInt("type"))
}

func (m *baseMission) getGoldReward() int {
	return m.attr.GetInt("gold")
}

func (m *baseMission) getTransportForage() int {
	return m.attr.GetInt("transportForage")
}

func (m *baseMission) getTransportGold() int {
	return m.attr.GetInt("transportGold")
}

func (m *baseMission) getDispatchGold() int {
	return m.attr.GetInt("dispatchGold")
}

func (m *baseMission) setDispatchGold(gold int) {
	m.attr.SetInt("dispatchGold", gold)
}

func (m *baseMission) getTransportCity() int {
	return m.attr.GetInt("transportCity")
}

func (m *baseMission) setTransportCity(cityID int) {
	m.attr.SetInt("transportCity", cityID)
}

func (m *baseMission) setTransportGold(gold int) {
	m.attr.SetInt("transportGold", gold)
}

func (m *baseMission) setTransportForage(forage int) {
	m.attr.SetInt("transportForage", forage)
}

func (m *baseMission) setTransportDistance(distance int) {
	m.attr.SetInt("transportDistance", distance)
}

func (m *baseMission) getTransportDistance() int {
	return m.attr.GetInt("transportDistance")
}

func (m *baseMission) getMaxTime() int {
	return m.attr.GetInt("maxTime")
}

type mission struct {
	baseMission
	cty *city
}

func newMission(cty *city, msType pb.CampaignMsType, amount, gold int) *mission {
	attr := attribute.NewMapAttr()
	attr.SetInt("type", int(msType))
	attr.SetInt("gold", gold)
	attr.SetInt("amount", amount)
	attr.SetInt("maxAmount", amount)
	attr.SetInt64("ptime", time.Now().Unix())
	return newMissionByAttr(cty, attr)
}

func newMissionByAttr(cty *city, attr *attribute.MapAttr) *mission {
	ms := &mission{
		cty: cty,
	}
	ms.attr = attr
	return ms
}

func (m *mission) String() string {
	return fmt.Sprintf("[ms type=%s, gold=%d, amount=%d]", m.getType(), m.getGoldReward(), m.getAmount())
}

func (m *mission) getAmount() int {
	return m.attr.GetInt("amount")
}

func (m *mission) getMaxAmount() int {
	return m.attr.GetInt("maxAmount")
}

func (m *mission) setAmount(amount int) {
	m.attr.SetInt("amount", amount)
}

func (m *mission) setGold(gold int) {
	m.attr.SetInt("gold", gold)
}

func (m *mission) getPublishTime() int64 {
	return m.attr.GetInt64("ptime")
}

func (m *mission) setPublishTime(t int64) {
	m.attr.SetInt64("ptime", t)
}

func (m *mission) getContribution() float64 {
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	switch m.getType() {
	case pb.CampaignMsType_Irrigation:
		return math.Ceil( paramGameData.IrrigationVic * m.cty.getGlory() )
	case pb.CampaignMsType_Trade:
		return math.Ceil( paramGameData.TradeVic * m.cty.getGlory() )
	case pb.CampaignMsType_Build:
		return math.Ceil( paramGameData.BuildVic * m.cty.getGlory() )
	case pb.CampaignMsType_Transport:
		return math.Ceil( paramGameData.TransportVic * float64(m.getTransportDistance()) * m.cty.getGlory() )
	default:
		return 0
	}
}

func (m *mission) accept(p *player, gcardIDs []uint32) *playerMission {
	amount := m.getAmount()
	if amount <= 0 {
		return nil
	}
	if p.getMission() != nil {
		return nil
	}
	amount--
	m.setAmount(amount)
	pm := newPlayerMission(m.getType(), m.getGoldReward(), m.cty.getCityID(), p, gcardIDs, m.getMaxTime())
	pm.setTransportCity(m.getTransportCity())
	pm.setTransportForage(m.getTransportForage())
	pm.setTransportGold(m.getTransportGold())
	pm.setContribution(m.getContribution())
	p.onAcceptMission(pm)

	if amount <= 0 {
		prefect := m.cty.getPrefect()
		if prefect != nil {
			noticeMgr.sendNoticeToPlayer(prefect.getUid(), pb.CampaignNoticeType_ClearMissionNt)
		}
	}
	return pm
}

func (m *mission) cancel() {
	amount := m.getAmount()
	if amount <= 0 {
		return
	}
	m.setAmount(0)
	gold := m.getGoldReward()
	transportGold := m.getTransportGold()
	transportForage := m.getTransportForage()
	dispatchGold := m.getDispatchGold()
	m.cty.modifyResource(resGold, float64((gold + transportGold + dispatchGold) * amount))
	m.cty.modifyResource(resForage, float64(transportForage * amount))
}

func (m *mission) packMsg() *pb.CampaignMission {
	msg := &pb.CampaignMission{
		Type: m.getType(),
		GoldReward: int32(m.getGoldReward()),
		Amount: int32(m.getAmount()),
		TransportTargetCity: int32(m.getTransportCity()),
		TransportMaxTime: int32(m.getMaxTime()),
		MaxAmount: int32(m.getMaxAmount()),
		Contribution: int32(m.getContribution()),
	}

	if m.getTransportForage() > 0 {
		msg.TransportType = pb.TransportTypeEnum_ForageTT
	} else if m.getTransportGold() > 0 {
		msg.TransportType = pb.TransportTypeEnum_GoldTT
	}
	return msg
}

func (m *mission) equal(pm *playerMission) bool {
	return m.getType() == pm.getType() && m.getGoldReward() == pm.getGoldReward() &&
		m.getTransportForage() == pm.getTransportForage() && m.getTransportGold() == pm.getTransportGold()
}

func (m *mission) setMaxTime(maxTime int) {
	m.attr.SetInt("maxTime", maxTime)
}

type playerMission struct {
	baseMission
	owner *player
	gcardIDsAttr *attribute.ListAttr
	completeTimer *timer.Timer
}

func newPlayerMission(msType pb.CampaignMsType, gold, cityID int, p *player, gcardIDs []uint32, maxTime int) *playerMission {
	attr := attribute.NewMapAttr()
	attr.SetInt("type", int(msType))
	attr.SetInt("gold", gold)
	attr.SetInt("cityID", cityID)
	gcardIDsAttr := attribute.NewListAttr()
	attr.SetListAttr("gcardIDs", gcardIDsAttr)

	if msType != pb.CampaignMsType_Transport && msType != pb.CampaignMsType_Dispatch {
		paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
		var timeRare float64
		var val float64
		switch msType {
		case pb.CampaignMsType_Irrigation:
			timeRare = paramGameData.IrrigationTime
		case pb.CampaignMsType_Trade:
			timeRare = paramGameData.TradeTime
		case pb.CampaignMsType_Build:
			fallthrough
		default:
			timeRare = paramGameData.BuildTime
		}

		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		for _, gcardID := range gcardIDs {
			cardData := poolGameData.GetCardByGid(gcardID)
			if cardData != nil {
				switch msType {
				case pb.CampaignMsType_Irrigation:
					val += cardData.Politics
				case pb.CampaignMsType_Trade:
					val += cardData.Intelligence
				case pb.CampaignMsType_Build:
					fallthrough
				default:
					val += cardData.Force
				}
			}
		}

		maxTime = int(timeRare / val * 3600)
	}

	attr.SetInt("maxTime", maxTime)
	for _, gcardID := range gcardIDs {
		gcardIDsAttr.AppendUInt32(gcardID)
	}

	attr.SetInt64("timeout", time.Now().Unix() + int64(attr.GetInt("maxTime")))
	return newPlayerMissionByAttr(p, attr)
}

func newPlayerMissionByAttr(p *player, attr *attribute.MapAttr) *playerMission {
	ms := &playerMission{
		owner: p,
		gcardIDsAttr: attr.GetListAttr("gcardIDs"),
	}
	ms.attr = attr

	if ms.getType() == pb.CampaignMsType_Dispatch {

	} else {
		remainTime := ms.getTimeout() - time.Now().Unix()
		if remainTime < 1 {
			remainTime = 1
		}
		ms.completeTimer = timer.AfterFunc(time.Duration(remainTime)*time.Second, ms.onComplete)
	}
	return ms
}

func (m *playerMission) String() string {
	return fmt.Sprintf("[pms cityID=%d, type=%s, gold=%d, transportCity=%d, transportForage=%d, transportGold=%d, " +
		"contribution=%f]", m.getCityID(), m.getType(), m.getGoldReward(), m.getTransportCity(), m.getTransportForage(),
		m.getTransportGold(), m.getContribution())
}

func (m *playerMission) getContribution() float64 {
	return m.attr.GetFloat64("contribution")
}

func (m *playerMission) setContribution(contribution float64) {
	m.attr.SetFloat64("contribution", contribution)
}

func (m *playerMission) getCityID() int {
	return m.attr.GetInt("cityID")
}

func (m *playerMission) onComplete() {
	m.completeTimer = nil
	if m.canReward() || m.hasReward() || m.isCancel() {
		return
	}
	glog.Infof("playerMission onComplete, uid=%d, %s", m.owner.getUid(), m)

	m.attr.SetBool("canReward", true)

	if m.getType() == pb.CampaignMsType_Dispatch {
		targetCityID := m.getTransportCity()
		targetCity := cityMgr.getCity(targetCityID)
		if targetCity == nil || targetCity.getCountryID() != m.owner.getCountryID() {
			return
		}

		m.owner.moveCity(targetCityID)
		return
	}

	utils.PlayerMqPublish(m.owner.getUid(), pb.RmqType_CampaignMissionDone, &pb.RmqCampaignMissionDone{
		Cards: m.getCards(),
	})

	cty := cityMgr.getCity(m.getCityID())
	if cty != nil {
		paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
		switch m.getType() {
		case pb.CampaignMsType_Irrigation:
			cty.modifyResource(resAgriculture, paramGameData.TaskTargetIrrigation)
		case pb.CampaignMsType_Trade:
			cty.modifyResource(resBusiness, paramGameData.TaskTargetTrade)
		case pb.CampaignMsType_Build:
			cty.modifyResource(resDefense, paramGameData.TaskTargetBuild)
		case pb.CampaignMsType_Transport:
			tCty := cityMgr.getCity(m.getTransportCity())
			if tCty != nil {
				forage := m.getTransportForage()
				gold := m.getTransportGold()
				tCty.modifyResource(resForage, float64(forage))
				tCty.modifyResource(resGold, float64(gold))

				transportType := pb.TransportTypeEnum_GoldTT
				amount := gold
				if forage > 0 {
					transportType = pb.TransportTypeEnum_ForageTT
					amount = forage
				}
				jobPlayers := tCty.getAllJobPlayers()
				for _, ps := range jobPlayers {
					for _, p := range ps {
						noticeMgr.sendNoticeToPlayer(p.getUid(), pb.CampaignNoticeType_TransportNt, cty.getCityID(),
							tCty.getCityID(), transportType, amount)
					}
				}
			}
		}
	}
}

func (m *playerMission) canReward() bool {
	return m.attr.GetBool("canReward")
}

func (m *playerMission) hasReward() bool {
	return m.attr.GetBool("hasReward")
}

func (m *playerMission) reward() {
	m.attr.SetBool("hasReward", true)
}

func (m *playerMission) isCancel() bool {
	return m.attr.GetBool("isCancel")
}

func (m *playerMission) cancel() {
	if m.completeTimer != nil {
		m.completeTimer.Cancel()
	}
	cty := cityMgr.getCity(m.getCityID())
	if cty != nil {
		cty.onMissionCancel(m)
	}
	m.attr.SetBool("isCancel", true)

	utils.PlayerMqPublish(m.owner.getUid(), pb.RmqType_CampaignMissionDone, &pb.RmqCampaignMissionDone{
		Cards: m.getCards(),
	})
}

func (m *playerMission) getCards() []uint32 {
	var cards []uint32
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	m.gcardIDsAttr.ForEachIndex(func(index int) bool {
		gcardID := m.gcardIDsAttr.GetUInt32(index)
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData != nil {
			cards = append(cards, cardData.CardID)
		}
		return true
	})
	return cards
}

func (m *playerMission) getRemainTime() time.Duration {
	if !m.canReward() && m.completeTimer != nil {
		return m.completeTimer.GetRemainTime()
	} else {
		return 0
	}
}

func (m *playerMission) getTimeout() int64 {
	return m.attr.GetInt64("timeout")
}

func (m *playerMission) packMsg() *pb.ExecutingCampaignMission {
	return &pb.ExecutingCampaignMission{
		Type: m.getType(),
		GoldReward: int32(m.getGoldReward()),
		Cards: m.getCards(),
		RemainTime: int32(m.getRemainTime().Seconds()),
		MaxTime: int32(m.getMaxTime()),
		Contribution: int32(m.getContribution()),
	}
}
