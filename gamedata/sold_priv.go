package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type RebornSoldPriv struct {
	ID int `json:"__id__"`
	PrivID int `json:"privId"`
	HonorPrice int `json:"honorPrice"`
	PrestigePrice int `json:"famePrice"`
	Cnt int `json:"cnt"`
}

type RebornSoldPrivGameData struct {
	baseGameData
	ID2Priv map[int]*RebornSoldPriv
}

func newRebornSoldPrivGameData() *RebornSoldPrivGameData {
	gd := &RebornSoldPrivGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornSoldPrivGameData) name() string {
	return consts.RebornSoldPriv
}

func (gd *RebornSoldPrivGameData) init(d []byte) error {
	var l []*RebornSoldPriv
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.ID2Priv = map[int]*RebornSoldPriv{}
	for _, p := range l {
		gd.ID2Priv[p.ID] = p
	}

	return nil
}
