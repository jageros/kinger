package gamedata

import (
	"encoding/json"
)

type ActivityLoginReward struct {
	ID       int      `json:"__id__"`
	LoginDay int      `json:"loginDay"`
	Reward   []string `json:"reward"`
}

type ActivityLoginRewardGameData struct {
	baseGameData
	name_                  string
	ActivityLoginRewardMap map[int]*ActivityLoginReward
}

func newActivityLoginRewardGameData(name_ string) *ActivityLoginRewardGameData {
	c := &ActivityLoginRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityLoginRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityLoginRewardGameData) init(d []byte) error {
	var l []*ActivityLoginReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityLoginRewardMap = map[int]*ActivityLoginReward{}
	for _, c := range l {
		ar.ActivityLoginRewardMap[c.ID] = c
	}
	return nil
}
