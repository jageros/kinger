package loginrecharge

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
	"strings"
)

var mod *activity

type activity struct {
	aTypes.BaseActivity
	id2Reward map[int]*gamedata.ActivityLoginRechargeRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfLoginRecharge)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("Login activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityLoginRechargeRewardGameData)
	}
}

func (a *activity) getRewardData(aid, rid int) *gamedata.ActivityLoginRechargeReward {
	if v, ok := a.id2Reward[aid]; ok {
		if rw, ok := v.ActivityLoginRechargeRewardMap[rid]; ok {
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
		glog.Errorf("Login activity getRewardMap getRewardData err=%s, activityID=%d, rewardID=%d", err, aid, rid)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("Login activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, aid, rid, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("Login activity getRewardMap reward num string to int err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, aid, rid, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}

	return rw
}

func (a *activity) getRewardIdList(activityID int) []int {
	var idList []int
	if rw, ok := a.id2Reward[activityID]; ok {
		for k, _ := range rw.ActivityLoginRechargeRewardMap {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(activityID, rewardID int) (int, string, error) {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ActivityLoginRechargeRewardMap[rewardID]; ok {
			return rw.LoginDay, rw.RechargeID, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return 0, "", err
}

func updateAllPlayerHint() {
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		p := newComponent(player)
		p.updateHint()
	})
}

func onRecharge(args ...interface{}) {
	player := args[0].(types.IPlayer)
	if goodsID, ok := args[2].(string); ok {
		p := newComponent(player)
		for _, aid := range mod.IAMod.GetActivityIdList() {
			activity := mod.IAMod.GetActivityByID(aid)
			if activity == nil {
				continue
			}

			gdata := gamedata.GetGameData(activity.GetRewardTableName())
			if gdata == nil {
				continue
			}

			if rewardData, ok := gdata.(gamedata.IActivityRechargeReward); !ok || !rewardData.IsRechargeID(goodsID) {
				continue
			}

			if p.ipc.Conform(aid) {
				p.setLoginRecharge(aid, goodsID)
				rids := mod.getRewardIdList(aid)
				for _, rid := range rids {
					rData := mod.getRewardData(aid, rid)
					if rData.RechargeID == goodsID {
						rd := &pb.Reward{
							RewardList: map[string]int32{},
						}
						err := ReceiveReward(player, aid, rid, rd)
						if err == nil {
							player.GetAgent().PushClient(pb.MessageID_S2C_GRANT_ACTIVITY_REWARD, rd)
						}
					}
				}
			}
		}
		p.updateHint()
	}
}
