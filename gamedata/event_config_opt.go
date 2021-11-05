package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type IActivityRechargeReward interface {
	IsRechargeID(goods string) bool
}

type Activity struct {
	ID           int    `json:"__id__"`
	ActivityType int    `json:"type"`
	TimeID       int    `json:"timeType"`
	ConditionID  int    `json:"condition"`
	RewardTable  string `json:"reward"`
	ItemMaxDaily int    `json:"itemMaxDaily"`
	Version      int    `json:"version"`
}

type ActivityGameData struct {
	baseGameData
	ActivityMap map[int]*Activity
}

func newActivityGameData() *ActivityGameData {
	c := &ActivityGameData{}
	c.i = c
	return c
}

func (ad *ActivityGameData) name() string {
	return consts.ActivityConfig
}

func (ad *ActivityGameData) init(d []byte) error {
	var l []*Activity
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ad.ActivityMap = map[int]*Activity{}
	for _, c := range l {
		ad.ActivityMap[c.ID] = c
		if GetGameData(c.RewardTable) == nil {
			switch c.ActivityType {
			case consts.ActivityOfLogin:
				addGameData(newActivityLoginRewardGameData(c.RewardTable))
			case consts.ActivityOfRecharge:
				addGameData(newActivityRechargeRewardGameData(c.RewardTable))
			case consts.ActivityOfOnline:
				addGameData(newActivityOnlineRewardGameData(c.RewardTable))
			case consts.ActivityOfFight:
				addGameData(newActivityFightRewardGameData(c.RewardTable))
			case consts.ActivityOfVictory:
				addGameData(newActivityWinRewardGameData(c.RewardTable))
			case consts.ActivityOfRank:
				addGameData(newActivityRankRewardGameData(c.RewardTable))
			case consts.ActivityOfConsume:
				addGameData(newActivityConsumeRewardGameData(c.RewardTable))
			case consts.ActivityOfLoginRecharge:
				addGameData(newActivityLoginRechargeRewardGameData(c.RewardTable))
			case consts.ActivityOfFirstRecharge:
				addGameData(newActivityFirstRechargeRewardGameData(c.RewardTable))
			case consts.ActivityOfGrowPlan:
				addGameData(newActivityGrowPlanRewardGameData(c.RewardTable))
			case consts.ActivityOfDailyRecharge:
				addGameData(newActivityDailyRechargeRewardGameData(c.RewardTable))
			case consts.ActivityOfDailyShare:
				addGameData(newActivityDailyShareRewardGameData(c.RewardTable))
			}
		}
	}
	return nil
}
