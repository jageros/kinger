package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type RebornTreausre struct {
	Team     int    `json:"__id__"`
	Treasure string `json:"treasure"`
	Gold     int    `json:"gold"`
}

type RebornTreausreGameData struct {
	baseGameData
	Team2Treausre map[int]string
	Team2Gold     map[int]int
}

func newRebornTreausreGameData() *RebornTreausreGameData {
	gd := &RebornTreausreGameData{}
	gd.i = gd
	return gd
}

func (gd *RebornTreausreGameData) name() string {
	return consts.RebornTreausre
}

func (gd *RebornTreausreGameData) init(d []byte) error {
	var l []*RebornTreausre
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Team2Treausre = map[int]string{}
	gd.Team2Gold = map[int]int{}
	for _, t := range l {
		gd.Team2Treausre[t.Team] = t.Treasure
		gd.Team2Gold[t.Team] = t.Gold
	}

	return nil
}
