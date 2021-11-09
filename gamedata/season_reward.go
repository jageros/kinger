package gamedata

import (
	"encoding/json"
	"kinger/common/config"
	"kinger/common/consts"
	"sort"
)

type SeasonReward struct {
	ID        int     `json:"__id__"`
	Team      int     `json:"team"`
	Ranking   int     `json:"ranking"`
	CardSkin  string  `json:"skin"`
	HeadFrame string  `json:"headFrame"`
	Treasure  string  `json:"treasure"`
	WinAmount int     `json:"winFlag"`
	Areas     [][]int `json:"areas"`

	areaLimit *AreaLimitConfig
}

func (sr *SeasonReward) init() {
	sr.areaLimit = newAreaLimitConfig(sr.Areas)

	/*
		sr.areaLimit = newAreaLimitConfig(sr.Areas)
		if sr.CardSkin != "" {
			if _, ok := GetGameData(consts.CardSkin).(*CardSkinGameData).ID2CardSkin[sr.CardSkin]; !ok {
				sr.CardSkin = ""
			}
		}

		if sr.HeadFrame != "" {
			if _, ok := GetGameData(consts.HeadFrame).(*HeadFrameGameData).ID2HeadFrame[sr.HeadFrame]; !ok {
				sr.HeadFrame = ""
			}
		}

		if sr.Treasure != "" {
			if _, ok := GetGameData(consts.Treasure).(*TreasureGameData).Treasures[sr.Treasure]; !ok {
				sr.Treasure = ""
			}
		}
	*/
}

type SeasonRewardGameData struct {
	baseGameData
	area2Rewards map[int][]*SeasonReward
}

func newSeasonRewardGameData() *SeasonRewardGameData {
	gd := &SeasonRewardGameData{}
	gd.i = gd
	return gd
}

func (gd *SeasonRewardGameData) name() string {
	return consts.SeasonReward
}

func (gd *SeasonRewardGameData) init(d []byte) error {
	var l []*SeasonReward

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	isMultiLan := config.GetConfig().IsMultiLan
	sort.Slice(l, func(i, j int) bool {
		r1 := l[i]
		r2 := l[j]
		if r1.Ranking > r2.Ranking {
			return true
		} else if r1.Ranking < r2.Ranking {
			return false
		}

		if isMultiLan {

			if r1.Team > r2.Team {
				return true
			} else if r1.Team < r2.Team {
				return false
			}

		} else {
			if r1.WinAmount > r2.WinAmount {
				return true
			} else if r1.WinAmount < r2.WinAmount {
				return false
			}
		}
		return r1.ID >= r2.ID
	})

	gd.area2Rewards = map[int][]*SeasonReward{}
	for _, r := range l {
		r.init()
		r.areaLimit.forEachArea(func(area int) {
			rewards := gd.area2Rewards[area]
			gd.area2Rewards[area] = append(rewards, r)
		})
	}

	return nil
}

func (gd *SeasonRewardGameData) GetRewards(area int) []*SeasonReward {
	return gd.area2Rewards[area]
}
