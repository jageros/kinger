package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type RebornGoldCacul struct {
	Gold int `json:"__id__"`
	Honor float32 `json:"honor"`
	Prestige float32 `json:"fame"`
}

type RebornGoldCaculGameData struct {
	baseGameData
	Cacul *RebornGoldCacul
}

func newRebornGoldCaculGameData() *RebornGoldCaculGameData {
	gd := &RebornGoldCaculGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornGoldCaculGameData) name() string {
	return consts.RebornGoldCacul
}

func (gd *RebornGoldCaculGameData) init(d []byte) error {
	var l []*RebornGoldCacul
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Cacul = l[0]
	return nil
}
