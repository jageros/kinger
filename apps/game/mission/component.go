package mission

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	//"kinger/gopuppy/common/glog"
)

var (
	_ types.IPlayerComponent = &missionComponent{}
	_ types.ICrossDayComponent = &missionComponent{}
)

type missionComponent struct {
	attr *attribute.MapAttr
	player types.IPlayer
	mgdata *gamedata.MissionGameData
	mtgdata *gamedata.MissionTreasureGameData
	missions []*missionSt
	treasureObj *treasure
}

func (mc *missionComponent) ComponentID() string {
	return consts.MissionCpt
}

func (mc *missionComponent) GetPlayer() types.IPlayer {
	return mc.player
}

func (mc *missionComponent) OnInit(player types.IPlayer) {
	mc.mgdata = gamedata.GetGameData(consts.Mission).(*gamedata.MissionGameData)
	mc.mtgdata = gamedata.GetGameData(consts.MissionTreasure).(*gamedata.MissionTreasureGameData)
	mc.player = player
	curDayno := timer.GetDayNo()
	missionsAttr := mc.attr.GetListAttr("missions")

	if missionsAttr == nil {

		mc.newbieInit()

	} else {

		missionsAttr.ForEachIndex(func(index int) bool {
			mAttr := missionsAttr.GetMapAttr(index)
			mo := newMissionByAttr(mAttr)
			mc.missions = append(mc.missions, mo)
			return true
		})

		mc.treasureObj = newTreasureByAttr(mc.attr.GetMapAttr("treasure"))
	}

	mc.OnCrossDay(curDayno)

	/*
	if config.GetConfig().IsXfServer() {
		for _, mo := range mc.missions {
			if mo.isReward() {
				mc.refreshMission(mo.getID(), true)
			}
		}
	}
	*/
}

func (mc *missionComponent) getCampStrKey(camp int) string {
	return fmt.Sprintf("camp_%d", camp)
}

func (mc *missionComponent) getCampUseNum(camp int) int {
	return mc.attr.GetInt(mc.getCampStrKey(camp))
}

func (mc *missionComponent) setCampUseNum(camp, num int) {
	mc.attr.SetInt(mc.getCampStrKey(camp), num)
}

func (mc *missionComponent) clearCampUseNum() {
	for i := 1; i <= 3; i++ {
		mc.setCampUseNum(i, 0)
	}
}

func (mc *missionComponent) refreshOftenUseCamp() {
	var camp, num int
	for i := 1; i <= 3; i++ {
		n := mc.getCampUseNum(i)
		if n > num {
			num = n
			camp = i
		}
	}
	if camp <= 0 {
		var maxlvl float64
		var maxlvlCamp int
		for i := 1; i <= 3; i++ {
			ics := mc.player.GetPvpCardPoolsByCamp(i)
			var lvl float64
			for _, ic := range ics {
				lvl = lvl + float64(ic.GetLevel())
			}
			lvl = lvl / float64(len(ics))
			if lvl > maxlvl {
				maxlvl = lvl
				maxlvlCamp = i
			}
		}
		camp = maxlvlCamp
	}
	mc.attr.SetInt("often_use_camp", camp)
	mc.clearCampUseNum()
}

func (mc *missionComponent) getOftenUseCamp() int {
	return mc.attr.GetInt("often_use_camp")
}

func (mc *missionComponent) getMissionFightCampType(missionGid int) int {
	return missionGid
}

func (mc *missionComponent) newbieInit() {
	mc.attr.SetInt("dayno", timer.GetDayNo())
	mc.attr.SetBool("canRefresh", true)
	missionsAttr := attribute.NewListAttr()
	mc.attr.SetListAttr("missions", missionsAttr)

	maxPvpLevel := mc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel()
	fightCamp := mc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()

	for missionGid := 1; missionGid <= maxMissionAmount; missionGid ++ {
		ms := mc.mgdata.GetCanAcceptMission(maxPvpLevel, mc.getMissionFightCampType(missionGid), fightCamp, []int{})
		for _, mdata2 := range ms {
			mdata := mdata2.(*gamedata.Mission)
			if mdata.Type == mtWxShare && !mc.player.IsWxgameAccount() {
				continue
			}

			m := newMission(mdata, missionGid)
			missionsAttr.AppendMapAttr(m.attr)
			mc.missions = append(mc.missions, m)
			break
		}
	}

	mc.treasureObj = newTreasure(mc.player)
	mc.attr.SetMapAttr("treasure", mc.treasureObj.attr)
}

func (mc *missionComponent) OnLogin(isRelogin, isRestore bool) {
}

func (mc *missionComponent) OnLogout() {

}

func (mc *missionComponent) OnCrossDay(curDayno int) {
	if curDayno == mc.attr.GetInt("dayno") {
		return
	}
	mc.attr.SetInt("dayno", curDayno)
	mc.attr.SetBool("canRefresh", true)
	var ignore []int
	refresh := map[int]int{}
	for i, m := range mc.missions {
		if m.isReward() {
			refresh[m.getID()] = i
		} else {
			ignore = append(ignore, m.getMissionID())
		}
	}

	if len(refresh) <= 0 {
		return
	}

	missionsAttr := mc.attr.GetListAttr("missions")
	maxPvpLevel := mc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel()
	fightCamp := mc.getOftenUseCamp()
	mc.refreshOftenUseCamp()
	if fightCamp <= 0 {
		fightCamp = mc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()
	}

	for missionGid, i := range refresh {
		ms := mc.mgdata.GetCanAcceptMission(maxPvpLevel, mc.getMissionFightCampType(missionGid), fightCamp, ignore)
		for _, mdata2 := range ms {
			mdata := mdata2.(*gamedata.Mission)
			if mdata.Type == mtWxShare && !mc.player.IsWxgameAccount() {
				continue
			}

			m := newMission(mdata, missionGid)
			missionsAttr.SetMapAttr(i, m.attr)
			mc.missions[i] = m
			module.Player.LogMission(mc.player, fmt.Sprintf("doMission_%d",  m.getMissionID()), 1)
			break
		}
	}

	agent := mc.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_MISSION_SHOW_RED_DOT, nil)
		agent.PushClient(pb.MessageID_S2C_UPDATE_MISSION_INFO, mc.packMsg())
	}
}

func (mc *missionComponent) refreshMissionAfterNewbieSelectCamp() {
	mc.attr.SetBool("canRefresh", true)
	var ignore []int
	refresh := map[int]int{}
	for i, m := range mc.missions {
		refresh[m.getID()] = i
	}

	if len(refresh) <= 0 {
		return
	}

	missionsAttr := mc.attr.GetListAttr("missions")
	maxPvpLevel := mc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel()
	fightCamp := mc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()

	for missionGid, i := range refresh {
		ms := mc.mgdata.GetCanAcceptMission(maxPvpLevel, mc.getMissionFightCampType(missionGid), fightCamp, ignore)
		for _, mdata2 := range ms {
			mdata := mdata2.(*gamedata.Mission)
			if mdata.Type == mtWxShare && !mc.player.IsWxgameAccount() {
				continue
			}

			m := newMission(mdata, missionGid)
			missionsAttr.SetMapAttr(i, m.attr)
			mc.missions[i] = m
			module.Player.LogMission(mc.player, fmt.Sprintf("doMission_%d",  m.getMissionID()), 1)
			break
		}
	}

	agent := mc.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_UPDATE_MISSION_INFO, mc.packMsg())
	}
}

func (mc *missionComponent) packMsg() *pb.MissionInfo {
	msg := &pb.MissionInfo{
		Treasure: mc.treasureObj.packMsg(),
		CanRefresh: mc.attr.GetBool("canRefresh"),
		RefreshRemainTime: int32(timer.TimeDelta(0, 0, 0).Seconds()),
	}

	for _, mo := range mc.missions {
		msg.Missions = append(msg.Missions, mo.packMsg())
	}

	return msg
}

func (mc *missionComponent) refreshMission(missionGID int, force bool) (*missionSt, error) {
	if !force && !mc.attr.GetBool("canRefresh") {
		return nil, gamedata.GameError(1)
	}

	var mo *missionSt
	var index int
	var ignore []int
	for i, mo2 := range mc.missions {
		ignore = append(ignore, mo2.getMissionID())
		if mo2.getID() == missionGID {
			mo = mo2
			index = i
		}
	}

	if mo == nil || (!force && mo.isReward()) {
		return nil, gamedata.GameError(2)
	}

	fightCamp := mc.getOftenUseCamp()
	if fightCamp <= 0 {
		fightCamp = mc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()
	}

	ms := mc.mgdata.GetCanAcceptMission(mc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel(),
		mc.getMissionFightCampType(missionGID), fightCamp, ignore)
	missionsAttr := mc.attr.GetListAttr("missions")
	for _, data := range ms {
		mdata := data.(*gamedata.Mission)
		if mdata.Type == mtWxShare && !mc.player.IsWxgameAccount() {
			continue
		}
		m := newMission(mdata, missionGID)
		mc.missions[index] = m
		missionsAttr.SetMapAttr(index, m.attr)
		if !force {
			mc.attr.SetBool("canRefresh", false)
		}

		module.Player.LogMission(mc.player, fmt.Sprintf("doMission_%d",  mdata.ID), 1)

		return m, nil
	}

	return nil, gamedata.GameError(3)
}

func (mc *missionComponent) getMissionReward(missionGID int) (jade, gold, bowlder int, nextMission *pb.Mission, err error) {
	var mo *missionSt
	for _, mo2 := range mc.missions {
		if mo2.getID() == missionGID {
			mo = mo2
			break
		}
	}

	jade, gold, bowlder, err = mo.getReward(mc.player)

	/*
	if err == nil && config.GetConfig().IsXfServer() {
		nextMo, _ := mc.refreshMission(mo.getID(), true)
		if nextMo != nil {
			nextMission = nextMo.packMsg()
		}
	}
	*/

	eventhub.Publish(consts.EvGetMissionReward, mc.player)
	return
}

func (mc *missionComponent) openTreasure() (*pb.OpenTreasureReply, *treasure, error) {
	reward, newT, err := mc.treasureObj.getReward(mc.player)
	if err != nil {
		return nil, nil, err
	}

	if newT != nil {
		mc.treasureObj = newT
		mc.attr.SetMapAttr("treasure", mc.treasureObj.attr)
	}

	return reward, mc.treasureObj, nil
}

func (mc *missionComponent) tryAddMissionCnt(callback func(tpl iMissionTemplate) int) {
	isComplete := false
	var updateMissions []*missionSt
	isTreasureUpdate := false
	for _, mo := range mc.missions {
		if mo.isReward() {
			//glog.Infof("tryAddMissionCnt 22222222222")
			continue
		}
		if mo.isComplete() {
			//glog.Infof("tryAddMissionCnt 333333333333")
			continue
		}

		tpl := mo.getMissionTpl()
		cnt := callback(tpl)
		//glog.Infof("tryAddMissionCnt 44444444444 %d", cnt)
		if cnt > 0 {
			isComplete2, isUpdate := mo.addCnt(cnt)
			if isComplete2 {
				isTreasureUpdate2 := mc.treasureObj.onMissionComplete(mc.player)
				isComplete = isComplete2
				if isTreasureUpdate2 {
					isTreasureUpdate = isTreasureUpdate2
				}
			}

			if isUpdate {
				updateMissions = append(updateMissions, mo)
			}
		}
	}

	if (isComplete || isTreasureUpdate || len(updateMissions) > 0) && mc.player.IsOnline() {
		if isComplete {
			mc.player.GetAgent().PushClient(pb.MessageID_S2C_MISSION_SHOW_RED_DOT, nil)
		}

		if len(updateMissions) > 0 {
			msg := &pb.UpdateMissionProcessArg{}
			for _, mo := range updateMissions {
				msg.Missions = append(msg.Missions, mo.packMsg())
			}
			mc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_MISSION_PROCESS, msg)
		}

		if isTreasureUpdate {
			mc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_MISSION_TREASURE_PROCESS, mc.treasureObj.packMsg())
		}
	}
}

func (mc *missionComponent) gmAddMission(mo *missionSt) {
	idx := mo.getID()
	mc.missions[idx-1] = mo
	missionsAttr := mc.attr.GetListAttr("missions")
	missionsAttr.SetMapAttr(idx-1, mo.attr)
	mc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_MISSION_INFO, mc.packMsg())
}

func (mc *missionComponent) gmCompleteMission() {
	for _, mo := range mc.missions {
		if mo.isComplete() {
			continue
		}

		mo.addCnt(1000)
		mc.treasureObj.onMissionComplete(mc.player)
	}

	msg := &pb.UpdateMissionProcessArg{}
	for _, mo := range mc.missions {
		msg.Missions = append(msg.Missions, mo.packMsg())
	}
	mc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_MISSION_PROCESS, msg)

	mc.player.GetAgent().PushClient(pb.MessageID_S2C_UPDATE_MISSION_TREASURE_PROCESS, mc.treasureObj.packMsg())
}
