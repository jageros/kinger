package seasonpvp

import (
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/attribute"
	"fmt"
	"time"
	"kinger/gopuppy/apps/logic"
	"kinger/gamedata"
	"kinger/gopuppy/common/evq"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/proto/pb"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	htypes "kinger/apps/game/huodong/types"
	"kinger/gopuppy/common/eventhub"
)

type seasonPvpHd struct {
	htypes.BaseHuodong
	area int
	sessionDatas map[int]*seasonPvpHdSessionData
	loadingSessionData map[int]chan struct{}
}

func (hd *seasonPvpHd) String() string {
	return fmt.Sprintf("[seasonPvpHd session=%d, beginTime=%s, endTime=%s, isOpen=%v, isClose=%v]",
		hd.GetVersion(), hd.GetBeginTime(), hd.GetEndTime(), hd.IsOpen(), hd.IsClose())
}

func (hd *seasonPvpHd) loadSessionData(session int) *seasonPvpHdSessionData {
	data, ok := hd.sessionDatas[session]
	if !ok {
		loading, ok2 := hd.loadingSessionData[session]
		if ok2 {
			evq.Await(func() {
				<- loading
			})
			return hd.sessionDatas[session]
		} else {
			c := make(chan struct{})
			hd.loadingSessionData[session] = c
			attr := attribute.NewAttrMgr(fmt.Sprintf("seasonPvpHd%d", hd.area), session, true)
			err := attr.Load()
			if err == attribute.NotExistsErr {
				attr.Save(false)
				err = nil
			}
			if err != nil {
				glog.Errorf("seasonPvpHd loadSessionData error, session=%d, area=%d, err=%s", session, hd.area, err)
				delete(hd.loadingSessionData, session)
				close(c)
				return hd.sessionDatas[session]
			}

			data = newSeasonPvpHdSessionDataByAttr(session, hd.area, attr)
			hd.sessionDatas[session] = data
			delete(hd.loadingSessionData, session)
			close(c)
			return data
		}
	} else {
		return data
	}
}

func (hd *seasonPvpHd) OnStart() {
	hd.BaseHuodong.OnStart()

	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		if player.GetArea() != hd.area {
			return
		}

		hdCpt := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent)
		hdData := hdCpt.GetOrNewHdData(pb.HuodongTypeEnum_HSeasonPvp)
		if hdData == nil {
			return
		}
		hd.OnPlayerLogin(player, hdData)
	})

	if module.Service.GetAppID() == 1 {
		logic.PushBackend("", 0, pb.MessageID_G2R_SEASON_PVP_BEGIN, &pb.TargetArea{Area: int32(hd.area)})
	}
}

func (hd *seasonPvpHd) OnStop() {
	if module.Service.GetAppID() == 1 {
		seasonHdData := hd.loadSessionData(hd.GetVersion())
		seasonHdData.onSeasonStop()
	} else {
		delete(hd.sessionDatas, hd.GetVersion())
	}

	hd.BaseHuodong.OnStop()

	session := hd.GetVersion()
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		if player.GetArea() != hd.area {
			return
		}

		hdCpt := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent)
		hdData := hdCpt.GetOrNewHdData(pb.HuodongTypeEnum_HSeasonPvp)
		if hdData == nil {
			return
		}
		data := hdData.(*seasonPvpHdPlayerData)
		if data.getSession() != session {
			return
		}
		hd.onReward(player, data, hd.getGameData())
	})

	if module.Service.GetAppID() == 1 {
		timer.AfterFunc(3 * time.Second, func() {
			var data interface{} = nil
			gdata := gamedata.GetSeasonPvpGameData().GetSeasonData(hd.area)
			if gdata != nil {
				data = gdata
			}

			if hd.Refresh(data) {
				logic.BroadcastBackend(pb.MessageID_G2G_ON_HUODONG_EVENT, &pb.HuodongEvent{
					HdType: pb.HuodongTypeEnum_HSeasonPvp,
					Event: pb.HuodongEventType_HetRefresh,
					Areas: []int32{ int32(hd.area) },
				})
			}
		})
	}
}

func (hd *seasonPvpHd) Refresh(gdata interface{}) bool {
	if gdata == nil {
		return false
	}

	data, ok := gdata.(gamedata.ISeasonPvp)
	if !ok || data == nil {
		return false
	}

	startTime := data.GetStartTime()
	stopTime := data.GetStopTime()
	if hd.GetBeginTime().Equal(startTime) && hd.GetEndTime().Equal(stopTime) {
		return false
	}

	if hd.IsClose() {
		hd.SetVersion(hd.GetVersion() + 1)
		hd.SetClose(false)
	}
	hd.SetBeginTime(startTime)
	hd.SetEndTime(stopTime)
	hd.Save()
	return true
}

func (hd *seasonPvpHd) NewPlayerData(player types.IPlayer) htypes.IHdPlayerData {
	attr := attribute.NewMapAttr()
	return hd.NewPlayerDataByAttr(player, attr)
}

func (hd *seasonPvpHd) NewPlayerDataByAttr(player types.IPlayer, attr *attribute.MapAttr) htypes.IHdPlayerData {
	hpd := &seasonPvpHdPlayerData{}
	hpd.Player = player
	hpd.Attr = attr
	return hpd
}

func (hd *seasonPvpHd) getGameData() gamedata.ISeasonPvp {
	return gamedata.GetSeasonPvpGameData().GetSeasonData(hd.area)
}

func (hd *seasonPvpHd) OnPlayerLogin(player types.IPlayer, hdData htypes.IHdPlayerData) {
	data, ok := hdData.(*seasonPvpHdPlayerData)
	if !ok {
		return
	}
	hdGameData := hd.getGameData()
	if hdGameData == nil {
		return
	}
	team := player.GetPvpTeam()
	if team < hdGameData.GetLimitPvpTeam() {
		return
	}

	curSession := hd.GetVersion()
	session := data.getSession()

	if hd.IsOpen() {
		if session != curSession {
			hd.beginSession(player, data)
		} else if data.isQuit() {
			data.join()
			agent := player.GetAgent()
			if agent != nil {
				agent.PushClient(pb.MessageID_S2C_SEASON_PVP_BEGIN, &pb.SeasonPvpLimitTime{
					LimitTime: module.Huodong.GetSeasonPvpLimitTime(player),
				})
			}
		}
	} else {

		if session > 0 && (session != curSession || hd.IsClose()) {
			hd.onReward(player, data, hdGameData)
		}
	}
}

func (hd *seasonPvpHd) onReward(player types.IPlayer, hdData *seasonPvpHdPlayerData, hdGameData gamedata.ISeasonPvp) {
	if hdData.isReward() {
		return
	}

	hdData.delHandCards()
	hdData.delChooseCards(hdGameData, []uint32{})

	sessionData := hd.loadSessionData(hdData.getSession())
	if sessionData == nil {
		glog.Errorf("seasonPvpHd onReward no sessionData %d", hdData.getSession())
		return
	}

	hdData.markReward()
	uid := player.GetUid()
	rank := sessionData.getRank(uid)
	pvpTeam := player.GetPvpTeam()
	winAmount := module.Player.GetResource(player, consts.WinDiff)
	glog.Infof("seasonPvpHd onReward uid=%d, pvpTeam=%d, rank=%d, winAmount=%d, session=%d", uid, pvpTeam, rank,
		winAmount, hdData.getSession())

	rankReward := false
	noRankReward := false
	isMultiLan := config.GetConfig().IsMultiLan
	var mailReward types.IMailReward = nil
	sender := module.Mail.NewMailSender(uid)
	for _, rw := range sessionData.rewards {
		//if rw.Team > 0 && rw.Team != pvpTeam  {
		//	continue
		//}
		if noRankReward && rankReward {
			break
		}

		if rw.Ranking > 0 {
			if rw.Ranking < rank || rankReward {
				continue
			}
		} else if noRankReward {
			continue
		} else {
			if isMultiLan {
				if rw.Team > 0 && rw.Team > pvpTeam {
					continue
				}
			} else if rw.WinAmount > 0 && rw.WinAmount > winAmount {
				continue
			}
		}

		if rw.Ranking > 0 {
			if !rankReward {
				rankReward = true
			}
		} else if !noRankReward {
			noRankReward = true
		}

		if mailReward == nil {
			mailReward = sender.GetRewardObj()
		}

		if rw.CardSkin != "" {
			mailReward.AddItem(pb.MailRewardType_MrtCardSkin, rw.CardSkin, 1)
		}
		if rw.HeadFrame != "" {
			mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, rw.HeadFrame, 1)
		}
		if rw.Treasure != "" {
			mailReward.AddItem(pb.MailRewardType_MrtTreasure, rw.Treasure, 1)
		}
	}

	if config.GetConfig().IsMultiLan {
		sender.SetTypeAndArgs(pb.MailTypeEnum_SeasonPvpEnd, player.GetPvpLevel())
	} else {
		sender.SetTypeAndArgs(pb.MailTypeEnum_SeasonPvpEnd, winAmount)
	}
	sender.Send()

	agent := player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SEASON_PVP_STOP, nil)
	}
}

func (hd *seasonPvpHd) beginSession(player types.IPlayer, hdData *seasonPvpHdPlayerData) {
	curSession := hd.GetVersion()
	hdData.beginSession(curSession)
	if hdData.isResetLevel() {
		return
	}

	hdData.markResetLevel()
	pvpCpt := player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	curLevel := pvpCpt.GetPvpLevel()
	var resetLevel int
	if curLevel < 16 {
		return
	}

	if config.GetConfig().IsMultiLan {
		/*
		if curLevel >= 31 {
			resetLevel = 21
		} else if curLevel >= 28 {
			resetLevel = 20
		} else if curLevel >= 25 {
			resetLevel = 19
		} else if curLevel >= 22 {
			resetLevel = 18
		} else if curLevel >= 19 {
			resetLevel = 17
		} else {
			resetLevel = 16
		}
		*/
		resetLevel = 16
		resetScore := module.Pvp.GetMinStarByPvpLevel(resetLevel)
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		modifyScore := resCpt.GetResource(consts.Score) - resetScore
		if modifyScore > 0 {
			resCpt.ModifyResource(consts.Score, - modifyScore)
		}
	}

	uid := player.GetUid()
	glog.Infof("seasonPvpHd beginSession curSession=%d, uid=%d, curLevel=%d, resetLevel=%d",
		curSession, uid, curLevel, resetLevel)
	sender := module.Mail.NewMailSender(uid)
	if config.GetConfig().IsMultiLan {
		sender.SetTypeAndArgs(pb.MailTypeEnum_SeasonPvpBegin, resetLevel)
	} else {
		sender.SetTypeAndArgs(pb.MailTypeEnum_SeasonPvpBegin)
	}
	sender.Send()

	agent := player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SEASON_PVP_BEGIN, &pb.SeasonPvpLimitTime{
			LimitTime: module.Huodong.GetSeasonPvpLimitTime(player),
		})
	}
}


func InitializeSeasonPvpHd() {
	registerRpc()

	eventhub.Subscribe(consts.EvEndPvpBattle, func(args ...interface{}) {
		player := args[0].(types.IPlayer)
		hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
		if hd == nil || !hd.IsOpen() {
			return
		}

		isWin := args[1].(bool)
		fighterData := args[2].(*pb.EndFighterData)
		hdData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(pb.HuodongTypeEnum_HSeasonPvp)
		if hdData == nil {
			return
		}
		data := hdData.(*seasonPvpHdPlayerData)
		if data.getSession() != hd.GetVersion() || data.isQuit() {
			return
		}

		data.onPvpEnd(isWin, fighterData.IsFirstHand, hd.(*seasonPvpHd).getGameData())
	})

	eventhub.Subscribe(consts.EvPvpLevelUpdate, func(args ...interface{}) {
		player := args[0].(types.IPlayer)
		hd := htypes.Mod.GetHuodong(player.GetArea(), pb.HuodongTypeEnum_HSeasonPvp)
		if hd == nil || !hd.IsOpen() {
			return
		}

		rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
		oldLevel := args[1].(int)
		oldTeam := rankGameData.Ranks[oldLevel].Team
		curLevel := args[2].(int)
		curTeam := rankGameData.Ranks[curLevel].Team

		seasonHd := hd.(*seasonPvpHd)
		gdata := seasonHd.getGameData()
		if gdata == nil {
			return
		}

		playerData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(
			pb.HuodongTypeEnum_HSeasonPvp).(*seasonPvpHdPlayerData)
		limitPvpTeam :=  gdata.GetLimitPvpTeam()
		if curTeam >= limitPvpTeam && oldTeam < limitPvpTeam {
			evq.CallLater(func() {
				seasonHd.OnPlayerLogin(player, playerData)
			})
		} else if curTeam < limitPvpTeam && oldTeam >= limitPvpTeam {
			playerData.quit(gdata)
		}
	})

	if module.Service.GetAppID() != 1 {
		return
	}

	gamedata.GetSeasonPvpGameData().AddReloadCallback(func(data gamedata.IGameData) {

		seasonGameData := data.(gamedata.ISeasonPvpGameData)
		areaGameData := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
		var arg *pb.HuodongEvent

		for _, areaCfg := range areaGameData.Areas {
			area := areaCfg.Area
			var needRefresh bool
			hd := htypes.Mod.GetHuodong(area, pb.HuodongTypeEnum_HSeasonPvp)
			if hd == nil {
				hd = htypes.Mod.NewHuodong(area, pb.HuodongTypeEnum_HSeasonPvp)
				if hd != nil {
					htypes.Mod.AddHuodong(area, hd)
					hd.Save()
					needRefresh = true
				}
			} else {
				needRefresh = hd.Refresh(seasonGameData.GetSeasonData(area))
			}

			if needRefresh {
				if arg == nil {
					arg = &pb.HuodongEvent{HdType: pb.HuodongTypeEnum_HSeasonPvp, Event: pb.HuodongEventType_HetRefresh}
				}
				arg.Areas = append(arg.Areas, int32(area))
			}
		}

		if arg != nil {
			logic.BroadcastBackend(pb.MessageID_G2G_ON_HUODONG_EVENT, arg)
		}

	})

}

func getSeasonPvpRefreshChooseCardJade() int {
	if config.GetConfig().IsMultiLan {
		return 10
	} else {
		return gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).RankSeasonRefreshPrice
	}
}
