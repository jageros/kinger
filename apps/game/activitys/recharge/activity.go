package recharge

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
	id2Reward map[int]*gamedata.ActivityRechargeRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfRecharge)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("Recharge activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityRechargeRewardGameData)
	}
}

func (a *activity) getRewardIdList(activityID int) []int {
	var idList []int
	if rw, ok := a.id2Reward[activityID]; ok {
		for k, _ := range rw.ActivityRechargeRewardMap {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(activityID, rewardID int) (int, error) {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ActivityRechargeRewardMap[rewardID]; ok {
			return rw.Recharge, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return 0, err
}

func (a *activity) getRewardData(activityID, rewardID int) *gamedata.ActivityRechargeReward {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ActivityRechargeRewardMap[rewardID]; ok {
			return rw
		}
	}
	return nil
}

func (a *activity) getRewardMap(activityID, rewardID int) map[string]int32 {
	rw := map[string]int32{}
	rd := a.getRewardData(activityID, rewardID)
	if rd == nil {
		err := gamedata.GameError(aTypes.GetRewardError)
		glog.Errorf("Recharge activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, activityID, rewardID)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("Recharge activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, activityID, rewardID, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("Recharge activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, activityID, rewardID, strl[1])
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
			p.setTotalRechargeAmount(id, money)
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