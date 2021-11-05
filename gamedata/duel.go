package gamedata

import (
	"encoding/json"
	"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Duel struct {
	ID    int     `json:"__id__"`
	Level int     `json:"level"`
	Score int     `json:"score"`
	Win   [][]int `json:"win"`
	Win2  [][]int `json:"win2"`
	Win3  [][]int `json:"win3"`
	Win4  [][]int `json:"win4"`
	Win5  [][]int `json:"win5"`
	Lose  int     `json:"lose"`
}

type DuelGameData struct {
	baseGameData
	duelMap  map[int]*Duel // map[level]*Duel
	duelList []*Duel
}

func newDuelGameData() *DuelGameData {
	d := &DuelGameData{}
	d.i = d
	return d
}

func (dg *DuelGameData) name() string {
	return consts.Duel
}

func (dg *DuelGameData) init(d []byte) error {
	dg.duelMap = make(map[int]*Duel)

	err := json.Unmarshal(d, &dg.duelList)
	if err != nil {
		return err
	}

	for _, d := range dg.duelList {
		dg.duelMap[d.Level] = d
	}

	glog.Infof("duelMap = %s", dg.duelMap)

	return nil
}

func (dg *DuelGameData) GetDuelData(level int) *Duel {
	if d, ok := dg.duelMap[level]; ok {
		return d
	} else {
		return nil
	}
}

func (dg *DuelGameData) GetAllDuel() []*Duel {
	return dg.duelList
}
