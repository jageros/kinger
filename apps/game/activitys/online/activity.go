package online

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"strconv"
	"strings"
	"time"
)

var mod *activity

var aTim = map[int]*timer.Timer{}

type activity struct {
	aid int
	aTypes.BaseActivity
	id2Reward map[int]*gamedata.ActivityOnlineRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfOnline)
	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("Online activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityOnlineRewardGameData)
	}
	a.refreshTimerEvent()
}

func (a *activity) getRewardIdList(aid int) []int {
	var idList []int
	if rw, ok := a.id2Reward[aid]; ok {
		for k, _ := range rw.ActivityOnlineRewardMap {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(aid, rid int) (string, error) {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ActivityOnlineRewardMap[rid]; ok {
			return rw.Times, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return "", err
}

func (a *activity) getRewardData(aid, rid int) *gamedata.ActivityOnlineReward {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ActivityOnlineRewardMap[rid]; ok {
			return rw
		}
	}
	return nil
}

func (a *activity) getRewardMap(aid, rid int) map[string]int32 {
	rw := map[string]int32{}
	rd := a.getRewardData(aid, rid)
	if rd == nil {
		err := gamedata.GameError(aTypes.GetRewardError)
		glog.Errorf("Online activity getRewardMap getRewardData err=%s, activityID=%d, rewardID=%d", err, aid, rid)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("Online activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, aid, rid, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("Online activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, aid, rid, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}
	return rw
}

func changeStartEndHour2Time(timCondition string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	strl := strings.Split(timCondition, ":")
	if len(strl) != 2 {
		err := gamedata.GameError(aTypes.GetTimeConditionError)
		return startTime, endTime, err
	}
	start, err := strconv.Atoi(strl[0])
	if err != nil {
		return startTime, endTime, err
	}
	startTime, err = utils.HourToTodayTime(start)
	if err != nil {
		return startTime, endTime, err
	}

	end, err := strconv.Atoi(strl[1])
	if err != nil {
		return startTime, endTime, err
	}
	endTime, err = utils.HourToTodayTime(end)
	if err != nil {
		return startTime, endTime, err
	}
	return startTime, endTime, nil
}

func (m *activity) getStartTime(aid, rid int) int {
	if rd, ok := m.id2Reward[aid]; ok {
		if rw, ok := rd.ActivityOnlineRewardMap[rid]; ok {
			str := strings.Split(rw.Times, ":")
			if len(str) < 2 {
				return -1
			}
			start, err := strconv.Atoi(str[0])
			if err != nil {
				return -1
			}
			return start
		}
	}
	return -1
}

func (m *activity) getEndTime(aid, rid int) int {
	if rd, ok := m.id2Reward[aid]; ok {
		if rw, ok := rd.ActivityOnlineRewardMap[rid]; ok {
			str := strings.Split(rw.Times, ":")
			if len(str) < 2 {
				return -1
			}
			end, err := strconv.Atoi(str[1])
			if err != nil {
				return -1
			}
			return end
		}
	}
	return -1
}

func (a *activity) getAllTime() []int {
	var t []int
	for aid, rds := range a.id2Reward {
		for rid, _ := range rds.ActivityOnlineRewardMap {
			st := a.getStartTime(aid, rid)
			et := a.getEndTime(aid, rid)
			if st == 24 {
				st = 0
			}
			if et == 24 {
				et = 0
			}
			t = append(t, st, et)
		}
	}
	return t
}

func (a *activity) refreshTimerEvent() {
	for _, t := range aTim {
		t.Cancel()
	}

	ts := a.getAllTime()
	aTim = map[int]*timer.Timer{}
	for _, t := range ts {
		if _, ok := aTim[t]; !ok {
			f := timer.RunEveryDay(t, 0, 1, updateAllPlayerInfo)
			aTim[t] = f
		}
	}
}

func updateAllPlayerInfo() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := newComponent(player)
		p.updateHint()
		p.updateReceiveStatus()
	})
}
