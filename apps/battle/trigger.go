package main

import (
	"kinger/proto/pb"
	"github.com/gogo/protobuf/proto"
	"kinger/gopuppy/common"
	"kinger/gopuppy/attribute"
	"fmt"
	//"kinger/gopuppy/common/glog"
)

type triggerResult struct {
	cantTurnCards common.IntSet   // 不能被翻面的牌

	modifyValues map[int]map[int]int   //  上次改变了多少点数 map[targetID]map[objID]value
	cantAddValTargets common.IntSet  // 哪些目标不能加点
	cantSubValTargets common.IntSet  // 哪些目标不能减点
	additionalAddValTargets map[int]int  // 哪些目标额外加多少点
	additionalSubValTargets map[int]int  // 哪些目标额外减多少点

	forbidAddSkills map[int][]int32  // 不能获得哪些技能

	drawFighters map[common.UUid]*fighter  // 被改变的抓牌目标

	forbidMoveCards common.IntSet  // 不能移动的卡
	moveCards []iTarget     // 刚移动的卡
}

func (tr *triggerResult) isCanTurn(card *fightCard) bool {
	if tr.cantTurnCards == nil {
		return true
	}
	return !tr.cantTurnCards.Contains(card.getObjID())
}

func (tr *triggerResult) addCantTurnCard(card *fightCard) {
	if tr.cantTurnCards == nil {
		tr.cantTurnCards = common.IntSet{}
	}
	tr.cantTurnCards.Add(card.getObjID())
}

func (tr *triggerResult) getCardModifyValue(card *fightCard, value int) int {
	if value == 0 {
		return 0
	}
	objID := card.getObjID()
	if value > 0 {
		if tr.cantAddValTargets != nil && tr.cantAddValTargets.Contains(objID) {
			return 0
		}
		if tr.additionalAddValTargets != nil {
			if n, ok := tr.additionalAddValTargets[objID]; ok {
				return value + n
			}
		}
	} else {
		if tr.cantSubValTargets != nil && tr.cantSubValTargets.Contains(objID) {
			return 0
		}
		if tr.additionalSubValTargets != nil {
			if n, ok := tr.additionalSubValTargets[objID]; ok {
				return value - n
			}
		}
	}

	return value
}

func (tr *triggerResult) addAdditionalAddVal(t iTarget, val int) {
	if tr.additionalAddValTargets == nil {
		tr.additionalAddValTargets = map[int]int{}
	}
	n := tr.additionalAddValTargets[t.getObjID()]
	tr.additionalAddValTargets[t.getObjID()] = n + val
}

func (tr *triggerResult) addAdditionalSubVal(t iTarget, val int) {
	if tr.additionalSubValTargets == nil {
		tr.additionalSubValTargets = map[int]int{}
	}
	n := tr.additionalSubValTargets[t.getObjID()]
	tr.additionalSubValTargets[t.getObjID()] = n + val
}

func (tr *triggerResult) setCantSubValTarget(t iTarget) {
	if tr.cantSubValTargets == nil {
		tr.cantSubValTargets = common.IntSet{}
	}
	tr.cantSubValTargets.Add(t.getObjID())
}

func (tr *triggerResult) setCantAddValTarget(t iTarget) {
	if tr.cantAddValTargets == nil {
		tr.cantAddValTargets = common.IntSet{}
	}
	tr.cantAddValTargets.Add(t.getObjID())
}

func (tr *triggerResult) setCardModifyValue(targetID, objID, value int) {
	if tr.modifyValues == nil {
		tr.modifyValues = map[int]map[int]int{}
	}
	obj2Value, ok := tr.modifyValues[targetID]
	if !ok {
		obj2Value = map[int]int{}
		tr.modifyValues[targetID] = obj2Value
	}
	obj2Value[objID] = value
}

func (tr *triggerResult) getCardLastModifyValue(targetID, objID int) int {
	if tr.modifyValues == nil {
		return 0
	}
	obj2Value, ok := tr.modifyValues[targetID]
	if !ok {
		return 0
	}
	return obj2Value[objID]
}

func (tr *triggerResult) canAddSkill(objID int, skillID int32) bool {
	if tr.forbidAddSkills == nil {
		return true
	}

	skillIDs, ok := tr.forbidAddSkills[objID]
	if ok {
		for _, id := range skillIDs {
			if id < 0 || id == skillID {
				return false
			}
		}
	}
	return true
}

func (tr *triggerResult) addForbidSkill(objID int, skillID int32) {
	if tr.forbidAddSkills == nil {
		tr.forbidAddSkills = map[int][]int32{}
	}
	sks := tr.forbidAddSkills[objID]
	tr.forbidAddSkills[objID] = append(sks, skillID)
}

func (tr *triggerResult) getDrawFighter(f *fighter) *fighter {
	if tr.drawFighters == nil {
		return f
	}
	f2, ok := tr.drawFighters[f.getUid()]
	if ok {
		return f2
	} else {
		return f
	}
}

func (tr *triggerResult) setDrawFighter(f1 *fighter, f2 *fighter) {
	if tr.drawFighters == nil {
		tr.drawFighters = map[common.UUid]*fighter{}
	}
	tr.drawFighters[f1.getUid()] = f2
}

func (tr *triggerResult) canMove(card *fightCard) bool {
	if tr.forbidMoveCards == nil {
		return true
	}
	return !tr.forbidMoveCards.Contains(card.getObjID())
}

func (tr *triggerResult) isMoveCard(t iTarget) bool {
	for _, c := range tr.moveCards {
		if c == t {
			return true
		}
	}
	return false
}

func (tr *triggerResult) addCantMoveCard(card *fightCard) {
	if tr.forbidMoveCards == nil {
		tr.forbidMoveCards = common.IntSet{}
	}
	tr.forbidMoveCards.Add(card.getObjID())
}

type turnCaster struct {
	c iCaster
	preTurnSit int	 // 翻面前的sit
}

func newTurnCaster(c iCaster, sit int) *turnCaster {
	return &turnCaster{
		c:  c,
		preTurnSit: sit,
	}
}

func (tc *turnCaster) String() string {
	return fmt.Sprintf("%s", tc.c)
}

type attackOutcomeRecord struct {
	records map[int][]int
}

func (aor *attackOutcomeRecord) forEachRecord(callback func(oppObjID, bat int)) {
	if aor.records == nil {
		return
	}
	for oppObjID, bats := range aor.records {
		callback(oppObjID, bats[len(bats) - 1])
	}
}

func (aor *attackOutcomeRecord) addRecord(oppObjID, bat int) {
	if aor.records == nil {
		aor.records = map[int][]int{}
	}
	old := aor.records[oppObjID]
	aor.records[oppObjID] = append(old, bat)
}

func (aor *attackOutcomeRecord) getLastRecord(oppObjID int) int {
	if aor.records == nil {
		return 0
	}
	bats := aor.records[oppObjID]
	return bats[len(bats) - 1]
}

func (aor *attackOutcomeRecord) isOnceWin(oppObjID int) bool {
	if aor.records == nil {
		return false
	}
	bats := aor.records[oppObjID]
	for _, b := range bats {
		if b == bWin {
			return true
		}
	}
	return false
}

type triggerContext struct {
	triggerType int
	// fuck this 这个是触发技能时技能拥有者的阵营，技能结算时技能拥有者阵营会变，但技能填表时是按触发时阵营去填的
	skillOwnerTriggerSits map[int]int
	enterBattleCard *fightCard

	attackCxt *attackContext
	attackCard   *fightCard   // 进攻卡
	beAttackCards []*fightCard  // 受击卡
	attackResult map[int]map[int]int  // 比点大小结果  map[objID]map[objID]bat
	attackOutcome map[int]*attackOutcomeRecord // 比点胜负结果  map[objID]*attackOutcomeRecord

	turner    *turnCaster         // 翻面者
	beTurners map[int]*turnCaster // 被翻面者
	notPreBeTurnner common.IntSet
	cantTurnCards []iTarget   // 不能被翻面的牌

	destoryer    iTarget   // 根除者
	beDestoryers []iTarget // 被根除者

	preMoveCards []iTarget  // 准备移动的卡
	moveCards []iTarget     // 刚移动的卡

	drawCards []iTarget  // 抓上来的卡
	preAddSkillIDs []int32  // 将要获得的技能

	triggerTargetSit map[int]int  // 目标在技能触发时的阵营 map[objID]sit
	actionTargetSit map[int]int  // 目标在技能结算前的阵营  map[objID]sit
	targetID2ActTargets map[int][]iTarget  // 这次技能触发的行为对象

	modifyValueTargets []iTarget  // 将要改变点数的target
	modifyValue int   //  将要改变多少点数

	returnCards []iTarget  // 回手的牌

	surrenderor *fighter  // 投降者
	beSurrenderor *fighter  // 被投降者
}

func (tc *triggerContext) isCanTurn(card iTarget) bool {
	objID := card.getObjID()
	for _, t := range tc.cantTurnCards {
		if t.getObjID() == objID {
			return false
		}
	}
	return true
}

func (tc *triggerContext) addCantTurnCard(card *fightCard) {
	objID := card.getObjID()
	for _, t := range tc.cantTurnCards {
		if t.getObjID() == objID {
			return
		}
	}
	tc.cantTurnCards = append(tc.cantTurnCards, card)
}

func (tc *triggerContext) setAttackResult(card *fightCard, oppCard *fightCard, bat int) {
	objID := card.getObjID()
	oppObjID := oppCard.getObjID()
	if tc.attackResult == nil {
		tc.attackResult = map[int]map[int]int{}
	}

	myAttackResult, ok := tc.attackResult[objID]
	if !ok {
		myAttackResult = map[int]int{}
		tc.attackResult[objID] = myAttackResult
	}

	if bat == 0 {
		delete(myAttackResult, oppObjID)
	} else {
		myAttackResult[oppObjID] = bat
	}
}

func (tc *triggerContext) setAttackOutcome(card *fightCard, oppCard *fightCard, bat int) {
	objID := card.getObjID()
	oppObjID := oppCard.getObjID()
	if tc.attackOutcome == nil {
		tc.attackOutcome = map[int]*attackOutcomeRecord{}
	}

	myAttackOutcome, ok := tc.attackOutcome[objID]
	if !ok {
		myAttackOutcome = &attackOutcomeRecord{}
		tc.attackOutcome[objID] = myAttackOutcome
	}

	myAttackOutcome.addRecord(oppObjID, bat)
}

func (tc *triggerContext) getAttackResult(card *fightCard, oppCard *fightCard) int {
	if tc.attackResult == nil {
		return 0
	}
	if myAttackResult, ok := tc.attackResult[card.getObjID()]; !ok {
		return 0
	} else {
		return myAttackResult[oppCard.getObjID()]
	}
}

func (tc *triggerContext) getAttackOutcome(card *fightCard, oppCard *fightCard) int {
	if tc.attackOutcome == nil {
		return 0
	}
	if myAttackOutcome, ok := tc.attackOutcome[card.getObjID()]; !ok {
		return 0
	} else {
		return myAttackOutcome.getLastRecord(oppCard.getObjID())
	}
}

func (tc *triggerContext) isOnceAttackWin(card *fightCard, oppCard *fightCard) bool {
	if tc.attackOutcome == nil {
		return false
	}

	if myAttackOutcome, ok := tc.attackOutcome[card.getObjID()]; !ok {
		return false
	} else {
		return myAttackOutcome.isOnceWin(oppCard.getObjID())
	}
}

func (tc *triggerContext) getAttackBatInfo(t iTarget) common.IntSet {
	info := common.IntSet{}
	if tc.attackResult != nil {
		if myAttackResult, ok := tc.attackResult[t.getObjID()]; ok {
			for _, b := range myAttackResult {
				info.Add(b)
			}
		}
	}

	if tc.attackOutcome != nil {
		if myAttackOutcome, ok := tc.attackOutcome[t.getObjID()]; ok {
			myAttackOutcome.forEachRecord(func(oppObjID, bat int) {
				info.Add(bat)
			})
		}
	}

	return info
}

func (tc *triggerContext) getAttackOutcomeRecord(card *fightCard) *attackOutcomeRecord {
	if tc.attackOutcome == nil {
		return nil
	}
	if myAttackOutcome, ok := tc.attackOutcome[card.getObjID()]; !ok {
		return nil
	} else {
		return myAttackOutcome
	}
}

func (tc *triggerContext) delPreBeTurnner(card *fightCard) {
	if tc.triggerType != preBeTurnTrigger {
		return
	}
	if tc.notPreBeTurnner == nil {
		tc.notPreBeTurnner = common.IntSet{}
	}
	tc.notPreBeTurnner.Add(card.getObjID())
}

func (tc *triggerContext) addBeTurners(card *fightCard, preTurnSit int) {
	if tc.beTurners == nil {
		tc.beTurners = map[int]*turnCaster{}
	}
	tc.beTurners[card.getObjID()] = newTurnCaster(card, preTurnSit)
}

func (tc *triggerContext) setTurner(c iCaster, preTurnSit int) {
	tc.turner = newTurnCaster(c, preTurnSit)
}

func (tc *triggerContext) getPreTurnSit(t iTarget) int {
	if tc.turner != nil && tc.turner.c.getObjID() == t.getObjID() {
		return tc.turner.preTurnSit
	}

	for _, tn := range tc.beTurners {
		if tn.c.getObjID() == t.getObjID() {
			return tn.preTurnSit
		}
	}

	return t.getSit()
}

func (tc *triggerContext) isTurnner(t iTarget) bool {
	return tc.turner != nil && tc.turner.c.getObjID() == t.getObjID()
}

func (tc *triggerContext) isBeTurnner(t iTarget) bool {
	if tc.beTurners == nil {
		return false
	}
	for _, tn := range tc.beTurners {
		if tn.c.getObjID() == t.getObjID() {
			return true
		}
	}
	return false
}

func (tc *triggerContext) isPreBeTurnner(t iTarget) bool {
	if tc.notPreBeTurnner != nil && tc.notPreBeTurnner.Contains(t.getObjID()) {
		return false
	}
	return tc.isBeTurnner(t)
}

func (tc *triggerContext) isDestoryer(t iTarget) bool {
	return tc.destoryer == t
}

func (tc *triggerContext) isBeDestoryer(t iTarget) bool {
	for _, c := range tc.beDestoryers {
		if c == t {
			return true
		}
	}
	return false
}

func (tc *triggerContext) isPreMoveCard(t iTarget) bool {
	for _, c := range tc.preMoveCards {
		if c == t {
			return true
		}
	}
	return false
}

func (tc *triggerContext) isMoveCard(t iTarget) bool {
	for _, c := range tc.moveCards {
		if c == t {
			return true
		}
	}
	return false
}

func (tc *triggerContext) isDrawCard(t iTarget) bool {
	for _, c := range tc.drawCards {
		if c == t {
			return true
		}
	}
	return false
}

func (tc *triggerContext) isEnterBattleCard(card *fightCard) bool {
	return card != nil && tc.enterBattleCard == card
}

func (tc *triggerContext) getAttackCard() *fightCard {
	return tc.attackCard
}

func (tc *triggerContext) isAttackCard(t iTarget) bool {
	return tc.attackCard != nil && tc.attackCard.getObjID() == t.getObjID()
}

func (tc *triggerContext) isBeAttackCard(t iTarget) bool {
	if tc.beAttackCards == nil {
		return false
	}

	for _, c := range tc.beAttackCards {
		if c.getObjID() == t.getObjID() {
			return true
		}
	}

	return false
}

func (tc *triggerContext) setTriggerTargetSit(situation *battleSituation) {
	if tc.triggerTargetSit != nil {
		return
	}
	tc.triggerTargetSit = map[int]int{}

	for _, objID := range situation.grids {
		t := situation.getTargetMgr().getTarget(objID)
		if t.getType() != stEmptyGrid {
			tc.triggerTargetSit[objID] = t.getSit()
		}
	}
}

func (tc *triggerContext) getTriggerTargetSit(t iTarget) int {
	if tc.triggerTargetSit == nil {
		return t.getSit()
	}
	if sit, ok := tc.triggerTargetSit[t.getObjID()]; ok {
		return sit
	} else {
		return t.getSit()
	}
}

func (tc *triggerContext) addActionTargetSit(t iTarget) {
	if tc.actionTargetSit == nil {
		tc.actionTargetSit = map[int]int{}
	}
	objID := t.getObjID()
	if _, ok := tc.actionTargetSit[objID]; !ok {
		tc.actionTargetSit[objID] = t.getSit()
	}
}

func (tc *triggerContext) getActionTargetSit(t iTarget) int {
	if tc.actionTargetSit == nil {
		return t.getSit()
	}
	if sit, ok := tc.actionTargetSit[t.getObjID()]; ok {
		return sit
	} else {
		return t.getSit()
	}
}

func (tc *triggerContext) addActionTargets(targetID int, targets []iTarget) {
	if tc.targetID2ActTargets == nil {
		tc.targetID2ActTargets = map[int][]iTarget{}
	}
	tc.targetID2ActTargets[targetID] = targets
}

func (tc *triggerContext) copyActionTargets() map[int][]iTarget {
	cpy := map[int][]iTarget{}
	if tc.targetID2ActTargets == nil {
		return cpy
	}

	for targetID, targets := range tc.targetID2ActTargets {
		cpy[targetID] = targets
	}
	return cpy
}

func (tc *triggerContext) delActionTargets() {
	tc.targetID2ActTargets = nil
}

func (tc *triggerContext) getActonTargetID(t iTarget) int {
	for targetID, ts := range tc.targetID2ActTargets {
		for _, t2 := range ts {
			if t2 == t {
				return targetID
			}
		}
	}
	return 0
}

func (tc *triggerContext) isReturnCard(t iTarget) bool {
	for _, t2 := range tc.returnCards {
		if t2.getObjID() == t.getObjID() {
			return true
		}
	}
	return false
}

type clientAction struct {
	actID pb.ClientAction_ActionID
	actMsg proto.Marshaler
}

func (act *clientAction) packMsg() *pb.ClientAction {
	msg := &pb.ClientAction{
		ID: act.actID,
	}
	msg.Data, _ = act.actMsg.Marshal()
	return msg
}

type triggerSkill struct {
	*skill
	triggerType int
}

func (ts *triggerSkill) String() string {
	return fmt.Sprintf("[caster=%s, skillID=%d]", ts.getOwner(), ts.getID())
}

type skillTriggerMgr struct {
	casters []int   // 施法者objID, 按进场顺序排序
	situation *battleSituation
}

func newSkillTriggerMgr(situation *battleSituation) *skillTriggerMgr {
	return &skillTriggerMgr{
		situation: situation,
	}
}

func (stm *skillTriggerMgr) copy(situation *battleSituation) *skillTriggerMgr {
	m := *stm
	cpy := &m
	cpy.situation = situation
	cpy.casters = make([]int, len(stm.casters))
	copy(cpy.casters, stm.casters)
	return cpy
}

func (stm *skillTriggerMgr) packAttr() *attribute.ListAttr {
	attr := attribute.NewListAttr()
	for _, objID := range stm.casters {
		attr.AppendInt(objID)
	}
	return attr
}

func (stm *skillTriggerMgr) restoredFromAttr(attr *attribute.ListAttr, situation *battleSituation) {
	stm.situation = situation
	attr.ForEachIndex(func(index int) bool {
		stm.casters = append(stm.casters, attr.GetInt(index))
		return true
	})
}

func (stm *skillTriggerMgr) boutEnd() []*clientAction {
	var actions []*clientAction
	for _, objID := range stm.casters {
		t := stm.situation.getTargetMgr().getTarget(objID)
		if t == nil {
			continue
		}
		c := t.(iCaster)
		actions = append(actions, c.boutEnd()...)
	}
	return actions
}

func (stm *skillTriggerMgr) casterEnterBattle(caster iCaster) []*clientAction {
	stm.casters = append(stm.casters, caster.getObjID())
	return caster.enterBattle()
}

func (stm *skillTriggerMgr) delCaster(objID int) {
	for i, objID2 := range stm.casters {
		if objID2 == objID {
			stm.casters = append(stm.casters[:i], stm.casters[i+1:]...)
			return
		}
	}
}

func (stm *skillTriggerMgr) getCanTriggerSkill(triggerTargets map[int][]iTarget, triggerCxt *triggerContext) ([]*triggerSkill,
	[]*clientAction) {

	// 结算顺序，1.触发对象，2.己方卡牌，3.对方卡牌。有多个时按出场顺序
	var triggerCard []iCaster
	var mysideCard []iCaster
	var enemyCard []iCaster
	curBoutSit := stm.situation.getCurBoutFighter().getSit()
L1:
	for _, objID := range stm.casters {
		t := stm.situation.getTargetMgr().getTarget(objID)
		if t == nil {
			continue
		}
		c, ok := t.(iCaster)
		if !ok {
			continue
		}

		for _, targets := range triggerTargets {
			for _, t := range targets {
				if t.getObjID() == c.getObjID() {
					triggerCard = append(triggerCard, c)
					continue L1
				}
			}
		}

		if c.getSit() == curBoutSit {
			mysideCard = append(mysideCard, c)
		} else {
			enemyCard = append(enemyCard, c)
		}
	}

	triggerCard = append(triggerCard, mysideCard...)
	triggerCard = append(triggerCard, enemyCard...)

	var actions []*clientAction
	var skills []*triggerSkill
	for _, c := range triggerCard {
		cacheTargets := map[int][]iTarget{}
		sks := c.getAllIgnoreLostCanTriggerSkills()
	    	for _, sk := range sks {
			for triggerType, targets := range triggerTargets {
				if !c.canTriggerSkill(sk) {
					continue
				}

				can, isTargetTriggerTimesLimit, acts := sk.canTrigger(triggerType, targets, triggerCxt, cacheTargets)
				actions = append(actions, acts...)

				if can {
					skills = append(skills, &triggerSkill{
						skill: sk,
						triggerType: triggerType,
					})
					break
				} else if card, ok := c.(*fightCard); ok && triggerCxt.isEnterBattleCard(card) &&
					triggerType == enterBattleTrigger && sk.getData().TriggerOpp == enterBattleTrigger &&
					sk.getData().TriggerObj == 1 {
					// 进场时机，没成功触发进场技，删掉
					actions = append(actions, card.delSkill(sk)...)
					break
				} else if isTargetTriggerTimesLimit && sk.getData().TriggerTimes > 0 && 
					  sk.targetTriggerTimes >= sk.getData().TriggerTimes {
					// target触发次数到限制了，且这个时机还不能触发，以后没机会了
					actions = append(actions, c.delSkill(sk)...)
					break
				}
			}

		}
	}

	// 根据优先级排序
	scnt := len(skills)
	if scnt > 1 {
		for i := scnt - 1; i > 0; i-- {
			for j := 0; j < i; j++ {
				if skills[j].getData().Priority > skills[j+1].getData().Priority {
					skills[j], skills[j+1] = skills[j+1], skills[j]
				}
			}
		}
	}

	return skills, actions
}

func (stm *skillTriggerMgr) trigger(triggerTargets map[int][]iTarget, cxt *triggerContext) ([]*clientAction,
	*triggerResult, bool) {

	var actions []*clientAction
	result := &triggerResult{}
	if len(triggerTargets) <= 0 {
		return actions, result, false
	}

	var needTalk bool
	triggerTargets2 := map[int][]iTarget{}
	for triggerType, targets := range triggerTargets {
		triggerTargets2[triggerType] = append([]iTarget{}, targets...)
	}

	skills, acts := stm.getCanTriggerSkill(triggerTargets, cxt)
	actions = append(actions, acts...)
	//glog.Infof("skillTriggerMgr trigger triggerTargets=%v, skills=%v", triggerTargets, skills)

	for _, sk := range skills {
		//glog.Infof("skillTriggerMgr trigger 11111111111 skill=%s", sk)
		owner := sk.getOwner()
		if owner == nil || !owner.canTriggerSkill(sk.skill) {
			continue
		}
		//glog.Infof("skillTriggerMgr trigger 222222222222 skill=%s", sk)
		var skillAct *pb.SkillAct
		var acts []*clientAction
		var needTalk2 bool
		triggerType := sk.triggerType
		targets := triggerTargets2[triggerType]
		skData := sk.getData()
		if len(targets) <= 0 && skData.TriggerObj != 0 {
			continue
		} else if len(targets) != len(triggerTargets[triggerType]) {
			// 触发对象改变了，再判断下是否满足条件
			if !sk.isTargetTriggerObj(triggerType, targets, cxt, nil) {
				continue
			}
		}

		//glog.Infof("skillTriggerMgr trigger 3333333333333 skill=%s", sk)

		skillAct, acts, triggerTargets2[triggerType], needTalk2 = sk.trigger(triggerType, targets, cxt, result)
		if skillAct != nil {
			actions = append(actions, &clientAction{
				actID: pb.ClientAction_Skill,
				actMsg: skillAct,
			})
		}
		actions = append(actions, acts...)

		if (skData.TriggerOpp == enterBattleTrigger && skData.TriggerObj == 1) ||
			(skData.TriggerTimes > 0 && sk.targetTriggerTimes >= skData.TriggerTimes) {
			skOwner := sk.getOwner()
			if skOwner != nil {
				actions = append(actions, skOwner.delSkill(sk.skill)...)
			}
		}

		if needTalk2 {
			needTalk = true
		}
	}

	cxt.triggerTargetSit = nil
	cxt.actionTargetSit = nil

	return actions, result, needTalk
}
