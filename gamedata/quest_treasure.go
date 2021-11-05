package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type MissionTreasure struct {
	ID int `json:"__id__"`
	RareLoop []int `json:"rareLoop"`
}

type MissionTreasureGameData struct {
	baseGameData
	ID2MissionTreasures map[int]*MissionTreasure
	MissionTreasures []*MissionTreasure
}

func newMissionTreasureGameData() *MissionTreasureGameData {
	mgd := &MissionTreasureGameData{}
	mgd.i = mgd
	return mgd
}

func (mgd *MissionTreasureGameData) name() string {
	return consts.MissionTreasure
}

func (mgd *MissionTreasureGameData) init(d []byte) error {
	var l []*MissionTreasure
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	mgd.MissionTreasures = l
	mgd.ID2MissionTreasures = map[int]*MissionTreasure{}
	for _, mt := range l {
		mgd.ID2MissionTreasures[mt.ID] = mt
	}

	return nil
}
