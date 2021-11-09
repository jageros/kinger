package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RebornCnt struct {
	Cnt     int `json:"__id__"`
	SkyBook int `json:"book"`
}

type RebornCntGameData struct {
	baseGameData
	Cnt2BookAmount map[int]int
}

func newRebornCntGameData() *RebornCntGameData {
	gd := &RebornCntGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornCntGameData) name() string {
	return consts.RebornCnt
}

func (gd *RebornCntGameData) init(d []byte) error {
	var l []*RebornCnt
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Cnt2BookAmount = map[int]int{}
	for _, t := range l {
		gd.Cnt2BookAmount[t.Cnt] = t.SkyBook
	}

	return nil
}
