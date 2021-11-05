package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type TreasureRewardFake struct {
	Cnt         int `json:"cnt"`
	TreaureRare int `json:""treaureRare"`
}

type TreasureRewardFakeGameData struct {
	baseGameData
	Cnt2TreaureRare map[int]int
}

func newTreasureRewardFakeGameData() *TreasureRewardFakeGameData {
	t := &TreasureRewardFakeGameData{}
	t.i = t
	return t
}

func (t *TreasureRewardFakeGameData) name() string {
	return consts.TreasureRewardFake
}

func (t *TreasureRewardFakeGameData) init(d []byte) error {
	var l []*TreasureRewardFake

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("treasures reward fake = %s", l)
	t.Cnt2TreaureRare = map[int]int{}

	for _, t2 := range l {
		t.Cnt2TreaureRare[t2.Cnt] = t2.TreaureRare
	}

	return nil
}
