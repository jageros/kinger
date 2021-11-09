package gamedata

import (
	"encoding/json"
)

type ActivityDailyShareReward struct {
	ID       int      `json:"__id__"`
	ShareCnt int      `json:"shareCnt"`
	Reward   []string `json:"reward"`
}

type ActivityDailyShareRewardGameData struct {
	baseGameData
	name_                       string
	ID2ActivityDailyShareReward map[int]*ActivityDailyShareReward
}

func newActivityDailyShareRewardGameData(name_ string) *ActivityDailyShareRewardGameData {
	c := &ActivityDailyShareRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityDailyShareRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityDailyShareRewardGameData) init(d []byte) error {
	var l []*ActivityDailyShareReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ID2ActivityDailyShareReward = map[int]*ActivityDailyShareReward{}
	for _, c := range l {
		ar.ID2ActivityDailyShareReward[c.ID] = c
	}
	return nil
}
