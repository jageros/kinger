package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
	"bytes"
	"errors"
	"sort"
)

type SoldGoldGift struct {
	ID string `json:"__id__"`
	TreasureID string `json:"treasureId"`
	JadePrice int `json:"jadePrice"`
	Order int `json:"order"`
	Areas [][]int `json:"areas"`

	areaLimit *AreaLimitConfig
}

func (g *SoldGoldGift) init() {
	g.areaLimit = newAreaLimitConfig(g.Areas)
}

type SoldGoldGiftGameData struct {
	baseGameData
	areaVersion int
	rawData []byte
	areaToGiftList map[int][]*SoldGoldGift
}

func newSoldGoldGiftGameData() *SoldGoldGiftGameData {
	sg := &SoldGoldGiftGameData{}
	sg.i = sg
	return sg
}

func (sg *SoldGoldGiftGameData) name() string {
	return consts.SoldGoldGift
}

func (sg *SoldGoldGiftGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if sg.areaVersion == areaVersion && bytes.Equal(sg.rawData, d) {
		return errors.New("no update")
	}

	var l []*SoldGoldGift
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	sg.areaVersion = areaVersion
	sg.rawData = d
	sg.areaToGiftList = map[int][]*SoldGoldGift{}
	for _, g := range l {
		g.init()
		g.areaLimit.forEachArea(func(area int) {
			giftList := sg.areaToGiftList[area]
			sg.areaToGiftList[area] = append(giftList, g)
		})
	}

	for area, gifts := range sg.areaToGiftList {
		sort.Slice(gifts, func(i, j int) bool {
			return gifts[i].Order <= gifts[j].Order
		})
		sg.areaToGiftList[area] = gifts
	}

	return nil
}

func (sg *SoldGoldGiftGameData) GetGift(area, idx int) *SoldGoldGift {
	if gifts, ok := sg.areaToGiftList[area]; ok {
		for i, g := range gifts {
			if i == idx {
				return g
			}
		}
	}
	return nil
}
