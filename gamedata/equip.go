package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type Equip struct {
	ID string `json:"__id__"`
	Skills []int32 `json:"skill"`
	Name int `json:"name"`
	Version int `json:"version"`
}

func (e *Equip) GetName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(e.Name)
}

type EquipGameData struct {
	baseGameData
	Equips []*Equip
	ID2Equip map[string]*Equip
	AllEquipIDs []interface{}
}

func newEquipGameData() *EquipGameData {
	gd := &EquipGameData{}
	gd.i = gd
	return gd
}

func (gd *EquipGameData) name() string {
	return consts.Equip
}

func (gd *EquipGameData) init(d []byte) error {
	var l []*Equip
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Equips = l
	gd.ID2Equip = map[string]*Equip{}
	gd.AllEquipIDs = []interface{} {}
	for _, e := range l {
		gd.AllEquipIDs = append(gd.AllEquipIDs, e.ID)
		gd.ID2Equip[e.ID] = e
	}

	return nil
}
