package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var campaignMgr = &campaignMgrSt{}

type campaignMgrSt struct {
	countryJobs  []pb.CampaignJob
	cityJobs     []pb.CampaignJob
	sortCountrys common.UInt32Set
	sortCitys    common.IntSet
	info         *pb.GCampaignInfo
}

func (cm *campaignMgrSt) initialize() {
	cm.countryJobs = []pb.CampaignJob{pb.CampaignJob_YourMajesty, pb.CampaignJob_Counsellor, pb.CampaignJob_General}
	cm.cityJobs = []pb.CampaignJob{pb.CampaignJob_Prefect, pb.CampaignJob_DuWei, pb.CampaignJob_FieldOfficer}
	cm.sortCountrys = common.UInt32Set{}
	cm.sortCitys = common.IntSet{}

	timer.AddTicker(time.Second, cm.onHeartBeat)
	eventhub.Subscribe(evWarReady, cm.onWarStateUpdate)
	eventhub.Subscribe(evWarBegin, cm.onWarStateUpdate)
	eventhub.Subscribe(evWarEnd, cm.onWarStateUpdate)
	eventhub.Subscribe(evUnified, cm.onWarStateUpdate)
}

func (cm *campaignMgrSt) genAttrName(name string) string {
	return fmt.Sprintf("%s%d", name, version)
}

func (cm *campaignMgrSt) onHeartBeat() {
	cm.sortCountrys.ForEach(func(countryID uint32) bool {
		cry := countryMgr.getCountry(countryID)
		if cry != nil {
			cry.sortPlayers()
		}
		return true
	})
	cm.sortCountrys = common.UInt32Set{}

	cm.sortCitys.ForEach(func(cityID int) {
		cty := cityMgr.getCity(cityID)
		if cty != nil {
			cty.sortInCityPlayers()
		}
	})
	cm.sortCitys = common.IntSet{}
}

func (cm *campaignMgrSt) addSortCity(cityID int) {
	cm.sortCitys.Add(cityID)
}

func (cm *campaignMgrSt) addSortCountry(countryID uint32) {
	cm.sortCountrys.Add(countryID)
}

func (cm *campaignMgrSt) isCityJob(job pb.CampaignJob) bool {
	for _, j := range cm.cityJobs {
		if j == job {
			return true
		}
	}
	return false
}

func (cm *campaignMgrSt) isCountryJob(job pb.CampaignJob) bool {
	for _, j := range cm.countryJobs {
		if j == job {
			return true
		}
	}
	return false
}

func (cm *campaignMgrSt) getCountryJobs() []pb.CampaignJob {
	return cm.countryJobs
}

func (cm *campaignMgrSt) getCityJobs() []pb.CampaignJob {
	return cm.cityJobs
}

func (cm *campaignMgrSt) broadcastClient(msgID pb.MessageID, arg interface{}) {
	api.BroadcastClient(msgID, arg, &gpb.BroadcastClientFilter{
		OP:  gpb.BroadcastClientFilter_EQ,
		Key: "campaign",
		Val: "all",
	})
}

func (cm *campaignMgrSt) setClientFilter(agent *logic.PlayerAgent) {
	agent.SetClientFilter("campaign", "all")
	p := playerMgr.getPlayer(agent.GetUid())
	if p != nil {
		if p.getCountryID() > 0 {
			agent.SetClientFilter("campaign_country", strconv.Itoa(int(p.getCountryID())))
		}
		if p.getLocationCityID() > 0 {
			agent.SetClientFilter("campaign_lcity", strconv.Itoa(p.getLocationCityID()))
		}
	}
}

func (cm *campaignMgrSt) delClientFilter(agent *logic.PlayerAgent) {
	agent.SetClientFilter("campaign", "")
	agent.SetClientFilter("campaign_country", "")
	agent.SetClientFilter("campaign_lcity", "")
}

func (cm *campaignMgrSt) getCampaignMissionInfo(cty *city, p *player) *pb.CampaignMissionInfo {
	reply := &pb.CampaignMissionInfo{}
	allMissions := cty.getAllMissions()
	for _, ms := range allMissions {
		if ms.getAmount() > 0 {
			reply.Missions = append(reply.Missions, ms.packMsg())
		}
	}

	if p != nil {
		ms := p.getMission()
		if ms != nil {
			reply.ExecutingMission = ms.packMsg()
		}
	}
	return reply
}

func (cm *campaignMgrSt) getCampaignInfo() *pb.GCampaignInfo {
	if cm.info == nil {
		cm.info = &pb.GCampaignInfo{
			Version:       version,
			CampaignState: int32(warMgr.getState()),
		}
	}
	return cm.info
}

func (cm *campaignMgrSt) onWarStateUpdate(_ ...interface{}) {
	cm.info = nil
	logic.BroadcastBackend(pb.MessageID_CA2G_UPDATE_CAMPAIGN_INFO, cm.getCampaignInfo())
}

func (cm *campaignMgrSt) getContributionPerByJob(job pb.CampaignJob) float64 {
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	switch job {
	case pb.CampaignJob_YourMajesty:
		return paramGameData.KingPer
	case pb.CampaignJob_Counsellor:
		return paramGameData.JunshiPer
	case pb.CampaignJob_General:
		return paramGameData.ZhonglangjiangPer
	case pb.CampaignJob_Prefect:
		return paramGameData.TaishouPer
	case pb.CampaignJob_DuWei:
		return paramGameData.DuweiPer
	case pb.CampaignJob_FieldOfficer:
		return paramGameData.XiaoweiPer
	default:
		return 0
	}
}
