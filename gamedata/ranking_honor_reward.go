package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RankHonorReward struct {
	ID       int    `json:"__id__"`
	Rank     []int  `json:"rank"`
	Treasure string `json:"treasure"`
	Gold     int    `json:"gold"`
	Jade     int    `json:"jade"`
}

type RankHonorRewardData struct {
	baseGameData
	RankReward map[int]*RankHonorReward
}

func (rr *RankHonorRewardData) name() string {
	return consts.RankHonorReward
}

func newRankHonorRewardData() *RankHonorRewardData {
	gd := &RankHonorRewardData{}
	gd.i = gd
	return gd
}

func (rr *RankHonorRewardData) init(d []byte) error {
	var l []*RankHonorReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	rr.RankReward = make(map[int]*RankHonorReward)
	for _, d := range l {
		rr.RankReward[d.ID] = d
	}
	return nil
}
