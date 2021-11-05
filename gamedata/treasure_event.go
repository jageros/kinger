package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type TreasureEvent struct {
	Rare         int        `json:"__id__"`
	UpRarePrice  int     `json:"upRarePrice"`
	AddCardPrice    int     `json:"addCardPrice"`
	AddCardCnt     int     `json:"addCardCnt"`
	UpTreasure int `json:"upTreasure"`
	Double int `json:"double"`
}

type TreasureEventGameData struct {
	baseGameData
	configs   []*TreasureEvent
}

func newTreasureEventGameData() *TreasureEventGameData {
	t := &TreasureEventGameData{}
	t.i = t
	return t
}

func (t *TreasureEventGameData) name() string {
	return consts.TreasureEvent
}

func (t *TreasureEventGameData) init(d []byte) error {
	var l []*TreasureEvent

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	t.configs = l

	return nil
}

func (t *TreasureEventGameData) GetConfigByRare(rare int) *TreasureEvent {
	for _, c := range t.configs {
		if c.Rare == rare {
			return c
		}
	}
	return nil
}
