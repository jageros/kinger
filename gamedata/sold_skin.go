package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RebornSoldSkin struct {
	ID         int    `json:"__id__"`
	SkinID     string `json:"skinId"`
	HonorPrice int    `json:"honorPrice"`
	Cnt        int    `json:"cnt"`
}

type RebornSoldSkinGameData struct {
	baseGameData
	ID2Skin map[int]*RebornSoldSkin
}

func newRebornSoldSkinGameData() *RebornSoldSkinGameData {
	gd := &RebornSoldSkinGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornSoldSkinGameData) name() string {
	return consts.RebornSoldSkin
}

func (gd *RebornSoldSkinGameData) init(d []byte) error {
	var l []*RebornSoldSkin
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.ID2Skin = map[int]*RebornSoldSkin{}
	for _, sk := range l {
		gd.ID2Skin[sk.ID] = sk
	}

	return nil
}
