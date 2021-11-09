package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type League struct {
	ID          int      `json:"__id__"`
	LeagueLevel int      `json:"leagueLevel"`
	RankScore   int      `json:"rank_integral"`
	Reward      []string `json:"reward"`

	NextScore int
}

type LeagueGameData struct {
	baseGameData
	LeagueList []*League
	ID2League  map[int]*League
	MaxScore   int
}

func newLeagueGameData() *LeagueGameData {
	r := &LeagueGameData{}
	r.i = r
	return r
}

func (r *LeagueGameData) name() string {
	return consts.League
}

func (r *LeagueGameData) init(d []byte) error {
	var l []*League

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	r.LeagueList = l
	r.ID2League = map[int]*League{}
	r.MaxScore = 0
	ll := len(l)
	for k, t := range l {
		r.ID2League[t.ID] = t
		if t.RankScore > r.MaxScore {
			r.MaxScore = t.RankScore
		}
		if k+1 < ll {
			t.NextScore = l[k+1].RankScore
		}
	}

	return nil
}

func (r *LeagueGameData) getLeagueByScore(score int) *League {
	for _, l := range r.LeagueList {
		if score >= l.RankScore && (score < l.NextScore || l.NextScore <= 0) {
			return l
		}
	}
	return nil
}

func (r *LeagueGameData) GetRewardById(id int) []string {
	if l, ok := r.ID2League[id]; ok {
		return l.Reward
	}
	return nil
}

func (r *LeagueGameData) GetLeagueEndRewardLvlByScore(score int) int {
	league := r.getLeagueByScore(score)
	if league != nil {
		return league.LeagueLevel
	}
	return 0
}

func (r *LeagueGameData) GetLeagueIdByScore(score int) int {
	league := r.getLeagueByScore(score)
	if league != nil {
		return league.ID
	}
	return 0
}

func (r *LeagueGameData) GetIdList() []int {
	var ids []int
	for _, l := range r.LeagueList {
		ids = append(ids, l.ID)
	}
	return ids
}

func (r LeagueGameData) GetScoreById(id int) int {
	if l, ok := r.ID2League[id]; ok {
		return l.RankScore
	}
	return 0
}
