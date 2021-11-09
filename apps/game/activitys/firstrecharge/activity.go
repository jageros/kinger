package firstrecharge

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
	id2Reward map[int]*gamedata.ActivityFirstRechargeRewardGameData
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfFirstRecharge)

	idList := a.IAMod.GetActivityIdList()
	for _, aid := range idList {

		iA := a.IAMod.GetActivityByID(aid)
		rewardTableName := iA.GetRewardTableName()
		rwData := gamedata.GetGameData(rewardTableName)
		if rwData == nil {
			err := gamedata.GameError(aTypes.GetRewardError)
			glog.Errorf("First recharge activity initActivityData reward GetGameData err=%s, activityID=%d", err, aid)
			continue
		}
		a.id2Reward[aid] = rwData.(*gamedata.ActivityFirstRechargeRewardGameData)
	}
}

func (a *activity) getRewardIdList(activityID int) []int {
	var idList []int
	if rw, ok := a.id2Reward[activityID]; ok {
		for k, _ := range rw.ID2ActivityFirstRechargeReward {
			idList = append(idList, k)
		}
	}
	return idList
}

func (a *activity) getRewardFinshCondition(activityID, rewardID int) (string, error) {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ID2ActivityFirstRechargeReward[rewardID]; ok {
			return rw.GoodsID, nil
		}
	}
	err := gamedata.GameError(aTypes.GetRewardError)
	return "", err
}

func (a *activity) getRewardData(activityID, rewardID int) *gamedata.ActivityFirstRechargeReward {
	if v, ok := a.id2Reward[activityID]; ok {
		if rw, ok := v.ID2ActivityFirstRechargeReward[rewardID]; ok {
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
		glog.Errorf("First recharge activity getRewardMap get reward data err=%s, activityID=%d, rewardID=%d", err, activityID, rewardID)
		return rw
	}
	for _, str := range rd.Reward {
		strl := strings.Split(str, ":")
		if len(strl) < 2 {
			err := gamedata.GameError(aTypes.RewardArgError)
			glog.Errorf("First recharge activity getRewardMap reward arg err=%s, activityID=%d, rewardID=%d, reward=%s", err, activityID, rewardID, str)
			return nil
		}
		num, err := strconv.Atoi(strl[1])
		if err != nil {
			glog.Errorf("First recharge activity getRewardMap reward num arg err=%s, activityID=%d, rewardID=%d, rewardNum=%s", err, activityID, rewardID, strl[1])
			return nil
		}
		rw[strl[0]] = int32(num)
	}
	return rw
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
			if p.conform(aid) {
				p.setBuyGoods(aid, goodsID)
				rids := mod.getRewardIdList(aid)
				for _, rid := range rids {
					rData := mod.getRewardData(aid, rid)
					if rData.GoodsID == goodsID {
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
