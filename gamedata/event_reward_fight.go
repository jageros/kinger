package gamedata

import (
	"encoding/json"
)

type ActivityFightReward struct {
	ID       int      `json:"__id__"`
	FightCnt int      `json:"fightCnt"`
	Reward   []string `json:"reward"`
}

type ActivityFightRewardGameData struct {
	baseGameData
	name_ string
	ActivityFightRewardMap map[int]*ActivityFightReward
}

func newActivityFightRewardGameData(name_ string) *ActivityFightRewardGameData {
	c := &ActivityFightRewardGameData{name_:name_}
	c.i = c
	return c
}

func (ar *ActivityFightRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityFightRewardGameData) init(d []byte) error {
	var l []*ActivityFightReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityFightRewardMap = map[int]*ActivityFightReward{}
	for _, c := range l {
		ar.ActivityFightRewardMap[c.ID] = c
	}
	return nil
}
