package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"math/rand"
	"sort"
	"time"
)

var playerMgr = &playerMgrSt{}

type playerMgrSt struct {
	allPlayers     map[common.UUid]*player
	loadingPlayer  map[common.UUid]chan struct{}
	onlinePlayers  map[common.UUid]*player
	kickOutPlayers map[common.UUid]*player
}

func (pm *playerMgrSt) initialize() {
	pm.allPlayers = map[common.UUid]*player{}
	pm.loadingPlayer = map[common.UUid]chan struct{}{}
	pm.onlinePlayers = map[common.UUid]*player{}
	pm.kickOutPlayers = map[common.UUid]*player{}

	attrs, err := attribute.LoadAll(campaignMgr.genAttrName("campaign_player"))
	if err != nil {
		panic(err)
	}
	for _, attr := range attrs {
		id, ok := attr.GetAttrID().(int64)
		if !ok {
			panic(fmt.Sprintf("wrong uid %s", attr.GetAttrID()))
		}
		uid := common.UUid(id)
		p := newPlayerByAttr(uid, attr)

		cty := p.getCity()
		lcty := p.getLocationCity()
		cry := p.getCountry()
		ccApply := p.getCcApply()
		if cty == nil && lcty == nil && cry == nil {
			// 在野玩家
			continue
		}

		pm.allPlayers[uid] = p
		var cityCountry uint32
		var playerCountry uint32
		if cty != nil {
			cityCountry = cty.getCountryID()
			cty.addPlayer(p)
			if ccApply != nil {
				cty.addCcApply(ccApply)
			}

			if p.isKickOut() {
				pm.kickOutPlayers[uid] = p
			}
		} else if ccApply != nil {
			p.delCcApply(false)
		}
		if lcty != nil {
			lcty.addInCityPlayer(p)
		}
		if cry != nil {
			playerCountry = cry.getID()
			cry.addPlayer(p)
		}

		captiveCityID := p.getCaptiveCityID()
		if captiveCityID > 0 {
			captiveCity := cityMgr.getCity(captiveCityID)
			if captiveCity == nil {
				p.setCaptive(nil, false)
			}
		}

		if cityCountry > 0 && playerCountry > 0 && cityCountry != playerCountry {
			glog.Infof("what the fuck, uid=%d, cityID=%d, cityCountry=%d, playerCountry=%d", uid, p.getCityID(),
				cityCountry, playerCountry)
		}
	}

	cityMgr.sortPlayers()
	cityMgr.sortCaPlayers()
	countryMgr.sortPlayers()

	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, pm.onPlayerLogout)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, pm.onPlayerLogout)
	eventhub.Subscribe(logic.RESTORE_AGENT_EV, pm.onRestoreAgent)

	timer.AddTicker(time.Duration(rand.Intn(20)+290)*time.Second, func() {
		pm.save(false)
	})
	timer.AddTicker(time.Minute, pm.checkKickOut)
}

func (pm *playerMgrSt) onWarEnd() {
	for _, p := range pm.allPlayers {
		p.supportCardIDs = []uint32{}
		t := p.getTeam()
		if t != nil {
			t.retreat(true)
			p.teamDisappear = nil
		}

		if !p.isCaptive() {
			p.setCity(p.getCityID(), p.getCityID(), true)
		}

		if !p.isOnline() {
			continue
		}

		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, p.getMyState())

		record := p.fetchWarEndRecord()
		if record == nil {
			continue
		}

		msg := &pb.CampaignState{
			State: pb.CampaignState_WarEnd,
		}
		msg.Arg, _ = record.Marshal()
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_STATE, msg)
	}
}

func (pm *playerMgrSt) onUnified(uid2Rank map[common.UUid]int, yourMajestyName, countryName string) {
	for _, p := range pm.allPlayers {
		p.cancelMission()
		t := p.getTeam()
		if t != nil {
			t.retreat(true)
			p.teamDisappear = nil
		}

		if !p.isCaptive() {
			p.setCity(p.getCityID(), p.getCityID(), true)
		}

		uid := p.getUid()
		utils.PlayerMqPublish(uid, pb.RmqType_UnifiedReward, &pb.RmqUnifiedReward{
			Rank:            int32(uid2Rank[uid]),
			Contribution:    int32(p.getMaxContribution() * 0.1),
			YourMajestyName: yourMajestyName,
			CountryName:     countryName,
		})

		if !p.isOnline() {
			continue
		}

		msg := &pb.CampaignState{
			State: pb.CampaignState_Unified,
			Arg:   warMgr.getUnifiedInfo(),
		}
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_STATE, msg)
	}
}

func (pm *playerMgrSt) onRestoreAgent(args ...interface{}) {
	clients := args[0].([]*gpb.PlayerClient)
	for _, cli := range clients {
		if cli.GateID <= 0 || cli.ClientID <= 0 || cli.Uid <= 0 {
			continue
		}

		p, _ := pm.loadPlayer(common.UUid(cli.Uid))
		if p == nil {
			continue
		}
		pm.onPlayerLogin(p, logic.NewPlayerAgent(cli))
		glog.Infof("onRestoreAgent %d", cli.Uid)
	}
}

func (pm *playerMgrSt) checkKickOut() {
	now := time.Now().Unix()
	for uid, p := range pm.kickOutPlayers {
		if p.isKickOut() {
			if now >= p.getKickOutTime() {
				p.quitCountry(0)
			}
		} else {
			delete(pm.kickOutPlayers, uid)
		}
	}
}

func (pm *playerMgrSt) addKickOutPlayer(p *player) {
	pm.kickOutPlayers[p.getUid()] = p
}

func (pm *playerMgrSt) delKickOutPlayer(p *player) {
	delete(pm.kickOutPlayers, p.getUid())
}

func (pm *playerMgrSt) save(isStopServer bool) {
	for _, p := range pm.allPlayers {
		p.save(isStopServer)
	}
}

func (pm *playerMgrSt) onPlayerLeaveCampaign(uid common.UUid) {
	if p, ok := pm.onlinePlayers[uid]; ok {
		delete(pm.onlinePlayers, uid)
		p.onLogout()
		p.save(false)
	}
}

func (pm *playerMgrSt) onPlayerLogout(args ...interface{}) {
	pa := args[0].(*logic.PlayerAgent)
	uid := pa.GetUid()
	if _, ok := pm.onlinePlayers[uid]; ok {
		pm.onPlayerLeaveCampaign(uid)
	}

	if p, ok := pm.allPlayers[uid]; ok {
		p.agent = nil
	}
}

func (pm *playerMgrSt) onPlayerInfoUpdate(arg *pb.UpdateSimplePlayerArg) {
	uid := common.UUid(arg.Uid)
	p := pm.getPlayer(uid)
	if p == nil {
		return
	}
	p.setName(arg.Name)
	p.setHeadImg(arg.HeadImgUrl)
	p.setHeadFrame(arg.HeadFrame)
	p.setPvpScore(int(arg.PvpScore))
}

func (pm *playerMgrSt) onPlayerLogin(p *player, agent *logic.PlayerAgent) {
	uid := p.getUid()
	if _, ok := pm.onlinePlayers[uid]; ok {
		return
	}
	pm.onlinePlayers[uid] = p
	p.onLogin(agent)
	eventhub.Publish(evPlayerLogin, p)
}

func (pm *playerMgrSt) getPlayer(uid common.UUid) *player {
	if p, ok := pm.allPlayers[uid]; ok {
		if p.isOnline() {
			noticeMgr.onPlayerLogin(p)
		}
		return p
	} else {
		return nil
	}
}

func (pm *playerMgrSt) createPlayer(uid common.UUid, name, headImg, headFrame string, pvpScore int) *player {
	p, ok := pm.allPlayers[uid]
	if ok {
		return p
	}

	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("campaign_player"), uid)
	p = newPlayerByAttr(uid, attr)
	p.setName(name)
	p.setHeadImg(headImg)
	p.setHeadFrame(headFrame)
	p.setPvpScore(pvpScore)
	pm.allPlayers[uid] = p
	attr.Save(false)
	return p
}

func (pm *playerMgrSt) loadPlayer(uid common.UUid) (*player, error) {
	p := pm.getPlayer(uid)
	if p != nil {
		return p, nil
	}

	c, ok := pm.loadingPlayer[uid]
	if ok {
		evq.Await(func() {
			<-c
		})
		return pm.allPlayers[uid], nil
	}

	attr := attribute.NewAttrMgr(campaignMgr.genAttrName("campaign_player"), uid)
	c = make(chan struct{})
	pm.loadingPlayer[uid] = c
	defer func() {
		delete(pm.loadingPlayer, uid)
		close(c)
	}()

	err := attr.Load()
	p, ok = pm.allPlayers[uid]
	if ok {
		return p, nil
	}

	if err != nil {
		return nil, err
	}

	p = newPlayerByAttr(uid, attr)
	pm.allPlayers[uid] = p
	return p, nil
}

func (pm *playerMgrSt) getPlayersByPage(players []*player, page int, isCaptive bool) []*pb.CampaignPlayer {
	var ret []*pb.CampaignPlayer
	beginIdx := page * pageAmount
	endIdx := beginIdx + pageAmount
	totalAmount := len(players)
	if beginIdx >= totalAmount {
		return ret
	}
	if endIdx > totalAmount {
		endIdx = totalAmount
	}

	for i := beginIdx; i < endIdx; i++ {
		ret = append(ret, players[i].packMsg(isCaptive))
	}
	return ret
}

type player struct {
	uid  common.UUid
	attr *attribute.AttrMgr

	// 正在进行的任务
	ms *playerMission
	// 申请创建势力
	ccApply        *createCountryApply
	myTeam         *team
	teamDisappear  *pb.MyTeamDisappear
	supportCardIDs []uint32
	agent          *logic.PlayerAgent
	// 修养
	restTimer *timer.Timer
	// 整顿
	rectifyTimer *timer.Timer
}

func newPlayerByAttr(uid common.UUid, attr *attribute.AttrMgr) *player {
	p := &player{
		uid:  uid,
		attr: attr,
	}

	msAttr := attr.GetMapAttr("ms")
	if msAttr != nil {
		p.ms = newPlayerMissionByAttr(p, msAttr)
	}

	ccApplyAttr := attr.GetMapAttr("ccApply")
	if ccApplyAttr != nil {
		p.ccApply = newCreateCountryApplyByAttr(uid, ccApplyAttr)
	}

	supportCardIDsAttr := attr.GetListAttr("supportCardIDs")
	if supportCardIDsAttr != nil {
		supportCardIDsAttr.ForEachIndex(func(index int) bool {
			p.supportCardIDs = append(p.supportCardIDs, supportCardIDsAttr.GetUInt32(index))
			return true
		})
		attr.Del("supportCardIDs")
	}

	teamDisappearAttr := p.attr.GetStr("teamDisappear")
	if teamDisappearAttr != "" {
		p.attr.Del("teamDisappear")
		p.teamDisappear = &pb.MyTeamDisappear{}
		p.teamDisappear.Unmarshal([]byte(teamDisappearAttr))
	}

	now := time.Now().Unix()
	restTimeout := p.getRestTimeout()
	rectifyTimeout := p.getRectifyTimeout()
	if restTimeout > 0 {
		t := restTimeout - now
		if t < 2 {
			t = 2
		}
		p.restTimer = timer.AfterFunc(time.Duration(t)*time.Second, p.onRestComplete)
	} else if rectifyTimeout > 0 {
		t := rectifyTimeout - now
		if t < 2 {
			t = 2
		}
		p.rectifyTimer = timer.AfterFunc(time.Duration(t)*time.Second, p.onRectifyComplete)
	}

	return p
}

func (p *player) beginRestState() {
	p.attr.SetInt64("restTimeout", time.Now().Unix()+90)
	p.restTimer = timer.AfterFunc(90*time.Second, p.onRestComplete)
	if p.isOnline() {
		arg := &pb.CampaignPlayerState{
			State: pb.CampaignPlayerState_Rest,
		}
		arg.Arg, _ = (&pb.CpStateLoadingArg{
			MaxTime:    90,
			RemainTime: 90,
		}).Marshal()
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, arg)
	}
}

func (p *player) beginRectifyState() {
	p.attr.SetInt64("rectifyTimeout", time.Now().Unix()+20)
	p.rectifyTimer = timer.AfterFunc(20*time.Second, p.onRectifyComplete)
	if p.isOnline() {
		arg := &pb.CampaignPlayerState{
			State: pb.CampaignPlayerState_Rectify,
		}
		arg.Arg, _ = (&pb.CpStateLoadingArg{
			MaxTime:    20,
			RemainTime: 20,
		}).Marshal()
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, arg)
	}
}

func (p *player) onRestComplete() {
	p.restTimer = nil
	p.attr.Del("restTimeout")

	if p.isOnline() {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, p.getMyState())
	}
}

func (p *player) onRectifyComplete() {
	p.rectifyTimer = nil
	p.attr.Del("rectifyTimeout")
	if p.myTeam == nil || !p.isOnline() {
		return
	}
	if p.isCaptive() {
		return
	}

	if p.isOnline() {
		if p.myTeam.type_ == ttDefCity {
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, &pb.CampaignPlayerState{
				State: pb.CampaignPlayerState_DefCity,
			})
		} else if p.myTeam.type_ == ttExpedition {
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, &pb.CampaignPlayerState{
				State: pb.CampaignPlayerState_Expedition,
			})
		}
	}
}

func (p *player) getRestTimeout() int64 {
	return p.attr.GetInt64("restTimeout")
}

func (p *player) getRectifyTimeout() int64 {
	return p.attr.GetInt64("rectifyTimeout")
}

// 撤退
func (p *player) onTeamDisappear(reason pb.MyTeamDisappear_ReasonEnum, t *team, disappearMsg *pb.MyTeamDisappear, needSync bool) {
	if p.myTeam == t {
		p.delTeam(t)
		if needSync {
			if p.isOnline() {
				msg := &pb.UpdateMyTeamStateArg{
					State: pb.TeamState_DisappearTS,
				}
				msg.Arg, _ = disappearMsg.Marshal()
				p.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_TEAM_STATE, msg)
			} else {
				p.teamDisappear = disappearMsg
			}
		}

		if p.isOnline() {
			state := p.getMyState()
			if state.State == pb.CampaignPlayerState_Normal {
				p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, state)
			}
		}
	}
}

func (p *player) fetchTeamDisappear() *pb.MyTeamDisappear {
	msg := p.teamDisappear
	p.teamDisappear = nil
	return msg
}

func (p *player) modifyForage(val int) {
	if p.myTeam != nil {
		p.myTeam.modifyForage(val)
	}
}

func (p *player) getForage() int {
	if p.myTeam != nil {
		return p.myTeam.getForage()
	} else {
		return 0
	}
}

func (p *player) getTeam() *team {
	return p.myTeam
}

func (p *player) setTeam(t *team) {
	p.myTeam = t
	p.attr.SetInt("team", t.getID())
}

func (p *player) initTeam(t *team) bool {
	tid := p.attr.GetInt("team")
	if tid != t.getID() {
		return false
	}
	p.myTeam = t
	return true
}

func (p *player) delTeam(t *team) {
	if p.myTeam == t {
		p.myTeam = nil
		p.attr.Del("team")

		if p.rectifyTimer != nil {
			p.rectifyTimer.Cancel()
			p.rectifyTimer = nil
			p.attr.Del("rectifyTimeout")
		}
	}
}

func (p *player) quitCountry(newYourMajestyUid common.UUid) error {
	var newYourMajesty *player
	lastCountryID := p.getLastCountryID()
	cry := p.getCountry()
	job := p.getCountryJob()
	cityJob := p.getCityJob()
	cty := p.getCity()
	cityID := p.getCityID()
	p.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
	p.setCity(0, 0, true)
	if lastCountryID <= 0 {
		p.subContribution(p.getMaxContribution()*0.1, true)
	}

	if job == pb.CampaignJob_YourMajesty {
		newYourMajesty = playerMgr.getPlayer(newYourMajestyUid)
		if newYourMajesty == nil || newYourMajesty.getCountryID() != cry.getID() {
			newYourMajesty = cry.chooseNewYourMajesty()
		}
	} else {

		if job != pb.CampaignJob_UnknowJob || cityJob == pb.CampaignJob_Prefect {
			yourMajesty := cry.getYourMajesty()
			if yourMajesty != nil {
				noticeMgr.sendNoticeToPlayer(yourMajesty.getUid(), pb.CampaignNoticeType_ResignNt, job, p.getName(),
					cityID)
			}
		}

		if cityJob != pb.CampaignJob_UnknowJob && cityJob != pb.CampaignJob_Prefect && cty != nil {
			prefect := cty.getPrefect()
			if prefect != nil {
				noticeMgr.sendNoticeToPlayer(prefect.getUid(), pb.CampaignNoticeType_ResignNt, cityJob, p.getName(),
					cityID)
			}
		}
	}

	if newYourMajesty != nil {
		cry.changeYourMajesty(newYourMajesty, p)
	} else {
		cry.checkPlayerAmount()
	}
	p.setCountryID(0, true)

	glog.Infof("quitCountry uid=%d", p.getUid())

	return nil
}

func (p *player) isCaptive() bool {
	return p.attr.GetInt("captiveCity") > 0
}

func (p *player) getCaptiveCityID() int {
	return p.attr.GetInt("captiveCity")
}

func (p *player) getCaptiveTimeout() int64 {
	return p.attr.GetInt64("captiveTimeout")
}

func (p *player) setCaptiveTimeout(t int64) {
	p.attr.SetInt64("captiveTimeout", t)
}

func (p *player) setCaptive(cty *city, needSync bool) {
	oldCityID := p.attr.GetInt("captiveCity")
	if cty != nil {
		if oldCityID == cty.getCityID() {
			return
		}
		p.attr.SetInt("captiveCity", cty.getCityID())
		p.setCaptiveTimeout(time.Now().Unix() + captiveTimeout)
		p.setKickOut(false)
		if needSync && p.isOnline() {
			state := &pb.CampaignPlayerState{
				State: pb.CampaignPlayerState_BeCaptive,
			}
			remainTime := p.getCaptiveTimeout() - time.Now().Unix()
			if remainTime < 0 {
				remainTime = 0
			}
			state.Arg, _ = (&pb.CpStateBeCaptiveArg{RemainTime: int32(remainTime)}).Marshal()
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, state)
		}

		cty.addInCityPlayer(p)
		glog.Infof("player setCaptive, uid=%d, cityID=%d", p.getUid(), cty.getCityID())
	} else if oldCityID > 0 {
		p.attr.Del("captiveCity")
		if needSync && p.isOnline() {
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_PLAYER_STATE, p.getMyState())
		}

		oldCity := cityMgr.getCity(oldCityID)
		if oldCity != nil {
			oldCity.delCaptive(p)
		}

		lCity := p.getLocationCity()
		if lCity == nil {
			lCity = oldCity
		}

		if lCity != nil {
			lCity.addInCityPlayer(p)
		}

		glog.Infof("player usetCaptive, uid=%d, cityID=%d", p.getUid(), oldCityID)
	}
}

func (p *player) isKickOut() bool {
	return p.attr.GetInt64("kickOutTime") > 0 && p.getCityJob() == pb.CampaignJob_UnknowJob &&
		p.getCountryJob() == pb.CampaignJob_UnknowJob
}

func (p *player) getKickOutTime() int64 {
	if p.getCityJob() != pb.CampaignJob_UnknowJob || p.getCountryJob() != pb.CampaignJob_UnknowJob {
		return 0
	}
	return p.attr.GetInt64("kickOutTime")
}

func (p *player) getKickOutRemainTime() int64 {
	t := p.getKickOutTime()
	if t <= 0 {
		return 0
	}

	t = t - time.Now().Unix()
	if t < 0 {
		t = 0
	}
	return t
}

func (p *player) setKickOut(val bool) {
	if val {
		p.attr.SetInt64("kickOutTime", time.Now().Unix()+24*60*60)
		playerMgr.addKickOutPlayer(p)
	} else {
		p.attr.Del("kickOutTime")
		playerMgr.delKickOutPlayer(p)
	}
}

func (p *player) getCcApply() *createCountryApply {
	return p.ccApply
}

func (p *player) delCcApply(returnGold bool) {
	if p.ccApply != nil {
		ca := p.ccApply
		p.ccApply = nil
		if returnGold {
			utils.PlayerMqPublish(p.getUid(), pb.RmqType_Bonus, &pb.RmqBonus{
				ChangeRes: []*pb.Resource{&pb.Resource{Type: int32(consts.Gold), Amount: int32(ca.getGold())}},
			})
		}
	}
	p.attr.Del("ccApply")
}

func (p *player) addCcApply(ccApply *createCountryApply) {
	if p.ccApply != nil {
		cty := p.getCity()
		if cty != nil {
			cty.delCcApply(ccApply)
		}
	}
	p.ccApply = ccApply
	p.attr.SetMapAttr("ccApply", ccApply.attr)
}

func (p *player) onLogin(agent *logic.PlayerAgent) {
	p.agent = agent
	p.attr.SetBool("isOnline", true)
}

func (p *player) onLogout() {
	p.attr.SetBool("isOnline", false)
}

func (p *player) save(isStopServer bool) {
	if isStopServer {
		supportCardIDsAttr := attribute.NewListAttr()
		for _, cardID := range p.supportCardIDs {
			supportCardIDsAttr.AppendUInt32(cardID)
		}
		p.attr.SetListAttr("supportCardIDs", supportCardIDsAttr)

		if p.teamDisappear != nil {
			data, _ := p.teamDisappear.Marshal()
			p.attr.SetStr("teamDisappear", string(data))
		}
	}
	p.attr.Save(isStopServer)
}

func (p *player) isOnline() bool {
	return p.agent != nil && p.attr.GetBool("isOnline")
}

func (p *player) getMission() *playerMission {
	return p.ms
}

func (p *player) onAcceptMission(ms *playerMission) {
	p.attr.SetMapAttr("ms", ms.attr)
	p.ms = ms

	utils.PlayerMqPublish(p.getUid(), pb.RmqType_CampaignAcceptMission, &pb.RmqCampaignAcceptMission{
		Cards: ms.getCards(),
	})

	glog.Infof("onAcceptMission uid=%d, cityID=%d, ms=%s", p.getUid(), p.getCityID(), ms)
}

func (p *player) getUid() common.UUid {
	return p.uid
}

func (p *player) getName() string {
	return p.attr.GetStr("name")
}

func (p *player) setName(name string) {
	old := p.getName()
	if old != name {
		p.attr.SetStr("name", name)
		cry := p.getCountry()
		if cry != nil && !cry.isKingdom() && p.getCountryJob() == pb.CampaignJob_YourMajesty {
			cry.setName(createCountryMgr.genCountryName(name))
			campaignMgr.broadcastClient(pb.MessageID_S2C_UPDATE_COUNTRY_NAME, &pb.UpdateCountryNameArg{
				CountryID: cry.getID(),
				Name:      cry.getName(),
			})
		}
	}
}

func (p *player) getHeadImg() string {
	return p.attr.GetStr("headImg")
}

func (p *player) setHeadImg(headImg string) {
	p.attr.SetStr("headImg", headImg)
}

func (p *player) getHeadFrame() string {
	return p.attr.GetStr("headFrame")
}

func (p *player) setHeadFrame(headFrame string) {
	p.attr.SetStr("headFrame", headFrame)
}

func (p *player) getPvpScore() int {
	return p.attr.GetInt("pvpScore")
}

func (p *player) setPvpScore(score int) {
	p.attr.SetInt("pvpScore", score)
}

// 所属城市
func (p *player) getCity() *city {
	cityID := p.getCityID()
	if cityID <= 0 {
		return nil
	}
	return cityMgr.getCity(cityID)
}

func (p *player) getCityID() int {
	return p.attr.GetInt("city")
}

// 所在城市
func (p *player) getLocationCity() *city {
	cityID := p.getLocationCityID()
	if cityID <= 0 {
		return nil
	}
	return cityMgr.getCity(cityID)
}

func (p *player) moveCity(cityID int) {
	p.setJob(pb.CampaignJob_UnknowJob, p.getCountryJob(), true)
	ca := p.getCcApply()
	cty := p.getCity()
	oldCityID := p.getCityID()
	if ca != nil && cty != nil {
		cty.delCcApply(ca)
		p.delCcApply(false)
	}
	p.setCity(cityID, cityID, true)
	glog.Infof("moveCity uid=%d, cityID=%d, targetCityID=%d", oldCityID, cityID)
}

func (p *player) setCity(cityID, locationCityID int, needSync bool) {
	oldCityID := p.getCityID()
	if oldCityID != cityID {
		oldCty := cityMgr.getCity(oldCityID)
		if oldCty != nil {
			if p.getCityJob() != pb.CampaignJob_UnknowJob {
				p.setJob(pb.CampaignJob_UnknowJob, p.getCountryJob(), false)
			}
			oldCty.delPlayer(p)
		}

		newCty := cityMgr.getCity(cityID)
		var countryID uint32
		if newCty != nil {
			newCty.addPlayer(p)
			countryID = newCty.getCountryID()
		}

		if p.ms != nil && !(p.ms.canReward() && !p.ms.hasReward()) {
			p.cancelMission()
		}
		p.setKickOut(false)
		p.attr.SetInt("city", cityID)
		p.setCountryID(countryID, true)
	}

	oldLCityID := p.getLocationCityID()
	if oldLCityID != locationCityID {
		oldLCty := cityMgr.getCity(oldLCityID)
		if oldLCty != nil {
			oldLCty.delInCityPlayer(p)
		}

		newLCty := cityMgr.getCity(locationCityID)
		if newLCty != nil {
			newLCty.addInCityPlayer(p)
		}
		p.attr.SetInt("lcity", locationCityID)
	}

	if oldCityID != cityID || oldLCityID != locationCityID {
		if needSync && p.isOnline() {
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_CITY, &pb.UpdateMyCityArg{
				CityID:         int32(cityID),
				LocationCityID: int32(locationCityID),
			})
		}

		p.syncInfoToGame()
	}

	if oldCityID != cityID && cityID != 0 {
		eventhub.Publish(evPlayerChangeCity, p, oldCityID)
	}

	if oldCityID != cityID && cityID == locationCityID {
		if len(p.supportCardIDs) > 0 {
			p.updateSupportCards([]uint32{})
		}
	}
}

func (p *player) updateSupportCards(cardIDs []uint32) {
	p.supportCardIDs = cardIDs
	if p.isOnline() && !warMgr.isNormalState() {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_CAMPAIGN_SUPPORT_CARDS, &pb.CampaignSupportCard{
			CardIDs: p.supportCardIDs,
		})
	}
}

func (p *player) getSupportCards() []uint32 {
	return p.supportCardIDs
}

func (p *player) getLocationCityID() int {
	return p.attr.GetInt("lcity")
}

func (p *player) getCountryID() uint32 {
	return p.attr.GetUInt32("country")
}

func (p *player) setCountryID(countryID uint32, needSync bool) {
	oldCountryID := p.getCountryID()
	if oldCountryID == countryID {
		return
	}

	oldCry := countryMgr.getCountry(oldCountryID)
	if oldCry != nil {
		if p.getCountryJob() != pb.CampaignJob_UnknowJob {
			p.setJob(p.getCityJob(), pb.CampaignJob_UnknowJob, false)
		}
		oldCry.delPlayer(p)
	}

	p.attr.SetUInt32("country", countryID)
	lastCountryID := p.getLastCountryID()
	if lastCountryID != 0 {
		lastCountryID = 0
		p.setLastCountryID(0, false)
	}

	newCry := countryMgr.getCountry(countryID)
	if newCry != nil {
		newCry.addPlayer(p)
	}

	if needSync && p.isOnline() {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_COUNTRY, &pb.UpdateMyCountryArg{
			CountryID:     countryID,
			LastCountryID: lastCountryID,
		})
	}

	if countryID != 0 {
		eventhub.Publish(evPlayerChangeCountry, p)
	} else {
		p.setJob(pb.CampaignJob_UnknowJob, pb.CampaignJob_UnknowJob, true)
	}

	p.syncInfoToGame()
}

func (p *player) getLastCountryID() uint32 {
	return p.attr.GetUInt32("lcountry")
}

func (p *player) setLastCountryID(countryID uint32, needSync bool) {
	if p.getLastCountryID() == countryID {
		return
	}
	p.attr.SetUInt32("lcountry", countryID)
	if needSync && p.isOnline() {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_MY_COUNTRY, &pb.UpdateMyCountryArg{
			CountryID:     p.getCountryID(),
			LastCountryID: countryID,
		})
	}
}

func (p *player) getCountry() *country {
	countryID := p.getCountryID()
	if countryID <= 0 {
		return nil
	}
	return countryMgr.getCountry(countryID)
}

func (p *player) isApplyCreateCountry() bool {
	return p.attr.GetBool("isApplyCCry")
}

func (p *player) setApplyCreateCountry(val bool) {
	if !val {
		p.attr.Del("isApplyCCry")
	} else {
		p.attr.SetBool("isApplyCCry", val)
	}
}

func (p *player) getJob() pb.CampaignJob {
	job := p.getCountryJob()
	if job <= 0 {
		return p.getCityJob()
	}
	return job
}

func (p *player) getCountryJob() pb.CampaignJob {
	return pb.CampaignJob(p.attr.GetInt("job"))
}

func (p *player) getCityJob() pb.CampaignJob {
	return pb.CampaignJob(p.attr.GetInt("cityJob"))
}

func (p *player) setJob(cityJob, countryJob pb.CampaignJob, needSync bool) {
	oldCityJob := p.getCityJob()
	oldCountryJob := p.getCountryJob()
	if oldCityJob != cityJob {
		p.attr.SetInt("cityJob", int(cityJob))
		cty := p.getCity()
		if cty != nil {
			cty.onJobUpdate(p, oldCityJob, cityJob)
			campaignMgr.addSortCity(cty.getCityID())
		}

		countryID := p.getCountryID()
		if countryID > 0 {
			campaignMgr.addSortCountry(countryID)
		}
		eventhub.Publish(evPlayerChangeCityJob, p, oldCityJob, cityJob)
	}

	if oldCountryJob != countryJob {
		p.attr.SetInt("job", int(countryJob))
		cry := p.getCountry()
		if cry != nil {
			cry.onJobUpdate(p, oldCountryJob, countryJob)
			campaignMgr.addSortCountry(cry.getID())
		}
	}

	if cityJob != pb.CampaignJob_UnknowJob || countryJob != pb.CampaignJob_UnknowJob {
		p.setKickOut(false)
	}

	if oldCityJob != cityJob || oldCountryJob != countryJob {
		if needSync && p.isOnline() {
			p.agent.PushClient(pb.MessageID_S2C_CAMPAIGN_UPDATE_JOB, &pb.CampaignUpdateJobArg{
				CityJob:    p.getCityJob(),
				CountryJob: p.getCountryJob(),
			})
		}
		p.syncInfoToGame()
	}
}

func (p *player) getMaxContribution() float64 {
	max := p.attr.GetFloat64("mcontribution")
	cur := p.getContribution()
	if cur > max {
		max = cur
		p.setMaxContribution(max)
	}
	return max
}

func (p *player) setMaxContribution(val float64) {
	p.attr.SetFloat64("mcontribution", val)
}

func (p *player) getContribution() float64 {
	return p.attr.GetFloat64("contribution")
}

func (p *player) setContribution(val float64, needLog bool) {
	old := p.getContribution()
	if old != val {
		p.attr.SetFloat64("contribution", val)
		max := p.getMaxContribution()
		if val > max {
			max = val
			p.setMaxContribution(max)
		}

		if p.agent != nil && int32(old) != int32(val) {
			p.agent.PushClient(pb.MessageID_S2C_UPDATE_CONTRIBUTION, &pb.UpdateContributionArg{
				Contribution:    int32(val),
				MaxContribution: int32(max),
			})
		}

		if needLog {
			glog.Infof("setContribution uid=%d, old=%f, cur=%f", p.getUid(), old, val)
		}
	}
}

func (p *player) addContribution(val float64, canPer, needLog bool) {
	if val <= 0 {
		return
	}

	if warMgr.isInWar() {
		oldLcontribution := p.attr.GetFloat64("lcontribution")
		lcontribution := oldLcontribution + val
		if lcontribution >= maxWarContribution {
			lcontribution = maxWarContribution
		}
		val = lcontribution - oldLcontribution

		if val <= 0 {
			return
		}

		p.attr.SetFloat64("lcontribution", lcontribution)
	}

	p.setContribution(p.getContribution()+val, needLog)

	if canPer {
		cty := p.getCity()
		if cty != nil {
			cty.onPlayerAddContribution(p, val)
		}
		cry := p.getCountry()
		if cry != nil {
			cry.onPlayerAddContribution(p, val)
		}
	}
}

func (p *player) subContribution(val float64, needLog bool) {
	if val <= 0 {
		return
	}

	cur := p.getContribution() - val
	p.setContribution(cur, needLog)
}

// 俸禄
func (p *player) getSalary() int {
	cty := p.getCity()
	if cty == nil {
		return 0
	}

	var salary float64
	countryJob := p.getCountryJob()
	paramGameData := gamedata.GetGameData(consts.CampaignParam).(*gamedata.CampaignParamGameData)
	if countryJob != pb.CampaignJob_UnknowJob {
		cry := cty.getCountry()
		if cry != nil {
			salary += float64(cry.getSalaryByJob(countryJob))
		}
	}

	cityJob := p.getCityJob()
	if cityJob != pb.CampaignJob_UnknowJob {
		profit, _ := cty.getCurBusinessGold()
		switch cityJob {
		case pb.CampaignJob_Prefect:
			salary += paramGameData.TaishouSalary * float64(profit)
		case pb.CampaignJob_DuWei:
			salary += paramGameData.DuweiSalary * float64(profit)
		case pb.CampaignJob_FieldOfficer:
			salary += paramGameData.XiaoweiSalary * float64(profit)
		}
	}
	return int(salary)
}

func (p *player) battleThan(oth *player) bool {
	job1 := p.getJob()
	job2 := oth.getJob()
	if job1 != pb.CampaignJob_UnknowJob || job2 != pb.CampaignJob_UnknowJob {
		if job1 == pb.CampaignJob_UnknowJob {
			return false
		} else if job2 == pb.CampaignJob_UnknowJob {
			return true
		}

		if job1 < job2 {
			return true
		} else if job1 > job2 {
			return false
		}
	}

	contribution1 := p.getContribution()
	contribution2 := oth.getContribution()
	if contribution1 > contribution2 {
		return true
	} else if contribution2 > contribution1 {
		return false
	}

	score1 := p.getPvpScore()
	score2 := p.getPvpScore()
	if score1 > score2 {
		return true
	} else if score2 > score1 {
		return false
	}

	return p.getUid() < oth.getUid()
}

func (p *player) packMsg(isCaptive bool) *pb.CampaignPlayer {
	return &pb.CampaignPlayer{
		Uid:          uint64(p.getUid()),
		Name:         p.getName(),
		HeadImg:      p.getHeadImg(),
		HeadFrame:    p.getHeadFrame(),
		CityJob:      p.getCityJob(),
		CountryJob:   p.getCountryJob(),
		PvpScore:     int32(p.getPvpScore()),
		State:        p.getOthState(isCaptive),
		Contribution: int32(p.getContribution()),
	}
}

func (p *player) cancelMission() error {
	if p.ms == nil {
		return gamedata.GameError(1)
	}
	p.ms.cancel()
	ms := p.ms
	p.ms = nil
	p.attr.Del("ms")
	glog.Infof("cancelMission uid=%d, ms=%s", p.getUid(), ms)
	return nil
}

func (p *player) getMissionReward() (*pb.GGetCampaignMissionRewardReply, error) {
	if p.ms == nil {
		return nil, gamedata.GameError(1)
	}
	if !p.ms.canReward() || p.ms.hasReward() || p.ms.isCancel() {
		return nil, gamedata.GameError(2)
	}
	p.ms.reward()
	ms := p.ms
	p.ms = nil
	p.attr.Del("ms")
	glog.Infof("getMissionReward uid=%d, ms=%s", p.getUid(), ms)
	p.addContribution(ms.getContribution(), true, true)
	return &pb.GGetCampaignMissionRewardReply{
		Gold:    int32(ms.getGoldReward()),
		CardIDs: ms.getCards(),
	}, nil
}

func (p *player) syncInfoToGame() {
	if p.agent != nil {
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

		p.agent.PushBackend(pb.MessageID_CA2G_UPDATE_CAMPAIGN_PLAYER_INFO, &pb.GCampaignPlayerInfo{
			CountryID:   p.getCountryID(),
			CityID:      int32(p.getCityID()),
			CityJob:     p.getCityJob(),
			CountryJob:  p.getCountryJob(),
			CountryName: countryName,
		})
	}
}

func (p *player) packSimpleMsg() *pb.CampaignSimplePlayer {
	return &pb.CampaignSimplePlayer{
		Uid:       uint64(p.getUid()),
		Name:      p.getName(),
		HeadImg:   p.getHeadImg(),
		HeadFrame: p.getHeadFrame(),
		PvpScore:  int32(p.getPvpScore()),
	}
}

func (p *player) getWarVersion() int {
	return p.attr.GetInt("warVer")
}

func (p *player) fetchWarEndRecord() *pb.CaStateWarEndArg {
	curWarVer := warMgr.getVersion()
	ver := p.getWarVersion()
	if curWarVer == ver {
		return nil
	}

	p.attr.SetInt("warVer", curWarVer)
	if ver != curWarVer-1 {
		p.attr.Del("lcontribution")
		return nil
	}

	cry := p.getCountry()
	var record *pb.CaStateWarEndArg
	if cry != nil {
		record = cry.getLastWarRecord()
	} else {
		record = &pb.CaStateWarEndArg{}
	}

	record.Contribution = int32(p.attr.GetFloat64("lcontribution"))
	record.NextWarRemainTime = int32(warMgr.getStateRemainTime())
	p.attr.Del("lcontribution")
	return record
}

func (p *player) getMyState() *pb.CampaignPlayerState {
	state := &pb.CampaignPlayerState{
		State: pb.CampaignPlayerState_Normal,
	}
	if p.isCaptive() {
		state.State = pb.CampaignPlayerState_BeCaptive
		remainTime := p.getCaptiveTimeout() - time.Now().Unix()
		if remainTime < 0 {
			remainTime = 0
		}
		state.Arg, _ = (&pb.CpStateBeCaptiveArg{RemainTime: int32(remainTime)}).Marshal()
	} else if warMgr.isInWar() || warMgr.isReadyWar() {
		if p.restTimer != nil {

			state.State = pb.CampaignPlayerState_Rest
			state.Arg, _ = (&pb.CpStateLoadingArg{
				MaxTime:    90,
				RemainTime: int32(p.restTimer.GetRemainTime().Seconds()),
			}).Marshal()
		} else if p.rectifyTimer != nil {

			state.State = pb.CampaignPlayerState_Rectify
			state.Arg, _ = (&pb.CpStateLoadingArg{
				MaxTime:    20,
				RemainTime: int32(p.rectifyTimer.GetRemainTime().Seconds()),
			}).Marshal()
		} else if p.myTeam != nil {

			if p.myTeam.type_ == ttSupport {
				state.State = pb.CampaignPlayerState_Support
			} else if p.myTeam.type_ == ttExpedition {
				state.State = pb.CampaignPlayerState_Expedition
			} else if p.myTeam.type_ == ttDefCity {
				state.State = pb.CampaignPlayerState_DefCity
			}
		}

	}
	//} else if p.isKickOut() {
	//	state.State = pb.CampaignPlayerState_KickOut
	//	state.Arg, _ = (&pb.CpStateKickOutArg{RemainTime: int32(p.getKickOutRemainTime())}).Marshal()
	//}
	return state
}

func (p *player) getOthState(isCaptive bool) *pb.CampaignPlayerState {
	state := &pb.CampaignPlayerState{
		State: pb.CampaignPlayerState_Normal,
	}

	if isCaptive && p.isCaptive() {
		state.State = pb.CampaignPlayerState_BeCaptive
		remainTime := p.getCaptiveTimeout() - time.Now().Unix()
		if remainTime < 0 {
			remainTime = 0
		}
		state.Arg, _ = (&pb.CpStateBeCaptiveArg{RemainTime: int32(remainTime)}).Marshal()

	} else if p.isKickOut() {
		state.State = pb.CampaignPlayerState_KickOut
		state.Arg, _ = (&pb.CpStateKickOutArg{RemainTime: int32(p.getKickOutRemainTime())}).Marshal()
	}
	return state
}

func (p *player) onCountryDestory(countryID uint32) {
	if p.getCountryID() != countryID {
		return
	}

	p.setCountryID(0, true)
	if p.myTeam != nil {
		if p.myTeam.type_ == ttDefCity {
			sceneMgr.delTeam(p.myTeam)
		} else {
			p.myTeam.disappear(pb.MyTeamDisappear_CountryDestory, true)
		}
	}
}

func (p *player) getCityGlory() float64 {
	cty := p.getCity()
	if cty == nil {
		cty = p.getLocationCity()
		if cty == nil {
			return 1
		}
	}
	return cty.getGlory()
}

func sortPlayers(players []*player) []*player {
	sort.Slice(players, func(i, j int) bool {
		return players[i].battleThan(players[j])
	})
	return players
}
