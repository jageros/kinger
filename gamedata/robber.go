package gamedata

import (
	"encoding/json"

	"kinger/common/consts"
	"kinger/gopuppy/common/glog"
)

type Robber struct {
	ID               int      `json:"__id__"`
	RobbeMasterrCntr []int    `json:"robbeMasterrCntr"`
	RobberMaster     []uint32 `json:"robberMaster"`
	RobberGeneral    []uint32 `json:"robberGeneral"`
	Condition        [][]int  `json:"condition"`
	Ignore           [][]int  `json:"ignore"`
	Colddown         int      `json:"colddown"`
	Period           int      `json:"period"`
	BattleScale      int      `json:"battleScale"`
	FieldReward      [][]int  `json:"fieldReward"`
}

type RobberGameData struct {
	baseGameData
	Id2Robber map[int]*Robber
	Robbers   []*Robber
}

func newRobberGameData() *RobberGameData {
	r := &RobberGameData{}
	r.i = r
	return r
}

func (rd *RobberGameData) name() string {
	return consts.Robber
}

func (rd *RobberGameData) init(d []byte) error {
	var l []*Robber
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	glog.Infof("RobberGameData = %s", l)

	rd.Id2Robber = map[int]*Robber{}
	for _, r := range l {
		rd.Id2Robber[r.ID] = r
	}
	rd.Robbers = l

	return nil
}
