package main

import (
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"strconv"
)

type iCaster interface {
	iTarget
	enterBattle() []*clientAction
	hasSkillByID(skillID int32) bool
	hasSkill(sk *skill) bool
	hasNoSkill() bool
	hasSkillByIDIgnoreEquip(skillID int32) bool
	hasSkillIgnoreEquip(sk *skill) bool
	hasNoSkillIgnoreEquip() bool
	addSkill(skillID int32, boutTimeout int) []*clientAction
	delSkill(sk *skill) []*clientAction
	getAllIgnoreLostCanTriggerSkills() []*skill
	canTriggerSkill(sk *skill) bool
	triggerSkill(skillID int32)
	lostSkill(skillID int32, untilBout int) ([]*clientAction, bool)
	boutEnd() []*clientAction
}

type baseCaster struct {
	baseTarget
	i                     iCaster
	skills                []*skill
	boutSkillTriggerTimes map[int32]int
	equip                 *equipment
}

func (bc *baseCaster) copyCaster(cpy *baseCaster) {
	cpy.boutSkillTriggerTimes = nil
	cpy.skills = make([]*skill, len(bc.skills))
	for i, sk := range bc.skills {
		cpy.skills[i] = sk.copy(cpy.situation)
	}

	if bc.equip != nil {
		cpy.equip = bc.equip.copy(cpy.situation)
	}
}

func (bc *baseCaster) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetMapAttr("baseTarget", bc.baseTarget.packAttr())
	if bc.boutSkillTriggerTimes != nil {
		boutSkillTriggerTimesAttr := attribute.NewMapAttr()
		for skillID, times := range bc.boutSkillTriggerTimes {
			boutSkillTriggerTimesAttr.SetInt(strconv.Itoa(int(skillID)), times)
		}
		attr.SetMapAttr("boutSkillTriggerTimes", boutSkillTriggerTimesAttr)
	}

	skillsAttr := attribute.NewListAttr()
	attr.SetListAttr("skills", skillsAttr)
	for _, sk := range bc.skills {
		skillsAttr.AppendMapAttr(sk.packAttr())
	}

	if bc.equip != nil {
		attr.SetMapAttr("equip", bc.equip.packAttr())
	}

	return attr
}

func (bc *baseCaster) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	(&bc.baseTarget).restoredFromAttr(attr.GetMapAttr("baseTarget"), situation)
	boutSkillTriggerTimesAttr := attr.GetMapAttr("boutSkillTriggerTimes")
	if boutSkillTriggerTimesAttr != nil {
		boutSkillTriggerTimesAttr.ForEachKey(func(key string) {
			skillID, _ := strconv.Atoi(key)
			bc.boutSkillTriggerTimes = map[int32]int{}
			bc.boutSkillTriggerTimes[int32(skillID)] = boutSkillTriggerTimesAttr.GetInt(key)
		})
	}

	skillsAttr := attr.GetListAttr("skills")
	skillsAttr.ForEachIndex(func(index int) bool {
		sk := &skill{}
		sk.restoredFromAttr(skillsAttr.GetMapAttr(index), situation)
		bc.skills = append(bc.skills, sk)
		return true
	})

	equipAttr := attr.GetMapAttr("equip")
	if equipAttr != nil {
		bc.equip = &equipment{}
		bc.equip.restoredFromAttr(equipAttr, situation)
	}
}

func (bc *baseCaster) enterBattle() []*clientAction {
	var acts []*clientAction
	for _, sk := range bc.skills {
		acts = append(acts, sk.onAdd(bc.i.getSit())...)
	}
	if bc.equip != nil {
		acts = append(acts, bc.equip.enterBattle(bc.i)...)
	}
	return acts
}

func (bc *baseCaster) boutBegin() []*clientAction {
	bc.boutSkillTriggerTimes = nil
	return bc.baseTarget.boutBegin()
}

func (bc *baseCaster) hasSkillByID(skillID int32) bool {
	has := bc.hasSkillByIDIgnoreEquip(skillID)
	if has {
		return true
	}
	return bc.equip != nil && bc.equip.hasSkillByID(skillID)
}

func (bc *baseCaster) hasSkill(sk *skill) bool {
	has := bc.hasSkillIgnoreEquip(sk)
	if has {
		return true
	}
	return bc.equip != nil && bc.equip.hasSkill(sk)
}

func (bc *baseCaster) hasSkillByIDIgnoreEquip(skillID int32) bool {
	for _, sk := range bc.skills {
		if sk.getID() == skillID && sk.isEffective() {
			return true
		}
	}
	return false
}

func (bc *baseCaster) hasSkillIgnoreEquip(sk *skill) bool {
	for _, sk2 := range bc.skills {
		if sk2 == sk {
			return true
		}
	}
	return false
}

func (bc *baseCaster) hasNoSkill() bool {
	has := bc.hasNoSkillIgnoreEquip()
	if !has {
		return false
	}
	return bc.equip == nil || bc.equip.hasNoSkill()
}

func (bc *baseCaster) hasNoSkillIgnoreEquip() bool {
	for _, sk := range bc.skills {
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

func (bc *baseCaster) addSkill(skillID int32, boutTimeout int) []*clientAction {
	return bc.casterAddSkill(skillID, boutTimeout, false)
}

func (bc *baseCaster) casterAddSkill(skillID int32, boutTimeout int, isAddInHand bool) []*clientAction {
	sk := newSkill(skillID, bc.situation, bc.i)
	if sk == nil {
		return []*clientAction{}
	}
	sk.isAddInHand = isAddInHand
	if boutTimeout == 0 {
		boutTimeout = sk.behavior.round
	}

	for _, sk2 := range bc.skills {
		if sk2.getID() == skillID && sk2.isEffective() {
			remainBout := sk2.getRemainBout()
			sk2.setPlayCardIdx(bc.situation.getPlayCardQueueIdx())
			if remainBout >= 0 {
				if boutTimeout < 0 || remainBout < boutTimeout {
					sk2.setBoutTimeout(boutTimeout)
				}
			}
			return []*clientAction{}
		}
	}

	sk.setBoutTimeout(boutTimeout)
	bc.skills = append(bc.skills, sk)
	acts := sk.onAdd(bc.getSit())
	sk.setPlayCardIdx(bc.situation.getPlayCardQueueIdx())
	acts = append(acts, &clientAction{
		actID:  pb.ClientAction_AddSkill,
		actMsg: &pb.ModifySkillAct{CardObjID: int32(bc.i.getObjID()), SkillID: skillID},
	})
	return acts
}

func (bc *baseCaster) delSkill(sk *skill) []*clientAction {
	index := -1
	isEffective := false
	if sk.isEquip && bc.equip != nil {
		for i, sk2 := range bc.equip.skills {
			if sk == sk2 {
				index = i
			} else if sk.getID() == sk2.getID() && sk2.isEffective() {
				isEffective = true
			}
		}
	} else if !sk.isEquip {
		for i, sk2 := range bc.skills {
			if sk == sk2 {
				index = i
			} else if sk.getID() == sk2.getID() && sk2.isEffective() {
				isEffective = true
			}
		}
	}

	var acts []*clientAction
	if index >= 0 {
		if sk.isEquip {
			bc.equip.skills = append(bc.equip.skills[:index], bc.equip.skills[index+1:]...)
		} else {
			bc.skills = append(bc.skills[:index], bc.skills[index+1:]...)
		}
		acts = []*clientAction{&clientAction{
			actID:  pb.ClientAction_DelSkill,
			actMsg: &pb.ModifySkillAct{CardObjID: int32(bc.i.getObjID()), SkillID: sk.getID(), IsEquip: sk.isEquip},
		}}
	}

	if !isEffective {
		acts = append(acts, sk.onInvalid()...)
	}
	return acts
}

func (bc *baseCaster) getAllEffectiveSkills() []*skill {
	var sks []*skill
L:
	for _, sk := range bc.skills {
		for _, sk2 := range sks {
			if sk.getID() == sk2.getID() {
				continue L
			}
		}

		if sk.isEffective() {
			sks = append(sks, sk)
		}
	}
	return sks
}

func (bc *baseCaster) triggerSkill(skillID int32) {
	if bc.boutSkillTriggerTimes == nil {
		bc.boutSkillTriggerTimes = map[int32]int{}
	}
	if n, ok := bc.boutSkillTriggerTimes[skillID]; ok {
		bc.boutSkillTriggerTimes[skillID] = n + 1
	} else {
		bc.boutSkillTriggerTimes[skillID] = 1
	}
}

func (bc *baseCaster) canTriggerSkill(sk *skill) bool {
	times := sk.getData().Times
	if times > 0 && bc.boutSkillTriggerTimes != nil {
		if n, ok := bc.boutSkillTriggerTimes[sk.getID()]; ok {
			if n >= times {
				return false
			}
		}
	}
	return true
}

func (bc *baseCaster) getAllIgnoreLostCanTriggerSkills() []*skill {
	var sks []*skill
	var checkTrigger = func(sk *skill) {
		for _, sk2 := range sks {
			if sk2.getID() == sk.getID() && sk2.isEffective() {
				return
			}
		}

		if sk.isEffectiveIgnoreLost() && bc.canTriggerSkill(sk) {
			sks = append(sks, sk)
		}
	}

	for _, sk := range bc.skills {
		checkTrigger(sk)
	}

	if bc.equip != nil {
		for _, sk := range bc.equip.skills {
			checkTrigger(sk)
		}
	}

	return sks
}

func (bc *baseCaster) lostSkill(skillID int32, untilBout int) ([]*clientAction, bool) {
	var actions []*clientAction
	isDel := false
	for _, sk := range bc.skills {
		isTargetSkill := sk.getID() == skillID
		if skillID < 0 || isTargetSkill {
			acts, ok := sk.lost(untilBout, isTargetSkill)
			actions = append(actions, acts...)
			if ok {
				if sk.isShownSkill() {
					isDel = true
				}
				actions = append(actions, &clientAction{
					actID:  pb.ClientAction_DelSkill,
					actMsg: &pb.ModifySkillAct{CardObjID: int32(bc.i.getObjID()), SkillID: sk.getID()},
				})
			}
		}
	}
	return actions, isDel
}

func (bc *baseCaster) setPlayCardIdx(idx int) {
	for _, sk := range bc.skills {
		sk.setPlayCardIdx(idx)
	}
	if bc.equip != nil {
		bc.equip.setPlayCardIdx(idx)
	}
}

func (bc *baseCaster) boutEnd() []*clientAction {
	var actions []*clientAction
	size1 := len(bc.skills)
	var size2 int
	if bc.equip != nil {
		size2 = len(bc.equip.skills)
	}
	sks := make([]*skill, size1+size2)
	copy(sks, bc.skills)
	if size2 > 0 {
		copy(sks[size1:], bc.equip.skills)
	}

	for _, sk := range sks {
		actions = append(actions, sk.boutEnd()...)
	}
	return actions
}

func (bc *baseCaster) forEachSkill(callback func(sk *skill)) {
	for _, sk := range bc.skills {
		callback(sk)
	}
	if bc.equip != nil {
		for _, sk := range bc.equip.skills {
			callback(sk)
		}
	}
}

func (bc *baseCaster) setObjID(objID int) {
	bc.objID = objID
	for _, sk := range bc.skills {
		sk.ownerObjID = objID
	}
	if bc.equip != nil {
		for _, sk := range bc.equip.skills {
			sk.ownerObjID = objID
		}
	}
}
