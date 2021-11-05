package gamedata

import (
	"encoding/json"
	"errors"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type TutorialBattle struct {
	ID             int      `json:"__id__"`
	EnemyHand      []uint32 `json:"enemy_hand"`
	Country        int      `json:"country"`
	TreasureReward []string `json:"treasureReward"`
	Offensive      int      `json:"offensive"`
	Name           int      `json:"name"`
	OwnSide        [][]int  `json:"own_side"`
	EnemySide      [][]int  `json:"enemy_side"`
	Head string `json:"head"`
	HeadFrame string `json:"headFrame"`
	NormalAI int `json:"normalAl"`
}

type TutorialGameData struct {
	baseGameData
	BattlesOfCamp map[int]map[int]*TutorialBattle
}

func newTutorialGameData() *TutorialGameData {
	t := &TutorialGameData{}
	t.i = t
	return t
}

func (t *TutorialGameData) name() string {
	return consts.Tutorial
}

func (t *TutorialGameData) init(d []byte) error {
	var l []*TutorialBattle

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("tutorial battles = %s", l)
	if len(l) < 12 {
		return errors.New("fuck tutorial battles")
	}

	t.BattlesOfCamp = map[int]map[int]*TutorialBattle{}

	idx := 0
	for _, camp := range []int{consts.Wei, consts.Shu, consts.Wu} {
		battles, ok := t.BattlesOfCamp[camp]
		if !ok {
			battles = map[int]*TutorialBattle{}
			t.BattlesOfCamp[camp] = battles
		}

		for id := 1; id <= consts.MaxGuidePro; id++ {
			battles[id] = l[idx]
			idx++
		}
	}

	/*
		for _, t2 := range l {
			battles := t.BattlesOfCamp[t2.Country]

			if battles == nil {
				battles = map[int]*TutorialBattle{}
				t.BattlesOfCamp[t2.Country] = battles
			}

			id, ok := campBattleIDs[t2.Country]
			if !ok {
				id = 1
			} else {
				id++
			}
			campBattleIDs[t2.Country] = id

			battles[id] = t2
		}
	*/

	return nil
}
