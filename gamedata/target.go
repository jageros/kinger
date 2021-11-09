package gamedata

import (
	"encoding/json"
	"fmt"
	"kinger/common/consts"
)

type Target struct {
	ID               int      `json:"__id__"`
	Relative         int      `json:"relative"`
	Types            []int    `json:"type"`
	Side             int      `json:"targetCamp"`
	TargetBat        []int    `json:"targetBat"`
	Turn             int      `json:"targeTurn"`
	Poses            []int    `json:"targetPos"`
	Camp             []int    `json:"targetNation"`
	TargetSkill      []int32  `json:"targetSkill"`
	TargetSkillFog   []int32  `json:"targetSkill_fog"`
	InitSide         []int    `json:"targetIniCamp"`
	Random           int      `json:"random"`
	TargetAtt        []int    `json:"targetAtt"`
	CardType         int      `json:"cardType"`
	PreTurnSide      int      `json:"preTurnCamp"`
	Type2            []int    `json:"type2"`
	Sequential       [][]int  `json:"sequential"`
	NoTargetSkill    []int32  `json:"noTargetSkill"`
	NoTargetSkillFog []int32  `json:"notargetSkill_fog"`
	TargetCard       []uint32 `json:"targetCard"`
	NotargetCard     []uint32 `json:"notargetCard"`
	TargetClean      int      `json:"targeRoot"`
	Condition        []string `json:"condition"`
	TargetSummon     int      `json:"targetSummon"`
	Surrender        int      `json:"surrender"`
	TargetEquip      int      `json:"targetItem"`
}

func (t *Target) String() string {
	return fmt.Sprintf("id=%d, type=%s, side=%d, TargetBat=%s, Turn=%d, pos=%s, camp=%d",
		t.ID, t.Types, t.Side, t.TargetBat, t.Turn, t.Poses, t.Camp)
}

type TargetGameData struct {
	baseGameData
	allTarget map[int]*Target
}

func newTargetGameData() *TargetGameData {
	t := &TargetGameData{}
	t.i = t
	return t
}

func (tg *TargetGameData) name() string {
	return consts.Target
}

func (tg *TargetGameData) init(data []byte) error {
	var _list []*Target
	err := json.Unmarshal(data, &_list)
	if err != nil {
		return err
	}

	tg.allTarget = make(map[int]*Target)

	for _, target := range _list {
		tg.allTarget[target.ID] = target
	}

	//glog.Infof("TargetGameData = %s", tg.allTarget)

	return nil
}

func (tg *TargetGameData) GetTargetDataByID(targetID int) *Target {
	if t, ok := tg.allTarget[targetID]; ok {
		return t
	} else {
		return nil
	}
}

func (tg *TargetGameData) GetAllTarget() map[int]*Target {
	return tg.allTarget
}
