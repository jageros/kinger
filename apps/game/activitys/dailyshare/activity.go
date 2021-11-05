package dailyshare

import (
	"kinger/gopuppy/common/glog"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"strconv"
	"strings"
)

var mod *activity

type activity struct {
	aTypes.BaseActivity
	id2Reward map[int]*gamedata.ActivityDailyShareRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfDailyShare)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("DailyShare activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityDailyShareRewardGameData)
	}
}

func (a *activity) getRewardIdList(aid int) []int {
	var idList []int
	if rw, ok := a.id2Reward[aid]; ok {
		for k, _ := range rw.ID2ActivityDailyShareReward {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(aid, rid int) (int, error) {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityDailyShareReward[rid]; ok {
			return rw.ShareCnt, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return 0, err
}

func (a *activity) getRewardData(aid, rid int) *gamedata.ActivityDailyShareReward {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityDailyShareReward[rid]; ok {
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
		glog.Errorf("Recharge activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, aid, rid)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("Recharge activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, aid, rid, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("Recharge activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, aid, rid, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}
	return rw
}

func onShare(args ...interface{}) {
	player := args[0].(types.IPlayer)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			p.setTodayShareAmount(id, 1)
		}
	}
	p.updateHint()
}

func updateAllPlayerHint() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := newComponent(player)
		p.updateHint()
	})
}