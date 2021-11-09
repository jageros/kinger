package main

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/wordfilter"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
	"strconv"
	"strings"
)

func rpc_G2CA_FetchCampaignInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, err := playerMgr.loadPlayer(uid)
	arg2 := arg.(*pb.CampaignSimplePlayer)
	if err != nil {
		if err == attribute.NotExistsErr {
			p = playerMgr.createPlayer(uid, arg2.Name, arg2.HeadImg, arg2.HeadFrame, int(arg2.PvpScore))
		} else {
			return nil, err
		}
	}

	p.setName(arg2.Name)
	p.setHeadFrame(arg2.HeadFrame)
	p.setHeadImg(arg2.HeadImg)
	p.setPvpScore(int(arg2.PvpScore))
	playerMgr.onPlayerLogin(p, agent)

	reply := &pb.CampaignInfo{
		//State: &pb.CampaignState{State: warMgr.getState()},
		State:            &pb.CampaignState{State: pb.CampaignState_Unified},
		DefPlayerAmounts: sceneMgr.getDefAmount(),
	}

	/*
		warRecord := p.fetchWarEndRecord()
		switch reply.State.State {
		case pb.CampaignState_ReadyWar:
			fallthrough
		case pb.CampaignState_InWar:
			reply.State.Arg, _ = (&pb.CaStateWarArg{RemainTime: int32(warMgr.getStateRemainTime())}).Marshal()
		case pb.CampaignState_Unified:
			reply.State.Arg = warMgr.getUnifiedInfo()
		case pb.CampaignState_Pause:
			fallthrough
		case pb.CampaignState_Normal:
			if warRecord != nil {
				reply.State.State = pb.CampaignState_WarEnd
				reply.State.Arg, _ = warRecord.Marshal()
			} else {
				reply.State.Arg, _ = (&pb.CaStateWarArg{RemainTime: int32(warMgr.getStateRemainTime())}).Marshal()
			}

		}
	*/

	if p != nil {
		reply.LastCountryID = p.getLastCountryID()
		reply.MyCityID = int32(p.getCityID())
		reply.MyLocationCityID = int32(p.getLocationCityID())
		reply.MyCityJob = p.getCityJob()
		reply.MyCountryJob = p.getCountryJob()
		reply.Forage = int32(p.getForage())
		reply.Contribution = int32(p.getContribution())
		reply.MaxContribution = int32(p.getMaxContribution())
		m := p.getMission()
		reply.HasCompleteMission = m != nil && m.canReward()
		n := noticeMgr.getPlayerNotice(uid)
		reply.HasNewNotice = n != nil && n.hasNew()
		reply.MyCountryID = p.getCountryID()
		reply.TeamDisappear = p.fetchTeamDisappear()
		reply.MyState = p.getMyState()
		reply.SupportCards = p.getSupportCards()
		t := p.getTeam()
		if t != nil {
			reply.MyTeam = t.packMsg()
		}
	}

	allCountrys := countryMgr.getAllCountrys()
	for _, cry := range allCountrys {
		reply.Countrys = append(reply.Countrys, cry.packSimpleMsg())
	}

	cityMgr.forEachCity(func(cty *city) bool {
		if cty.getCountryID() > 0 {
			reply.Citys = append(reply.Citys, cty.packSimpleMsg())
		}
		return true
	})

	teams := sceneMgr.getSyncClientTeams()
	for _, t := range teams {
		reply.Teams = append(reply.Teams, t.packMsg())
	}

	campaignMgr.setClientFilter(agent)
	return reply, nil
}

func rpc_G2CA_SettleCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.GSettleCityArg)
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.InternalErr
	}

	p.setName(arg2.Player.Name)
	p.setHeadImg(arg2.Player.HeadImg)
	p.setHeadFrame(arg2.Player.HeadFrame)
	p.setPvpScore(int(arg2.Player.PvpScore))
	c := cityMgr.getCity(int(arg2.CityID))
	if c == nil {
		return nil, gamedata.GameError(2)
	}

	return nil, c.playerSettle(p, true)
}

func rpc_C2S_FetchCityData(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetCity)
	c := cityMgr.getCity(int(arg2.CityID))
	if c == nil {
		return nil, gamedata.InternalErr
	}

	return c.packMsg(agent.GetUid()), nil
}

func rpc_G2CA_CreateCountry(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.InternalErr
	}

	c := p.getCity()
	if c == nil {
		return nil, gamedata.GameError(1)
	}

	cry := c.getCountry()
	if cry != nil {
		return nil, gamedata.GameError(2)
	}

	if !createCountryMgr.canApplyCreateCountry() {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CreateCountryArg)
	c.applyCreateCountry(p, int(arg2.Gold))

	glog.Infof("rpc_G2CA_CreateCountry uid=%d, gold=%d, cityID=%d", uid, arg2.Gold, c.getCityID())
	reply := &pb.ApplyCreateCountryData{
		RemainTime: int32(createCountryMgr.getApplyCreateCountryRemainTime().Seconds()),
		Players:    c.getCreateCoutryApplysByPage(0),
	}
	ca := c.getCreateCountryApply(agent.GetUid())
	if ca != nil {
		reply.MyApplyMoney = int32(ca.getGold())
	}
	return reply, nil
}

func rpc_C2S_LeaveCampaignScene(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	campaignMgr.delClientFilter(agent)
	playerMgr.onPlayerLeaveCampaign(agent.GetUid())
	return nil, nil
}

func rpc_C2S_FetchApplyCreateCountryInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetCity)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	if cty.getCountry() != nil {
		return nil, gamedata.GameError(2)
	}

	reply := &pb.ApplyCreateCountryData{
		RemainTime: int32(createCountryMgr.getApplyCreateCountryRemainTime().Seconds()),
		Players:    cty.getCreateCoutryApplysByPage(0),
	}
	ca := cty.getCreateCountryApply(agent.GetUid())
	if ca != nil {
		reply.MyApplyMoney = int32(ca.getGold())
	}
	return reply, nil
}

func rpc_C2S_FetchApplyCreateCountryPlayers(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchApplyCreateCountryPlayersArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	if cty.getCountry() != nil {
		return nil, gamedata.GameError(2)
	}
	return cty.getCreateCoutryApplysByPage(int(arg2.Page)), nil
}

func rpc_C2S_FetchCampaignMissionInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetCity)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}

	p, _ := playerMgr.loadPlayer(agent.GetUid())
	return campaignMgr.getCampaignMissionInfo(cty, p), nil
}

func rpc_G2CA_AcceptCampaignMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(5)
	}

	c := p.getCity()
	if c == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getMission() != nil {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.AcceptCampaignMissionArg)
	ms := c.acceptMission(p, arg2.Type, int(arg2.TransportTargetCity), arg2.Cards)
	if ms == nil {
		return nil, gamedata.GameError(4)
	}

	reply := &pb.GAcceptCampaignMissionReply{
		AcceptReply: &pb.AcceptCampaignMissionReply{
			RemainTime: int32(ms.getRemainTime().Seconds()),
			Missions:   campaignMgr.getCampaignMissionInfo(c, nil).Missions,
		},
	}

	if ms.getType() == pb.CampaignMsType_Dispatch {
		ms.onComplete()
		reward, _ := p.getMissionReward()
		if reward != nil {
			reply.RewardGold = reward.Gold
		}
	}

	return reply, nil
}

func rpc_G2CA_CancelCampaignMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	err := p.cancelMission()
	if err != nil {
		return nil, err
	}
	return campaignMgr.getCampaignMissionInfo(p.getCity(), nil), nil
}

func rpc_G2CA_GetCampaignMissionReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	return p.getMissionReward()
}

func rpc_C2S_CampaignPublishMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(5)
	}

	job := p.getCityJob()
	if job != pb.CampaignJob_Prefect && job != pb.CampaignJob_DuWei {
		return nil, gamedata.GameError(2)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CampaignPublishMissionArg)
	if arg2.GoldReward < 0 {
		return nil, gamedata.GameError(4)
	}

	var transportCityPath []int
	for _, cityID := range arg2.TransportCityPath {
		transportCityPath = append(transportCityPath, int(cityID))
	}
	err := cty.publishMission(uid, arg2.Type, int(arg2.GoldReward), int(arg2.Amount), arg2.TransportType, transportCityPath)
	if err != nil {
		return nil, err
	}
	return &pb.CampaignPublishMissionReply{
		MissionInfo: campaignMgr.getCampaignMissionInfo(cty, p),
		Gold:        int32(cty.getResource(resGold)),
		Forage:      int32(cty.getResource(resForage)),
	}, nil
}

func rpc_C2S_PatrolCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.TargetCity)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	return &pb.PatrolCityReply{
		Contribution: int32(p.getContribution()),
		//Salary: int32(p.getSalary()),
		//IsBeAttack: cty.isBeAttack(),
	}, nil
}

func rpc_C2S_FetchInCityPlayers(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchCityPlayersArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	return &pb.CampaignPlayerList{
		Players: playerMgr.getPlayersByPage(cty.getAllInCityPlayers(), int(arg2.Page), false),
	}, nil
}

func rpc_C2S_FetchCityPlayers(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchCityPlayersArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	cty.sortPlayers()
	return &pb.CampaignPlayerList{
		Players: playerMgr.getPlayersByPage(cty.getAllPlayers(), int(arg2.Page), false),
	}, nil
}

func rpc_C2S_FetchCityCaptives(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchCityPlayersArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	return &pb.CampaignPlayerList{
		Players: playerMgr.getPlayersByPage(cty.getCaptives(), int(arg2.Page), true),
	}, nil
}

func rpc_C2S_SetForagePrice(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	job := p.getCityJob()
	if job != pb.CampaignJob_Prefect && job != pb.CampaignJob_DuWei {
		return nil, gamedata.GameError(2)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(3)
	}

	price := int(arg.(*pb.SetForagePriceArg).Price)
	if price < 0 {
		return nil, gamedata.GameError(4)
	}
	cty.setForagePrice(price)
	return nil, nil
}

func rpc_C2S_FetchForagePrice(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getLocationCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	return &pb.FetchForagePriceReply{
		ForageAmount: int32(cty.getResource(resForage)),
		Price:        int32(cty.getForagePrice()),
	}, nil
}

func rpc_C2S_FetchAllCityPlayerAmount(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	reply := &pb.AllCityPlayerAmount{}
	cityMgr.forEachCity(func(cty *city) bool {
		msg := &pb.CityPlayerAmount{
			CityID:           int32(cty.getCityID()),
			PlayerAmount:     int32(cty.getPlayerAmount()),
			Glory:            int32(cty.getGlory() * 10),
			AvgMissionReward: int32(cty.getMaxMissionReward()),
		}
		ca := cty.getTopCreateCoutryApply()
		if ca != nil {
			msg.MaxApplyCountryGold = int32(ca.getGold())
		}
		reply.PlayerAmounts = append(reply.PlayerAmounts, msg)
		return true
	})
	return reply, nil
}

func rpc_C2S_FetchCountryJobPlayers(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	cry := cty.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(3)
	}
	return &pb.CampaignPlayerList{
		Players: cry.getCountryJobPlayers(),
	}, nil
}

func rpc_C2S_FetchCountryPlayers(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchCountryPlayersArg)
	cry := countryMgr.getCountry(arg2.CountryID)
	if cry == nil {
		return nil, gamedata.GameError(1)
	}
	cry.sortPlayers()
	return &pb.CampaignPlayerList{
		Players: playerMgr.getPlayersByPage(cry.getAllPlayers(), int(arg2.Page), false),
	}, nil
}

func rpc_C2S_AppointJob(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	cry := cty.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.AppointJobArg)
	if arg2.Job == pb.CampaignJob_YourMajesty {
		return nil, gamedata.GameError(4)
	}

	targetUid := common.UUid(arg2.Uid)
	targetPlayer, _ := playerMgr.loadPlayer(targetUid)
	if targetPlayer == nil {
		return nil, gamedata.GameError(5)
	}

	tCty := targetPlayer.getCity()
	if tCty == nil {
		return nil, gamedata.GameError(6)
	}

	if tCty.getCountry() != cry {
		return nil, gamedata.GameError(7)
	}

	return nil, cry.appointJob(p, targetPlayer, arg2.Job, common.UUid(arg2.OldUid))
}

func rpc_C2S_RecallJob(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	cry := p.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.RecallJobArg)
	if arg2.Job == pb.CampaignJob_YourMajesty {
		return nil, gamedata.GameError(4)
	}

	targetUid := common.UUid(arg2.Uid)
	targetPlayer, _ := playerMgr.loadPlayer(targetUid)
	if targetPlayer == nil {
		return nil, gamedata.GameError(5)
	}

	if p.getCountry() != cry {
		return nil, gamedata.GameError(7)
	}

	glog.Infof("rpc_C2S_RecallJob, uid=%d, targetUid=%d, job=%s", uid, targetUid, arg2.Job)
	return nil, cry.recallJob(p, targetPlayer, arg2.Job, p == targetPlayer, true, true)
}

func rpc_C2S_FetchCampaignNotice(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	nb := noticeMgr.getPlayerNotice(uid)
	ns := nb.fetchNotices()
	reply := &pb.CampaignNoticeInfo{}
	for _, n := range ns {
		reply.Notices = append(reply.Notices, n.packMsg())
	}
	return reply, nil
}

func rpc_G2CA_CityCapitalInjection(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.CityCapitalInjectionArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	if cty.getCountry() == nil {
		return nil, gamedata.GameError(3)
	}

	cty.capitalInjection(p, int(arg2.Gold))
	glog.Infof("rpc_G2CA_CityCapitalInjection, uid=%d, cityID=%d, gold=%d", uid, arg2.CityID, arg2.Gold)
	return &pb.CityCapitalInjectionReply{
		CurGold: int32(cty.getResource(resGold)),
	}, nil
}

func rpc_G2CA_MoveCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(3)
	}

	if p.getCityID() != p.getLocationCityID() {
		return nil, gamedata.GameError(4)
	}

	arg2 := arg.(*pb.TargetCity)
	cityID := cty.getCityID()
	if cityID == int(arg2.CityID) {
		return nil, gamedata.GameError(5)
	}

	cry := cty.getCountry()
	targetCity := cityMgr.getCity(int(arg2.CityID))
	targetCry := targetCity.getCountry()
	if cry != nil && cry != targetCry {
		return nil, gamedata.GameError(6)
	}

	p.moveCity(int(arg2.CityID))
	return &pb.MoveCityRelpy{
		NeedGold: cty.getCountryID() > 0,
	}, nil
}

func rpc_G2CA_GetMyCountry(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}
	return &pb.GetMyCountryReply{
		CountryID: p.getCountryID(),
	}, nil
}

func rpc_C2S_QuitCountry(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	cry := p.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CampaignTargetPlayer)
	err := p.quitCountry(common.UUid(arg2.Uid))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func rpc_C2S_KickOutCityPlayer(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if p.getCityJob() == pb.CampaignJob_UnknowJob {
		return nil, gamedata.GameError(2)
	}

	targetPlayer, _ := playerMgr.loadPlayer(common.UUid(arg.(*pb.CampaignTargetPlayer).Uid))
	if targetPlayer == nil {
		return nil, gamedata.GameError(3)
	}

	if targetPlayer.isKickOut() || targetPlayer.getCity() != p.getCity() ||
		targetPlayer.getCityJob() != pb.CampaignJob_UnknowJob ||
		targetPlayer.getCountryJob() != pb.CampaignJob_UnknowJob {
		return nil, gamedata.GameError(4)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(5)
	}

	glog.Infof("rpc_C2S_KickOutCityPlayer uid=%d, targetUid=%d", uid, targetPlayer.getUid())
	targetPlayer.setKickOut(true)
	noticeMgr.sendNoticeToPlayer(targetPlayer.getUid(), pb.CampaignNoticeType_KickOutNt, p.getCityJob(), p.getName(),
		targetPlayer.getCityID())
	return nil, nil
}

func rpc_C2S_CancelKickOutCityPlayer(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	if p.getCityJob() == pb.CampaignJob_UnknowJob {
		return nil, gamedata.GameError(2)
	}

	targetPlayer, _ := playerMgr.loadPlayer(common.UUid(arg.(*pb.CampaignTargetPlayer).Uid))
	if targetPlayer == nil {
		return nil, gamedata.GameError(3)
	}

	if targetPlayer.getCity() != p.getCity() {
		return nil, gamedata.GameError(4)
	}

	glog.Infof("rpc_C2S_CancelKickOutCityPlayer uid=%d, targetUid=%d", uid, targetPlayer.getUid())
	targetPlayer.setKickOut(false)
	return nil, nil
}

func rpc_G2CA_CountryModifyName(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	cry := p.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getCountryJob() != pb.CampaignJob_YourMajesty {
		return nil, gamedata.GameError(3)
	}

	if cry.getCityAmount() < 10 {
		return nil, gamedata.GameError(4)
	}

	arg2 := arg.(*pb.CountryModifyNameArg)
	//cry.setFlag(arg2.Flag)
	cry.setName(arg2.Name)
	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_COUNTRY_NAME, &pb.UpdateCountryNameArg{
		CountryID: cry.getID(),
		//Flag: arg2.Flag,
		Name: arg2.Name,
	})
	return nil, nil
}

func rpc_C2S_CancelPublishMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	cityJob := p.getCityJob()
	if cityJob != pb.CampaignJob_Prefect && cityJob != pb.CampaignJob_DuWei {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CancelPublishMissionArg)
	cty.cancelMission(arg2.Type, int(arg2.TransportTargetCity))
	glog.Infof("rpc_C2S_CancelPublishMission uid=%d, msType=%s, TransportTargetCity=%d", uid, arg2.Type,
		arg2.TransportTargetCity)
	return &pb.CancelPublishMissionReply{
		Gold:   int32(cty.getResource(resGold)),
		Forage: int32(cty.getResource(resForage)),
	}, nil
}

func rpc_C2S_Autocephaly(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	if p.getCityJob() != pb.CampaignJob_Prefect {
		return nil, gamedata.GameError(3)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(4)
	}

	cry := cty.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(5)
	}

	yourMajesty := cry.getYourMajesty()
	if yourMajesty != nil && yourMajesty.getCityID() == cty.getCityID() {
		return nil, gamedata.GameError(12)
	}

	return nil, cty.autocephaly(false)
}

func rpc_C2S_SupportCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	targetCty := cityMgr.getCity(int(arg.(*pb.TargetCity).CityID))
	if targetCty == nil || cty == targetCty {
		return nil, gamedata.GameError(2)
	}

	cry := targetCty.getCountry()
	if cry == nil || cry != p.getCountry() {
		return nil, gamedata.GameError(3)
	}

	if !warMgr.isReadyWar() {
		return nil, gamedata.GameError(4)
	}

	p.setCity(p.getCityID(), targetCty.getCityID(), true)
	return nil, nil
}

/*
func rpc_G2CA_CampaignExpedition(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.CampaignExpeditionArg)
	lcty := p.getLocationCity()
	if lcty == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getTeam() != nil {
		return nil, gamedata.GameError(7)
	}

	if int(arg2.CityPath[0]) != lcty.getCityID() {
		return nil, gamedata.GameError(3)
	}

	myCountryID := p.getCountryID()
	pathLen := len(arg2.CityPath)
	var fighterData *pb.FighterData
	if myCountryID != cityMgr.getCity(int(arg2.CityPath[pathLen-1])).getCountryID() {
		// 最终目标城是敌人，需要取我方队伍
		reply, err := agent.CallBackend(pb.MessageID_L2G_GET_PVP_FIGHTER_DATA, nil)
		if err != nil {
			return nil, err
		}
		fighterData = reply.(*pb.FighterData)
	}

	roadGameData := gamedata.GetGameData(consts.Road).(*gamedata.RoadGameData)
	var cityPath []int
	var distance int
	var paths []*roadPath
	for i, ctyID := range arg2.CityPath {
		cityID := int(ctyID)
		cty := cityMgr.getCity(cityID)
		if cty == nil {
			return nil, gamedata.GameError(4)
		}
		cityPath = append(cityPath, cityID)

		if i <= pathLen - 2 {
			if myCountryID != cty.getCountryID() {
				return nil, gamedata.GameError(5)
			}

			rs, ok := roadGameData.City2Road[cityID]
			if !ok {
				return nil, gamedata.GameError(6)
			}
			cityID2 := int(arg2.CityPath[i+1])
			r, ok := rs[cityID2]
			if !ok {
				return nil, gamedata.GameError(6)
			}

			distance += r.Distance
			paths = append(paths, newRoadPath(distance, cityID2, r))
		} else {
			break
		}
	}

	if arg2.Forage > 0 {
		paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
		forageAmount := int(arg2.Forage) * paramGameData.BundleForage
		if lcty.getResource(resForage) < forageAmount {
			return nil, gamedata.GameError(8)
		}
		lcty.modifyResource(resForage, - forageAmount)
		p.modifyForage(forageAmount)
	}

	t := newTeam(p, cityPath, paths, fighterData)
	p.setTeam(t)
	return t.packMsg(), nil
}
*/

func rpc_G2CA_CampaignOnBattleEnd(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || warMgr.isUnified() || warMgr.isPause() {
		return nil, nil
	}

	t := p.getTeam()
	if t == nil {
		return nil, nil
	}

	t.onBattleEnd(arg.(*pb.CampaignBattleEnd).IsWin)
	return nil, nil
}

func rpc_C2S_DefCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil || t.type_ != ttDefCity || t.fighterData == nil {
		return nil, gamedata.GameError(2)
	}

	var cardIDs []uint32
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range t.fighterData.HandCards {
		cardData := poolGameData.GetCardByGid(card.GCardID)
		cardIDs = append(cardIDs, cardData.CardID)
	}

	reply, err := agent.CallBackend(pb.MessageID_L2G_GET_PVP_FIGHTER_DATA, &pb.GetFighterDataArg{
		CardIDs: cardIDs,
	})
	if err != nil {
		return nil, err
	}
	t.fighterData = reply.(*pb.FighterData)
	t.beginDefCity()
	return nil, nil
}

func rpc_C2S_CancelDefCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil || t.type_ != ttDefCity {
		return nil, gamedata.GameError(2)
	}

	sceneMgr.delTeam(t)

	if p.isOnline() {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, p.getMyState())
	}
	return nil, nil
}

func rpc_C2S_BeginAttackCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil || t.getState() != pb.TeamState_CanAttackCityTS {
		return nil, gamedata.GameError(2)
	}

	targetCty := t.getTargetCity()
	if targetCty == nil || targetCty.getCountryID() == p.getCountryID() {
		return nil, gamedata.GameError(3)
	}
	t.beginAttack(true)
	return nil, nil
}

func rpc_C2S_FetchAutocephalyInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	aa := createCountryMgr.getCityAutocephaly(p.getCityID())
	if aa == nil {
		return &pb.AutocephalyInfo{}, nil
	} else {
		return aa.packMsg(), nil
	}
}

func rpc_C2S_VoteAutocephaly(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	aa := createCountryMgr.getCityAutocephaly(p.getCityID())
	if aa == nil {
		return nil, gamedata.GameError(2)
	}

	arg2 := arg.(*pb.VoteAutocephalyArg)
	nb := noticeMgr.getPlayerNotice(uid)
	n := nb.getNoticeByID(int(arg2.NoticeID))
	if n == nil {
		return nil, gamedata.GameError(3)
	}

	if n.isOp() || n.getAutocephalyPlayer() != aa.getUid() || n.getAutocephalyCity() != aa.getCityID() {
		return nil, gamedata.GameError(4)
	}
	n.op()
	aa.vote(uid, arg2.IsAgree)
	glog.Infof("rpc_C2S_VoteAutocephaly uid=%d, aaUid=%d, aaCityID=%d, isAgree=%v", uid, aa.getUid(), aa.getCityID(),
		arg2.IsAgree)
	return nil, nil
}

func rpc_G2CA_FetchCampaignPlayerInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return &pb.GCampaignPlayerInfo{}, nil
	}

	noticeMgr.onPlayerLogin(p)
	p.agent = agent
	notices := noticeMgr.getPlayerNotice(uid).getNotices()

	cry := p.getCountry()
	var countryName string
	if cry != nil {
		countryName = cry.getName()
		if !cry.isKingdom() {
			yourMajesty := cry.getYourMajesty()
			if yourMajesty != nil {
				countryName = yourMajesty.getName()
			}
		}
	}
	reply := &pb.GCampaignPlayerInfo{
		CountryID:   p.getCountryID(),
		CityID:      int32(p.getCityID()),
		CityJob:     p.getCityJob(),
		CountryJob:  p.getCountryJob(),
		CountryName: countryName,
	}

	for _, n := range notices {
		reply.Notices = append(reply.Notices, n.packMsg())
	}
	return reply, nil
}

func rpc_C2S_CountryModifyFlag(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cry := p.getCountry()
	if cry == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getCountryJob() != pb.CampaignJob_YourMajesty {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CountryModifyFlagArg)
	cry.setFlag(arg2.Flag)
	campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_COUNTRY_FLAG, &pb.UpdateCountryFlagArg{
		CountryID: cry.getID(),
		Flag:      arg2.Flag,
	})
	return nil, nil
}

func rpc_C2S_FetchCityCapitalInjectionHistory(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchCityCapitalInjectionArg)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}

	return &pb.CityCapitalInjectionHistory{
		Records: cty.getCapitalInjections(int(arg2.Page)),
	}, nil
}

func rpc_C2S_UpdateCityNotice(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getCityJob() != pb.CampaignJob_Prefect {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.CityNotice)
	notice, _, _, _ := wordfilter.ContainsDirtyWords(arg2.Notice, true)
	cty.setNotice(notice)
	return nil, nil
}

func rpc_C2S_FetchCityNotice(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetCity)
	cty := cityMgr.getCity(int(arg2.CityID))
	if cty == nil {
		return nil, gamedata.GameError(1)
	}
	return &pb.CityNotice{
		Notice: cty.getNotice(),
	}, nil
}

func rpc_C2S_PublishMilitaryOrders(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	cry := p.getCountry()
	cty := p.getCity()
	if cry == nil || cty == nil || cty.getCountryID() != cry.getID() {
		return nil, gamedata.GameError(3)
	}

	if p.getCityJob() != pb.CampaignJob_Prefect {
		return nil, gamedata.GameError(4)
	}

	var cityPath []int
	arg2 := arg.(*pb.PublishMilitaryOrderArg)
	for _, cityID := range arg2.CityPath {
		cityPath = append(cityPath, int(cityID))
	}
	reply, err := cty.publishMilitaryOrder(arg2.Type, int(arg2.Forage), int(arg2.Amount), cityPath)
	if err != nil {
		return nil, err
	}

	glog.Infof("rpc_C2S_PublishMilitaryOrders, uid=%d, cityID=%d, moType=%s, cityPath=%v, forage=%d, amount=%d",
		uid, cty.getCityID(), arg2.Type, cityPath, arg2.Forage, arg2.Amount)

	return reply, nil
}

func rpc_C2S_CancelMilitaryOrder(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	cty := p.getCity()
	if cty == nil {
		return nil, gamedata.GameError(2)
	}

	if p.getCityJob() != pb.CampaignJob_Prefect {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.TargetMilitaryOrder)
	cty.cancelMilitaryOrder(arg2.Type, int(arg2.TargetCity))

	glog.Infof("rpc_C2S_CancelMilitaryOrder, uid=%d, cityID=%d, moType=%s, targetCityID=%d", uid, cty.getCityID(),
		arg2.Type, arg2.TargetCity)

	return &pb.CancelMilitaryOrderReply{
		Forage: int32(cty.getResource(resForage)),
	}, nil
}

func rpc_C2S_FetchMilitaryOrders(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.TargetCity)
	cityID := int(arg2.CityID)
	if p.getCityJob() != pb.CampaignJob_Prefect || cityID != p.getCityID() {
		cityID = p.getLocationCityID()
	}

	cty := cityMgr.getCity(cityID)
	if cty == nil {
		return nil, gamedata.GameError(2)
	}
	return cty.getMilitaryOrdersInfo(), nil
}

func rpc_G2CA_AcceptMilitaryOrder(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	if warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	cty := p.getLocationCity()
	if cty == nil || cty.getCountryID() != p.getCountryID() {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.AcceptMilitaryOrderArg)
	return cty.acceptMilitaryOrder(p, arg2.Type, int(arg2.TargetCity), arg2.CardIDs)
}

func rpc_C2S_MyTeamMarch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() || warMgr.isUnified() || warMgr.isPause() {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil || t.type_ != ttExpedition || t.getState() != pb.TeamState_FieldBattleEndTS {
		return nil, gamedata.GameError(2)
	}

	var cardIDs []uint32
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range t.fighterData.HandCards {
		cardData := poolGameData.GetCardByGid(card.GCardID)
		cardIDs = append(cardIDs, cardData.CardID)
	}

	reply, err := agent.CallBackend(pb.MessageID_L2G_GET_PVP_FIGHTER_DATA, &pb.GetFighterDataArg{
		CardIDs: cardIDs,
	})
	if err != nil {
		return nil, err
	}
	t.fighterData = reply.(*pb.FighterData)

	t.setState(pb.TeamState_NormalTS)
	agent.PushClient(pb.MessageID_S2C_UPDATE_MY_TEAM_STATE, &pb.UpdateMyTeamStateArg{
		State: pb.TeamState_NormalTS,
	})
	return nil, nil
}

func rpc_C2S_MyTeamRetreat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil {
		return nil, gamedata.GameError(2)
	}

	return t.retreat(false), nil
}

func rpc_C2S_CampaignBackCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t != nil {
		return nil, gamedata.GameError(2)
	}

	if p.getCity() == nil {
		return nil, gamedata.GameError(3)
	}

	p.setCity(p.getCityID(), p.getCityID(), true)
	return nil, nil
}

func rpc_G2CA_GmCommand(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.GmCommand)
	commandInfo := strings.Split(arg2.Command, " ")
	switch commandInfo[1] {
	case "next":
		warMgr.nextState()
	case "addcforage":
		amount, _ := strconv.Atoi(commandInfo[2])
		lcity := p.getLocationCity()
		if lcity != nil {
			lcity.modifyResource(resForage, float64(amount))
		}
	case "addcdef":
		amount, _ := strconv.Atoi(commandInfo[2])
		lcity := p.getLocationCity()
		if lcity != nil {
			lcity.modifyResource(resDefense, float64(amount))
		}
	case "addforage":
		amount, _ := strconv.Atoi(commandInfo[2])
		p.modifyForage(amount)
	case "addco":
		amount, _ := strconv.Atoi(commandInfo[2])
		p.setContribution(p.getContribution()+float64(amount), true)
	default:
		return nil, gamedata.GameError(2)
	}
	return nil, nil
}

func rpc_G2CA_EscapedFromJail(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	if !p.isCaptive() {
		return nil, nil
	}

	reply := &pb.TargetCity{}
	p.setCaptive(nil, true)
	cry := p.getCountry()
	if cry == nil {
		p.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
		p.setCity(0, 0, true)
		p.setCountryID(0, true)
	} else {
		lCityID := p.getLocationCityID()
		cty := p.getCity()
		if cty == nil || cty.getCountryID() != p.getCountryID() {
			cty = cry.randomCity()
		}
		p.setCity(cty.getCityID(), cty.getCityID(), true)
		reply.CityID = int32(cty.getCityID())

		noticeMgr.sendNoticeToCity(lCityID, pb.CampaignNoticeType_EscapedNt, p.getName())
		noticeMgr.sendNoticeToCity(cty.getCityID(), pb.CampaignNoticeType_EscapedReturnNt, p.getName())
	}

	glog.Infof("rpc_G2CA_EscapedFromJail, uid=%d, cityID=%d, lCityID=%d, countryID=%d, cityJob=%s, countryJob=%s",
		uid, p.getCityID(), p.getLocationCityID(), p.getCountryID(), p.getCityJob(), p.getCountryJob())
	return reply, nil
}

func rpc_C2S_CampaignSurrender(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	if !p.isCaptive() {
		return nil, nil
	}

	oldCryJob := p.getCountryJob()
	oldCry := p.getCountry()

	defer func() {
		if oldCry == nil {
			return
		}

		if oldCryJob == pb.CampaignJob_YourMajesty {
			newYourMajesty := oldCry.chooseNewYourMajesty()
			if newYourMajesty != nil {
				oldCry.changeYourMajesty(newYourMajesty, p)
			}
		}

		oldCry.checkPlayerAmount()
	}()

	p.setCaptive(nil, true)
	p.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
	lcty := p.getLocationCity()
	if lcty == nil || lcty.getCountry() == nil {
		p.setCity(0, 0, true)
		p.setCountryID(0, true)
		return nil, nil
	}

	cityID := lcty.getCityID()
	oldCityID := p.getCityID()
	p.subContribution(p.getMaxContribution()*0.1, true)
	p.setCity(cityID, cityID, true)
	p.setCountryID(lcty.getCountryID(), true)

	noticeMgr.sendNoticeToCity(cityID, pb.CampaignNoticeType_SurrenderNt, p.getName())
	if oldCityID > 0 {
		noticeMgr.sendNoticeToCity(oldCityID, pb.CampaignNoticeType_BetrayNt, p.getName())
	}

	glog.Infof("rpc_C2S_CampaignSurrender, uid=%d, cityID=%d, lCityID=%d, countryID=%d, cityJob=%s, countryJob=%s",
		uid, p.getCityID(), p.getLocationCityID(), p.getCountryID(), p.getCityJob(), p.getCountryJob())

	return nil, nil
}

func rpc_C2S_SurrenderCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil || p.isCaptive() {
		return nil, gamedata.GameError(1)
	}

	if !warMgr.isNormalState() {
		return nil, gamedata.GameError(2)
	}

	cty := p.getCity()
	cry := p.getCountry()
	if cty == nil || cry == nil || cty.getCountryID() != cry.getID() {
		return nil, gamedata.GameError(3)
	}

	arg2 := arg.(*pb.SurrenderCityArg)
	targetCry := countryMgr.getCountry(arg2.CountryID)
	if targetCry == nil || cry.getID() == arg2.CountryID {
		return nil, gamedata.GameError(4)
	}

	if p.getCountryJob() == pb.CampaignJob_YourMajesty {
		return nil, cry.surrender(p, targetCry)
	} else if p.getCityJob() == pb.CampaignJob_Prefect {
		return nil, cty.surrender(p, targetCry, false)
	} else {
		return nil, gamedata.GameError(5)
	}
}

func rpc_C2S_FetchContribution(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	p := playerMgr.getPlayer(uid)
	if p == nil {
		return &pb.ContributionReply{}, nil
	} else {
		p.agent = agent
		return &pb.ContributionReply{
			Contribution: int32(p.getContribution()),
		}, nil
	}
}

func rpc_G2CA_GetCampaignInfo(_ *network.Session, arg interface{}) (interface{}, error) {
	return campaignMgr.getCampaignInfo(), nil
}

func rpc_G2CA_IsInCampaignMatch(_ *network.Session, arg interface{}) (interface{}, error) {
	if !warMgr.isInWar() {
		return nil, gamedata.GameError(1)
	}

	uid := common.UUid(arg.(*pb.TargetPlayer).Uid)
	p := playerMgr.getPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	t := p.getTeam()
	if t == nil || t.type_ == ttSupport {
		return nil, gamedata.GameError(1)
	}

	state := t.getState()
	if state == pb.TeamState_NormalTS || state == pb.TeamState_AttackingCityTS {
		return nil, nil
	}

	return nil, gamedata.GameError(1)
}

func rpc_G2CA_FetchCampaignTargetPlayerInfo(_ *network.Session, arg interface{}) (interface{}, error) {
	uid := common.UUid(arg.(*pb.TargetPlayer).Uid)
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return &pb.GCampaignPlayerInfo{}, nil
	}

	cry := p.getCountry()
	var countryName string
	if cry != nil {
		countryName = cry.getName()
		if !cry.isKingdom() {
			yourMajesty := cry.getYourMajesty()
			if yourMajesty != nil {
				countryName = yourMajesty.getName()
			}
		}
	}
	return &pb.GCampaignPlayerInfo{
		CountryID:   p.getCountryID(),
		CityID:      int32(p.getCityID()),
		CityJob:     p.getCityJob(),
		CountryJob:  p.getCountryJob(),
		CountryName: countryName,
	}, nil
}

func rpc_G2CA_ModifyContribution(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.ModifyContributionArg)
	uid := common.UUid(arg2.Uid)
	p, _ := playerMgr.loadPlayer(uid)
	if p == nil {
		return nil, gamedata.GameError(1)
	}

	amount := float64(arg2.Amount)
	old := p.getContribution()
	if amount < 0 && old < -amount {
		return nil, gamedata.GameError(2)
	}

	p.setContribution(old+amount, true)
	return nil, nil
}

func rpc_L2L_UpdateSimplePlayer(_ *network.Session, arg interface{}) (interface{}, error) {
	playerMgr.onPlayerInfoUpdate(arg.(*pb.UpdateSimplePlayerArg))
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_FETCH_CAMPAIGN_INFO, rpc_G2CA_FetchCampaignInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_SETTLE_CITY, rpc_G2CA_SettleCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CITY_DATA, rpc_C2S_FetchCityData)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_CREATE_COUNTRY, rpc_G2CA_CreateCountry)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LEAVE_CAMPAIGN_SCENE, rpc_C2S_LeaveCampaignScene)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_APPLY_CREATE_COUNTRY_INFO, rpc_C2S_FetchApplyCreateCountryInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_APPLY_CREATE_COUNTRY_PLAYERS, rpc_C2S_FetchApplyCreateCountryPlayers)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CAMPAIGN_MISSION_INFO, rpc_C2S_FetchCampaignMissionInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_ACCEPT_CAMPAIGN_MISSION, rpc_G2CA_AcceptCampaignMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_CANCEL_CAMPAIGN_MISSION, rpc_G2CA_CancelCampaignMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_GET_CAMPAIGN_MISSION_REWARD, rpc_G2CA_GetCampaignMissionReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CAMPAIGN_PUBLISH_MISSION, rpc_C2S_CampaignPublishMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_PATROL_CITY, rpc_C2S_PatrolCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CITY_PLAYERS, rpc_C2S_FetchCityPlayers)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_IN_CITY_PLAYERS, rpc_C2S_FetchInCityPlayers)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SET_FORAGE_PRICE, rpc_C2S_SetForagePrice)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_FORAGE_PRICE, rpc_C2S_FetchForagePrice)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_ALL_CITY_PLAYER_AMOUNT, rpc_C2S_FetchAllCityPlayerAmount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_COUNTRY_JOB_PLAYERS, rpc_C2S_FetchCountryJobPlayers)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_COUNTRY_PLAYERS, rpc_C2S_FetchCountryPlayers)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_APPOINT_JOB, rpc_C2S_AppointJob)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_RECALL_JOB, rpc_C2S_RecallJob)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CAMPAIGN_NOTICE, rpc_C2S_FetchCampaignNotice)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_CITY_CAPITAL_INJECTION, rpc_G2CA_CityCapitalInjection)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_MOVE_CITY, rpc_G2CA_MoveCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_GET_MY_COUNTRY, rpc_G2CA_GetMyCountry)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_QUIT_COUNTRY, rpc_C2S_QuitCountry)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_KICK_OUT_CITY_PLAYER, rpc_C2S_KickOutCityPlayer)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_KICK_OUT_CITY_PLAYER, rpc_C2S_CancelKickOutCityPlayer)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_COUNTRY_MODIFY_NAME, rpc_G2CA_CountryModifyName)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_PUBLISH_MISSION, rpc_C2S_CancelPublishMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_AUTOCEPHALY, rpc_C2S_Autocephaly)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_CAMPAIGN_ON_BATTLE_END, rpc_G2CA_CampaignOnBattleEnd)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DEF_CITY, rpc_C2S_DefCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BEGIN_ATTACK_CITY, rpc_C2S_BeginAttackCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_AUTOCEPHALY_INFO, rpc_C2S_FetchAutocephalyInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_VOTE_AUTOCEPHALY, rpc_C2S_VoteAutocephaly)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_FETCH_CAMPAIGN_PLAYER_INFO, rpc_G2CA_FetchCampaignPlayerInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_COUNTRY_MODIFY_FLAG, rpc_C2S_CountryModifyFlag)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CITY_CAPITAL_INJECTION_HISTORY, rpc_C2S_FetchCityCapitalInjectionHistory)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_CITY_NOTICE, rpc_C2S_UpdateCityNotice)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CITY_NOTICE, rpc_C2S_FetchCityNotice)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_PUBLISH_MILITARY_ORDERS, rpc_C2S_PublishMilitaryOrders)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_MILITARY_ORDER, rpc_C2S_CancelMilitaryOrder)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_MILITARY_ORDERS, rpc_C2S_FetchMilitaryOrders)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_ACCEPT_MILITARY_ORDER, rpc_G2CA_AcceptMilitaryOrder)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_DEF_CITY, rpc_C2S_CancelDefCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MY_TEAM_MARCH, rpc_C2S_MyTeamMarch)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MY_TEAM_RETREAT, rpc_C2S_MyTeamRetreat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CAMPAIGN_BACK_CITY, rpc_C2S_CampaignBackCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_GM_COMMAND, rpc_G2CA_GmCommand)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2CA_ESCAPED_FROM_JAIL, rpc_G2CA_EscapedFromJail)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CAMPAIGN_SURRENDER, rpc_C2S_CampaignSurrender)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CITY_CAPTIVES, rpc_C2S_FetchCityCaptives)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SURRENDER_CITY, rpc_C2S_SurrenderCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CONTRIBUTION, rpc_C2S_FetchContribution)

	logic.RegisterRpcHandler(pb.MessageID_G2CA_GET_CAMPAIGN_INFO, rpc_G2CA_GetCampaignInfo)
	logic.RegisterRpcHandler(pb.MessageID_G2CA_IS_IN_CAMPAIGN_MATCH, rpc_G2CA_IsInCampaignMatch)
	logic.RegisterRpcHandler(pb.MessageID_G2CA_FETCH_CAMPAIGN_TARGET_PLAYER_INFO, rpc_G2CA_FetchCampaignTargetPlayerInfo)
	logic.RegisterRpcHandler(pb.MessageID_G2CA_MODIFY_CONTRIBUTION, rpc_G2CA_ModifyContribution)
	logic.RegisterRpcHandler(pb.MessageID_L2L_UPDATE_SIMPLE_PLAYER, rpc_L2L_UpdateSimplePlayer)
}
