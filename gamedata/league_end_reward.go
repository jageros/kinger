package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type LeagueReward struct {
	ID          int      `json:"__id__"`
	Reward     []string `json:"reward"`
}

type LeagueRewardGameData struct {
	baseGameData
	ID2LeagueReward  map[int]*LeagueReward
}

func newLeagueRewardGameData() *LeagueRewardGameData {
	r := &LeagueRewardGameData{}
	r.i = r
	return r
}

func (r *LeagueRewardGameData) name() string {
	return consts.LeagueReward
}

func (r *LeagueRewardGameData) init(d []byte) error {
	var l []*LeagueReward

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	r.ID2LeagueReward = map[int]*LeagueReward{}

	for _, t := range l {
		r.ID2LeagueReward[t.ID] = t
	}
	return nil
}

func (r *LeagueRewardGameData) GetRewardByLeagueLvl(lv int) []string {
	if rw, ok := r.ID2LeagueReward[lv]; ok {
		return rw.Reward
	}
	return nil
}