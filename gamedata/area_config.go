package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/consts"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/utils"
	"sort"
	"time"
)

type AreaConfig struct {
	Area     int    `json:"__id__"`
	OpenTime string `json:"openTime"`

	openTime time.Time
}

func (ac *AreaConfig) init() (err error) {
	ac.openTime, err = utils.StringToTime(ac.OpenTime, utils.TimeFormat2)
	return
}

func (ac *AreaConfig) IsOpen() bool {
	return !time.Now().Before(ac.openTime)
}

type AreaConfigGameData struct {
	baseGameData
	rawData      []byte
	Version      int
	AreaToConfig map[int]*AreaConfig
	Areas        []*AreaConfig
	curArea      *AreaConfig
	MaxArea      *AreaConfig
}

func newAreaConfigGameData() *AreaConfigGameData {
	r := &AreaConfigGameData{}
	r.i = r
	return r
}

func (gd *AreaConfigGameData) name() string {
	return consts.AreaConfig
}

func (gd *AreaConfigGameData) init(d []byte) error {
	if bytes.Equal(gd.rawData, d) {
		return errors.New("no update")
	}

	var l []*AreaConfig
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.rawData = d
	gd.Version += 1
	gd.AreaToConfig = map[int]*AreaConfig{}
	gd.MaxArea = nil
	gd.curArea = nil
	gd.Areas = l
	for _, c := range l {
		err := c.init()
		if err != nil {
			return err
		}

		if gd.MaxArea == nil || c.Area > gd.MaxArea.Area {
			gd.MaxArea = c
		}
		if c.IsOpen() {
			gd.curArea = c
		}

		gd.AreaToConfig[c.Area] = c
	}

	if gd.curArea == nil {
		return errors.New("no opened area")
	}

	return nil
}

func (gd *AreaConfigGameData) GetCurArea() *AreaConfig {
	if gd.curArea.Area >= gd.MaxArea.Area {
		return gd.curArea
	}

	for a := gd.MaxArea.Area; a > gd.curArea.Area; a++ {
		if area, ok := gd.AreaToConfig[a]; !ok || !area.IsOpen() {
			continue
		} else {
			gd.curArea = area
			return area
		}
	}
	return gd.curArea
}

func (gd *AreaConfigGameData) ForEachOpenedArea(callback func(config *AreaConfig)) {
	for _, a := range gd.Areas {
		if a.IsOpen() {
			callback(a)
		}
	}
}

type AreaLimitConfig struct {
	effectiveAreas common.IntSet // 有效的区，为nil时，所有区都有效
}

func newAreaLimitConfig(areasInfo [][]int) *AreaLimitConfig {
	alc := &AreaLimitConfig{}
	if len(areasInfo) <= 0 {
		// 所有区都有效
		return alc
	}

	for _, areaSec := range areasInfo {
		if len(areaSec) <= 0 {
			continue
		}

		sort.Ints(areaSec)
		minArea := areaSec[0]
		maxArea := areaSec[len(areaSec)-1]

		for area := minArea; area <= maxArea; area++ {
			if area <= 0 {
				// 所有区都无效
				alc.effectiveAreas = common.IntSet{}
				return alc
			}

			if alc.effectiveAreas == nil {
				alc.effectiveAreas = common.IntSet{}
			}
			alc.effectiveAreas.Add(area)
		}
	}
	return alc
}

func (alc *AreaLimitConfig) IsEffective(area int) bool {
	if alc.effectiveAreas == nil {
		return true
	}
	return alc.effectiveAreas.Contains(area)
}

func (alc *AreaLimitConfig) forEachArea(callback func(area int)) {
	if alc.effectiveAreas == nil {
		areaGameData := GetGameData(consts.AreaConfig).(*AreaConfigGameData)
		for _, areaCfg := range areaGameData.Areas {
			callback(areaCfg.Area)
		}
		return
	}

	alc.effectiveAreas.ForEach(callback)
}
