package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type LeagueRankReward struct {
	ID         int      `json:"__id__"`
	Rank       []int    `json:"rank"`
	LeagueTeam int      `json:"leagueTeam"`
	Reward     []string `json:"reward"`
	KingFlag   int      `json:"kingFlag"`
}

type LeagueRankRewardGameData struct {
	baseGameData
	ID2LeagueRankReward  map[int]*LeagueRankReward
}

func newLeagueRankRewardGameData() *LeagueRankRewardGameData {
	r := &LeagueRankRewardGameData{}
	r.i = r
	return r
}

func (r *LeagueRankRewardGameData) name() string {
	return consts.LeagueRankReward
}

func (r *LeagueRankRewardGameData) init(d []byte) error {
	var l []*LeagueRankReward

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	r.ID2LeagueRankReward = map[int]*LeagueRankReward{}

	for _, t := range l {
		r.ID2LeagueRankReward[t.ID] = t
	}
	return nil
}

func (r *LeagueRankRewardGameData) GetRewardByRank(rank int) (lvl int, rwl []string, kingFlag int) {
	for _, rw := range r.ID2LeagueRankReward {
		if rank >= rw.Rank[0] && rank <= rw.Rank[1] {
			return rw.LeagueTeam, rw.Reward, rw.KingFlag
		}
	}
	return 0, nil, 0
}