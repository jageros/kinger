package activitys

/*****************************************************************
 *
 *  活动相关GM指令
 *  act 关键字可进入到TestARpc函数
 * 【act relogin】: 执行角色在服务器中活动模块的重登陆操作
 * 【act del 0/活动ID】: 删除全部/指定活动的角色数据
 * 【act rc 钱数】: 充值对应的钱数，紧用于活动模块中
 * 【act lc 物品ID】: 登录充值活动中第三天和第六天的充值
 *
 ****************************************************************/

import (
	atype "kinger/apps/game/activitys/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/eventhub"
	"kinger/proto/pb"
	"strconv"
)

func (m *activityModule) TestARpc(agent *logic.PlayerAgent, com []string) (rsp interface{}, err error) {

	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		err := gamedata.GameError(404)
		return nil, err
	}
	//活动重登操作
	if len(com) == 2 {
		if com[1] == "relogin" {
			p := player.GetComponent(consts.ActivityCpt).(*activityComponent)
			p.OnLogin(false, false)
			return
		}
	}

	if len(com) == 3 {
		if com[1] == "kickout" {
			if com[2] == "me" {
				player.OnForbidLogin()
			}
			if com[2] == "all" {
				module.Player.ForEachOnlinePlayer(func(p types.IPlayer) {
					p.OnForbidLogin()
				})
			}
			return
		}
	}

	if len(com) == 2 {
		if com[1] == "share" {
			eventhub.Publish(consts.EvShare, player)
			return
		}
	}

	//清除角色的活动数据
	if len(com) == 3 {
		if com[1] == "del" {
			act := map[int]string{1: atype.LoginActivity, 2: atype.RechargeActivity, 3: atype.OnlineActivity, 4: atype.OnlineActivity, 5: atype.FightActivity,
				6: atype.WinActivity, 7: atype.RankActivity, 8: atype.ConsumeActivity, 9: atype.LoginRechargeActivity, 10: atype.FirstRechargeActivity,
				11: atype.GrowPlanActivity, 12: atype.LoginActivity, 13: atype.DailyRechargeActivity, 14: atype.DailyShareActivity}
			if aid, err1 := strconv.Atoi(com[2]); err1 == nil {
				if aid == 0 {
					p := player.GetComponent(consts.ActivityCpt).(*activityComponent)
					for id, aidStr := range act {
						attr := p.attr.GetMapAttr(aidStr)
						attr.Del(strconv.Itoa(id))
					}
					return
				}
				if aidStr, ok := act[aid]; ok {
					p := player.GetComponent(consts.ActivityCpt).(*activityComponent)
					attr := p.attr.GetMapAttr(aidStr)
					attr.Del(com[2])
					return
				}
			}
		}
	}

	//充值
	if len(com) == 3 {
		if com[1] == "rc" {
			if num, err1 := strconv.Atoi(com[2]); err1 == nil {
				eventhub.Publish(consts.EvRecharge, player, num)
				return
			}
		}

		if com[1] == "lc" {
			eventhub.Publish(consts.EvRecharge, player, 0, com[2])
			return
		}
	}

	//活动操作，放最后（用于服务端自测）
	var ty, aid, rid int
	ty, err = strconv.Atoi(com[1])
	if err != nil {
		return
	}
	if len(com) >= 3 {
		aid, err = strconv.Atoi(com[2])
		if err != nil {
			return
		}
	}

	if len(com) >= 4 {
		rid, err = strconv.Atoi(com[3])
		if err != nil {
			return
		}
	}

	msid := pb.MessageID(ty)
	switch msid {
	case pb.MessageID_C2S_FETCH_ACTIVITY_DETAIL:
		arg := &pb.ActivityID{}
		arg.ActivityID = int32(aid)
		rsp, err = rpc_C2S_FetchActivityDETAIL(agent, arg)

	case pb.MessageID_C2S_RECEIVE_ACTIVITY_REWARD:
		arg := &pb.TargetActivity{}
		arg.ID = int32(aid)
		arg.RewardID = int32(rid)
		rsp, err = rpc_C2S_ReceiveActivityReward(agent, arg)
	case pb.MessageID_C2S_FETCH_ACTIVITY_LABEL_LIST:
		var arg int
		rsp, err = rpc_C2S_FetchActivityLabelList(agent, arg)
	case pb.MessageID_C2S_FETCH_FIRST_RECHARGE_ACTIVITY_DETAIL:
		var arg interface{}
		rsp, err = rpc_C2S_FetchFirstRechargeActivityDetail(agent, arg)
	}

	return
}
