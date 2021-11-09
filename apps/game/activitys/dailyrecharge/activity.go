package dailyrecharge

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"strconv"
	"strings"
)

var mod *activity

type activity struct {
	aTypes.BaseActivity
	id2Reward map[int]*gamedata.ActivityDailyRechargeRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfDailyRecharge)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("DailyRecharge activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityDailyRechargeRewardGameData)
	}
}

func (a *activity) getRewardIdList(aid int) []int {
	var idList []int
	if rw, ok := a.id2Reward[aid]; ok {
		for k, _ := range rw.ID2ActivityDailyRechargeReward {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(aid, rid int) (int, error) {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityDailyRechargeReward[rid]; ok {
			return rw.Recharge, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return 0, err
}

func (a *activity) getRewardData(aid, rid int) *gamedata.ActivityDailyRechargeReward {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ID2ActivityDailyRechargeReward[rid]; ok {
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
		glog.Errorf("DailyRecharge activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, aid, rid)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("DailyRecharge activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, aid, rid, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("DailyRecharge activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, aid, rid, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}
	return rw
}

func onRecharge(args ...interface{}) {
	player := args[0].(types.IPlayer)
	money := args[1].(int)
	p := newComponent(player)
	for _, id := range mod.IAMod.GetActivityIdList() {
		if p.ipc.Conform(id) {
			p.setTodayRechargeAmount(id, money)
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
