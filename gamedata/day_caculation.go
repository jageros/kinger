package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RebornDayCacul struct {
	ID       int     `json:"__id__"`
	Prestige float32 `json:"fame"`
}

type RebornDayCaculGameData struct {
	baseGameData
	Prestige float32
}

func newRebornDayCaculGameData() *RebornDayCaculGameData {
	gd := &RebornDayCaculGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornDayCaculGameData) name() string {
	return consts.RebornDayCacul
}

func (gd *RebornDayCaculGameData) init(d []byte) error {
	var l []*RebornDayCacul
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	for _, d := range l {
		gd.Prestige = d.Prestige
	}

	return nil
}
