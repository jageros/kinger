package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type TreasureShare struct {
	ID         int        `json:"__id__"`
	BeginTime  string     `json:"beginTime"`
	EndTime    string     `json:"endTime"`
	Reward     string     `json:"reward"`
	TreasureId [][]string `json:"treasureId"`
}

type TreasureShareGameData struct {
	baseGameData
	TreasureShares   []*TreasureShare
	Id2TreasureShare map[int]*TreasureShare
}

func newTreasureShareGameData() *TreasureShareGameData {
	t := &TreasureShareGameData{}
	t.i = t
	return t
}

func (t *TreasureShareGameData) name() string {
	return consts.TreasureShare
}

func (t *TreasureShareGameData) init(d []byte) error {
	var l []*TreasureShare

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	t.TreasureShares = l
	t.Id2TreasureShare = make(map[int]*TreasureShare)
	for _, ts := range l {
		t.Id2TreasureShare[ts.ID] = ts
	}

	return nil
}
