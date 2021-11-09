package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RebornSoldCard struct {
	ID         int    `json:"__id__"`
	CardID     uint32 `json:"cardId"`
	HonorPrice int    `json:"honorPrice"`
	Cnt        int    `json:"cnt"`
}

type RebornSoldCardGameData struct {
	baseGameData
	ID2Card     map[int]*RebornSoldCard
	CardID2Card map[uint32]*RebornSoldCard
}

func newRebornSoldCardGameData() *RebornSoldCardGameData {
	gd := &RebornSoldCardGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornSoldCardGameData) name() string {
	return consts.RebornSoldCard
}

func (gd *RebornSoldCardGameData) init(d []byte) error {
	var l []*RebornSoldCard
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.ID2Card = map[int]*RebornSoldCard{}
	gd.CardID2Card = map[uint32]*RebornSoldCard{}
	for _, c := range l {
		gd.ID2Card[c.ID] = c
		gd.CardID2Card[c.CardID] = c
	}

	return nil
}
