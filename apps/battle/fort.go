package main

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/attribute"
	"kinger/common/consts"
)

var _ iCaster = &fortCaster{}

type fortCaster struct {
	baseCaster
}

func newFortCaster(objID int, owner common.UUid, skillIDs []int32, situation *battleSituation) *fortCaster {
	fc := &fortCaster{}
	fc.objID = objID
	fc.situation = situation
	fc.gridID = -1
	fc.i = fc
	fc.targetType = stFort

	// 平局技能
	//skillIDs = append(skillIDs, 1010)
	for _, skillID := range skillIDs {
		s := newSkill(skillID, situation, fc)
		if s != nil {
			fc.skills = append(fc.skills, s)
		}
	}

	if owner == situation.fighter1.getUid() {
		fc.setSit(consts.SitOne)
	} else {
		fc.setSit(consts.SitTwo)
	}
	return fc
}

func (fc *fortCaster) copy(situation *battleSituation) iTarget {
	c := *fc
	cpy := &c
	cpy.situation = situation
	cpy.i = cpy
	fc.copyCaster(&cpy.baseCaster)
	return cpy
}

func (fc *fortCaster) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("attrType", attrFort)
	attr.SetMapAttr("baseCaster", fc.baseCaster.packAttr())
	return attr
}

func (fc *fortCaster) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	fc.i = fc
	(&fc.baseCaster).restoredFromAttr(attr.GetMapAttr("baseCaster"), situation)
}

func (fc *fortCaster) setSit(sit int) {
	fc.sit = sit
	fc.initSit = sit
}

func (fc *fortCaster) getCopyTarget() iTarget {
	return fc
}
