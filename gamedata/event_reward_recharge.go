package gamedata

import (
	"encoding/json"
)

type ActivityRechargeReward struct {
	ID       int      `json:"__id__"`
	Recharge int      `json:"recharge"`
	Reward   []string `json:"reward"`
}

type ActivityRechargeRewardGameData struct {
	baseGameData
	name_ string
	ActivityRechargeRewardMap map[int]*ActivityRechargeReward
}

func newActivityRechargeRewardGameData(name_ string) *ActivityRechargeRewardGameData {
	c := &ActivityRechargeRewardGameData{name_:name_}
	c.i = c
	return c
}

func (ar *ActivityRechargeRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityRechargeRewardGameData) init(d []byte) error {
	var l []*ActivityRechargeReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityRechargeRewardMap = map[int]*ActivityRechargeReward{}
	for _, c := range l {
		ar.ActivityRechargeRewardMap[c.ID] = c
	}
	return nil
}
