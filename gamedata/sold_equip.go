package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type RebornSoldEquip struct {
	ID int `json:"__id__"`
	EquipID string `json:"equipId"`
	Price int `json:"famePrice"`
}

type RebornSoldEquipGameData struct {
	baseGameData
	ID2Equip map[int]*RebornSoldEquip
	EquipID2Goods map[string]*RebornSoldEquip
}

func newRebornSoldEquipGameData() *RebornSoldEquipGameData {
	gd := &RebornSoldEquipGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornSoldEquipGameData) name() string {
	return consts.RebornSoldEquip
}

func (gd *RebornSoldEquipGameData) init(d []byte) error {
	var l []*RebornSoldEquip
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.ID2Equip = map[int]*RebornSoldEquip{}
	gd.EquipID2Goods = map[string]*RebornSoldEquip{}
	for _, eq := range l {
		gd.ID2Equip[eq.ID] = eq
		gd.EquipID2Goods[eq.EquipID] = eq
	}

	return nil
}

