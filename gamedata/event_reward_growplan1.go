package gamedata

import (
	"encoding/json"
)

//type IActivityLoginRewardGameData interface {
//	IGameData
//	GetLoginRewardByID(rewardID int) *ActivityLoginReward
//}

type ActivityGrowPlanReward struct {
	ID            int      `json:"__id__"`
	ConditionType int      `json:"condition"`
	ConditionVal  []int   `json:"value"`
	Purchase      string  `json:"purchase"`
	Reward        []string `json:"reward"`
}

type ActivityGrowPlanRewardGameData struct {
	baseGameData
	name_                          string
	ID2ActivityGrowPlanReward map[int]*ActivityGrowPlanReward
}

func newActivityGrowPlanRewardGameData(name_ string) *ActivityGrowPlanRewardGameData {
	c := &ActivityGrowPlanRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityGrowPlanRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityGrowPlanRewardGameData) init(d []byte) error {
	var l []*ActivityGrowPlanReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ID2ActivityGrowPlanReward = map[int]*ActivityGrowPlanReward{}
	for _, c := range l {
		ar.ID2ActivityGrowPlanReward[c.ID] = c
	}
	return nil
}
