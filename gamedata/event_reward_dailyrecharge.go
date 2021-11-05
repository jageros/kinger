package gamedata

import (
	"encoding/json"
)

type ActivityDailyRechargeReward struct {
	ID       int      `json:"__id__"`
	Recharge int      `json:"recharge"`
	Reward   []string `json:"reward"`
}

type ActivityDailyRechargeRewardGameData struct {
	baseGameData
	name_ string
	ID2ActivityDailyRechargeReward map[int]*ActivityDailyRechargeReward
}

func newActivityDailyRechargeRewardGameData(name_ string) *ActivityDailyRechargeRewardGameData {
	c := &ActivityDailyRechargeRewardGameData{name_:name_}
	c.i = c
	return c
}

func (ar *ActivityDailyRechargeRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityDailyRechargeRewardGameData) init(d []byte) error {
	var l []*ActivityDailyRechargeReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ID2ActivityDailyRechargeReward = map[int]*ActivityDailyRechargeReward{}
	for _, c := range l {
		ar.ID2ActivityDailyRechargeReward[c.ID] = c
	}
	return nil
}
