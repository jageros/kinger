package main

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"strconv"
	"strings"
)

var allTargetFilters = map[int]*targetFilter{}

// 技能目标
type iTarget interface {
	getObjID() int // 当前战斗中的唯一id
	getType() int
	getSit() int // 座位，判断是否跟我同边
	getGrid() int
	getCamp() int
	getInitSit() int
	getCardType() int
	copy(situation *battleSituation) iTarget
	packAttr() *attribute.MapAttr
	restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation)
	boutBegin() []*clientAction
	addEffect(ownerObjID int, movieID string, playType, boutTimeout int) *clientAction
	delEffect(movieID string) *clientAction
	getCopyTarget() iTarget
	isPublicEnemy() bool

	//addBuff(buff iBuff) *pb.ClientAction
	//delBuff(buff iBuff) *pb.ClientAction
}

type mcMovieEffect struct {
	movieID     string
	playType    int
	invalidBout int
	ownerObjID  int
}

func (m *mcMovieEffect) packMsg() *pb.MovieEffect {
	return &pb.MovieEffect{
		MovieID:    m.movieID,
		PlayType:   int32(m.playType),
		OwnerObjID: int32(m.ownerObjID),
	}
}

func (m *mcMovieEffect) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetStr("movieID", m.movieID)
	attr.SetInt("playType", m.playType)
	attr.SetInt("invalidBout", m.invalidBout)
	attr.SetInt("ownerObjID", m.ownerObjID)
	return attr
}

func (m *mcMovieEffect) restoredFromAttr(attr *attribute.MapAttr) {
	m.movieID = attr.GetStr("movieID")
	m.playType = attr.GetInt("playType")
	m.invalidBout = attr.GetInt("invalidBout")
	m.ownerObjID = attr.GetInt("ownerObjID")
}

type baseTarget struct {
	objID      int
	targetType int
	sit        int
	gridID     int
	camp       int
	initSit    int
	//buffs []iBuff
	effects   []*mcMovieEffect
	situation *battleSituation
}

func newBaseTarget(situation *battleSituation) *baseTarget {
	return &baseTarget{
		situation: situation,
		gridID:    -1,
	}
}

func (bt *baseTarget) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("objID", bt.objID)
	attr.SetInt("targetType", bt.targetType)
	attr.SetInt("sit", bt.sit)
	attr.SetInt("gridID", bt.gridID)
	attr.SetInt("camp", bt.camp)
	attr.SetInt("initSit", bt.initSit)

	effectsAttr := attribute.NewListAttr()
	attr.SetListAttr("effects", effectsAttr)
	for _, e := range bt.effects {
		effectsAttr.AppendMapAttr(e.packAttr())
	}

	return attr
}

func (bt *baseTarget) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	bt.situation = situation
	bt.objID = attr.GetInt("objID")
	bt.targetType = attr.GetInt("targetType")
	bt.sit = attr.GetInt("sit")
	bt.gridID = attr.GetInt("gridID")
	bt.camp = attr.GetInt("camp")
	bt.initSit = attr.GetInt("initSit")
	effectsAttr := attr.GetListAttr("effects")
	effectsAttr.ForEachIndex(func(index int) bool {
		e := &mcMovieEffect{}
		e.restoredFromAttr(effectsAttr.GetMapAttr(index))
		bt.effects = append(bt.effects, e)
		return true
	})
}

func (bt *baseTarget) getType() int {
	return bt.targetType
}

func (bt *baseTarget) setTargetType(targetType int) {
	bt.targetType = targetType
}

func (bt *baseTarget) copy(situation *battleSituation) *baseTarget {
	c := *bt
	cpy := &c
	cpy.situation = situation
	//for _, b := range bt.buffs {
	//	if cb := b.copy(); cb != nil {
	//		cpy.buffs = append(cpy.buffs, cb)
	//	}
	//}
	return cpy
}

func (bt *baseTarget) getObjID() int {
	return bt.objID
}

func (bt *baseTarget) getSit() int {
	return bt.sit
}

func (bt *baseTarget) setSit(sit int) {
	bt.sit = sit
}

func (bt *baseTarget) getGrid() int {
	return bt.gridID
}

func (bt *baseTarget) getCamp() int {
	return bt.camp
}

func (bt *baseTarget) setCamp(camp int) {
	bt.camp = camp
}

func (bt *baseTarget) getInitSit() int {
	return bt.initSit
}

func (bt *baseTarget) getCardType() int {
	return -1
}

func (bt *baseTarget) getCopyTarget() iTarget {
	return nil
}

func (bt *baseTarget) isPublicEnemy() bool {
	return false
}

/*
func (bt *baseTarget) addBuff(buff iBuff) *pb.ClientAction {
	bt.buffs = append(bt.buffs, buff)
	return buff.onAdd()
}

func (bt *baseTarget) delBuff(buff iBuff) *pb.ClientAction {
	index := -1
	for i, b := range bt.buffs {
		if b == buff {
			index = i
			break
		}
	}

	if index >= 0 {
		bt.buffs = append(bt.buffs[:index], bt.buffs[index+1:]...)
		return buff.onDel()
	} else {
		return nil
	}
}
*/

func (bt *baseTarget) boutBegin() []*clientAction {
	var actions []*clientAction
	curBout := bt.situation.getCurBout()
	for i := 0; i < len(bt.effects); {
		e := bt.effects[i]
		if e.invalidBout >= curBout {
			bt.effects = append(bt.effects[:i], bt.effects[i+1:]...)
			actions = append(actions, &clientAction{
				actID: pb.ClientAction_Movie,
				actMsg: &pb.MovieAct{
					MovieID:  e.movieID,
					Targets:  []int32{int32(bt.getObjID())},
					PlayType: 0,
				},
			})
		} else {
			i++
		}
	}
	return actions
}

func (bt *baseTarget) addEffect(ownerObjID int, movieID string, playType, boutTimeout int) *clientAction {
	if boutTimeout > 0 {
		boutTimeout += bt.situation.getCurBout()
	}
	bt.effects = append(bt.effects, &mcMovieEffect{
		movieID:     movieID,
		playType:    playType,
		invalidBout: boutTimeout,
		ownerObjID:  ownerObjID,
	})
	return &clientAction{
		actID: pb.ClientAction_Movie,
		actMsg: &pb.MovieAct{
			MovieID:    movieID,
			Targets:    []int32{int32(bt.getObjID())},
			PlayType:   int32(playType),
			OwnerObjID: int32(ownerObjID),
		},
	}
}

func (bt *baseTarget) delEffect(movieID string) *clientAction {
	for i, movieInfo := range bt.effects {
		if movieInfo.movieID == movieID && movieInfo.invalidBout <= 0 {
			bt.effects = append(bt.effects[:i], bt.effects[i+1:]...)
			return &clientAction{
				actID: pb.ClientAction_Movie,
				actMsg: &pb.MovieAct{
					MovieID:  movieID,
					Targets:  []int32{int32(bt.getObjID())},
					PlayType: 0,
				},
			}
		}
	}
	return nil
}

type targetCondition struct {
	operator string
	targetID int
	amount   int
}

func newTargetCondition(condition string) *targetCondition {
	op := "=="
	i := strings.Index(condition, op)
	if i < 0 {
		op = "<="
		i = strings.Index(condition, op)
	}
	if i < 0 {
		op = ">="
		i = strings.Index(condition, op)
	}
	if i < 0 {
		op = "<"
		i = strings.Index(condition, op)
	}
	if i < 0 {
		op = ">"
		i = strings.Index(condition, op)
	}
	if i < 0 {
		glog.Errorf("newTargetCondition i %d", i)
		return nil
	}

	targetID, err := strconv.Atoi(strings.TrimSpace(condition[:i]))
	if err != nil {
		glog.Errorf("newTargetCondition targetID %s %s", condition, err)
		return nil
	}

	amount, err := strconv.Atoi(strings.TrimSpace(condition[i+len(op):]))
	if err != nil {
		glog.Errorf("newTargetCondition amount %s %s", condition, err)
		return nil
	}

	return &targetCondition{
		operator: op,
		targetID: targetID,
		amount:   amount,
	}
}

func (c *targetCondition) check(sk *skill, skillOwner iTarget, triggerCxt *triggerContext, situation *battleSituation,
	cacheTargets map[int][]iTarget) bool {

	var targets []iTarget = nil
	if cacheTargets != nil {
		if ts, ok := cacheTargets[c.targetID]; ok {
			targets = ts
		}
	}

	if targets == nil {
		targets = situation.getTargetMgr().findTarget(sk, skillOwner, c.targetID, triggerCxt, cacheTargets)
	}

	amount := len(targets)
	switch c.operator {
	case "==":
		if amount != c.amount {
			return false
		}
	case "<":
		if amount >= c.amount {
			return false
		}
	case ">":
		if amount <= c.amount {
			return false
		}
	case "<=":
		if amount > c.amount {
			return false
		}
	case ">=":
		if amount < c.amount {
			return false
		}
	}
	return true
}

type targetFilter struct {
	data       *gamedata.Target
	matchers   []iTargetMatcher
	conditions []*targetCondition
}

func newTargetFilter(data *gamedata.Target) *targetFilter {
	tf := &targetFilter{
		data: data,
	}

	if data.TargetSummon != 0 {
		tf.matchers = append(tf.matchers, &targetSummonMatcher{summon: data.TargetSummon})
	}
	if data.TargetClean != 0 {
		tf.matchers = append(tf.matchers, &targetCleanDestoryMatcher{clean: data.TargetClean})
	}
	if len(data.TargetCard) > 0 {
		tf.matchers = append(tf.matchers, &targetCardMatcher{cardIDs: data.TargetCard})
	}
	if len(data.NotargetCard) > 0 {
		tf.matchers = append(tf.matchers, &targetNoCardMatcher{cardIDs: data.NotargetCard})
	}
	if data.Side != 0 {
		tf.matchers = append(tf.matchers, &targetSideMatcher{side: data.Side})
	}
	if data.PreTurnSide != 0 {
		tf.matchers = append(tf.matchers, &targetPreTurnSideMatcher{preTurnSide: data.PreTurnSide})
	}
	if len(data.InitSide) > 0 {
		tf.matchers = append(tf.matchers, &targetInitSideMatcher{initSides: data.InitSide})
	}
	if len(data.Types) > 0 {
		tf.matchers = append(tf.matchers, &targetTypeMatcher{types: data.Types})
	}
	if len(data.TargetBat) > 0 {
		tf.matchers = append(tf.matchers, &targetBatMatcher{bats: data.TargetBat})
	}
	if data.Turn != 0 {
		tf.matchers = append(tf.matchers, &targetTurnMatcher{turn: data.Turn})
	}
	if len(data.Camp) > 0 {
		tf.matchers = append(tf.matchers, &targetCampMatcher{camps: data.Camp})
	}
	if len(data.TargetSkill) > 0 {
		tf.matchers = append(tf.matchers, &targetSkillMatcher{skills: data.TargetSkill})
	}
	if len(data.TargetSkillFog) > 0 {
		tf.matchers = append(tf.matchers, &targetSkillFogMatcher{skills: data.TargetSkillFog})
	}
	if len(data.NoTargetSkill) > 0 {
		tm := &targetNoSkillMatcher{}
		tm.skills = data.NoTargetSkill
		tf.matchers = append(tf.matchers, tm)
	}
	if len(data.NoTargetSkillFog) > 0 {
		tf.matchers = append(tf.matchers, &targetNoSkillFogMatcher{skills: data.NoTargetSkillFog})
	}
	if len(data.TargetAtt) > 0 {
		tf.matchers = append(tf.matchers, &targetAttackMatcher{attTypes: data.TargetAtt})
	}
	if data.CardType != 0 {
		tf.matchers = append(tf.matchers, &targetCardTypeMatcher{cardType: data.CardType})
	}
	if len(data.Sequential) > 0 {
		tf.matchers = append(tf.matchers, &targetSequentialMatcher{sequentials: data.Sequential})
	}
	if len(data.Poses) > 0 {
		tf.matchers = append(tf.matchers, &targetPosMatcher{posTypes: data.Poses})
	}
	if len(data.Type2) > 0 {
		tf.matchers = append(tf.matchers, &targetType2Matcher{targetIDs: data.Type2})
	}
	if data.Surrender > 0 {
		tf.matchers = append(tf.matchers, &targetSurrenderMatcher{isSurrenderor: data.Surrender == surrenderor})
	}
	if data.TargetEquip != 0 {
		tf.matchers = append(tf.matchers, &targetEquipMatcher{hasEquip: data.TargetEquip == 1})
	}

	for _, condition := range data.Condition {
		c := newTargetCondition(condition)
		if c != nil {
			tf.conditions = append(tf.conditions, c)
		}
	}
	return tf
}

func (tf *targetFilter) getRelativeTargetID() int {
	return tf.data.Relative
}

func (tf *targetFilter) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	if cacheTargets == nil {
		cacheTargets = map[int][]iTarget{}
	}

	target = target.getCopyTarget()
	var ok bool
	for _, m := range tf.matchers {
		target, ok = m.checkMatch(sk, skillOwner, relative, target, triggerCxt, situation, cacheTargets)
		if !ok {
			return target, false
		}
	}

	if len(tf.conditions) > 0 {
		for _, c := range tf.conditions {
			if !c.check(sk, target, triggerCxt, situation, map[int][]iTarget{}) {
				return target, false
			}
		}
	}
	return target, true
}

func (tf *targetFilter) isTarget(sk *skill, skillOwner, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) bool {

	if tf.data.ID == 1 && skillOwner == target {
		// 技能拥有者
		return true
	}

	if cacheTargets != nil {
		if ts, ok := cacheTargets[tf.data.ID]; ok {
			for _, t := range ts {
				if t.getObjID() == target.getObjID() {
					return true
				}
			}
			return false
		}
	}

	var relatives []iTarget
	var ok bool
	relativeTargetID := tf.getRelativeTargetID()
	if relativeTargetID <= 0 {
		return false
	} else if relativeTargetID == 1 {
		// 技能拥有者
		relatives = []iTarget{skillOwner}
	} else {
		if cacheTargets == nil {
			cacheTargets = map[int][]iTarget{}
		} else {
			relatives, ok = cacheTargets[relativeTargetID]
		}
		if !ok {
			relatives = situation.getTargetMgr().findTarget(sk, skillOwner, relativeTargetID, triggerCxt, cacheTargets)
		}
	}

	for _, relative := range relatives {
		t, ok := tf.checkMatch(sk, skillOwner, relative, target, triggerCxt, situation, cacheTargets)
		if ok && t == target {
			return true
		}
	}

	return false
}

// --------------------------- target matcher begin ------------------------------
type iTargetMatcher interface {
	checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext, situation *battleSituation,
		cacheTargets map[int][]iTarget) (iTarget, bool)
}

// 是否是召唤物
type targetSummonMatcher struct {
	summon int
}

func (tm *targetSummonMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetCard, isCard := target.(*fightCard)
	if !isCard {
		return target, false
	}

	//if targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
	//	return target, false
	//}

	if tm.summon == 1 && !targetCard.isSummon {
		return target, false
	} else if tm.summon != 1 && targetCard.isSummon {
		return target, false
	}

	return target, true
}

// 是否是根除者或被根除者
type targetCleanDestoryMatcher struct {
	clean int
}

func (tm *targetCleanDestoryMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetCard, isCard := target.(*fightCard)
	if !isCard {
		return target, false
	}
	if tm.clean == 1 {
		if !triggerCxt.isDestoryer(targetCard) {
			return target, false
		}
	} else {
		if !triggerCxt.isBeDestoryer(targetCard) {
			return target, false
		}
	}

	return target, true
}

// 是否是目标卡
type targetCardMatcher struct {
	cardIDs []uint32
}

func (tm *targetCardMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetCard, isCard := target.(*fightCard)
	if !isCard {
		return target, false
	}
	//if targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
	//	return target, false
	//}
	for _, cardID := range tm.cardIDs {
		if cardID == targetCard.cardID {
			return target, true
		}
	}
	return target, false
}

// 是否不是目标卡
type targetNoCardMatcher struct {
	cardIDs []uint32
}

func (tm *targetNoCardMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetCard, isCard := target.(*fightCard)
	if !isCard {
		return target, false
	}
	if targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
		return target, false
	}
	for _, cardID := range tm.cardIDs {
		if cardID == targetCard.cardID {
			return target, false
		}
	}

	return target, true
}

// 是否是目标阵营
type targetSideMatcher struct {
	side int
}

func (tm *targetSideMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetSit := target.getSit()
	if targetSit <= 0 {
		return target, false
	}
	relativeSit := relative.getSit()
	if relative == skillOwner {
		relativeSit = triggerCxt.getTriggerTargetSit(skillOwner)
	}
	if relativeSit <= 0 {
		return target, false
	}

	if tm.side == sOwn {
		if relative == target {
			return target, true
		}
		if targetSit != relativeSit || target.isPublicEnemy() {
			return target, false
		}
	} else {
		if relative == target {
			return target, false
		}
		if target.isPublicEnemy() {
			return target, true
		}
		if targetSit == relativeSit {
			return target, false
		}
	}

	return target, true
}

// 是否是目标翻面前阵营
type targetPreTurnSideMatcher struct {
	preTurnSide int
}

func (tm *targetPreTurnSideMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	preTurnSit := triggerCxt.getPreTurnSit(target)
	relativePreTurnSit := triggerCxt.getPreTurnSit(relative)
	if tm.preTurnSide == sOwn {
		if relative == target {
			return target, true
		}
		if preTurnSit != relativePreTurnSit || target.isPublicEnemy() {
			return target, false
		}
	} else {
		if relative == target {
			return target, false
		}
		if target.isPublicEnemy() {
			return target, true
		}
		if preTurnSit == relativePreTurnSit {
			return target, false
		}
	}

	return target, true
}

// 是否是目标初始阵营
type targetInitSideMatcher struct {
	initSides []int
}

func (tm *targetInitSideMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetSit := target.getSit()
	targetInitSit := target.getInitSit()
	relativeInitSit := relative.getInitSit()
	for _, initSide := range tm.initSides {
		switch initSide {
		case sInitOwn1:
			if relative == target {
				return target, true
			}
			if target.isPublicEnemy() {
				return target, false
			}
			if targetInitSit == relativeInitSit {
				return target, true
			}
		case sInitEnemy1:
			if relative == target {
				return target, false
			}
			if target.isPublicEnemy() {
				return target, true
			}
			if targetInitSit != relativeInitSit {
				return target, true
			}
		case sInitOwn:
			if targetInitSit == targetSit {
				return target, true
			}
		case sInitEnemy:
			if targetInitSit != targetSit {
				return target, true
			}
		case sRelativeInitOwn:
			if relative != target && target.isPublicEnemy() {
				return target, false
			}
			if targetSit == relativeInitSit {
				return target, true
			}
		case sRelativeInitEnemy:
			if relative != target && target.isPublicEnemy() {
				return target, true
			}
			if targetSit != relativeInitSit {
				return target, true
			}
		}
	}

	return target, false
}

// 是否是目标类型
type targetTypeMatcher struct {
	types []int
}

func (tm *targetTypeMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	isFindGrid := false
	typeSet := common.IntSet{}
	targetCard, isCard := target.(*fightCard)
	if target.getObjID() == skillOwner.getObjID() {
		typeSet.Add(stOwner)
		if triggerCxt.isPreMoveCard(target) {
			typeSet.Add(stPreMoveCard)
		} else if triggerCxt.isMoveCard(target) {
			typeSet.Add(stMoveCard)
		}
	} else {

		if triggerCxt.isPreMoveCard(target) {
			typeSet.Add(stPreMoveCard)
			typeSet.Add(stInDesk)
			typeSet.Add(target.getType())
		} else if triggerCxt.isMoveCard(target) {
			typeSet.Add(stMoveCard)
			typeSet.Add(stInDesk)
			typeSet.Add(target.getType())
		} else if triggerCxt.isDrawCard(target) {
			typeSet.Add(stHand)
			typeSet.Add(stDrawCard)
		} else if triggerCxt.isReturnCard(target) {
			typeSet.Add(stReturnCard)
			typeSet.Add(target.getType())
		} else if isCard && triggerCxt.isEnterBattleCard(targetCard) {
			typeSet.Add(stEnterDesk)
		} else {
			typeSet.Add(target.getType())
		}
	}

	for _, ty := range tm.types {
		if typeSet.Contains(ty) {
			return target, true
		} else if ty == stGrid {
			isFindGrid = true
		}
	}

	if isFindGrid {
		// TODO 找有牌的格子
		if isCard {
			gridObj := targetCard.getGridObj()
			if gridObj == nil {
				return target, false
			}
			return gridObj, true
		} else {
			return target, false
		}
	} else {
		return target, false
	}
}

// 是否是目标比点结果
type targetBatMatcher struct {
	bats []int
}

func (tm *targetBatMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	batInfo := triggerCxt.getAttackBatInfo(target)
	for _, b := range tm.bats {
		if batInfo.Contains(b) {
			return target, true
		}
	}

	return target, false
}

// 是否是翻面者或被翻面者
type targetTurnMatcher struct {
	turn int
}

func (tm *targetTurnMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	var ok bool
	switch tm.turn {
	case tTurner:
		ok = triggerCxt.triggerType != preBeTurnTrigger && triggerCxt.isTurnner(target)
	case tBeTurner:
		ok = triggerCxt.triggerType != preBeTurnTrigger && triggerCxt.isBeTurnner(target)
	case tPreTurner:
		ok = triggerCxt.triggerType == preBeTurnTrigger && triggerCxt.isTurnner(target)
	case tPreBeTurner:
		ok = triggerCxt.triggerType == preBeTurnTrigger && triggerCxt.isPreBeTurnner(target)
	case tCantBeTurner:
		ok = !triggerCxt.isCanTurn(target)
	}

	//if sk.getID() == 1234 {
	//	glog.Infof("targetTurnMatcher target=%s, tm.turn=%d, ok=%v, triggerType=%d, turner=%s, beTurner=%v",
	//		target, tm.turn, ok, triggerCxt.turner, triggerCxt.beTurners)
	//}

	return target, ok
}

// 是否是目标国家
type targetCampMatcher struct {
	camps []int
}

func (tm *targetCampMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	targetCard, ok := target.(*fightCard)
	if ok && targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
		return target, false
	}

	camp := target.getCamp()
	for _, camp2 := range tm.camps {
		if camp == camp2 {
			return target, true
		}
	}
	return target, false
}

// 是否有目标技能
type targetSkillMatcher struct {
	skills []int32
}

func (tm *targetSkillMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	c, ok := target.(iCaster)
	if !ok {
		return target, false
	}

	targetCard, ok := target.(*fightCard)
	if ok && targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
		return target, false
	}

	for _, skID := range tm.skills {
		if c.hasSkillByIDIgnoreEquip(skID) {
			return target, true
		}
	}
	return target, false
}

// 是否有目标技能（不收大雾影响）
type targetSkillFogMatcher struct {
	skills []int32
}

func (tm *targetSkillFogMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	c, ok := target.(iCaster)
	if !ok {
		return target, false
	}

	for _, skID := range tm.skills {
		if c.hasSkillByIDIgnoreEquip(skID) {
			return target, true
		}
	}
	return target, false
}

// 是否没有目标技能（不收大雾影响）
type targetNoSkillFogMatcher struct {
	skills []int32
}

func (tm *targetNoSkillFogMatcher) isMatch(c iCaster) bool {
	for _, skID := range tm.skills {
		if skID == -1 {
			if !c.hasNoSkillIgnoreEquip() {
				return false
			}
		} else if c.hasSkillByIDIgnoreEquip(skID) {
			return false
		}
	}
	return true
}

func (tm *targetNoSkillFogMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	c, ok := target.(iCaster)
	if !ok {
		return target, false
	}
	return target, tm.isMatch(c)
}

// 是否没有目标技能
type targetNoSkillMatcher struct {
	targetNoSkillFogMatcher
}

func (tm *targetNoSkillMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	c, ok := target.(iCaster)
	if !ok {
		return target, false
	}

	targetCard, ok := target.(*fightCard)
	if ok && targetCard.isInFog() && targetCard.getSit() != sk.getOwner().getSit() {
		return target, false
	}

	return target, tm.isMatch(c)
}

// 是否是攻击者或受击者
type targetAttackMatcher struct {
	attTypes []int
}

func (tm *targetAttackMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	//if sk.getID() == 1234 {
	//	glog.Infof("targetAttackMatcher begin target=%s, tm.attTypes=%v", target, tm.attTypes)
	//}

	targetCard, ok := target.(*fightCard)
	if !ok {
		return target, false
	}
	for _, at := range tm.attTypes {
		switch at {
		case atAttacker:
			if triggerCxt.isAttackCard(target) {
				return target, true
			}
		case atBeAttacker:
			if triggerCxt.isBeAttackCard(target) {
				return target, true
			}
		case atForceAttacker:
			if !triggerCxt.isAttackCard(target) {
				return target, false
			}

			batInfo := triggerCxt.getAttackBatInfo(target)
			if batInfo.Contains(bWin) || batInfo.Contains(bGt) || targetCard.isForceAttack() {
				return target, true
			}

		case atForceBeAttacker:
			if !triggerCxt.isBeAttackCard(target) {
				return target, false
			}

			attackCard := triggerCxt.getAttackCard()
			if attackCard == nil {
				return target, false
			}

			if attackCard.isForceAttack() {
				return target, true
			}

			attackResult := triggerCxt.getAttackResult(attackCard, targetCard)
			if attackResult == bWin || attackResult == bGt {
				return target, true
			}

			attackOutcome := triggerCxt.getAttackOutcome(attackCard, targetCard)
			if attackOutcome == bWin || attackOutcome == bGt {
				return target, true
			}

			return target, false
		}
	}

	//if sk.getID() == 1234 {
	//	glog.Infof("targetAttackMatcher target=%s, tm.attTypes=%d, attacker=%s, beattacker=%v",
	//		target, tm.attTypes, triggerCxt.attackCard, triggerCxt.beAttackCards)
	//}

	return target, false
}

// 是否是将军或小兵 (pc 端)
type targetCardTypeMatcher struct {
	cardType int
}

func (tm *targetCardTypeMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	return target, target.getCardType() == tm.cardType
}

type targetSequentialMatcher struct {
	sequentials [][]int
}

func (tm *targetSequentialMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	if _, ok := target.(*fightCard); !ok {
		return target, false
	}
	relativeCard, ok := relative.(iCaster)
	if !ok {
		return target, false
	}

	playCardIdx := sk.getPlayCardIdx()
	targetObjID := target.getObjID()
	for _, seq := range tm.sequentials {
		var objID int
		side := 0
		op := seq[0]
		n := seq[1]
		if len(seq) >= 3 {
			side = seq[0]
			op = seq[1]
			n = seq[2]
		}
		if op == 1 {
			objID = situation.getPrePlayCard(relativeCard, triggerCxt.triggerType, side, n)
		} else {
			objID = situation.getNextPlayCard(relativeCard, playCardIdx, side, n)
		}

		if targetObjID == objID {
			return target, true
		}
	}

	return target, false
}

// 是否是目标位置
type targetPosMatcher struct {
	posTypes []int
}

func (tm *targetPosMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	grid := target.getGrid()
	if grid < 0 {
		return target, false
	}
	relativeGrid := relative.getGrid()
	if relativeGrid < 0 {
		return target, false
	}
	column := situation.getGridColumn()

L:
	for _, posType := range tm.posTypes {
		switch posType {
		case pAll:
			return target, true
		case pOwn:
			if grid == relativeGrid {
				return target, true
			}
		case pAdjoin:
			if grid+column == relativeGrid || grid-column == relativeGrid ||
				(grid+1 == relativeGrid && grid/column == relativeGrid/column) ||
				(grid-1 == relativeGrid && grid/column == relativeGrid/column) {
				return target, true
			}
		case pApart:
			if grid/column == relativeGrid/column {
				apartGrids := grid - relativeGrid
				if apartGrids >= 2 || apartGrids <= -2 {
					return target, true
				}
			} else {
				targetRow := grid % column
				relativeRow := relativeGrid % column
				if targetRow == relativeRow {
					apartGrids := grid/column - relativeGrid/column
					if apartGrids >= 2 || apartGrids <= -2 {
						return target, true
					}
				}
			}
		case pApartEmpty:
			if grid/column == relativeGrid/column {
				apartGrids := grid - relativeGrid
				if apartGrids >= 2 || apartGrids <= -2 {
					grid1 := grid
					grid2 := relativeGrid
					if grid1 > grid2 {
						grid1, grid2 = grid2, grid1
					}
					for g := grid1 + 1; g < grid2; g++ {
						t := situation.getTargetInGrid(g)
						if t.getType() != stEmptyGrid {
							continue L
						}
					}
					return target, true
				}
			} else {
				targetRow := grid % column
				relativeRow := relativeGrid % column
				if targetRow == relativeRow {
					apartGrids := grid/column - relativeGrid/column
					if apartGrids >= 2 || apartGrids <= -2 {
						grid1 := grid
						grid2 := relativeGrid
						if grid1 > grid2 {
							grid1, grid2 = grid2, grid1
						}
						for g := grid1 + column; g < grid2; g += column {
							t := situation.getTargetInGrid(g)
							if t.getType() != stEmptyGrid {
								continue L
							}
						}
						return target, true
					}
				}
			}

		case pNotAdjoin:
			if !(grid+column == relativeGrid || grid-column == relativeGrid ||
				(grid+1 == relativeGrid && grid/column == relativeGrid/column) ||
				(grid-1 == relativeGrid && grid/column == relativeGrid/column)) {

				return target, true
			}
		}
	}

	return target, false
}

type targetType2Matcher struct {
	targetIDs []int
}

func (tm *targetType2Matcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	for _, targetID := range tm.targetIDs {
		ts, ok := cacheTargets[targetID]
		if !ok {
			ts = situation.getTargetMgr().findTarget(sk, skillOwner, targetID, triggerCxt, cacheTargets)
			cacheTargets[targetID] = ts
		}

		for _, t := range ts {
			if t == target {
				return target, true
			}
		}
	}

	return target, false
}

type targetSurrenderMatcher struct {
	isSurrenderor bool
}

func (tm *targetSurrenderMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	if tm.isSurrenderor {
		return target, triggerCxt.surrenderor != nil && triggerCxt.surrenderor.getObjID() == target.getObjID()
	} else {
		return target, triggerCxt.beSurrenderor != nil && triggerCxt.beSurrenderor.getObjID() == target.getObjID()
	}
}

type targetEquipMatcher struct {
	hasEquip bool
}

func (tm *targetEquipMatcher) checkMatch(sk *skill, skillOwner, relative, target iTarget, triggerCxt *triggerContext,
	situation *battleSituation, cacheTargets map[int][]iTarget) (iTarget, bool) {

	c, ok := target.(*fightCard)
	if !ok || c.isInFog() {
		return target, false
	}

	return target, c.hasEquip() == tm.hasEquip
}

// --------------------------- target matcher end ------------------------------

func doInitSkillTarget(gdata gamedata.IGameData) {
	targetGameData := gdata.(*gamedata.TargetGameData)
	filters := map[int]*targetFilter{}
	for targetID, data := range targetGameData.GetAllTarget() {
		filters[targetID] = newTargetFilter(data)
	}
	allTargetFilters = filters
}

func initSkillTarget() {
	gdata := gamedata.GetGameData(consts.Target)
	gdata.AddReloadCallback(doInitSkillTarget)
	doInitSkillTarget(gdata)
}

type skillTargetMgr struct {
	situation *battleSituation
	// 当前战斗的所有target
	allTargets map[int]iTarget
	// 可以成为技能目标的target
	skillTargets map[int]iTarget
}

func newSkillTargetMgr(situation *battleSituation) *skillTargetMgr {
	return &skillTargetMgr{
		situation:    situation,
		allTargets:   map[int]iTarget{},
		skillTargets: map[int]iTarget{},
	}
}

func (stm *skillTargetMgr) copy(situation *battleSituation) *skillTargetMgr {
	cpy := &skillTargetMgr{
		situation:    situation,
		allTargets:   map[int]iTarget{},
		skillTargets: map[int]iTarget{},
	}

	for objID, t := range stm.allTargets {
		cpy.allTargets[objID] = t.copy(situation)
	}
	for objID, _ := range stm.skillTargets {
		cpy.skillTargets[objID] = cpy.allTargets[objID]
	}
	return cpy
}

func (stm *skillTargetMgr) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	targetsAttr := attribute.NewMapAttr()
	attr.SetMapAttr("targets", targetsAttr)
	for objID, t := range stm.allTargets {
		targetsAttr.SetMapAttr(strconv.Itoa(objID), t.packAttr())
	}

	skillTargetsAttr := attribute.NewListAttr()
	attr.SetListAttr("skillTargets", skillTargetsAttr)
	for objID, _ := range stm.skillTargets {
		skillTargetsAttr.AppendInt(objID)
	}

	return attr
}

func (stm *skillTargetMgr) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	stm.situation = situation
	stm.allTargets = map[int]iTarget{}
	stm.skillTargets = map[int]iTarget{}

	targetsAttr := attr.GetMapAttr("targets")
	targetsAttr.ForEachKey(func(key string) {
		objID, _ := strconv.Atoi(key)
		tAttr := targetsAttr.GetMapAttr(key)
		attrType := tAttr.GetInt("attrType")
		var t iTarget
		switch attrType {
		case attrGrid:
			t = &deskGrid{}
		case attrCard:
			t = &fightCard{}
		case attrFort:
			t = &fortCaster{}
		case attrFighter:
			t = &fighter{}
		default:
			return
		}
		t.restoredFromAttr(tAttr, situation)
		stm.allTargets[objID] = t
	})

	skillTargetsAttr := attr.GetListAttr("skillTargets")
	skillTargetsAttr.ForEachIndex(func(index int) bool {
		objID := skillTargetsAttr.GetInt(index)
		stm.skillTargets[objID] = stm.allTargets[objID]
		return true
	})
}

func (stm *skillTargetMgr) boutBegin() []*clientAction {
	var actions []*clientAction
	for _, t := range stm.allTargets {
		actions = append(actions, t.boutBegin()...)
	}
	return actions
}

func (stm *skillTargetMgr) replaceTarget(target iTarget) {
	objID := target.getObjID()
	if _, ok := stm.skillTargets[objID]; ok {
		stm.skillTargets[objID] = target
	}
	if _, ok := stm.allTargets[objID]; ok {
		stm.allTargets[objID] = target
	}
}

func (stm *skillTargetMgr) addTarget(target iTarget) {
	objID := target.getObjID()
	stm.skillTargets[objID] = target
	stm.allTargets[objID] = target
}

func (stm *skillTargetMgr) getTarget(objID int) iTarget {
	if t, ok := stm.allTargets[objID]; ok {
		return t
	} else {
		return nil
	}
}

func (stm *skillTargetMgr) getTargetCard(objID int) *fightCard {
	if t, ok := stm.allTargets[objID]; ok {
		if c, ok2 := t.(*fightCard); ok2 {
			return c
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (stm *skillTargetMgr) getTargetGrid(objID int) *deskGrid {
	if t, ok := stm.allTargets[objID]; ok {
		if g, ok2 := t.(*deskGrid); ok2 {
			return g
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (stm *skillTargetMgr) delSkillTarget(objID int) {
	delete(stm.skillTargets, objID)
}

func (stm *skillTargetMgr) delTarget(objID int) {
	delete(stm.allTargets, objID)
	delete(stm.skillTargets, objID)
}

func (stm *skillTargetMgr) findTarget(sk *skill, skillOwner iTarget, targetID int, triggerCxt *triggerContext,
	cacheTargets map[int][]iTarget) []iTarget {

	if targetID == 1 {
		// 技能拥有者
		return []iTarget{skillOwner}
	}

	if cacheTargets != nil {
		if ts, ok := cacheTargets[targetID]; ok {
			return ts
		}
	}

	filter, ok := allTargetFilters[targetID]
	if !ok {
		return []iTarget{}
	}

	relativeTargetID := filter.getRelativeTargetID()
	var relatives []iTarget
	ok = false
	if relativeTargetID <= 0 {
		return []iTarget{}
	} else if relativeTargetID == 1 {
		// 技能拥有者
		relatives = []iTarget{skillOwner}
	} else {
		if cacheTargets == nil {
			cacheTargets = map[int][]iTarget{}
		} else {
			relatives, ok = cacheTargets[relativeTargetID]
		}
		if !ok {
			relatives = stm.findTarget(sk, skillOwner, relativeTargetID, triggerCxt, cacheTargets)
		}
	}

	var ts []interface{}
	targetSet := common.IntSet{}
	var allTargets []iTarget = nil

	if filter.data.TargetClean != 0 {
		if filter.data.TargetClean == 1 {
			if triggerCxt.destoryer != nil {
				allTargets = []iTarget{triggerCxt.destoryer}
			} else {
				allTargets = []iTarget{}
			}
		} else {
			allTargets = triggerCxt.beDestoryers
		}
	} else if filter.data.Turn != 0 {
		allTargets = []iTarget{}
		if filter.data.Turn == tTurner || filter.data.Turn == tPreTurner {
			if triggerCxt.turner != nil {
				allTargets = []iTarget{stm.getTarget(triggerCxt.turner.c.getObjID())}
			}
		} else if filter.data.Turn == tCantBeTurner {
			allTargets = triggerCxt.cantTurnCards
		} else {
			if triggerCxt.beTurners != nil {
				for _, t := range triggerCxt.beTurners {
					allTargets = append(allTargets, t.c)
				}
			}
		}
	} else if len(filter.data.TargetAtt) > 0 || len(filter.data.TargetBat) > 0 {
		allTargets = []iTarget{}
		if triggerCxt.attackCard != nil {
			allTargets = append(allTargets, triggerCxt.attackCard)
		}
		for _, c := range triggerCxt.beAttackCards {
			allTargets = append(allTargets, c)
		}
	} else if filter.data.Surrender == surrenderor {
		allTargets = []iTarget{}
		if triggerCxt.surrenderor != nil {
			allTargets = []iTarget{triggerCxt.surrenderor}
		}
	} else if filter.data.Surrender == beSurrenderor {
		allTargets = []iTarget{}
		if triggerCxt.beSurrenderor != nil {
			allTargets = []iTarget{triggerCxt.beSurrenderor}
		}
	}

	if allTargets == nil {

		for _, relative := range relatives {
			for objID, target := range stm.skillTargets {
				if targetSet.Contains(objID) {
					continue
				}
				target2, ok := filter.checkMatch(sk, skillOwner, relative, target, triggerCxt, stm.situation, cacheTargets)
				if ok {
					targetSet.Add(objID)
					ts = append(ts, target2)
				}
			}
		}

	} else {

		for _, relative := range relatives {
			for _, target := range allTargets {
				if target == nil {
					continue
				}
				objID := target.getObjID()
				if targetSet.Contains(objID) {
					continue
				}
				target2, ok := filter.checkMatch(sk, skillOwner, relative, target, triggerCxt, stm.situation, cacheTargets)
				if ok {
					targetSet.Add(objID)
					ts = append(ts, target2)
				}
			}
		}

	}

	if filter.data.Random > 0 && len(ts) > filter.data.Random {
		utils.Shuffle(ts)
		ts = ts[:filter.data.Random]
	}

	retTargets := make([]iTarget, len(ts))
	for i, t := range ts {
		retTargets[i] = t.(iTarget)
	}

	if cacheTargets != nil {
		cacheTargets[targetID] = retTargets
	}
	return retTargets
}
