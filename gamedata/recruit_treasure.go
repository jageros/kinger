package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/consts"
	"math"
	"sort"
	"time"
)

type RecruitTreasure struct {
	ID           string  `json:"__id__"`
	TreasureID   string  `json:"treasureId"`
	JadePrice    int     `json:"jadePrice"`
	Team         int     `json:"team"`
	Switch       int     `json:"switch"`
	OpenWeekDays []int   `json:"openWeekDays"`
	Areas        [][]int `json:"areas"`

	areaLimit          *AreaLimitConfig
	openWeekDaysAmount int
}

func (r *RecruitTreasure) init() {
	r.areaLimit = newAreaLimitConfig(r.Areas)
	sort.Ints(r.OpenWeekDays)
	r.openWeekDaysAmount = len(r.OpenWeekDays)
	for i, weekDay := range r.OpenWeekDays {
		if weekDay == 7 {
			r.OpenWeekDays[i] = 0
		}
	}
}

func (r *RecruitTreasure) IsOpen() bool {
	if r.openWeekDaysAmount <= 0 {
		return false
	}

	curWeekDay := int(time.Now().Weekday())
	for _, weekDay := range r.OpenWeekDays {
		if curWeekDay == weekDay {
			return true
		}
	}
	return false
}

func (r *RecruitTreasure) GetNextOpenRemainTime() int {
	if r.openWeekDaysAmount <= 0 {
		return math.MaxInt32
	}

	now := time.Now()
	curWeekDay := int(now.Weekday())
	remainDay := math.MaxInt32
	for _, weekDay := range r.OpenWeekDays {
		if curWeekDay == weekDay {
			return 0
		}

		dayDiff := (weekDay - curWeekDay + 7) % 7
		if dayDiff < remainDay {
			remainDay = dayDiff
		}
	}

	return remainDay*24*3600 - (now.Hour()*3600 + now.Minute()*60 + now.Second())
}

type RecruitTreasureGameData struct {
	baseGameData
	areaVersion    int
	rawData        []byte
	area2Treausre  map[int]map[int]map[int]*RecruitTreasure // map[area]map[team]map[switch]*RecruitTreasure
	area2MaxSwitch map[int]int
}

func newRecruitTreasureGameData() *RecruitTreasureGameData {
	gd := &RecruitTreasureGameData{}
	gd.i = gd
	return gd
}

func (gd *RecruitTreasureGameData) name() string {
	return consts.RecruitTreausre
}

func (gd *RecruitTreasureGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if gd.areaVersion == areaVersion && bytes.Equal(gd.rawData, d) {
		return errors.New("no update")
	}

	var l []*RecruitTreasure
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.areaVersion = areaVersion
	gd.rawData = d
	gd.area2Treausre = map[int]map[int]map[int]*RecruitTreasure{}
	gd.area2MaxSwitch = map[int]int{}
	for _, t := range l {
		t.init()
		t.areaLimit.forEachArea(func(area int) {
			team2Treausre, ok := gd.area2Treausre[area]
			if !ok {
				team2Treausre = map[int]map[int]*RecruitTreasure{}
				gd.area2Treausre[area] = team2Treausre
			}
			swch2Treasure, ok := team2Treausre[t.Team]
			if !ok {
				swch2Treasure = map[int]*RecruitTreasure{}
				team2Treausre[t.Team] = swch2Treasure
			}
			swch2Treasure[t.Switch] = t
			if sw, ok := gd.area2MaxSwitch[area]; ok {
				if t.Switch > sw {
					gd.area2MaxSwitch[area] = t.Switch
				}
			} else {
				gd.area2MaxSwitch[area] = t.Switch
			}
		})
	}

	return nil
}

func (gd *RecruitTreasureGameData) GetTeam2Treausre(area int) map[int]map[int]*RecruitTreasure {
	if team2Treausre, ok := gd.area2Treausre[area]; ok {
		return team2Treausre
	} else {
		return map[int]map[int]*RecruitTreasure{}
	}
}

func (gd *RecruitTreasureGameData) GetMaxSwitchByArea(area int) int {
	if sw, ok := gd.area2MaxSwitch[area]; ok {
		return sw
	}
	return 0
}
