package gamedata

import (
	"encoding/json"
)

type ActivityWinReward struct {
	ID      int      `json:"__id__"`
	Country []int    `json:"country"`
	Wid     int      `json:"wid"`
	Reward  []string `json:"reward"`
}

type ActivityWinRewardGameData struct {
	baseGameData
	name_ string
	ActivityWinRewardMap map[int]*ActivityWinReward
}

func newActivityWinRewardGameData(name_ string) *ActivityWinRewardGameData {
	c := &ActivityWinRewardGameData{name_:name_}
	c.i = c
	return c
}

func (ar *ActivityWinRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityWinRewardGameData) init(d []byte) error {
	var l []*ActivityWinReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityWinRewardMap = map[int]*ActivityWinReward{}
	for _, c := range l {
		ar.ActivityWinRewardMap[c.ID] = c
	}
	return nil
}
