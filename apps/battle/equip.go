package main

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
)

type equipment struct {
	data   *gamedata.Equip
	skills []*skill
}

func newEquipment(equipID string, owner iCaster, situation *battleSituation) *equipment {
	data := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData).ID2Equip[equipID]
	if data == nil {
		return nil
	}

	eq := &equipment{data: data}
	for _, skillID := range data.Skills {
		s := newSkill(skillID, situation, owner)
		if s != nil {
			s.isEquip = true
			eq.skills = append(eq.skills, s)
		}
	}
	return eq
}

func (eq *equipment) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetStr("equipID", eq.data.ID)
	skillsAttr := attribute.NewListAttr()
	for _, sk := range eq.skills {
		skillsAttr.AppendMapAttr(sk.packAttr())
	}
	attr.SetListAttr("skills", skillsAttr)
	return attr
}

func (eq *equipment) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	equipID := attr.GetStr("equipID")
	eq.data = gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData).ID2Equip[equipID]
	skillsAttr := attr.GetListAttr("skills")
	skillsAttr.ForEachIndex(func(index int) bool {
		sk := &skill{}
		sk.restoredFromAttr(skillsAttr.GetMapAttr(index), situation)
		eq.skills = append(eq.skills, sk)
		return true
	})
}

func (eq *equipment) copy(situation *battleSituation) *equipment {
	cpy := &equipment{data: eq.data}
	cpy.skills = make([]*skill, len(eq.skills))
	for i, sk := range eq.skills {
		cpy.skills[i] = sk.copy(situation)
	}
	return cpy
}

func (eq *equipment) enterBattle(owner iCaster) []*clientAction {
	var acts []*clientAction
	for _, sk := range eq.skills {
		acts = append(acts, sk.onAdd(owner.getSit())...)
	}
	return acts
}

func (eq *equipment) leaveBattle() []*clientAction {
	var acts []*clientAction
	for _, sk := range eq.skills {
		acts = append(acts, sk.onInvalid()...)
	}
	return acts
}

func (eq *equipment) hasSkillByID(skillID int32) bool {
	for _, sk := range eq.skills {
		if sk.getID() == skillID && sk.isEffective() {
			return true
		}
	}
	return false
}

func (eq *equipment) hasSkill(sk *skill) bool {
	for _, sk2 := range eq.skills {
		if sk2 == sk {
			return true
		}
	}
	return false
}

func (eq *equipment) hasNoSkill() bool {
	for _, sk := range eq.skills {
		data := sk.getData()
		if data.Name <= 0 && data.SkillGroup == "" {
			continue
		}
		if sk.isEffective() {
			return false
		}
	}
	return true
}

func (eq *equipment) setPlayCardIdx(idx int) {
	for _, sk := range eq.skills {
		sk.setPlayCardIdx(idx)
	}
}

func (eq *equipment) packMsg() *pb.BattleEquip {
	e := &pb.BattleEquip{EquipID: eq.data.ID}
L:
	for _, sk := range eq.skills {
		for _, skID := range e.Skills {
			if sk.getID() == skID {
				continue L
			}
		}

		if sk.isEffective() {
			e.Skills = append(e.Skills, sk.getID())
		}
	}
	return e
}
