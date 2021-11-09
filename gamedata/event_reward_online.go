package gamedata

import (
	"encoding/json"
)

type ActivityOnlineReward struct {
	ID     int      `json:"__id__"`
	Times  string   `json:"times"`
	Reward []string `json:"reward"`
}

type ActivityOnlineRewardGameData struct {
	baseGameData
	name_                   string
	ActivityOnlineRewardMap map[int]*ActivityOnlineReward
}

func newActivityOnlineRewardGameData(name_ string) *ActivityOnlineRewardGameData {
	c := &ActivityOnlineRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ao *ActivityOnlineRewardGameData) name() string {
	return ao.name_
}

func (ao *ActivityOnlineRewardGameData) init(d []byte) error {
	var l []*ActivityOnlineReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ao.ActivityOnlineRewardMap = map[int]*ActivityOnlineReward{}
	for _, c := range l {
		ao.ActivityOnlineRewardMap[c.ID] = c
	}
	return nil
}
