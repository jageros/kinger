package gamedata

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"kinger/common/consts"
	"strconv"
)

type Skill struct {
	ID             int32      `json:"__id__"`
	Round          int        `json:"round"`
	TriggerObj     int        `json:"triggerObj"`
	TriggerOpp     int        `json:"triggerOpp"`
	Variable       []int      `json:"variable"`
	Condition      []string   `json:"condition"`
	TargetAct      []int      `json:"targetAct"`
	ActionValue    [][]string `json:"actionValue"`
	SkillChange    [][]int    `json:"skillChange"`
	PointChange    [][]int    `json:"pointChange"`
	TurnOver       [][]int    `json:"turnOver"`
	Attack         [][]int    `json:"attack"`
	ActionOth      [][]string `json:"actionOth"`
	Effect         [][]int    `json:"effect"`
	MovieEffect    [][]string `json:"movie_effect"`
	StatusEffect   [][]string `json:"status_effect"`
	Times          int        `json:"times"`
	Move           [][]string `json:"move"`
	Priority       int        `json:"priority"`
	Type           [][]int    `json:"type"`
	SkillCom       []int32    `json:"skillCom"`
	ActRate        int        `json:"actRate"`
	DrawCard       [][]string `json:"draw"`
	Discard        [][]string `json:"discard"`
	Summon         [][]int    `json:"summon"`
	Destroy        [][]int    `json:"destroy"`
	Return         int        `json:"return"`
	TriggerTimes   int        `json:"triggerTimes"`
	SkillGroup     string     `json:"skillGroup"`
	Name           int        `json:"name"`
	Copy           [][]int    `json:"copy"`
	GoldRob        [][]int    `json:"goldRob"`
	ChangeBoutTime int        `json:"changeBoutTime"`
	ValueEffect    [][]string `json:"Value_effect"`
	SwitchCard     [][]string `json:"switch_card"`
	RemoveEquip    []int      `json:"remove_item"`

	// [][targetid, effectid, playtype]
	statusEffect [][]interface{}
}

func (s *Skill) init() error {

	for _, einfo := range s.StatusEffect {
		// 状态特效
		var movieID string
		var targetIDs []int
		var playType int
		var err1 error
		var err2 error
		switch len(einfo) {
		case 2:
			movieID = einfo[0]
			targetIDs = s.TargetAct
			playType, err1 = strconv.Atoi(einfo[1])
		case 3:
			movieID = einfo[1]
			playType, err1 = strconv.Atoi(einfo[2])
			targetID, err := strconv.Atoi(einfo[0])
			err2 = err
			targetIDs = []int{targetID}
		default:
			return errors.Errorf("int StatusEffect error %d", s.ID)
		}

		if err1 != nil || err2 != nil {
			return errors.Errorf("int StatusEffect error %d, %s, %s", s.ID, err1, err2)
		}
		s.statusEffect = append(s.statusEffect, []interface{}{targetIDs, movieID, playType})
	}

	return nil
}

func (s *Skill) GetStatusEffect() [][]interface{} {
	return s.statusEffect
}

func (s *Skill) String() string {
	return fmt.Sprintf("[Skill gamedata id=%d]", s.ID)
}

// implement data.IGameRes
type SkillGameData struct {
	baseGameData
	allSkill map[int32]*Skill
}

func newSkillGameData() *SkillGameData {
	s := &SkillGameData{}
	s.i = s
	return s
}

func (sg *SkillGameData) name() string {
	return consts.Skill
}

func (sg *SkillGameData) init(data []byte) error {
	var _list []*Skill
	err := json.Unmarshal(data, &_list)
	if err != nil {
		return err
	}

	sg.allSkill = make(map[int32]*Skill)

	for _, s := range _list {
		if err := s.init(); err != nil {
			return err
		}
		sg.allSkill[s.ID] = s
	}

	return nil
}

func (sg *SkillGameData) GetAllSkill() map[int32]*Skill {
	return sg.allSkill
}
