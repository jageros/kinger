package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type LuckBagReward struct {
	ID         int `json:"__id__"`
	Reward []string `json:"reward"`
	Prop int `json:"prop"`
}

type LuckBagRewardGameData struct {
	baseGameData
	Rewards []*LuckBagReward
}

func newLuckBagRewardGameData() *LuckBagRewardGameData {
	r := &LuckBagRewardGameData{}
	r.i = r
	return r
}

func (gd *LuckBagRewardGameData) name() string {
	return consts.LuckyBagReward
}

func (gd *LuckBagRewardGameData) init(d []byte) error {
	var l []*LuckBagReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Rewards = l

	return nil
}
