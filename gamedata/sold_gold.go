package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/config"
	"kinger/common/consts"
)

type SoldGold struct {
	GoodsID      string  `json:"__id__"`
	Gold         int     `json:"soldGold"`
	JadePrice    int     `json:"jadePrice"`
	BowlderPrice int     `json:"bowlderPrice"`
	Areas        [][]int `json:"areas"`

	areaLimit *AreaLimitConfig
}

func (g *SoldGold) init() {
	g.areaLimit = newAreaLimitConfig(g.Areas)
}

type ISoldGoldGameData interface {
	IGameData
	GetGoodsList(area int) []*SoldGold
}

type SoldGoldGameData struct {
	baseGameData
	areaVersion     int
	rawData         []byte
	areaToGoodsList map[int][]*SoldGold
}

func newSoldGoldGameData() *SoldGoldGameData {
	sg := &SoldGoldGameData{}
	sg.i = sg
	return sg
}

func (sg *SoldGoldGameData) name() string {
	return consts.SoldGold
}

func (sg *SoldGoldGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if sg.areaVersion == areaVersion && bytes.Equal(sg.rawData, d) {
		return errors.New("no update")
	}

	var l []*SoldGold
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	sg.areaVersion = areaVersion
	sg.rawData = d
	sg.areaToGoodsList = map[int][]*SoldGold{}
	for _, g := range l {
		g.init()
		g.areaLimit.forEachArea(func(area int) {
			goodsList := sg.areaToGoodsList[area]
			sg.areaToGoodsList[area] = append(goodsList, g)
		})
	}

	return nil
}

func (sg *SoldGoldGameData) GetGoodsList(area int) []*SoldGold {
	return sg.areaToGoodsList[area]
}

type SoldGoldHandjoyGameData struct {
	SoldGoldGameData
}

func newSoldGoldHandjoyGameData() *SoldGoldHandjoyGameData {
	sg := &SoldGoldHandjoyGameData{}
	sg.i = sg
	return sg
}

func (sg *SoldGoldHandjoyGameData) name() string {
	return consts.SoldGoldHandjoy
}

func GetSoldGoldGameData() ISoldGoldGameData {
	if config.GetConfig().IsMultiLan {
		return GetGameData(consts.SoldGoldHandjoy).(*SoldGoldHandjoyGameData)
	} else {
		return GetGameData(consts.SoldGold).(*SoldGoldGameData)
	}
}
