package main

import (
	"fmt"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"math"
)

var _ iCaster = &fightCard{}

type fightCard struct {
	baseCaster
	gcardID                uint32
	cardID                 uint32
	skin                   string
	gridObjID              int
	up                     int
	initUp                 int
	down                   int
	initDown               int
	left                   int
	initLeft               int
	right                  int
	initRight              int
	upValueRate            float32
	downValueRate          float32
	leftValueRate          float32
	rightValueRate         float32
	adjFValue              float32
	cardValue              float32
	owner                  common.UUid
	controller             common.UUid
	initController         common.UUid
	level                  int
	cardType               int
	isPlayInHand           bool
	isSummon               bool
	isDestroy              bool
	hasTurnOverCauseByOth  bool // 曾经被其他人翻过
	hasTurnOver            bool // 曾经被翻过
	forceAttackSkillAmount int
	theCopy                *fightCard
	publicEnemySkillAmount int
	fogAmount              int
	hideEquip              string // 文丑的宝物，召唤颜良时用
}

func newCardByData(objID int, cardData types.IFightCardData, cardSkin, equipID string, situation *battleSituation) *fightCard {
	c := &fightCard{
		gcardID:        cardData.GetGCardID(),
		cardID:         cardData.GetCardID(),
		skin:           cardSkin,
		up:             cardData.RandomUp(),
		down:           cardData.RandomDown(),
		left:           cardData.RandomLeft(),
		right:          cardData.RandomRight(),
		upValueRate:    cardData.GetUpValueRate(),
		downValueRate:  cardData.GetDownValueRate(),
		leftValueRate:  cardData.GetLeftValueRate(),
		rightValueRate: cardData.GetRightValueRate(),
		adjFValue:      cardData.GetAdjFValue(),
		cardValue:      cardData.GetCardValue(),
		level:          cardData.GetLevel(),
		cardType:       cardData.GetCardType(),
	}
	c.i = c
	c.camp = cardData.GetCamp()
	c.initUp = c.up
	c.initDown = c.down
	c.initLeft = c.left
	c.initRight = c.right
	c.objID = objID
	c.situation = situation
	c.targetType = stHand
	c.gridID = -1

	for _, skillID := range cardData.GetSkillIds() {
		s := newSkill(skillID, situation, c)
		if s != nil {
			c.skills = append(c.skills, s)
		}
	}

	if equipID != "" {
		c.equip = newEquipment(equipID, c, situation)
	}

	return c
}

func newFightCard(objID int, gcardID uint32, cardSkin, equipID string, situation *battleSituation) *fightCard {
	poolGameData := bService.getPoolGameData()
	cardData := poolGameData.GetCardByGid(gcardID)
	if cardData == nil {
		return nil
	}
	return newCardByData(objID, cardData, cardSkin, equipID, situation)
}

func newCardByTemplate(templateCard *fightCard, situation *battleSituation, isCopy bool) *fightCard {
	var equipID string
	if !isCopy && templateCard.equip != nil {
		equipID = templateCard.equip.data.ID
	}
	card := newFightCard(situation.genObjID(), templateCard.gcardID, templateCard.skin, equipID, situation)
	card.up = templateCard.initUp
	card.down = templateCard.initDown
	card.left = templateCard.initLeft
	card.right = templateCard.initRight
	card.initUp = templateCard.initUp
	card.initDown = templateCard.initDown
	card.initLeft = templateCard.initLeft
	card.initRight = templateCard.initRight
	card.owner = templateCard.owner
	return card
}

func (fc *fightCard) String() string {
	return fmt.Sprintf("[cardID=%d, objID=%d]", fc.cardID, fc.objID)
}

func (fc *fightCard) copy(situation *battleSituation) iTarget {
	c := *fc
	cpy := &c
	cpy.situation = situation
	cpy.effects = []*mcMovieEffect{}
	cpy.i = cpy
	fc.copyCaster(&cpy.baseCaster)

	return cpy
}

func (fc *fightCard) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("attrType", attrCard)
	attr.SetMapAttr("baseCaster", fc.baseCaster.packAttr())
	attr.SetUInt32("gcardID", fc.gcardID)
	attr.SetUInt32("cardID", fc.cardID)
	attr.SetInt("up", fc.up)
	attr.SetInt("initUp", fc.initUp)
	attr.SetInt("down", fc.down)
	attr.SetInt("initDown", fc.initDown)
	attr.SetInt("left", fc.left)
	attr.SetInt("initLeft", fc.initLeft)
	attr.SetInt("right", fc.right)
	attr.SetInt("initRight", fc.initRight)
	attr.SetFloat("upValueRate", fc.upValueRate)
	attr.SetFloat("downValueRate", fc.downValueRate)
	attr.SetFloat("leftValueRate", fc.leftValueRate)
	attr.SetFloat("rightValueRate", fc.rightValueRate)
	attr.SetFloat("adjFValue", fc.adjFValue)
	attr.SetFloat("cardValue", fc.cardValue)
	attr.SetUInt64("controller", uint64(fc.controller))
	attr.SetUInt64("initController", uint64(fc.initController))
	attr.SetUInt64("owner", uint64(fc.owner))
	attr.SetInt("gridObjID", fc.gridObjID)
	attr.SetInt("level", fc.level)
	attr.SetInt("cardType", fc.cardType)
	attr.SetBool("isPlayInHand", fc.isPlayInHand)
	attr.SetBool("isSummon", fc.isSummon)
	attr.SetBool("isDestroy", fc.isDestroy)
	attr.SetBool("hasTurnOverCauseByOth", fc.hasTurnOverCauseByOth)
	attr.SetBool("hasTurnOver", fc.hasTurnOver)
	attr.SetInt("forceAttackSkillAmount", fc.forceAttackSkillAmount)
	attr.SetInt("publicEnemySkillAmount", fc.publicEnemySkillAmount)
	attr.SetInt("fogAmount", fc.fogAmount)
	attr.SetStr("hideEquip", fc.hideEquip)
	/*
		if self.diyInfo != nil {
			diyAttr := attribute.NewMapAttr()
			attr.SetMapAttr("diy", diyAttr)
			diyAttr.SetUInt32("cardID", self.diyInfo.CardId)
			diyAttr.SetStr("name", self.diyInfo.Name)
			diyAttr.SetInt32("diySkillId1", self.diyInfo.DiySkillId1)
			diyAttr.SetInt32("diySkillId2", self.diyInfo.DiySkillId2)
			diyAttr.SetStr("weapon", self.diyInfo.Weapon)
		}
	*/

	return attr
}

func (fc *fightCard) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	fc.i = fc
	(&fc.baseCaster).restoredFromAttr(attr.GetMapAttr("baseCaster"), situation)
	fc.gcardID = attr.GetUInt32("gcardID")
	fc.cardID = attr.GetUInt32("cardID")
	fc.up = attr.GetInt("up")
	fc.initUp = attr.GetInt("initUp")
	fc.down = attr.GetInt("down")
	fc.initDown = attr.GetInt("initDown")
	fc.left = attr.GetInt("left")
	fc.initLeft = attr.GetInt("initLeft")
	fc.right = attr.GetInt("right")
	fc.initRight = attr.GetInt("initRight")
	fc.upValueRate = attr.GetFloat32("upValueRate")
	fc.downValueRate = attr.GetFloat32("downValueRate")
	fc.leftValueRate = attr.GetFloat32("leftValueRate")
	fc.rightValueRate = attr.GetFloat32("rightValueRate")
	fc.adjFValue = attr.GetFloat32("adjFValue")
	fc.cardValue = attr.GetFloat32("cardValue")
	fc.controller = common.UUid(attr.GetUInt64("controller"))
	fc.initController = common.UUid(attr.GetUInt64("initController"))
	fc.owner = common.UUid(attr.GetUInt64("owner"))
	fc.gridObjID = attr.GetInt("gridObjID")
	fc.level = attr.GetInt("level")
	fc.cardType = attr.GetInt("cardType")
	fc.isPlayInHand = attr.GetBool("isPlayInHand")
	fc.isSummon = attr.GetBool("isSummon")
	fc.isDestroy = attr.GetBool("isDestroy")
	fc.hasTurnOverCauseByOth = attr.GetBool("hasTurnOverCauseByOth")
	fc.hasTurnOver = attr.GetBool("hasTurnOver")
	fc.forceAttackSkillAmount = attr.GetInt("forceAttackSkillAmount")
	fc.publicEnemySkillAmount = attr.GetInt("publicEnemySkillAmount")
	fc.fogAmount = attr.GetInt("fogAmount")
	fc.hideEquip = attr.GetStr("hideEquip")
}

func (fc *fightCard) getSit() int {
	if fc.sit == 0 {
		return fc.situation.getFighter(fc.controller).getSit()
	} else {
		return fc.sit
	}
}

func (fc *fightCard) getInitSit() int {
	if fc.initSit == 0 {
		return fc.situation.getFighter(fc.initController).getSit()
	} else {
		return fc.initSit
	}
}

func (bt *baseTarget) setInitSit(sit int) {
	bt.initSit = sit
}

func (fc *fightCard) getCopyTarget() iTarget {
	if fc.theCopy != nil {
		return fc.theCopy
	} else {
		return fc
	}
}

func (fc *fightCard) setController(uid common.UUid) {
	fc.controller = uid
	//if fc.owner == 0 {
	//	fc.owner = uid
	//}
	if fc.initController == 0 {
		fc.initController = uid
	}
}

func (fc *fightCard) setInitController(uid common.UUid) {
	fc.initController = uid
}

func (fc *fightCard) setOwner(uid common.UUid) {
	if fc.owner == 0 {
		fc.owner = uid
	}
}

func (fc *fightCard) forceSetOwner(uid common.UUid) {
	fc.owner = uid
}

func (fc *fightCard) getController() *fighter {
	return fc.situation.getFighter(fc.controller)
}

func (fc *fightCard) getControllerUid() common.UUid {
	return fc.controller
}

func (fc *fightCard) getInitControllerUid() common.UUid {
	return fc.initController
}

func (fc *fightCard) setGrid(grid *deskGrid) {
	fc.gridID = grid.getGrid()
	fc.gridObjID = grid.getObjID()
}

func (fc *fightCard) getUp() int {
	return fc.up
}

func (fc *fightCard) getDown() int {
	return fc.down
}

func (fc *fightCard) getLeft() int {
	return fc.left
}

func (fc *fightCard) getRight() int {
	return fc.right
}

func (fc *fightCard) isInBattle() bool {
	return !fc.isDestroy && fc.gridObjID > 0
}

func (fc *fightCard) turnOver(turnner iTarget) []*clientAction {
	turnOverCauseByOth := turnner.getObjID() != fc.getObjID()
	if turnOverCauseByOth {
		fc.hasTurnOverCauseByOth = turnOverCauseByOth
	}
	fc.hasTurnOver = true
	var actions []*clientAction
	fight1 := fc.situation.getFighter1()
	fight2 := fc.situation.getFighter2()
	if fc.controller == fight1.getUid() {
		fc.controller = fight2.getUid()
		fc.sit = fight2.getSit()
	} else {
		fc.controller = fight1.getUid()
		fc.sit = fight1.getSit()
	}

	if fc.gridObjID == 0 {
		// 手牌
		return actions
	}

	if fc.situation.bonusObj != nil {
		fc.situation.bonusObj.onTurnOver(fc)
	}

	sit := fc.getSit()
	for _, sk := range fc.skills {
		actions = append(actions, sk.onOwnerTurnOver(sit)...)
	}
	return actions
}

func (fc *fightCard) packMsg() *pb.Card {
	c := &pb.Card{
		Id:      fc.gcardID,
		ObjId:   int32(fc.objID),
		Up:      int32(fc.up),
		Down:    int32(fc.down),
		Left:    int32(fc.left),
		Right:   int32(fc.right),
		IsInFog: fc.isInFog(),
		//DiyInfo: fc.diyInfo,
		Skin: fc.skin,
	}

	for _, effect := range fc.effects {
		c.Effect = append(c.Effect, effect.packMsg())
	}

	skills := fc.getAllEffectiveSkills()
	for _, sk := range skills {
		c.Skills = append(c.Skills, sk.getID())
	}

	if fc.equip != nil {
		c.Equip = fc.equip.packMsg()
	}
	return c
}

func (fc *fightCard) getLevel() int {
	return fc.level
}

func (fc *fightCard) getGridObj() *deskGrid {
	if fc.gridObjID <= 0 {
		return nil
	}
	return fc.situation.getTargetMgr().getTargetGrid(fc.gridObjID)
}

func (fc *fightCard) modifyValue(value, modifyType, additional int) (minValue, maxValue int) {
	pos := 0 // all
	if modifyType == mvtMinPos {
		// 只有最小的点加
		min := fc.up
		pos = consts.UP
		if fc.down < min {
			min = fc.down
			pos = consts.DOWN
		}
		if fc.left < min {
			min = fc.left
			pos = consts.LEFT
		}
		if fc.right < min {
			min = fc.right
			pos = consts.RIGHT
		}

	} else if modifyType == mvtAllBecome {
		// 变成最大值
		minValue = math.MaxInt32
		maxValue = fc.up

		for _, posValue := range []int{fc.down, fc.left, fc.right} {
			if posValue > maxValue {
				maxValue = posValue
			}
			if posValue < minValue {
				minValue = posValue
			}
		}
		maxValue += additional
	}

	if pos == 0 || pos == consts.UP {
		if modifyType == mvtAllBecome {
			fc.up = maxValue
		} else {
			fc.up += value
		}
		if fc.up < 0 {
			fc.up = 0
		}
	}
	if pos == 0 || pos == consts.DOWN {
		if modifyType == mvtAllBecome {
			fc.down = maxValue
		} else {
			fc.down += value
		}
		if fc.down < 0 {
			fc.down = 0
		}
	}
	if pos == 0 || pos == consts.LEFT {
		if modifyType == mvtAllBecome {
			fc.left = maxValue
		} else {
			fc.left += value
		}
		if fc.left < 0 {
			fc.left = 0
		}
	}
	if pos == 0 || pos == consts.RIGHT {
		if modifyType == mvtAllBecome {
			fc.right = maxValue
		} else {
			fc.right += value
		}
		if fc.right < 0 {
			fc.right = 0
		}
	}

	return
}

func (fc *fightCard) enterBattle() []*clientAction {
	fc.initController = fc.controller
	fc.initSit = fc.situation.getFighter(fc.initController).getSit()
	return (&fc.baseCaster).enterBattle()
}

func (fc *fightCard) isForceAttack() bool {
	return fc.forceAttackSkillAmount > 0
}

func (fc *fightCard) onSkillEffective(sk *skill) {
	if sk.behavior.isForceAttack {
		fc.forceAttackSkillAmount++
	}
	if sk.behavior.isPublicEnemy {
		fc.publicEnemySkillAmount++
	}
}

func (fc *fightCard) onSkillInvalid(sk *skill) {
	if sk.behavior.isForceAttack {
		if fc.forceAttackSkillAmount > 0 {
			fc.forceAttackSkillAmount--
		}
	}
	if sk.behavior.isPublicEnemy {
		if fc.publicEnemySkillAmount > 0 {
			fc.publicEnemySkillAmount--
		}
	}
}

func (fc *fightCard) onDestroy() []*clientAction {
	var acts []*clientAction
	fc.forEachSkill(func(sk *skill) {
		if sk.isEffective() {
			acts = append(acts, sk.onInvalid()...)
		}
	})
	fc.isDestroy = true
	return acts
}

func (fc *fightCard) copyCard(beCopyCard *fightCard) (*fightCard, []*clientAction) {
	initController := fc.initController
	initSit := fc.initSit

	acts := fc.onDestroy()
	c := newCardByTemplate(beCopyCard, fc.situation, true)
	fc.theCopy = c
	c.isPlayInHand = fc.isPlayInHand
	c.controller = beCopyCard.controller
	c.setSit(beCopyCard.getSit())
	c.setObjID(fc.getObjID())
	c.setGrid(fc.getGridObj())
	c.owner = fc.owner
	c.forceAttackSkillAmount = fc.forceAttackSkillAmount
	c.publicEnemySkillAmount = fc.publicEnemySkillAmount
	c.fogAmount = fc.fogAmount
	c.setTargetType(fc.getType())
	fc.situation.getTargetMgr().replaceTarget(c)
	c.enterBattle()
	c.initController = initController
	c.initSit = initSit
	return c, acts
}

func (fc *fightCard) isPublicEnemy() bool {
	return fc.publicEnemySkillAmount > 0
}

func (fc *fightCard) isInFog() bool {
	return fc.fogAmount > 0
}

func (fc *fightCard) onEnterFog() {
	fc.fogAmount++
}

func (fc *fightCard) onLeaveFog() {
	if fc.fogAmount > 0 {
		fc.fogAmount--
	}
}

func (fc *fightCard) addSkill(skillID int32, boutTimeout int) []*clientAction {
	return fc.casterAddSkill(skillID, boutTimeout, fc.gridObjID <= 0)
}

func (fc *fightCard) packDrawCard() *drawCard {
	var equip string
	if fc.equip != nil {
		equip = fc.equip.data.ID
	}
	return newDrawCard(fc.gcardID, fc.skin, equip)
}

func (fc *fightCard) hasEquip() bool {
	return fc.equip != nil
}

func (fc *fightCard) delEquip() []*clientAction {
	if fc.equip == nil {
		return []*clientAction{}
	}

	acts := fc.equip.leaveBattle()
	fc.equip = nil
	return acts
}
