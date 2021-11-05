package gamedata

import (
	"encoding/json"
)

type ActivityRankReward struct {
	ID       int      `json:"__id__"`
	Rank int      `json:"recharge"`
	Reward   []string `json:"reward"`
}

type ActivityRankRewardGameData struct {
	baseGameData
	name_ string
	ActivityRankRewardMap map[int]*ActivityRankReward
}

func newActivityRankRewardGameData(name_ string) *ActivityRankRewardGameData {
	c := &ActivityRankRewardGameData{name_:name_}
	c.i = c
	return c
}

func (ar *ActivityRankRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityRankRewardGameData) init(d []byte) error {
	var l []*ActivityRankReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityRankRewardMap = map[int]*ActivityRankReward{}
	for _, c := range l {
		ar.ActivityRankRewardMap[c.ID] = c
	}
	return nil
}
