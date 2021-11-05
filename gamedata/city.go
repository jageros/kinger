package gamedata

import (
	"encoding/json"

	"kinger/common/consts"
)

type City struct {
	ID              int      `json:"__id__"`
	Castle     []int32      `json:"castle"`        // 城防技
	AgricultureMax float64 `json:"agriculture_max"`  // 农业上限
	BusinessMax float64 `json:"business_max"`        // 商业上限
	DefenseMax float64 `json:"defense_max"`          // 城防上限
	MilitaryMax float64 `json:"military_max"`        // 兵役上限
	GloryBase float64 `json:"gloryBase"`         // 基础荣耀
}

type CityGameData struct {
	baseGameData
	ID2City map[int]*City
}

func newCityGameData() *CityGameData {
	c := &CityGameData{}
	c.i = c
	return c
}

func (cg *CityGameData) name() string {
	return consts.City
}

func (cg *CityGameData) init(d []byte) error {
	var l []*City
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	cg.ID2City = map[int]*City{}
	for _, c := range l {
		cg.ID2City[c.ID] = c
	}

	return nil
}
