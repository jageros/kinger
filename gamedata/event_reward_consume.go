package gamedata

import (
	"encoding/json"
)

type ActivityConsumeReward struct {
	ID      int      `json:"__id__"`
	Consume int      `json:"consume"`
	Reward  []string `json:"reward"`
}

type ActivityConsumeRewardGameData struct {
	baseGameData
	name_                    string
	ActivityConsumeRewardMap map[int]*ActivityConsumeReward
}

func newActivityConsumeRewardGameData(name_ string) *ActivityConsumeRewardGameData {
	c := &ActivityConsumeRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityConsumeRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityConsumeRewardGameData) init(d []byte) error {
	var l []*ActivityConsumeReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityConsumeRewardMap = map[int]*ActivityConsumeReward{}
	for _, c := range l {
		ar.ActivityConsumeRewardMap[c.ID] = c
	}
	return nil
}
