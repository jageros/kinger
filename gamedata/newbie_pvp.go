package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type NewbiePvpBattle struct {
	Camp     int        `json:"__id__"`
	Hand     []uint32   `json:"hand"`
	GridCard [][]uint32 `json:"grid_card"`
}

type NewbiePvpGameData struct {
	baseGameData
	Camp2Battle map[int]*NewbiePvpBattle
}

func newNewbiePvpGameData() *NewbiePvpGameData {
	t := &NewbiePvpGameData{}
	t.i = t
	return t
}

func (t *NewbiePvpGameData) name() string {
	return consts.NewbiePvp
}

func (t *NewbiePvpGameData) init(d []byte) error {
	var l []*NewbiePvpBattle

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	t.Camp2Battle = map[int]*NewbiePvpBattle{}
	for _, b := range l {
		t.Camp2Battle[b.Camp] = b
	}

	return nil
}
