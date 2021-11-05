package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type TreasureDailyFake struct {
	Cnt         int `json:"cnt"`
	TreaureRare int `json:"treaureRare"`
}

type TreasureDailyFakeGameData struct {
	baseGameData
	Cnt2TreaureRare map[int]int
}

func newTreasureDailyFakeGameData() *TreasureDailyFakeGameData {
	t := &TreasureDailyFakeGameData{}
	t.i = t
	return t
}

func (t *TreasureDailyFakeGameData) name() string {
	return consts.TreasureDailyFake
}

func (t *TreasureDailyFakeGameData) init(d []byte) error {
	var l []*TreasureDailyFake

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("treasures daily fake = %s", l)
	t.Cnt2TreaureRare = map[int]int{}

	for _, t2 := range l {
		t.Cnt2TreaureRare[t2.Cnt] = t2.TreaureRare
	}

	return nil
}
