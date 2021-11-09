package activitys

import (
	"kinger/apps/game/activitys/consume"
	"kinger/apps/game/activitys/dailyrecharge"
	"kinger/apps/game/activitys/dailyshare"
	"kinger/apps/game/activitys/fight"
	"kinger/apps/game/activitys/firstrecharge"
	"kinger/apps/game/activitys/growplan"
	"kinger/apps/game/activitys/login"
	"kinger/apps/game/activitys/loginrecharge"
	"kinger/apps/game/activitys/online"
	"kinger/apps/game/activitys/rank"
	"kinger/apps/game/activitys/recharge"
	"kinger/apps/game/activitys/spring"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/activitys/win"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
)

//MessageID_C2S_FETCH_ACTIVITY_DETAIL=335
func rpc_C2S_FetchActivityDETAIL(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		err := gamedata.GameError(aTypes.GetPlayerError)
		glog.Errorf("rpc_C2S_FetchActivityDETAIL get player err=%s, uid=%d", err, uid)
		return nil, err
	}

	playerCom := player.GetComponent(consts.ActivityCpt).(*activityComponent)
	if playerCom == nil {
		err := gamedata.GameError(aTypes.GetPlayComponentError)
		glog.Errorf("rpc_C2S_FetchActivityDETAIL get playerComponent err=%s, uid=%d", err, uid)
		return nil, err
	}

	act := &pb.ActivityID{}
	if at, ok := arg.(*pb.ActivityID); ok {
		act = at
	} else {
		err := gamedata.GameError(aTypes.GetArgError)
		glog.Errorf("rpc_C2S_FetchActivityDETAIL get arg err=%s, uid=%d, activityID=%d", err, uid)
		return nil, err
	}

	activityID := int(act.ActivityID)
	openBool := playerCom.ConformOpen(activityID)
	timeBool := playerCom.ConformTime(activityID)

	if !timeBool || !openBool {
		err := gamedata.GameError(aTypes.NotConformCondition)
		return nil, err
	}
	activity := mod.GetActivityByID(activityID)
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("rpc_C2S_FetchActivityDETAIL get activity err=%s, uid=%d, activityID=%d", err, uid, activityID)
		return nil, err
	}
	rspData := &pb.ActivityData{}
	var err error
	switch activity.GetActivityType() {
	case consts.ActivityOfLogin:
		rspData, err = login.FetchActivityList(player, activityID)
	case consts.ActivityOfRecharge:
		rspData, err = recharge.FetchActivityList(player, activityID)
	case consts.ActivityOfOnline:
		rspData, err = online.FetchActivityList(player, activityID)
	case consts.ActivityOfFight:
		rspData, err = fight.FetchActivityList(player, activityID)
	case consts.ActivityOfVictory:
		rspData, err = win.FetchActivityList(player, activityID)
	case consts.ActivityOfRank:
		rspData, err = rank.FetchActivityList(player, activityID)
	case consts.ActivityOfConsume:
		rspData, err = consume.FetchActivityList(player, activityID)
	case consts.ActivityOfLoginRecharge:
		rspData, err = loginrecharge.FetchActivityList(player, activityID)
	case consts.ActivityOfFirstRecharge:
		rspData, err = firstrecharge.FetchActivityList(player, activityID)
	case consts.ActivityOfGrowPlan:
		rspData, err = growplan.FetchActivityList(player, activityID)
	case consts.ActivityOfSpring:
		rspData, err = spring.FetchActivityList(player, activityID)
	case consts.ActivityOfDailyRecharge:
		rspData, err = dailyrecharge.FetchActivityList(player, activityID)
	case consts.ActivityOfDailyShare:
		rspData, err = dailyshare.FetchActivityList(player, activityID)
	}

	if err != nil {
		glog.Errorf("rpc_C2S_FetchActivityDETAIL fetch activity list err=%s, uid%d, activityID=%d: ", err, uid, activityID)
	}
	return rspData, nil
}

//MessageID_C2S_RECEIVE_ACTIVITY_REWARD=336
func rpc_C2S_ReceiveActivityReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		err := gamedata.GameError(111)
		glog.Errorf("rpc_C2S_ReceiveActivityReward get player err=%s, uid=%d", err, uid)
		return false, err
	}

	playerCom := player.GetComponent(consts.ActivityCpt).(*activityComponent)
	if playerCom == nil {
		err := gamedata.GameError(101)
		glog.Errorf("rpc_C2S_ReceiveActivityReward get playerComponent err=%s, uid=%d", err, uid)
		return false, err
	}
	rewardInfo := &pb.TargetActivity{}
	if r, ok := arg.(*pb.TargetActivity); ok {
		rewardInfo = r
	} else {
		err := gamedata.GameError(100)
		glog.Errorf("rpc_C2S_ReceiveActivityReward get arg err=%s, uid=%d", err, uid)
		return false, err
	}

	activityID := int(rewardInfo.ID)
	rewardID := int(rewardInfo.RewardID)
	activity := mod.GetActivityByID(int(activityID))
	if activity == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("rpc_C2S_ReceiveActivityReward get activity err=%s, uid=%d, activityID=%d", err, uid, activityID)
		return nil, err
	}
	rd := &pb.Reward{
		RewardList: map[string]int32{},
	}
	var err error
	switch activity.GetActivityType() {
	case consts.ActivityOfLogin:
		err = login.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfRecharge:
		err = recharge.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfOnline:
		err = online.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfFight:
		err = fight.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfVictory:
		err = win.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfRank:
		err = rank.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfConsume:
		err = consume.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfLoginRecharge:
		err = loginrecharge.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfFirstRecharge:
		err = firstrecharge.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfGrowPlan:
		err = growplan.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfDailyRecharge:
		err = dailyrecharge.ReceiveReward(player, activityID, rewardID, rd)
	case consts.ActivityOfDailyShare:
		err = dailyshare.ReceiveReward(player, activityID, rewardID, rd)
	}
	if err != nil {
		glog.Errorf("rpc_C2S_ReceiveActivityReward receive reward err=%s, uid=%d, activityID=%d, rewardID=%d", err, uid, activityID, rewardID)
		return nil, err
	}
	playerCom.LogActivity(activityID, activity.GetActivityVersion(), activity.GetActivityType(), rewardID, aTypes.ActivityHasReceive)
	return rd, err
}

//MessageID_C2S_FETCH_ACTIVITY_LABEL_LIST=337
func rpc_C2S_FetchActivityLabelList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		err := gamedata.GameError(111)
		glog.Errorf("rpc_C2S_FetchActivityTypeList get player err=%s, uid=%d", err, uid)
		return nil, err
	}
	activityLabelList := &pb.ActivityLabelList{}
	activityLabelList.IDList = []int32{}

	playerCom := player.GetComponent(consts.ActivityCpt).(*activityComponent)
	if playerCom == nil {
		err := gamedata.GameError(101)
		glog.Errorf("rpc_C2S_FetchActivityTypeList get playerComponent err=%s, uid=%d", err, uid)
		return nil, err
	}

	activityLabelList.IDList = playerCom.activityTagList
	return activityLabelList, nil
}

func rpc_C2S_FetchFirstRechargeActivityDetail(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		err := gamedata.GameError(111)
		glog.Errorf("rpc_C2S_FetchActivityTypeList get player err=%s, uid=%d", err, uid)
		return nil, err
	}
	aidl := mod.getActivityIdListByType(consts.ActivityOfFirstRecharge)
	if len(aidl) < 1 {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("rpc_C2S_FetchFirstRechargeActivityDetail getActivityIdListByType err=%s, uid=%d", err, uid)
		return nil, err
	} else if len(aidl) > 1 {
		err := gamedata.GameError(aTypes.GetActivityError)
		for _, aid := range aidl {
			glog.Errorf("rpc_C2S_FetchFirstRechargeActivityDetail Not only one first recharge activity err=%s, uid=%d, activityID=%d", err, uid, aid)
		}
	}
	arg2 := &pb.ActivityID{}
	arg2.ActivityID = int32(aidl[0])
	rsd, err := rpc_C2S_FetchActivityDETAIL(agent, arg2)
	return rsd, err
}

func registerRpc() {
	spring.RegisterRpc()
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_ACTIVITY_DETAIL, rpc_C2S_FetchActivityDETAIL)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_RECEIVE_ACTIVITY_REWARD, rpc_C2S_ReceiveActivityReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_ACTIVITY_LABEL_LIST, rpc_C2S_FetchActivityLabelList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_FIRST_RECHARGE_ACTIVITY_DETAIL, rpc_C2S_FetchFirstRechargeActivityDetail)
}
