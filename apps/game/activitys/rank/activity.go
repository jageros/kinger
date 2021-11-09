package rank

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
	id2Reward map[int]*gamedata.ActivityRankRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfRank)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("Rank activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityRankRewardGameData)

	}
}

func (a *activity) getRewardIdList(activityID int) []int {
	var idList []int
	if rw, ok := a.id2Reward[activityID]; ok {
		for k, _ := range rw.ActivityRankRewardMap {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(activityID, rewardID int) (int, error) {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ActivityRankRewardMap[rewardID]; ok {
			return rw.Rank, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return 0, err
}

func (a *activity) getRewardData(activityID, rewardID int) *gamedata.ActivityRankReward {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ActivityRankRewardMap[rewardID]; ok {
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
		glog.Errorf("Rank activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, activityID, rewardID)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("Rank activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, activityID, rewardID, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("Rank activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, activityID, rewardID, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}
	return rw
}

func onRankUpdate(args ...interface{}) {
	player := args[0].(types.IPlayer)
	oldLevel := args[1].(int)
	newLevel := args[2].(int)
	p := newComponent(player)
	p.pushFinshNum(oldLevel, newLevel)
	p.updateHint()
}

func updateAllPlayerHint() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := newComponent(player)
		p.updateHint()
	})
}

func (p *activityCom) pushFinshNum(oldLevel, newLevel int) {
	idl := mod.IAMod.GetActivityIdList()
	for _, aid := range idl {
		if p.ipc.Conform(aid) {
			p.ipc.PushFinshNum(aid, 0, p.player.GetMaxPvpLevel())
			p.updateReceiveStatus(aid, oldLevel, newLevel)
		}
	}
}
