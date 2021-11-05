package types

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
)

var IMod IActivityMod

type IActivityMod interface {
	GetActivityByID(int) IActivity
	//GetIActivityListByType(t pb.ActivityTypeEnum) IActivityMod
	GetActivityIdList() []int
	ForEachActivityDataByType(atype int, callback func(data IActivity))
	InitDataByType(int) IActivityMod
	NewPCM(player types.IPlayer) IPlayerCom
}

type IActivity interface {
	GetTimeCondition() *gamedata.ActivityTime
	GetOpenCondition() *gamedata.ActivityOpenCondition
	GetRewardTableName() string
	GetGameData() *gamedata.Activity
	GetActivityType() int
	GetActivityVersion() int
	GetActivityId() int
}

type IPlayerCom interface {
	InitAttr(string) *attribute.MapAttr
	GetActivityTagList() []int32
	UpdateActivityTagList([]int32)
	ConformOpen(int) bool
	ConformTime(int) bool
	Conform(int) bool
	GiveReward(string, int, *pb.Reward, int, int) error
	LogActivity(aid, vid, aty, rid, eid int)
	PushFinshNum(aid, rid, num int)
	PushReceiveStatus(aid, rid int, rst pb.ActivityReceiveStatus)
}

type BaseActivity struct {
	IAMod IActivityMod
}

func IsInArry(t int32, arry []int32) bool {
	for _, a := range arry {
		if a == t {
			return true
		}
	}
	return false
}

func IsInArryInt(t int, arry []int) bool {
	for _, a := range arry {
		if a == t {
			return true
		}
	}
	return false
}

func GenModifyResourceReason(activityID, activityType, rewardID int) string {
	return fmt.Sprintf("%s%d_%d_%d", consts.RmrActivityRewardPrefix, activityType, activityID, rewardID)
}