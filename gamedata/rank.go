package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Rank struct {
	ID           int      `json:"__id__"`
	Team         int      `json:"team"`
	LevelUpStar  int      `json:"levelUpStar"`  // 当前段位的满分，再拿一份就升段了（已废弃）
	Unlock       []uint32    `json:"unlock"`
	GoldReward   int      `json:"goldReward"`
	RankUpReward []uint32 `json:"rankUpReward"`
	MatchUpper int `json:"matchUpper"`
	MatchLower int `json:"matchLower"`
	Protection int `json:"protection"`
	IndexInterval float64 `json:"index_interval"`  // 匹配区间
	IndexRedline float64 `json:"index_redline"`   // 匹配红线
	Kvalue float64 `json:"k_value"`
	RankIntegral int `json:"rank_integra"`
	PositiveIQ int `json:"PositiveIQ"`
	NegativeIQ int `json:"NegativeIQ"`

	OriginUnlockCard []uint32
}

type RankGameData struct {
	baseGameData
	Ranks       map[int]*Rank
	RanksOfTeam map[int][]*Rank
	RankList    []*Rank
	MaxTeam     int
}

func newRankGameData() *RankGameData {
	r := &RankGameData{}
	r.i = r
	return r
}

func (r *RankGameData) name() string {
	return consts.Rank
}

func (r *RankGameData) init(d []byte) error {
	var l []*Rank

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("treasures = %s", l)
	r.RankList = l
	r.Ranks = map[int]*Rank{}
	r.RanksOfTeam = map[int][]*Rank{}
	r.MaxTeam = 0
	lastStar := 0
	var lastUnlock []uint32

	for _, t2 := range l {
		t2.LevelUpStar += lastStar
		lastStar = t2.LevelUpStar

		t2.OriginUnlockCard = make([]uint32, len(t2.Unlock))
		copy(t2.OriginUnlockCard, t2.Unlock)
		t2.Unlock = append(t2.Unlock, lastUnlock...)
		lastUnlock = t2.Unlock

		if t2.Team > r.MaxTeam {
			r.MaxTeam = t2.Team
		}

		r.Ranks[t2.ID] = t2
		ll, ok := r.RanksOfTeam[t2.Team]
		if !ok {
			ll = []*Rank{}
			r.RanksOfTeam[t2.Team] = ll
		}
		r.RanksOfTeam[t2.Team] = append(ll, t2)
	}

	return nil
}

func (r *RankGameData) GetPvpLevelByStar(star int) int {
	if star <= 0 {
		return 1
	}

	for i := 0; i < len(r.RankList); i++ {
		rank := r.RankList[i]
		if star <= rank.LevelUpStar {
			return rank.ID
		}
	}
	return r.RankList[len(r.RankList)-1].ID
}

func (r *RankGameData) GetTeamByStar(star int) int {
	level := r.GetPvpLevelByStar(star)
	return r.Ranks[level].Team
}
