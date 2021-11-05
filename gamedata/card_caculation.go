package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type RebornCardCacul struct {
	Star int `json:"__id__"`
	Honor float32 `json:"honor"`
}

type RebornCardCaculGameData struct {
	baseGameData
	Start2Feats map[int]float32
}

func newRebornCardCaculGameData() *RebornCardCaculGameData {
	gd := &RebornCardCaculGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornCardCaculGameData) name() string {
	return consts.RebornCardCacul
}

func (gd *RebornCardCaculGameData) init(d []byte) error {
	var l []*RebornCardCacul
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Start2Feats = map[int]float32{}
	for _, d := range l {
		gd.Start2Feats[d.Star] = d.Honor
	}

	return nil
}
