package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
	"kinger/gopuppy/common/utils"
	"math/rand"
	"time"
)

type Mission struct {
	ID int `json:"__id__"`
	Type int `json:"type"`
	Process int `json:"process"`
	Camp int `json:"country"`
	CardRare int `json:"cardRare"`
	Jade int `json:"jade"`
	UnlockRank int `json:"unlockRank"`
	Gold int `json:"gold"`
	Bowlder int `json:"bowlder"`
}

type MissionGameData struct {
	baseGameData
	Missions       map[int]*Mission
	Rank2Missions map[int][]*Mission
}

func newMissionGameData() *MissionGameData {
	mgd := &MissionGameData{}
	mgd.i = mgd
	return mgd
}

func (mgd *MissionGameData) name() string {
	return consts.Mission
}

func (mgd *MissionGameData) init(d []byte) error {
	var l []*Mission

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	mgd.Missions = map[int]*Mission{}
	mgd.Rank2Missions = map[int][]*Mission{}
	for _, m := range l {
		mgd.Missions[m.ID] = m
		rl := mgd.Rank2Missions[m.UnlockRank]
		mgd.Rank2Missions[m.UnlockRank] = append(rl, m)
	}

	return nil
}

func (mgd *MissionGameData) GetCanAcceptMission(pvpLevel, fightCampType, fightCamp int, ignore []int) []interface{} {
	var ms []interface{}
	if pvpLevel < 1 {
		pvpLevel = 1
	}
	if fightCamp <= 0 {
		fightCampType = 4
	}

	if pvpLevel < 9 && fightCampType == 2{
		rad := rand.NewSource(time.Now().Unix())
		n := rand.New(rad).Int()
		if n%2 == 0 {
			fightCampType = 1
		}else {
			fightCampType = 3
		}
	}

	for lv := 1; lv <= pvpLevel; lv++ {
		rl := mgd.Rank2Missions[lv]
L:		for _, m := range rl {
			for _, id := range ignore {
				if id == m.ID {
					continue L
				}
			}

			switch fightCampType {
			case 1:  // 跟当前阵容相关
				if m.Camp != fightCamp {
					continue
				}
			case 2:  // 跟其他阵容相关
				if m.Camp == 0 || m.Camp == fightCamp {
					continue
				}
			case 3:  // 跟阵容无关
				if m.Camp > 0 {
					continue
				}
			}

			ms = append(ms, m)
		}
	}

	utils.Shuffle(ms)
	return ms
}
