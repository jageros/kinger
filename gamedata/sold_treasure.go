package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/config"
	"kinger/common/consts"
	"sort"
)

type SoldTreasure struct {
	ID                string  `json:"__id__"`
	TreasureModelID   string  `json:"treasureId"`
	JadePrice         int     `json:"jadePrice"`
	Team              int     `json:"team"`
	PvpLevelCondition int     `json:"pvpLevelCondition"`
	BowlderPrice      int     `json:"bowlderPrice"`
	Areas             [][]int `json:"areas"`
	Camp              int     `json:"country"`
	Order             int     `json:"order"`

	areaLimit *AreaLimitConfig
}

func (st *SoldTreasure) init() {
	st.areaLimit = newAreaLimitConfig(st.Areas)
}

type ISoldTreasureGameData interface {
	IGameData
	GetTeam2Treasures(area int) map[int][]*SoldTreasure
	GetCampTreasure(area, camp, team, idx int) *SoldTreasure
	GetTreasureByID(treasureID string) *SoldTreasure
}

type SoldTreasureGameData struct {
	baseGameData
	areaVersion     int
	rawData         []byte
	areaToTreasures map[int]map[int]map[int][]*SoldTreasure // map[area]map[camp]map[team][]*SoldTreasure
	id2Treasure     map[string]*SoldTreasure
}

func newSoldTreasureGameData() *SoldTreasureGameData {
	sg := &SoldTreasureGameData{}
	sg.i = sg
	return sg
}

func (sg *SoldTreasureGameData) name() string {
	return consts.SoldTreasure
}

func (sg *SoldTreasureGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if sg.areaVersion == areaVersion && bytes.Equal(sg.rawData, d) {
		return errors.New("no update")
	}

	var l []*SoldTreasure
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	sg.areaVersion = areaVersion
	sg.rawData = d
	sg.areaToTreasures = map[int]map[int]map[int][]*SoldTreasure{}
	sg.id2Treasure = map[string]*SoldTreasure{}

	for _, st := range l {
		sg.id2Treasure[st.TreasureModelID] = st
		st.init()
		st.areaLimit.forEachArea(func(area int) {
			camp2Treasures, ok := sg.areaToTreasures[area]
			if !ok {
				camp2Treasures = map[int]map[int][]*SoldTreasure{}
				sg.areaToTreasures[area] = camp2Treasures
			}

			team2Treasures, ok := camp2Treasures[st.Camp]
			if !ok {
				team2Treasures = map[int][]*SoldTreasure{}
				camp2Treasures[st.Camp] = team2Treasures
			}

			ts := team2Treasures[st.Team]
			team2Treasures[st.Team] = append(ts, st)
		})
	}

	for _, camp2Treasures := range sg.areaToTreasures {
		for _, team2Treasures := range camp2Treasures {
			for team, ts := range team2Treasures {
				sort.Slice(ts, func(i, j int) bool {
					return ts[i].Order <= ts[j].Order
				})
				team2Treasures[team] = ts
			}
		}
	}

	return nil
}

func (sg *SoldTreasureGameData) GetTeam2Treasures(area int) map[int][]*SoldTreasure {
	return map[int][]*SoldTreasure{}
}

func (sg *SoldTreasureGameData) GetCampTreasure(area, camp, team, idx int) *SoldTreasure {
	if camp2Treasures, ok := sg.areaToTreasures[area]; ok {
		if team2Treasures, ok := camp2Treasures[camp]; ok {
			if ts, ok := team2Treasures[team]; ok {
				n := len(ts)
				if n <= 0 {
					return nil
				}

				for i, t := range ts {
					if i == idx {
						return t
					}
				}
				return ts[n-1]
			}
		}
	}
	return nil
}

func (sg *SoldTreasureGameData) GetTreasureByID(treasureID string) *SoldTreasure {
	return sg.id2Treasure[treasureID]
}

type SoldTreasureHandjoyGameData struct {
	SoldTreasureGameData
}

func newSoldTreasureHandjoyGameData() *SoldTreasureHandjoyGameData {
	sg := &SoldTreasureHandjoyGameData{}
	sg.i = sg
	return sg
}

func (sg *SoldTreasureHandjoyGameData) name() string {
	return consts.SoldTreasureHandjoy
}

func (sg *SoldTreasureHandjoyGameData) GetCampTreasure(area, camp, team, idx int) *SoldTreasure {
	return nil
}

func GetSoldTreasureGameData() ISoldTreasureGameData {
	if config.GetConfig().IsMultiLan {
		return GetGameData(consts.SoldTreasureHandjoy).(*SoldTreasureHandjoyGameData)
	} else {
		return GetGameData(consts.SoldTreasure).(*SoldTreasureGameData)
	}
}
