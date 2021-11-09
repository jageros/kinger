package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RecruitRefreshConfig struct {
	ID    int `json:"__id__"`
	Weeks int `json:"weeks"`
	Day   int `json:"day"`
	Hours int `json:"hours"`
	Min   int `json:"minute"`
	Sec   int `json:"second"`
}

type RecruitRefreshConfigGameData struct {
	baseGameData
	ID2RecruitRefreshConfig map[int]*RecruitRefreshConfig
}

func newRecruitRefreshConfigGameData() *RecruitRefreshConfigGameData {
	c := &RecruitRefreshConfigGameData{}
	c.i = c
	return c
}

func (ac *RecruitRefreshConfigGameData) name() string {
	return consts.RecruitRefreshConfig
}

func (ac *RecruitRefreshConfigGameData) init(d []byte) error {
	var l []*RecruitRefreshConfig
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ac.ID2RecruitRefreshConfig = map[int]*RecruitRefreshConfig{}
	for _, c := range l {
		ac.ID2RecruitRefreshConfig[c.ID] = c
	}
	return nil
}

func (r *RecruitRefreshConfig) InitCfg() {
	gd := GetGameData(consts.RecruitRefreshConfig).(*RecruitRefreshConfigGameData)
	if g, ok := gd.ID2RecruitRefreshConfig[1]; ok {
		r.ID = g.ID
		r.Weeks = g.Weeks
		r.Day = g.Day
		r.Hours = g.Hours
		r.Min = g.Min
		r.Sec = g.Sec
	}
}
