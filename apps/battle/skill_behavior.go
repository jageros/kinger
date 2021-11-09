package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"math"
	"strconv"
	"strings"
)

var (
	allSkillBehaviors = map[int32]*skillBehavior{}
	drawnGameAction   iSkillAction
)

type targetAmountCondition struct {
	operator string
	right    string
	left     string
}

func newTargetAmountCondition(condition string) *targetAmountCondition {
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
		return nil
	}

	return &targetAmountCondition{
		operator: op,
		left:     strings.TrimSpace(condition[:i]),
		right:    strings.TrimSpace(condition[i+len(op):]),
	}
}

func (c *targetAmountCondition) String() string {
	return fmt.Sprintf("[targetAmountCondition %s %s]", c.operator, c.right)
}

func (c *targetAmountCondition) check(targetCountX int, targetCountY int, targetCountZ int) bool {
	right := 0
	switch c.right {
	case "x":
		right = targetCountX
	case "y":
		right = targetCountY
	case "z":
		right = targetCountZ
	default:
		right, _ = strconv.Atoi(c.right)
	}

	targetCount := 0
	switch c.left {
	case "x":
		targetCount = targetCountX
	case "y":
		targetCount = targetCountY
	default:
		targetCount = targetCountZ
	}

	switch c.operator {
	case ">":
		return targetCount > right
	case ">=":
		return targetCount >= right
	case "<":
		return targetCount < right
	case "<=":
		return targetCount <= right
	case "==":
		return targetCount == right
	default:
		return false
	}
}

type iSkillAction interface {
	getTargetIDs() []int
	invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget, variableAmount int,
		triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs, acts []*clientAction,
		moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget)
}

type baseSkillAction struct {
	targetIDs []int
}

func (at *baseSkillAction) getTargetIDs() []int {
	return at.targetIDs
}

type valueAction struct {
	baseSkillAction
	value       string
	modifyType  int // 1:minValue  2:changeToMax
	mcMovieID   string
	textMovieID int
}

func newValueAction(skillData *gamedata.Skill, valueInfo []string, movieInfo []string) *valueAction {
	var targetIDs []int
	var modifyType, textMovieID int
	var value, mcMovieID string
	switch len(valueInfo) {
	case 3:
		targetID, err := strconv.Atoi(valueInfo[1])
		if err != nil {
			return nil
		}
		targetIDs = []int{targetID}
		value = valueInfo[2]
		modifyType, _ = strconv.Atoi(valueInfo[0])
	case 2:
		targetID, err := strconv.Atoi(valueInfo[0])
		if err != nil {
			return nil
		}
		targetIDs = []int{targetID}
		value = valueInfo[1]
	case 1:
		targetIDs = skillData.TargetAct
		value = valueInfo[0]
	default:
		return nil
	}

	switch len(movieInfo) {
	case 0:
	case 1:
		mcMovieID = movieInfo[0]
	default:
		mcMovieID = movieInfo[0]
		textMovieID, _ = strconv.Atoi(movieInfo[1])
	}

	return &valueAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		value:           value,
		modifyType:      modifyType,
		mcMovieID:       mcMovieID,
		textMovieID:     textMovieID,
	}
}

func (at *valueAction) getValue(variableAmount int) int {
	if at.value == "" {
		return 0
	}

	_value := strings.Replace(at.value, "x", strconv.Itoa(variableAmount), -1)
	n, err := strconv.Atoi(_value)
	if err == nil {
		return n
	} else {
		exps, err := parseExp(_value)
		if err != nil {
			glog.Errorf("parseExp err %s", err)
			return 0
		}
		exps2 := pre2stuf(exps)
		return caculate(exps2)
	}
}

func (at *valueAction) preModifyValue(variableAmount int, actionTargets []iTarget, situation *battleSituation) (
	acts []*clientAction, triggerRes *triggerResult, value int) {

	value = at.getValue(variableAmount)
	if at.modifyType == mvtAllBecome {
		value = 1
	}
	if value == 0 {
		return
	}

	preModifyValueTriggerType := preAddValueTrigger
	if value < 0 {
		preModifyValueTriggerType = preSubValueTrigger
	}
	acts, triggerRes, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{preModifyValueTriggerType: actionTargets}, &triggerContext{
		triggerType:        preModifyValueTriggerType,
		modifyValueTargets: actionTargets,
		modifyValue:        value,
	})
	return
}

func (at *valueAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var value int
	var triggerRes *triggerResult
	skillActs, triggerRes, value = at.preModifyValue(variableAmount, actionTargets, situation)
	if value == 0 {
		return
	}

	var targets []int32
	target2MaxValue := map[int]int{}
	value2Targets := map[int][]int32{}
	var afterSubValueTargets, afterAddValueTargets []iTarget

	for _, t := range actionTargets {
		if c, ok := t.(*fightCard); ok {
			value2 := triggerRes.getCardModifyValue(c, value)
			if value2 == 0 {
				continue
			}

			minValue, maxValue := c.modifyValue(value2, at.modifyType, value2-value)
			target2MaxValue[c.getObjID()] = maxValue

			if at.modifyType == mvtAllBecome {
				if maxValue > minValue {
					afterAddValueTargets = append(afterAddValueTargets, c)
				}
			} else if value2 > 0 {
				afterAddValueTargets = append(afterAddValueTargets, c)
			} else {
				afterSubValueTargets = append(afterSubValueTargets, c)
			}

			objID := int32(c.getObjID())
			targets = append(targets, objID)
			value2Targets[value2] = append(value2Targets[value2], objID)
		} else {
			glog.Errorf("valueAction target wrong type, %d", t.getType())
		}
	}

	if len(targets) <= 0 {
		return
	}

	modifyValueAct := &pb.ModifyValueAct{
		McMovieID:   at.mcMovieID,
		TextMovieID: int32(at.textMovieID),
		ModifyType:  int32(at.modifyType),
		OwnerObjID:  int32(skOwner.getObjID()),
	}
	if at.modifyType != 2 {

		for v, ts := range value2Targets {
			modifyValueAct.Items = append(modifyValueAct.Items, &pb.ModifyValueActItem{
				Value:   int32(v),
				Targets: ts,
			})

			for _, objID := range ts {
				for _, targetID := range at.targetIDs {
					result.setCardModifyValue(targetID, int(objID), v)
				}
			}
		}

	} else {

		for _, objID := range targets {
			v, ok := target2MaxValue[int(objID)]
			if ok {
				modifyValueAct.Items = append(modifyValueAct.Items, &pb.ModifyValueActItem{
					Value:   int32(v),
					Targets: []int32{objID},
				})
			}
		}
	}

	if len(afterAddValueTargets) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{afterAddValueTrigger: afterAddValueTargets}, &triggerContext{
			triggerType: afterAddValueTrigger,
		})
		acts = append(acts, acts2...)
	}

	if len(afterSubValueTargets) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{afterSubValueTrigger: afterSubValueTargets}, &triggerContext{
			triggerType: afterSubValueTrigger,
		})
		acts = append(acts, acts2...)
	}

	action := &clientAction{
		actID:  pb.ClientAction_ModifyValue,
		actMsg: modifyValueAct,
	}

	if sk.getData().TriggerOpp == afterMoveTrigger {
		afterMoveActs = skillActs
		afterMoveActs = append(afterMoveActs, action)
		skillActs = []*clientAction{}
	} else {
		skillActs = append(skillActs, action)
	}
	return
}

type textMovieAction struct {
	baseSkillAction
	movieID  int
	playType int
	boutTime int
}

func newTextMovieAction(skillData *gamedata.Skill, textMovieInfo []int) *textMovieAction {
	var targetIDs []int
	var movieID int
	var playerType int
	var boutTime int
	switch len(textMovieInfo) {
	case 2:
		targetIDs = skillData.TargetAct
		movieID = textMovieInfo[0]
		playerType = textMovieInfo[1]
	case 3:
		targetIDs = []int{textMovieInfo[0]}
		movieID = textMovieInfo[1]
		playerType = textMovieInfo[2]
	case 4:
		targetIDs = []int{textMovieInfo[0]}
		movieID = textMovieInfo[1]
		playerType = textMovieInfo[2]
		boutTime = textMovieInfo[3]
	}

	return &textMovieAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		movieID:         movieID,
		playType:        playerType,
		boutTime:        boutTime,
	}
}

func (at *textMovieAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var actions []*clientAction
	value2Act := map[int]*clientAction{}

	for _, t := range actionTargets {
		objID := t.getObjID()
		value := result.getCardLastModifyValue(at.targetIDs[0], objID)
		act, ok := value2Act[value]
		if !ok {
			act = &clientAction{
				actID: pb.ClientAction_TextMovie,
				actMsg: &pb.TextMovieAct{
					MovieID:     int32(at.movieID),
					PlayType:    int32(at.playType),
					TargetCount: int32(variableAmount),
					Value:       int32(value),
					OwnerObjID:  int32(skOwner.getObjID()),
				},
			}
			value2Act[value] = act
			actions = append(actions, act)
		}

		textMovieAct := act.actMsg.(*pb.TextMovieAct)
		textMovieAct.Targets = append(textMovieAct.Targets, int32(t.getObjID()))
	}

	if sk.getData().TriggerOpp == afterMoveTrigger {
		afterMoveActs = actions
	} else {
		skillActs = actions
	}
	return
}

type mcMovieAction struct {
	baseSkillAction
	movieID  string
	playType int
	boutTime int
}

func newMcMovieAction(skillData *gamedata.Skill, movieInfo []string) *mcMovieAction {
	var targetIDs []int
	targetID, err := strconv.Atoi(movieInfo[0])
	var movieID string
	var playerType int
	var boutTime int

	if err != nil {
		targetIDs = skillData.TargetAct
		movieID = movieInfo[0]
		playerType, _ = strconv.Atoi(movieInfo[1])
		if len(movieInfo) == 3 {
			boutTime, _ = strconv.Atoi(movieInfo[2])
		}
	} else {
		targetIDs = []int{targetID}
		movieID = movieInfo[1]
		playerType, _ = strconv.Atoi(movieInfo[2])
		if len(movieInfo) == 4 {
			boutTime, _ = strconv.Atoi(movieInfo[3])
		}
	}

	return &mcMovieAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		movieID:         movieID,
		playType:        playerType,
		boutTime:        boutTime,
	}
}

func (at *mcMovieAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var targets []int32
	var ownerObjID int
	if skOwner != nil {
		ownerObjID = skOwner.getObjID()
	}
	for _, t := range actionTargets {
		targets = append(targets, int32(t.getObjID()))
		if at.playType == -1 || at.playType == -2 {
			t.addEffect(ownerObjID, at.movieID, at.playType, at.boutTime)
		} else if at.playType == 0 {
			t.delEffect(at.movieID)
		}
	}
	skillActs = []*clientAction{&clientAction{
		actID: pb.ClientAction_Movie,
		actMsg: &pb.MovieAct{
			MovieID:    at.movieID,
			PlayType:   int32(at.playType),
			Targets:    targets,
			OwnerObjID: int32(ownerObjID),
		},
	}}
	return
}

type turnOverAction struct {
	baseSkillAction
	turn int
}

func newTurnOverActions(skillData *gamedata.Skill, turnInfo [][]int) []iSkillAction {
	var targetIDs []int
	var turn int
	var acts []iSkillAction
	for _, info := range turnInfo {
		switch len(info) {
		case 1:
			targetIDs = skillData.TargetAct
			turn = info[0]
		case 2:
			targetIDs = []int{info[0]}
			turn = info[1]
		default:
			glog.Errorf("newTurnOverActions error skillID=%d %v", skillData.ID, info)
			continue
		}

		acts = append(acts, &turnOverAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
			turn:            turn,
		})
	}
	return acts
}

func (at *turnOverAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	var preTurnOvers []iTarget
	triggerMgr := situation.getTriggerMgr()
	isAllHandCard := true
	for _, t := range actionTargets {
		c, ok := t.(*fightCard)
		if !ok {
			glog.Errorf("turnOverAction target wrong type, %d", t.getType())
			continue
		}

		if at.turn == cantTurn {
			triggerCxt.addCantTurnCard(c)
			triggerCxt.delPreBeTurnner(c)
			if sk.getData().TriggerOpp == preBeTurnTrigger {
				// 强撸，改变triggerTargets，让后面的技能不触发
				for i, t2 := range triggerTargets {
					if t2.getObjID() == t.getObjID() {
						triggerTargets = append(triggerTargets[:i], triggerTargets[i+1:]...)
						break
					}
				}
			}
		} else {
			oldSit := triggerCxt.getActionTargetSit(t)
			if oldSit != c.getSit() {
				// 触发到结算期间已经被翻过面
				continue
			}
			if c.isInBattle() {
				isAllHandCard = false
			}
			preTurnOvers = append(preTurnOvers, c)
		}
	}

	triggerTargets2 = triggerTargets
	if at.turn == cantTurn {
		return
	}

	// --------------------------- 翻面前 -------------------------------
	triggerCxt2 := &triggerContext{
		triggerType: preBeTurnTrigger,
	}
	triggerCxt2.setTurner(skOwner, triggerCxt.getPreTurnSit(skOwner))
	//var triggerRes *triggerResult
	preTurnSit := map[int]int{}
	if len(preTurnOvers) > 0 && !isAllHandCard {
		for _, t := range preTurnOvers {
			var cardPreTurnSit int
			if triggerCxt.isTurnner(t) {
				cardPreTurnSit = triggerCxt.getPreTurnSit(t)
			} else {
				cardPreTurnSit = t.getSit()
			}
			preTurnSit[t.getObjID()] = cardPreTurnSit
			triggerCxt2.addBeTurners(t.(*fightCard), cardPreTurnSit)
		}
		skillActs, _, _ = triggerMgr.trigger(map[int][]iTarget{preBeTurnTrigger: preTurnOvers}, triggerCxt2)
	}

	// --------------------------- 翻面后 -------------------------------
	var turnOvers []iTarget
	turnOverAct := &pb.TurnOverAct{}
	skillActs = append(skillActs, &clientAction{
		actID:  pb.ClientAction_TurnOver,
		actMsg: turnOverAct,
	})
	triggerCxt2.triggerType = turnTrigger
	triggerCxt2.beTurners = nil

	// hehe 自己把自己翻面有些技能有问题
	justTurnOverItSelf := true

	for _, t := range preTurnOvers {
		c := t.(*fightCard)
		if c.getType() != stHand && !c.isInBattle() {
			continue
		}
		if pSit, ok := preTurnSit[c.getObjID()]; triggerCxt2.isCanTurn(c) && (!ok || pSit == c.getSit()) {
			turnOvers = append(turnOvers, c)
			triggerCxt2.addBeTurners(c, c.getSit())
			skillActs = append(skillActs, c.turnOver(skOwner)...)
			turnOverAct.BeTurners = append(turnOverAct.BeTurners, int32(t.getObjID()))

			if t.getObjID() != skOwner.getObjID() {
				justTurnOverItSelf = false
			}
		}
	}

	var turnTriggerObj []iTarget
	if !justTurnOverItSelf && len(turnOvers) > 0 {
		turnTriggerObj = []iTarget{skOwner}
	} else {
		triggerCxt2.turner = nil
	}

	if !isAllHandCard {
		acts2, _, _ := triggerMgr.trigger(map[int][]iTarget{turnTrigger: turnTriggerObj, beTurnTrigger: turnOvers},
			triggerCxt2)
		skillActs = append(skillActs, acts2...)
	}

	return
}

type skillChange struct {
	skillID     int32
	boutTimeout int
}

type skillChangeAction struct {
	baseSkillAction
	addSkills []*skillChange
	delSkills []*skillChange
}

func newSkillChangeAction(skillData *gamedata.Skill, skillChangeInfo []int) *skillChangeAction {
	var targetIDs []int
	var skillID int32
	var modifyType int
	var boutTimeout int
	switch len(skillChangeInfo) {
	case 3:
		targetIDs = skillData.TargetAct
		modifyType = skillChangeInfo[0]
		skillID = int32(skillChangeInfo[1])
		boutTimeout = skillChangeInfo[2]
	case 4:
		targetIDs = []int{skillChangeInfo[0]}
		modifyType = skillChangeInfo[1]
		skillID = int32(skillChangeInfo[2])
		boutTimeout = skillChangeInfo[3]
	default:
		return nil
	}

	act := &skillChangeAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
	}

	sk := &skillChange{
		skillID:     skillID,
		boutTimeout: boutTimeout,
	}
	if modifyType == 1 {
		act.addSkills = []*skillChange{sk}
	} else {
		act.delSkills = []*skillChange{sk}
	}

	return act
}

func newSkillChangeActions(skillData *gamedata.Skill) []*skillChangeAction {
	var actions []*skillChangeAction
L:
	for _, skillChangeInfo := range skillData.SkillChange {
		action := newSkillChangeAction(skillData, skillChangeInfo)
		if action == nil {
			glog.Errorf("newSkillChangeActions error skillID=%d %v", skillData.ID, skillChangeInfo)
			continue
		}
		for _, act := range actions {
			if act.isTargetEqual(action) {
				// 目标相同的合并，减少触发添加技能前时机
				act.merge(action)
				continue L
			}
		}

		actions = append(actions, action)
	}
	return actions
}

func (at *skillChangeAction) isTargetEqual(oth *skillChangeAction) bool {
	if len(at.targetIDs) != len(oth.targetIDs) {
		return false
	}

	for i, id := range at.targetIDs {
		if oth.targetIDs[i] != id {
			return false
		}
	}
	return true
}

func (at *skillChangeAction) merge(oth *skillChangeAction) {
	at.addSkills = append(at.addSkills, oth.addSkills...)
	at.delSkills = append(at.delSkills, oth.delSkills...)
}

func (at *skillChangeAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var loseSkillCasters []iTarget
	var triggerRes *triggerResult

	if len(actionTargets) > 0 && len(at.addSkills) > 0 {
		var preAddSkillIDs []int32
		for _, sc := range at.addSkills {
			preAddSkillIDs = append(preAddSkillIDs, sc.skillID)
		}
		acts, triggerRes, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{preAddSkillTrigger: actionTargets}, &triggerContext{
			triggerType:    preAddSkillTrigger,
			preAddSkillIDs: preAddSkillIDs,
		})
	}

	for _, t := range actionTargets {
		c, ok := t.(iCaster)
		if !ok {
			glog.Errorf("skillChangeAction wrong target type %d", t.getType())
			continue
		}

		for _, sk := range at.delSkills {
			invalidBout := situation.getCurBout() + sk.boutTimeout
			if sk.boutTimeout < 0 {
				invalidBout = -1
			}

			acts2, ok := c.lostSkill(sk.skillID, invalidBout)
			acts = append(acts, acts2...)
			if ok {
				loseSkillCasters = append(loseSkillCasters, t)
			}
		}

		for _, sk := range at.addSkills {
			if triggerRes != nil && !triggerRes.canAddSkill(t.getObjID(), sk.skillID) {
				continue
			}

			acts = append(acts, c.addSkill(sk.skillID, sk.boutTimeout)...)
		}
	}

	if len(loseSkillCasters) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{loseSkillTrigger: loseSkillCasters}, &triggerContext{
			triggerType: loseSkillTrigger,
		})
		acts = append(acts, acts2...)
	}

	return
}

type disCardAction struct {
	baseSkillAction
	targetID2Amount  map[int]string
	targetID2MovieID map[int]int
}

func newDisCardAction(skillData *gamedata.Skill, disCardInfos [][]string) *disCardAction {
	targetIDSet := common.IntSet{}
	targetID2Amount := map[int]string{}
	targetID2MovieID := map[int]int{}

	for _, disCardInfo := range disCardInfos {

		infoLen := len(disCardInfo)
		if infoLen == 0 {
			continue
		}

		var targetIDs []int
		var amount string
		var movieID int

		if infoLen == 1 {
			targetIDs = skillData.TargetAct
			amount = disCardInfo[0]
		} else {
			targetID, err := strconv.Atoi(disCardInfo[0])
			if err != nil {
				glog.Errorf("newDisCardAction error %d", skillData.ID)
				continue
			}
			targetIDs = []int{targetID}
			amount = disCardInfo[1]

			if infoLen == 3 {
				movieID, err = strconv.Atoi(disCardInfo[2])
				if err != nil {
					glog.Errorf("newDisCardAction error %d", skillData.ID)
					continue
				}
			}
		}

		for _, targetID := range targetIDs {
			targetIDSet.Add(targetID)
			targetID2Amount[targetID] = amount
			targetID2MovieID[targetID] = movieID
		}
	}

	if targetIDSet.Size() <= 0 {
		return nil
	}

	return &disCardAction{
		baseSkillAction:  baseSkillAction{targetIDs: targetIDSet.ToList()},
		targetID2MovieID: targetID2MovieID,
		targetID2Amount:  targetID2Amount,
	}
}

func (at *disCardAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	var getAmountByTargetID = func(targetID int) int {
		var amount int
		strAmount, ok := at.targetID2Amount[targetID]
		if ok {
			if strAmount == "x" {
				amount = variableAmount
			} else {
				amount, _ = strconv.Atoi(strAmount)
			}
		}
		return amount
	}

	disCardAct := &pb.DisCardAct{}
	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			glog.Errorf("disCardAction Invoke target wrong type %d", t.getType())
			continue
		}

		var cardObjIds []int32

		targetID := triggerCxt.getActonTargetID(t)
		amount := getAmountByTargetID(targetID)
		for i := 0; i < amount; i++ {
			cardObjID := f.disCard()
			if cardObjID > 0 {
				cardObjIds = append(cardObjIds, int32(cardObjID))
			}
		}

		if len(cardObjIds) > 0 {
			disCardAct.Items = append(disCardAct.Items, &pb.DisCardActItem{
				Uid:        uint64(f.getUid()),
				CardObjIDs: cardObjIds,
				MovieID:    int32(at.targetID2MovieID[targetID]),
			})
		}
	}

	if len(disCardAct.Items) > 0 {
		skillActs = append(skillActs, &clientAction{
			actID:  pb.ClientAction_DisCard,
			actMsg: disCardAct,
		})
	}

	return
}

type drawCardAction struct {
	baseSkillAction
	targetID2Amount  map[int]string
	targetID2MovieID map[int]int
}

func newDrawCardAction(skillData *gamedata.Skill, drawCardInfos [][]string) *drawCardAction {
	var targetIDs []int
	act := &drawCardAction{
		targetID2Amount:  map[int]string{},
		targetID2MovieID: map[int]int{},
	}

	for _, drawCardInfo := range drawCardInfos {
		if len(drawCardInfo) != 3 {
			continue
		}

		targetID, err := strconv.Atoi(drawCardInfo[0])
		if err != nil {
			continue
		}
		movieID, err := strconv.Atoi(drawCardInfo[2])
		if err != nil {
			continue
		}

		targetIDs = append(targetIDs, targetID)
		act.targetID2MovieID[targetID] = movieID
		act.targetID2Amount[targetID] = drawCardInfo[1]
	}

	if len(targetIDs) <= 0 {
		return nil
	}
	act.targetIDs = targetIDs
	return act
}

func (at *drawCardAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var triggerRes *triggerResult
	skillActs, triggerRes, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{beforeDrawCardTrigger: actionTargets},
		&triggerContext{triggerType: beforeDrawCardTrigger})

	var drawTargets []iTarget
	var isSkOwnerHeroes bool
	casterCard, ok := skOwner.(*fightCard)
	if ok {
		isSkOwnerHeroes = casterCard.getCamp() == consts.Heroes
	}

	drawCardAct := &pb.DrawCardAct{}
	for _, t := range actionTargets {
		fter, ok := t.(*fighter)
		if !ok {
			glog.Errorf("drawCardAction Invoke target wrong type %d", t.getType())
			continue
		}

		f := triggerRes.getDrawFighter(fter)
		if f == nil {
			continue
		}

		var cardMsg []*pb.Card
		amount := variableAmount
		targetID := triggerCxt.getActonTargetID(t)
		strAmount := at.targetID2Amount[targetID]
		if strAmount != "x" {
			amount, _ = strconv.Atoi(strAmount)
		}
		amount2 := situation.getMaxHandCardAmount() - f.getHandAmount()
		if amount > amount2 {
			amount = amount2
		}

		//glog.Infof("drawCardAction t=%s, amount=%d, at.amount=%s", t, amount, at.amount)

		for i := 0; i < amount; i++ {
			var card *fightCard
			if isSkOwnerHeroes {
				card = f.popHeroesDrawCard(casterCard)
			} else {
				card = f.popDrawCard(casterCard)
			}
			if card == nil {
				continue
			}

			f.addHandCard(card)
			cardMsg = append(cardMsg, card.packMsg())
		}

		if len(cardMsg) > 0 {
			drawCardAct.Items = append(drawCardAct.Items, &pb.DrawCardItem{
				Uid:        uint64(f.getUid()),
				Cards:      cardMsg,
				MovieID:    int32(at.targetID2MovieID[targetID]),
				OwnerObjID: int32(skOwner.getObjID()),
			})

			drawTargets = append(drawTargets, f)
		}
	}

	if len(drawCardAct.Items) > 0 {
		skillActs = append(skillActs, &clientAction{
			actID:  pb.ClientAction_DrawCard,
			actMsg: drawCardAct,
		})
	}

	if len(drawTargets) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{afterDrawCardTrigger: drawTargets},
			&triggerContext{triggerType: afterDrawCardTrigger})
		acts = append(acts, acts2...)
	}
	return
}

// ActionOth

type pointChangeAction struct {
	baseSkillAction
	pointChangeInfo [][]int
}

func newPointChangeActions(skillData *gamedata.Skill) []*pointChangeAction {
	var actions []*pointChangeAction
	for _, pointChangeInfo := range skillData.PointChange {
		for _, targetID := range skillData.TargetAct {
			var action *pointChangeAction
			for _, act := range actions {
				if act.targetIDs[0] == targetID {
					action = act
					break
				}
			}

			if action == nil {
				action = &pointChangeAction{
					baseSkillAction: baseSkillAction{targetIDs: []int{targetID}},
				}
				actions = append(actions, action)
			}

			switch len(pointChangeInfo) {
			case 2:
				action.addPointChangeInfo(pointChangeInfo)
			case 3:
				action.addPointChangeInfo(pointChangeInfo[1:])
			default:
				glog.Errorf("newPointChangeActions error skillID=%d %v", skillData.ID, pointChangeInfo)
			}
		}
	}
	return actions
}

func (at *pointChangeAction) addPointChangeInfo(info []int) {
	at.pointChangeInfo = append(at.pointChangeInfo, info)
}

func (at *pointChangeAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets

	for _, t := range actionTargets {
		c, ok := t.(*fightCard)
		if !ok {
			glog.Errorf("pointChangeAction target wrong type, %d", t.getType())
			continue
		}

		record := triggerCxt.getAttackOutcomeRecord(c)
		if record == nil {
			continue
		}

		needChange := map[int]int{}
		originalRecord := map[int]int{}
		for _, pointChangeInfo := range at.pointChangeInfo {
			original := pointChangeInfo[0]
			change := pointChangeInfo[1]

			record.forEachRecord(func(oppObjID, bat int) {
				if bat == original {
					needChange[oppObjID] = change
					originalRecord[oppObjID] = original
				}
			})
		}

		for oppObjID, change := range needChange {
			record.addRecord(oppObjID, change)
			opp := situation.getTargetMgr().getTargetCard(oppObjID)
			if opp == nil {
				glog.Errorf("pointChangeAction cant find opp %d", oppObjID)
				continue
			}

			oppChange := bEq
			if change == bLose {
				oppChange = bWin
			} else if change == bWin {
				oppChange = bLose
			}

			triggerCxt.setAttackOutcome(opp, c, oppChange)
			if originalRecord[oppObjID] == bEq && change != bEq {
				triggerCxt.setAttackResult(opp, c, 0)
				triggerCxt.setAttackResult(c, opp, 0)
			}
		}

	}

	return
}

type moveAction struct {
	baseSkillAction
	fromTargetID int // 相对于谁移
	moveType     int // 1.靠近fromTargetID  2.远离fromTargetID  3.移到fromTargetID那个格子
	movieID      string
	n            int // 移多少格
}

func newMoveAction(skillData *gamedata.Skill, moveInfo []string) *moveAction {
	var targetIDs []int
	var fromTargetID int
	var moveType int
	var movieID string
	n := 1
	var err error

	switch len(moveInfo) {
	case 3:
		targetIDs = skillData.TargetAct
		fromTargetID, err = strconv.Atoi(moveInfo[0])
		if err != nil {
			return nil
		}
		moveType, err = strconv.Atoi(moveInfo[1])
		if err != nil {
			return nil
		}
		movieID = moveInfo[2]

	case 5:
		n, err = strconv.Atoi(moveInfo[4])
		if err != nil {
			return nil
		}
		fallthrough
	case 4:
		targetID, err := strconv.Atoi(moveInfo[0])
		if err != nil {
			return nil
		}
		targetIDs = []int{targetID}
		fromTargetID, err = strconv.Atoi(moveInfo[1])
		if err != nil {
			return nil
		}
		moveType, err = strconv.Atoi(moveInfo[2])
		if err != nil {
			return nil
		}
		movieID = moveInfo[3]
	default:
		return nil
	}

	return &moveAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		fromTargetID:    fromTargetID,
		moveType:        moveType,
		movieID:         movieID,
		n:               n,
	}
}

func (at *moveAction) moveToTargetGrid(skOwner iCaster, toTarget *deskGrid, actionTargets []iTarget, triggerCxt *triggerContext,
	situation *battleSituation) (acts []*clientAction, moveActs []*pb.MoveAct, moveTargets []iTarget) {

	var ownerObjID int
	if skOwner != nil {
		ownerObjID = skOwner.getObjID()
	}
	for _, t := range actionTargets {
		targetCard, ok := t.(*fightCard)
		if !ok {
			continue
		}

		oldGrid := targetCard.getGridObj()
		if oldGrid == nil {
			continue
		}

		var result *triggerResult
		acts, result, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{beforeMoveTrigger: []iTarget{targetCard}},
			&triggerContext{triggerType: beforeMoveTrigger, preMoveCards: []iTarget{targetCard}})

		if !result.canMove(targetCard) {
			break
		}

		situation.setGrid(targetCard.getGrid(), oldGrid)
		situation.setGrid(toTarget.getGrid(), targetCard)
		targetCard.setGrid(toTarget)
		newGridObjId := toTarget.getObjID()
		situation.getTargetMgr().delSkillTarget(newGridObjId)
		situation.getTargetMgr().addTarget(oldGrid)

		moveTargets = []iTarget{targetCard}
		moveActs = []*pb.MoveAct{&pb.MoveAct{
			Target:     int32(targetCard.getObjID()),
			TargetGrid: int32(newGridObjId),
			MovieID:    at.movieID,
			MovePos:    consts.UP,
			OwnerObjID: int32(ownerObjID),
		}}
		break
	}

	return
}

func (at *moveAction) moveFromTarget(skOwner iCaster, fromTarget *fightCard, actionTargets []iTarget, triggerCxt *triggerContext,
	situation *battleSituation) (acts []*clientAction, moveActs []*pb.MoveAct, moveTargets []iTarget) {

	fromGrid := fromTarget.getGrid()
	var preMoveTargets []iTarget
	column := situation.getGridColumn()
	gridAmount := situation.gridsAmount
	moveTargetPos := map[int]int{}
	moveTargetNewGird := map[int]*deskGrid{}
	var ownerObjID int
	if skOwner != nil {
		ownerObjID = skOwner.getObjID()
	}

	//glog.Infof("moveAction actionTargets %v", actionTargets)
	for _, t := range actionTargets {
		c, ok := t.(*fightCard)
		if !ok {
			glog.Errorf("moveFromTarget target wrong type, %d", t.getType())
			continue
		}

		movePos := -1
		var newGrid *deskGrid
		targetGrid := c.getGrid()

		if at.moveType == 1 {
			// 拉近

			if targetGrid%column == fromGrid%column {
				// 同一列
				if targetGrid < fromGrid {
					// 上
					movePos = consts.UP
					n := int(math.Min(float64(at.n), float64(fromGrid/column-targetGrid/column-1)))
					canMove := true
					for i := 0; i < n; i++ {
						objID := situation.grids[targetGrid+(i+1)*column]
						g := situation.getTargetMgr().getTargetGrid(objID)
						if g == nil {
							canMove = false
							break
						}
					}

					if canMove {
						newGrid = situation.getTargetInGrid(fromGrid - column).(*deskGrid)
					}
				} else if targetGrid > fromGrid {
					// 下
					movePos = consts.DOWN
					n := int(math.Min(float64(at.n), float64(targetGrid/column-fromGrid/column-1)))
					canMove := true
					for i := 0; i < n; i++ {
						objID := situation.grids[targetGrid-(i+1)*column]
						g := situation.getTargetMgr().getTargetGrid(objID)
						if g == nil {
							canMove = false
							break
						}
					}

					if canMove {
						newGrid = situation.getTargetInGrid(fromGrid + column).(*deskGrid)
					}
				} else {
					continue
				}

			} else if targetGrid/column == fromGrid/column {
				// 同一行
				if targetGrid < fromGrid {
					// 左
					movePos = consts.LEFT
					n := int(math.Min(float64(at.n), float64(fromGrid%column-targetGrid%column-1)))
					canMove := true
					for i := 0; i < n; i++ {
						objID := situation.grids[targetGrid+1+i]
						g := situation.getTargetMgr().getTargetGrid(objID)
						if g == nil {
							canMove = false
							break
						}
					}

					if canMove {
						newGrid = situation.getTargetInGrid(fromGrid - 1).(*deskGrid)
					}
				} else if targetGrid > fromGrid {
					// 右
					movePos = consts.RIGHT
					n := int(math.Min(float64(at.n), float64(targetGrid%column-fromGrid%column-1)))
					canMove := true
					for i := 0; i < n; i++ {
						objID := situation.grids[targetGrid-1-i]
						g := situation.getTargetMgr().getTargetGrid(objID)
						if g == nil {
							canMove = false
							break
						}
					}

					if canMove {
						newGrid = situation.getTargetInGrid(fromGrid + 1).(*deskGrid)
					}
				} else {
					continue
				}
			} else {
				continue
			}

		} else if at.moveType == 2 {
			// 推开
			if targetGrid%column == fromGrid%column {
				// 同一列

				if targetGrid < fromGrid {
					movePos = consts.UP
				} else {
					movePos = consts.DOWN
				}

				if targetGrid+column == fromGrid {
					// 上
					for i := 1; i <= at.n; i++ {
						g := targetGrid - i*column
						if g < 0 {
							break
						}
						gridObj, ok := situation.getTargetInGrid(g).(*deskGrid)
						if !ok {
							break
						} else {
							newGrid = gridObj
						}
					}
				} else if targetGrid-column == fromGrid {
					// 下
					for i := 1; i <= at.n; i++ {
						g := targetGrid + i*column
						if g >= gridAmount {
							break
						}
						gridObj, ok := situation.getTargetInGrid(g).(*deskGrid)
						if !ok {
							break
						} else {
							newGrid = gridObj
						}
					}
				}

			} else if targetGrid/column == fromGrid/column {
				// 同一行

				if targetGrid < fromGrid {
					movePos = consts.LEFT
				} else {
					movePos = consts.RIGHT
				}

				if targetGrid+1 == fromGrid {
					// 左
					for i := 1; i <= at.n; i++ {
						g := targetGrid - i
						if g < 0 || targetGrid/column != g/column {
							break
						}
						gridObj, ok := situation.getTargetInGrid(g).(*deskGrid)
						if !ok {
							break
						} else {
							newGrid = gridObj
						}
					}
				} else if targetGrid-1 == fromGrid {
					// 右
					for i := 1; i <= at.n; i++ {
						g := targetGrid + i
						if g >= gridAmount || targetGrid/column != g/column {
							break
						}
						gridObj, ok := situation.getTargetInGrid(g).(*deskGrid)
						if !ok {
							break
						} else {
							newGrid = gridObj
						}
					}
				}
			} else {
				continue
			}

		} else {
			continue
		}

		if movePos < 0 {
			continue
		}

		if newGrid != nil {
			preMoveTargets = append(preMoveTargets, c)
			moveTargetPos[c.getObjID()] = movePos
			moveTargetNewGird[c.getObjID()] = newGrid
		} else {
			moveActs = append(moveActs, &pb.MoveAct{
				Target:     int32(c.getObjID()),
				TargetGrid: -1,
				MovieID:    at.movieID,
				MovePos:    int32(movePos),
				OwnerObjID: int32(ownerObjID),
			})
		}

	}

	var triggerRes *triggerResult
	if len(preMoveTargets) > 0 {
		acts, triggerRes, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{beforeMoveTrigger: preMoveTargets},
			&triggerContext{triggerType: beforeMoveTrigger, preMoveCards: preMoveTargets})
	}

	for _, t := range preMoveTargets {
		c := t.(*fightCard)
		if !triggerRes.canMove(c) {
			continue
		}

		newGrid := moveTargetNewGird[c.getObjID()]
		oldGrid := c.getGridObj()
		situation.setGrid(c.getGrid(), oldGrid)
		situation.setGrid(newGrid.getGrid(), c)
		c.setGrid(newGrid)
		newGridObjID := newGrid.getObjID()
		situation.getTargetMgr().delSkillTarget(newGridObjID)
		situation.getTargetMgr().addTarget(oldGrid)

		moveTargets = append(moveTargets, c)
		movePos := moveTargetPos[c.getObjID()]
		moveActs = append(moveActs, &pb.MoveAct{
			Target:     int32(c.getObjID()),
			TargetGrid: int32(newGridObjID),
			MovieID:    at.movieID,
			MovePos:    int32(movePos),
			OwnerObjID: int32(ownerObjID),
		})
	}

	//glog.Infof("moveAction moveActs %s", moveActs)

	return
}

func (at *moveAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	fromTargets := situation.getTargetMgr().findTarget(sk, skOwner, at.fromTargetID, triggerCxt, nil)
	if len(fromTargets) <= 0 {
		return
	}

	var moveTargets []iTarget
	var moveToTarget *deskGrid
	var moveFromTarget *fightCard
	if at.moveType == 3 {
		targetGrid, ok := fromTargets[0].(*deskGrid)
		if !ok || targetGrid.getType() != stEmptyGrid {
			return
		}
		moveToTarget = targetGrid
	} else {
		from, ok := fromTargets[0].(*fightCard)
		if !ok {
			return
		}
		moveFromTarget = from
	}

	if at.moveType == 3 {
		acts, moveActs, moveTargets = at.moveToTargetGrid(skOwner, moveToTarget, actionTargets, triggerCxt, situation)
	} else {
		acts, moveActs, moveTargets = at.moveFromTarget(skOwner, moveFromTarget, actionTargets, triggerCxt, situation)
	}

	if len(moveTargets) > 0 {
		afterMoveActs, _, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{afterMoveTrigger: moveTargets}, &triggerContext{
			triggerType: afterMoveTrigger,
			moveCards:   moveTargets,
		})

		result.moveCards = moveTargets
	}
	return
}

type destroyAction struct {
	baseSkillAction
	isClean bool
}

func newDestroyActions(skillData *gamedata.Skill) ([]*destroyAction, []*reEnterBattleAction) {
	var destroyActions []*destroyAction
	isCleanDestory := true
	var reEnterBattleActions []*reEnterBattleAction
	for _, destroyInfo := range skillData.Destroy {
		if destroyInfo[0] == 1 {
			act := &destroyAction{
				baseSkillAction: baseSkillAction{[]int{destroyInfo[1]}},
			}
			destroyActions = append(destroyActions, act)
		} else {
			var girdTargetID int
			if len(destroyInfo) >= 3 {
				girdTargetID = destroyInfo[2]
			}
			reEnterBattleActions = append(reEnterBattleActions, &reEnterBattleAction{
				baseSkillAction: baseSkillAction{[]int{destroyInfo[1]}},
				gridTargetID:    girdTargetID,
			})
			isCleanDestory = false
		}
	}

	if len(skillData.Summon) > 0 {
		isCleanDestory = false
	}
	for _, act := range destroyActions {
		act.isClean = isCleanDestory
	}

	return destroyActions, reEnterBattleActions
}

func (at *destroyAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var targetIDs []int32
	var targets []iTarget
	for _, target := range actionTargets {
		card, ok := target.(*fightCard)
		if !ok {
			continue
		}

		gridObj := card.getGridObj()
		if gridObj != nil {
			situation.setGrid(card.getGrid(), gridObj)
			situation.getTargetMgr().addTarget(gridObj)
			situation.getTriggerMgr().delCaster(card.getObjID())
		}
		situation.getTargetMgr().delTarget(card.getObjID())

		targetIDs = append(targetIDs, int32(card.getObjID()))
		targets = append(targets, card)
		acts = append(acts, card.onDestroy()...)
	}

	if len(targetIDs) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{afterDestroyTrigger: targets}, &triggerContext{
			triggerType: afterDestroyTrigger,
		})
		acts = append(acts, acts2...)

		if at.isClean {
			acts2, _, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{afterCleanDestroyTrigger: targets}, &triggerContext{
				triggerType:  afterCleanDestroyTrigger,
				destoryer:    skOwner,
				beDestoryers: targets,
			})
			acts = append(acts, acts2...)
		}

		skillActs = []*clientAction{&clientAction{
			actID:  pb.ClientAction_Destroy,
			actMsg: &pb.DestroyAct{Targets: targetIDs},
		}}
	}

	return
}

type reEnterBattleAction struct {
	baseSkillAction
	gridTargetID int
}

func (at *reEnterBattleAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var targetGrid *deskGrid
	var actionTargetCards []*fightCard
	for _, target := range actionTargets {
		card, ok := target.(*fightCard)
		if !ok || !card.isInBattle() {
			continue
		}
		actionTargetCards = append(actionTargetCards, card)
	}

	if at.gridTargetID > 0 {
		if len(actionTargetCards) != 1 {
			return
		}

		gridTargets := situation.getTargetMgr().findTarget(sk, skOwner, at.gridTargetID, triggerCxt, nil)
		if len(gridTargets) == 0 {
			return
		}

		if targetGrid2, ok := gridTargets[0].(*deskGrid); !ok {
			return
		} else {
			targetGrid = targetGrid2
		}
	}

	var enterBattleActions []func()
	for _, card := range actionTargetCards {
		targetCard := card
		if !card.isInBattle() {
			continue
		}

		var gridObj *deskGrid
		myGrid := card.getGridObj()
		if targetGrid != nil {
			gridObj = targetGrid
		} else {
			gridObj = myGrid
		}
		if gridObj == nil {
			continue
		}

		situation.setGrid(card.getGrid(), myGrid)
		situation.getTargetMgr().addTarget(myGrid)
		situation.getTriggerMgr().delCaster(card.getObjID())
		situation.getTargetMgr().delTarget(card.getObjID())
		acts = append(acts, card.onDestroy()...)

		c := newCardByTemplate(card, situation, false)
		c.isSummon = true
		c.isPlayInHand = card.isPlayInHand
		fighter := card.getController()
		uid := fighter.getUid()
		c.setController(uid)
		situation.getTargetMgr().addTarget(c)
		cardMsg := c.packMsg()
		acts = append(acts, situation.preCardEnterBattle(c, gridObj, nil, card)...)
		c.setInitController(card.getInitControllerUid())
		c.setInitSit(card.getInitSit())

		enterBattleActions = append(enterBattleActions, func() {

			skillActs = append(skillActs, &clientAction{
				actID:  pb.ClientAction_Destroy,
				actMsg: &pb.DestroyAct{Targets: []int32{int32(targetCard.getObjID())}},
			})

			acts2, _, isInFog, isPublicEnemy := situation.cardEnterBattle(c, gridObj, nil, targetCard)
			acts = append(acts, acts2...)

			skillActs = append(skillActs, &clientAction{
				actID: pb.ClientAction_Summon,
				actMsg: &pb.SummonAct{
					Uid:           uint64(fighter.getUid()),
					GridObjID:     int32(gridObj.getObjID()),
					Card:          cardMsg,
					IsInFog:       isInFog,
					IsPublicEnemy: isPublicEnemy,
				},
			})

		})
	}

	for _, f := range enterBattleActions {
		f()
	}

	return
}

type returnAction struct {
	baseSkillAction
}

func newReturnAction(skillData *gamedata.Skill) *returnAction {
	return &returnAction{
		baseSkillAction: baseSkillAction{targetIDs: []int{skillData.Return}},
	}
}

func (at *returnAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	var desTargetIDs []int32
	var returnTargets []iTarget
	var returnCards []*fightCard
	var preReturnTargets []iTarget

	for _, target := range actionTargets {
		card, ok := target.(*fightCard)
		if !ok || !card.isInBattle() || card.owner <= 0 {
			continue
		}

		preReturnTargets = append(preReturnTargets, card)
	}

	if len(preReturnTargets) > 0 {
		acts, _, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{preReturnTrigger: preReturnTargets}, &triggerContext{
			triggerType: preReturnTrigger,
			returnCards: preReturnTargets,
		})
	}

	for _, t := range preReturnTargets {
		card := t.(*fightCard)
		gridObj := card.getGridObj()
		// 移除
		situation.setGrid(card.getGrid(), gridObj)
		situation.getTargetMgr().delTarget(card.getObjID())
		situation.getTargetMgr().addTarget(gridObj)
		situation.getTriggerMgr().delCaster(card.getObjID())
		acts = append(acts, card.onDestroy()...)

		// 补牌
		f := situation.getFighter(card.owner)
		if f.getUid() != card.owner || f.getHandAmount() >= situation.getMaxHandCardAmount() {
			// 不能补
			desTargetIDs = append(desTargetIDs, int32(card.getObjID()))
			continue
		}

		reCard := newCardByTemplate(card, situation, false)
		returnTargets = append(returnTargets, reCard)
		f.addHandCard(reCard)
		returnCards = append(returnCards, reCard)
		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_Return,
			actMsg: &pb.ReturnAct{
				Uid:       uint64(f.getUid()),
				CardObjID: int32(card.getObjID()),
				Card:      reCard.packMsg(),
			},
		})
	}

	if len(desTargetIDs) > 0 {
		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_Destroy,
			actMsg: &pb.DestroyAct{
				Targets: desTargetIDs,
			},
		})
	}

	if len(returnTargets) > 0 {
		acts2, _, _ := situation.getTriggerMgr().trigger(map[int][]iTarget{afterReturnTrigger: returnTargets}, &triggerContext{
			triggerType: afterReturnTrigger,
			returnCards: returnTargets,
		})
		acts = append(acts, acts2...)
	}

	return
}

type summonAction struct {
	baseSkillAction
	cardID uint32
	side   int
}

func newSummonAction(skillData *gamedata.Skill, summonInfo []int) *summonAction {
	act := &summonAction{
		baseSkillAction: baseSkillAction{targetIDs: []int{summonInfo[1]}},
		cardID:          uint32(summonInfo[0]),
	}
	if summonInfo[2] == 1 {
		act.side = sEnemy
	} else {
		act.side = sOwn
	}
	return act
}

func (at *summonAction) getSummonSkinAndEquip(skillOwnerCard *fightCard, skillOwnerCardData, summonCardData *gamedata.Card) (
	skin, equip, hideEquip string) {

	if skillOwnerCardData == nil {
		return
	}

	if skillOwnerCard.equip != nil {
		if (skillOwnerCardData.CardID == 28 && summonCardData.CardID == 93) ||
			(skillOwnerCardData.CardID == 93 && summonCardData.CardID == 28) {
			// 马超雁鹤
			equip = skillOwnerCard.equip.data.ID
		} else if skillOwnerCardData.CardID == 73 && summonCardData.CardID == 74 {
			// 颜良召文丑
			hideEquip = skillOwnerCard.equip.data.ID
		}

	} else if skillOwnerCard.hideEquip != "" && (skillOwnerCardData.CardID == 74 && summonCardData.CardID == 73) {
		// 文丑招颜良
		equip = skillOwnerCard.hideEquip
	}

	if skillOwnerCardData.Icon == summonCardData.Icon {
		skin = skillOwnerCard.skin
	} else if skillOwnerCard.skin != "" {
		skinGameData := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData)
		if skinData, ok := skinGameData.ID2CardSkin[skillOwnerCard.skin]; ok {
			if bindSkinData, ok := skinGameData.ID2CardSkin[skinData.Bind]; ok && bindSkinData.CardID == summonCardData.CardID {
				skin = skinData.Bind
			}
		}
	}
	return
}

func (at *summonAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, target := range actionTargets {
		gridObj, ok := target.(*deskGrid)
		if !ok || gridObj.getType() != stEmptyGrid {
			continue
		}

		skillOwnerCard, ok := skOwner.(*fightCard)
		skillOwnerCardLevel := 1
		var skillOwnerCardData *gamedata.Card
		var ownerFighter *fighter
		poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
		if ok {
			skillOwnerCardLevel = skillOwnerCard.getLevel()
			skillOwnerCardData = poolGameData.GetCardByGid(skillOwnerCard.gcardID)
			ownerFighter = skillOwnerCard.getController()
		} else {
			ownerFighter = situation.getFighterBySit(skOwner.getSit())
		}

		cardData := poolGameData.GetCard(at.cardID, skillOwnerCardLevel)
		if cardData == nil {
			cardData = poolGameData.GetCampaignCard(at.cardID, skillOwnerCardLevel)
			if cardData == nil {
				cardData = poolGameData.GetCard(at.cardID, 1)
				if cardData == nil {
					cardData = poolGameData.GetCampaignCard(at.cardID, 1)
					if cardData == nil {
						continue
					}
				}
			}
		}

		skin, equip, hideEquip := at.getSummonSkinAndEquip(skillOwnerCard, skillOwnerCardData, cardData)

		c := newCardByData(situation.genObjID(), cardData, skin, equip, situation)
		c.isSummon = true
		c.hideEquip = hideEquip
		if at.side != sOwn {
			ownerFighter = situation.getEnemyFighter(ownerFighter)
		}
		c.setController(ownerFighter.getUid())
		situation.getTargetMgr().addTarget(c)

		cardMsg := c.packMsg()
		acts = append(acts, situation.preCardEnterBattle(c, gridObj, nil, nil)...)
		c.setInitSit(skOwner.getInitSit())
		c.setInitController(situation.getFighterBySit(skOwner.getInitSit()).getUid())
		c.forceSetOwner(c.getInitControllerUid())

		acts2, _, isInFog, isPublicEnemy := situation.cardEnterBattle(c, gridObj, nil, nil)
		acts = append(acts, acts2...)

		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_Summon,
			actMsg: &pb.SummonAct{
				Uid:           uint64(ownerFighter.getUid()),
				GridObjID:     int32(gridObj.getObjID()),
				Card:          cardMsg,
				IsInFog:       isInFog,
				IsPublicEnemy: isPublicEnemy,
			},
		})
	}

	return
}

type attackAction struct {
	baseSkillAction
	attackTargetIDs []int
	attackType      int
}

func newAttackActions(skillData *gamedata.Skill) []*attackAction {
	var atkActions []*attackAction
	for _, atkInfo := range skillData.Attack {
		var targetIDs []int
		var attackTargetID int
		var attackType int
		switch len(atkInfo) {
		case 1:
			targetIDs = skillData.TargetAct
			attackTargetID = atkInfo[0]
		case 2:
			targetIDs = skillData.TargetAct
			attackTargetID = atkInfo[0]
			attackType = atkInfo[1]
		case 3:
			targetIDs = []int{atkInfo[0]}
			attackTargetID = atkInfo[1]
			attackType = atkInfo[2]
		default:
			glog.Errorf("newAttackActions error skillID=%d %v", skillData.ID, atkInfo)
			continue
		}

		for _, targetID := range targetIDs {
			var action *attackAction
			for _, act := range atkActions {
				if act.targetIDs[0] == targetID {
					action = act
					break
				}
			}

			if action == nil {
				action = &attackAction{
					baseSkillAction: baseSkillAction{targetIDs: []int{targetID}},
				}
				atkActions = append(atkActions, action)
			}

			exist := false
			for _, id := range action.attackTargetIDs {
				if id == attackTargetID {
					exist = true
					break
				}
			}
			if !exist {
				action.attackTargetIDs = append(action.attackTargetIDs, attackTargetID)
			}

			if attackType != 0 {
				action.attackType = attackType
			}
		}
	}

	return atkActions
}

func (at *attackAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	if triggerCxt.attackCxt == nil {
		triggerCxt.attackCxt = &attackContext{}
	}

	var ts []*fightCard
	targetPreAttackSit := map[int]int{}
	for _, t := range actionTargets {
		if sk.behavior.hasMoveAction && !result.isMoveCard(t) {
			continue
		}

		c, ok := t.(*fightCard)
		if !ok {
			glog.Errorf("attackAction target wrong type, %d", t.getType())
			continue
		}

		ts = append(ts, c)
		targetPreAttackSit[c.getObjID()] = c.getSit()
	}

	for _, c := range ts {
		if targetPreAttackSit[c.getObjID()] != c.getSit() {
			continue
		}

		var attackTargets []iTarget
		isAppoint := true
		var cacheTargets map[int][]iTarget
		for _, atkTargetID := range at.attackTargetIDs {
			if atkTargetID <= 0 {
				isAppoint = false
				continue
			}

			if cacheTargets == nil {
				// 指定目标攻击，如果目标的targetID跟行为对象一样，应该直接找行为对象，否则有些随机对象可能会找得不对
				cacheTargets = triggerCxt.copyActionTargets()
			}

			targets := situation.getTargetMgr().findTarget(sk, skOwner, atkTargetID, triggerCxt, cacheTargets)
		L:
			for _, target := range targets {
				for _, t := range attackTargets {
					if t.getObjID() == target.getObjID() {
						continue L
					}
				}
				attackTargets = append(attackTargets, target)
			}
		}

		if len(attackTargets) > 0 {
			targetFinder := newAppointAttackTargetFinderSt(attackTargets, at.attackType)
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), targetFinder)
			acts2, _ := situation.attack(c, triggerCxt.attackCxt, true)
			acts = append(acts, acts2...)
		} else if isAppoint {
			// 没有指定攻击的目标
			continue
		} else {
			acts2, _ := situation.attack(c, triggerCxt.attackCxt, true)
			acts = append(acts, acts2...)
		}
	}
	triggerCxt.attackCxt = nil
	return
}

type copyAction struct {
	baseSkillAction
	copyTargetID int
}

func newCopyAction(skillData *gamedata.Skill, copyInfo []int) *copyAction {
	act := &copyAction{}
	switch len(copyInfo) {
	case 1:
		act.targetIDs = skillData.TargetAct
		act.copyTargetID = copyInfo[0]
	case 2:
		act.copyTargetID = copyInfo[0]
		act.targetIDs = []int{copyInfo[1]}
	default:
		return nil
	}
	return act
}

func (at *copyAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	copyTargets := situation.getTargetMgr().findTarget(sk, skOwner, at.copyTargetID, triggerCxt, nil)
	if len(copyTargets) <= 0 {
		return
	}
	beCopyCard, ok := copyTargets[0].(*fightCard)
	if !ok {
		return
	}

	for _, target := range actionTargets {
		card, ok := target.(*fightCard)
		if !ok {
			continue
		}

		c, acts2 := card.copyCard(beCopyCard)
		acts = append(acts, acts2...)

		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_Copy,
			actMsg: &pb.CopyAct{
				Target:   int32(c.getObjID()),
				CopyCard: c.packMsg(),
				OwnerUid: uint64(c.getControllerUid()),
			},
		})
	}
	return
}

type goldGobAction struct {
	baseSkillAction
	gobType int // 1:获得，2:失去
	money   int
}

func newGoldGobAction(skillData *gamedata.Skill, goldGobInfo []int) *goldGobAction {
	act := &goldGobAction{}
	switch len(goldGobInfo) {
	case 2:
		act.targetIDs = skillData.TargetAct
		act.gobType = goldGobInfo[0]
		act.money = goldGobInfo[1]
	case 3:
		act.targetIDs = []int{goldGobInfo[0]}
		act.gobType = goldGobInfo[1]
		act.money = goldGobInfo[2]
	default:
		glog.Errorf("newGoldGobAction error %d", skillData.ID)
		return nil
	}
	return act
}

func (at *goldGobAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	uids := common.UInt64Set{}
	gold := at.money
	if gold <= 0 {
		return
	}
	if at.gobType == 2 {
		gold = -gold
	}
	isLadder := situation.battleType == consts.BtPvp || situation.battleType == consts.BtCampaign

	for _, target := range actionTargets {
		f, ok := target.(*fighter)
		if !ok || f.isRobot {
			continue
		}

		uid := uint64(f.getUid())
		uids.Add(uid)
		acts = append(acts, &clientAction{
			actID: pb.ClientAction_GoldGob,
			actMsg: &pb.GoldGobAct{
				Uid:      uid,
				Gold:     int32(gold),
				IsLadder: isLadder,
			},
		})
	}

	if uids.Size() > 0 && isLadder {
		evq.CallLater(func() {
			uids.ForEach(func(uid uint64) bool {
				utils.PlayerMqPublish(common.UUid(uid), pb.RmqType_Bonus, &pb.RmqBonus{
					ChangeRes: []*pb.Resource{&pb.Resource{Type: int32(consts.Gold), Amount: int32(gold)}},
				})
				return true
			})
		})
	}
	return
}

type changeBoutTimeAction struct {
	baseSkillAction
	change int
}

func newChangeBoutTimeAction(skillData *gamedata.Skill) *changeBoutTimeAction {
	act := &changeBoutTimeAction{
		change: skillData.ChangeBoutTime,
	}
	act.targetIDs = skillData.TargetAct
	return act
}

func (at *changeBoutTimeAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			continue
		}
		f.setBoutTimeout(f.getBoutTimeout() + at.change)
	}
	return
}

type removeEquipAction struct {
	baseSkillAction
}

func newRemoveEquipAction(removeEquipInfo []int) iSkillAction {
	return &removeEquipAction{
		baseSkillAction: baseSkillAction{targetIDs: removeEquipInfo},
	}
}

func (at *removeEquipAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	delEquipAct := &pb.DelEquipAct{}
	for _, t := range actionTargets {
		c, ok := t.(*fightCard)
		if !ok || !c.hasEquip() {
			continue
		}
		acts = append(acts, c.delEquip()...)
	}

	if len(delEquipAct.CardObjIDs) > 0 {
		skillActs = append(skillActs, &clientAction{
			actID:  pb.ClientAction_DelEquip,
			actMsg: delEquipAct,
		})
	}
	return
}

type tauntAction struct {
	baseSkillAction
}

func (at *tauntAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		if f, ok := t.(*fighter); ok {
			var tauntGrid []int
			ts := situation.getTargetMgr().findTarget(sk, skOwner, 16, triggerCxt, nil)
			for _, _t := range ts {
				if g, ok := _t.(*deskGrid); ok {
					tauntGrid = append(tauntGrid, g.getGrid())
				}
			}

			if len(tauntGrid) > 0 {
				situation.addTaunt(sk, skOwner, f, tauntGrid)
			}

		} else {
			glog.Errorf("tauntAction target wrong type, %d", t.getType())
		}
	}
	return
}

type switchPosAction struct {
	baseSkillAction
	switchTargetID int
}

func (at *switchPosAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	target := actionTargets[0]
	c, ok := target.(*fightCard)
	if !ok {
		glog.Errorf("switchPosAction target wrong type, %d", target.getType())
		return
	}

	ts := situation.getTargetMgr().findTarget(sk, skOwner, at.switchTargetID, triggerCxt, nil)
	if len(ts) != 1 {
		return
	}
	targetCard, ok := ts[0].(*fightCard)
	if !ok || !targetCard.isInBattle() {
		return
	}

	moveTargets := []iTarget{targetCard, c}
	var triggerRes *triggerResult
	skillActs, triggerRes, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{beforeMoveTrigger: moveTargets},
		&triggerContext{triggerType: beforeMoveTrigger, preMoveCards: moveTargets})

	if !(triggerRes.canMove(targetCard) && triggerRes.canMove(c)) {
		return
	}

	myGrid := c.getGridObj()
	targetGrid := targetCard.getGridObj()
	situation.setGrid(myGrid.getGrid(), targetCard)
	situation.setGrid(targetGrid.getGrid(), c)
	c.setGrid(targetGrid)
	targetCard.setGrid(myGrid)
	skillActs = append(skillActs, &clientAction{
		actID: pb.ClientAction_SwitchPos,
		actMsg: &pb.SwitchPosAct{
			Target:       int32(c.getObjID()),
			SwitchTarget: int32(targetCard.getObjID()),
		},
	})

	afterMoveActs, _, _ = situation.getTriggerMgr().trigger(map[int][]iTarget{afterMoveTrigger: moveTargets}, &triggerContext{
		triggerType: afterMoveTrigger, moveCards: moveTargets})
	result.moveCards = moveTargets
	return
}

type switchHandCardAction struct {
	baseSkillAction
	targetID2McMovieID   map[int]string
	targetID2TextMovieID map[int]int
}

func newSwitchHandCardAction(switchCardInfos [][]string) *switchHandCardAction {
	act := &switchHandCardAction{
		targetID2McMovieID:   map[int]string{},
		targetID2TextMovieID: map[int]int{},
	}

	for _, switchCardInfo := range switchCardInfos {
		if len(switchCardInfo) != 3 {
			continue
		}

		targetID, err := strconv.Atoi(switchCardInfo[0])
		if err != nil {
			continue
		}
		textMovieID, err := strconv.Atoi(switchCardInfo[1])
		if err != nil {
			continue
		}

		act.targetIDs = append(act.targetIDs, targetID)
		act.targetID2McMovieID[targetID] = switchCardInfo[2]
		act.targetID2TextMovieID[targetID] = textMovieID
	}

	if len(act.targetIDs) <= 0 {
		return nil
	}
	return act
}

func (at *switchHandCardAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	switchHandCardAct := &pb.SwitchHandCardAct{}
	casterCard, _ := skOwner.(*fightCard)
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			continue
		}

		disCardObjID, drawCard := f.switchHandCard(casterCard)
		if drawCard == nil {
			continue
		}

		f.addHandCard(drawCard)
		targetID := triggerCxt.getActonTargetID(t)
		switchHandCardAct.Items = append(switchHandCardAct.Items, &pb.SwitchHandCardItem{
			Uid:          uint64(f.getUid()),
			DisCardObjID: int32(disCardObjID),
			DrawCard:     drawCard.packMsg(),
			McMovieID:    at.targetID2McMovieID[targetID],
			TextMovieID:  int32(at.targetID2TextMovieID[targetID]),
		})
	}

	if len(switchHandCardAct.Items) > 0 {
		skillActs = append(skillActs, &clientAction{
			actID:  pb.ClientAction_SwitchHandCard,
			actMsg: switchHandCardAct,
		})
	}
	return
}

type biyueAction struct {
	baseSkillAction
}

func (at *biyueAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			glog.Errorf("biyueAction wrong type %d", t.getType())
			continue
		}
		result.setDrawFighter(f, nil)
	}
	return
}

type baiyueAction struct {
	baseSkillAction
}

func (at *baiyueAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	ownerFighter := situation.getFighter1()
	if skOwner.getSit() == consts.SitTwo {
		ownerFighter = situation.getFighter2()
	}
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			glog.Errorf("baiyueAction wrong type %d", t.getType())
			continue
		}
		result.setDrawFighter(f, ownerFighter)
	}
	return
}

type tianxiangAction struct {
	baseSkillAction
	movieID int
}

func newTianxiangAction(targetIDs []int, tianxiangInfo string) *tianxiangAction {
	act := &tianxiangAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
	}
	movieInfo := strings.Split(tianxiangInfo, "_")
	if len(movieInfo) != 2 {
		return act
	}

	movieID, err := strconv.Atoi(movieInfo[1])
	if err != nil {
		return act
	}

	act.movieID = movieID
	return act
}

func (at *tianxiangAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	if triggerCxt.modifyValue == 0 {
		return
	}

	needMovie := false
	if triggerCxt.modifyValue < 0 {
		for _, vt := range triggerCxt.modifyValueTargets {
			if vt.getSit() == skOwner.getSit() {
				needMovie = true
				result.setCantSubValTarget(vt)
			}
		}
	} else {
		for _, vt := range triggerCxt.modifyValueTargets {
			if vt.getSit() != skOwner.getSit() {
				needMovie = true
				result.setCantAddValTarget(vt)
			}
		}
	}

	if needMovie && at.movieID > 0 {
		skillActs = []*clientAction{&clientAction{
			actID: pb.ClientAction_TextMovie,
			actMsg: &pb.TextMovieAct{
				MovieID:    int32(at.movieID),
				PlayType:   1,
				Targets:    []int32{int32(skOwner.getObjID())},
				OwnerObjID: int32(skOwner.getObjID()),
			},
		}}
	}
	return
}

type guoseAction struct {
	baseSkillAction
	movieID int
}

func newGuoseAction(targetIDs []int, tianxiangInfo string) *guoseAction {
	act := &guoseAction{
		baseSkillAction: baseSkillAction{targetIDs: targetIDs},
	}
	movieInfo := strings.Split(tianxiangInfo, "_")
	if len(movieInfo) != 2 {
		return act
	}

	movieID, err := strconv.Atoi(movieInfo[1])
	if err != nil {
		return act
	}

	act.movieID = movieID
	return act
}

func (at *guoseAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	if triggerCxt.modifyValue == 0 {
		return
	}

	needMovie := false
	if triggerCxt.modifyValue < 0 {
		for _, vt := range triggerCxt.modifyValueTargets {
			if vt.getSit() != skOwner.getSit() {
				needMovie = true
				result.addAdditionalSubVal(vt, 1)
			}
		}
	} else {
		for _, vt := range triggerCxt.modifyValueTargets {
			if vt.getSit() == skOwner.getSit() {
				needMovie = true
				result.addAdditionalAddVal(vt, 1)
			}
		}
	}

	if needMovie && at.movieID > 0 {
		skillActs = []*clientAction{&clientAction{
			actID: pb.ClientAction_TextMovie,
			actMsg: &pb.TextMovieAct{
				MovieID:    int32(at.movieID),
				PlayType:   1,
				Targets:    []int32{int32(skOwner.getObjID())},
				OwnerObjID: int32(skOwner.getObjID()),
			},
		}}
	}
	return
}

type guanxingAction struct {
	baseSkillAction
}

func (at *guanxingAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			continue
		}
		f2 := situation.getEnemyFighter(f)
		f2.guanxing = true
		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_Guanxing,
			actMsg: &pb.GuanxingAct{
				Uids:            []uint64{uint64(f2.getUid())},
				SitOneDrawCards: situation.fighter1.packDrawCardShadow(),
				SitTwoDrawCards: situation.fighter2.packDrawCardShadow(),
			},
		})
	}
	return
}

type handShowAction struct {
	baseSkillAction
}

func (at *handShowAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, t := range actionTargets {
		f, ok := t.(*fighter)
		if !ok {
			continue
		}
		skillActs = append(skillActs, &clientAction{
			actID: pb.ClientAction_HandShow,
			actMsg: &pb.HandShowAct{
				Uid: uint64(f.getUid()),
			},
		})
	}
	return
}

type timewalkAction struct {
	baseSkillAction
}

func (at *timewalkAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	target := actionTargets[0]
	curFighter := situation.getCurBoutFighter()
	if f, ok := target.(*fighter); ok {
		if f == curFighter {
			situation.nextBoutFighter = f
		}
	} else if c, ok := target.(*fightCard); ok {
		f := c.getController()
		if f == curFighter {
			situation.nextBoutFighter = f
		}
	} else {
		glog.Errorf("timewalkAction target wrong type, %d", target.getType())
	}
	return
}

type forbidMoveAction struct {
	baseSkillAction
}

func (at *forbidMoveAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, target := range actionTargets {
		if c, ok := target.(*fightCard); ok {
			result.addCantMoveCard(c)
		}
	}
	return
}

type forbidAddSkillAction struct {
	baseSkillAction
	skillID int32
}

func (at *forbidAddSkillAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	for _, target := range actionTargets {
		if c, ok := target.(iCaster); ok {
			result.addForbidSkill(c.getObjID(), at.skillID)
		}
	}
	return
}

type attackBuffAction struct {
	baseSkillAction
	buff string
	arg  int
}

func (at *attackBuffAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	if triggerCxt.attackCxt == nil {
		triggerCxt.attackCxt = &attackContext{}
	}
	triggerTargets2 = triggerTargets
	for _, target := range actionTargets {
		c, ok := target.(*fightCard)
		if !ok {
			continue
		}

		switch at.buff {
		case "buff_arrow":
			// 箭矢
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), newArrowAttackTargetFinder(at.arg))
		case "buff_peerless":
			// 无双
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), &peerlessAttackTargetFinderSt{})
		case "buff_lightning":
			// 雷击
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), &lightningAttackTargetFinderSt{})
		case "buff_riprap":
			// 抛石
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), newRiprapAttackTargetFinder(at.arg))
		case "attack_target_finder":
			ts := situation.getTargetMgr().findTarget(sk, skOwner, at.arg, triggerCxt, nil)
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), newBuffAppointAttackTargetFinderSt(ts))
		case "buff_breakthrough":
			// 贯矢
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), &breakthroughAttackTargetFinderSt{})
		case "buff_aoe":
			// 阉党?
			triggerCxt.attackCxt.setAttackTargetFinder(c.getObjID(), &aoeAttackTargetFinderSt{})
		case "buff_pierce":
			// 点破
			triggerCxt.attackCxt.setAttacker(c.getObjID(), pierceAttacker)
		case "buff_shield":
			// 藤甲
			triggerCxt.attackCxt.setDefenser(c.getObjID(), shieldDefenser)
		case "buff_scuffle":
			// 乱斗
			triggerCxt.attackCxt.setAttackType(c.getObjID(), atScuffle)
		}
	}
	return
}

type fogAction struct {
	baseSkillAction
}

func (at *fogAction) invoke(sk *skill, skOwner iCaster, triggerType int, triggerTargets, actionTargets []iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (skillActs,
	acts []*clientAction, moveActs []*pb.MoveAct, afterMoveActs []*clientAction, triggerTargets2 []iTarget) {

	triggerTargets2 = triggerTargets
	acts = sk.onEnterFog(actionTargets)
	return
}

func newOthAction(sb *skillBehavior, skillData *gamedata.Skill, othActionInfo []string) (iSkillAction, int32) {
	var othName string
	var targetIDs []int
	var arg int

	switch len(othActionInfo) {
	case 1:
		targetIDs = skillData.TargetAct
		othName = othActionInfo[0]
	case 2:
		targetIDs = skillData.TargetAct
		othName = othActionInfo[0]
		arg, _ = strconv.Atoi(othActionInfo[1])
	case 3:
		targetID, err := strconv.Atoi(othActionInfo[0])
		if err != nil {
			return nil, 0
		}
		targetIDs = []int{targetID}
		othName = othActionInfo[0]
		arg, err = strconv.Atoi(othActionInfo[2])
		if err != nil {
			return nil, 0
		}
	default:
		return nil, 0
	}

	switch othName {
	case "buff_taunt":
		return &tauntAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "switch_pos":
		sb.hasMoveAction = true
		return &switchPosAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
			switchTargetID:  arg,
		}, 0
	case "biyue":
		return &biyueAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "baiyue":
		return &baiyueAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "guanxing":
		return &guanxingAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "buff_hand_show":
		return &handShowAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "timewalk":
		return &timewalkAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "forbidMove":
		return &forbidMoveAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0
	case "forbidAddSkill":
		return &forbidAddSkillAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
			skillID:         int32(arg),
		}, int32(arg)

	case "buff_arrow":
		fallthrough
	case "buff_peerless":
		fallthrough
	case "buff_lightning":
		fallthrough
	case "buff_riprap":
		fallthrough
	case "attack_target_finder":
		fallthrough
	case "buff_breakthrough":
		fallthrough
	case "buff_pierce":
		fallthrough
	case "buff_shield":
		fallthrough
	case "buff_aoe":
		fallthrough
	case "buff_scuffle":
		return &attackBuffAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
			buff:            othName,
			arg:             arg,
		}, 0
	case "force_attack":
		sb.isForceAttack = true
		return nil, 0
	case "public_enemy":
		sb.isPublicEnemy = true
		return nil, 0
	case "fog":
		if skillData.TriggerOpp == awaysTrigger {
			sb.isAwaysTriggerFogSkill = true
		}
		return &fogAction{
			baseSkillAction: baseSkillAction{targetIDs: targetIDs},
		}, 0

	default:
		if strings.HasPrefix(othName, "tianxiang") {
			return newTianxiangAction(targetIDs, othName), 0
		} else if strings.HasPrefix(othName, "guose") {
			return newGuoseAction(targetIDs, othName), 0
		} else {
			return nil, 0
		}
	}
}

type skillBehavior struct {
	data            *gamedata.Skill
	targetIDs       []int // 技能所有的目标
	usefulTargetIDs []int // 不包括只用于展示的目标
	conditions      []*targetAmountCondition
	actions         []iSkillAction

	cantDel                bool
	isTurnRecover          bool
	isTurnDel              bool
	totalTimes             int
	round                  int
	forbidAddSkillIDs      []int32
	isForceAttack          bool
	isPublicEnemy          bool
	isAwaysTriggerFogSkill bool
	hasMoveAction          bool
	effectiveBout          int // 1:我方回合有效，2.地方回合有效
}

func newSkillBehavior(skillData *gamedata.Skill) *skillBehavior {
	sb := &skillBehavior{
		data: skillData,
	}
	for _, condition := range skillData.Condition {
		c := newTargetAmountCondition(condition)
		if c != nil {
			sb.conditions = append(sb.conditions, c)
		} else {
			glog.Errorf("newTargetAmountCondition error skillID=%d %s", skillData.ID, condition)
		}
	}

	for _, typeInfo := range skillData.Type {
		switch typeInfo[0] {
		case 1:
			sb.round = typeInfo[1]
		case 2:
			sb.totalTimes = typeInfo[1]
		case 3:
			if typeInfo[1] == 1 {
				sb.isTurnDel = true
			} else if typeInfo[1] == 2 {
				sb.isTurnDel = true
				sb.isTurnRecover = true
			}
		case 4:
			sb.cantDel = true
		case 5:
			sb.effectiveBout = typeInfo[1]
		}
	}

	var lastValueEffect []string
	var valueEffect []string
	for i, valueInfo := range skillData.ActionValue {
		if i+1 > len(skillData.ValueEffect) {
			valueEffect = lastValueEffect
		} else {
			valueEffect = skillData.ValueEffect[i]
			lastValueEffect = valueEffect
		}

		act := newValueAction(skillData, valueInfo, valueEffect)
		if act == nil {
			glog.Errorf("newValueAction error skillID=%d %v", skillData.ID, valueInfo)
		} else {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	for _, textMovieInfo := range skillData.Effect {
		act := newTextMovieAction(skillData, textMovieInfo)
		sb.actions = append(sb.actions, act)
		sb.addTargetIDs(act.getTargetIDs())
	}

	for _, mcMovieInfo := range skillData.MovieEffect {
		act := newMcMovieAction(skillData, mcMovieInfo)
		sb.actions = append(sb.actions, act)
		sb.addTargetIDs(act.getTargetIDs())
	}

	if len(skillData.TurnOver) > 0 {
		acts := newTurnOverActions(skillData, skillData.TurnOver)
		for _, act := range acts {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.SkillChange) > 0 {
		actions := newSkillChangeActions(skillData)
		for _, act := range actions {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.Discard) > 0 {
		act := newDisCardAction(skillData, skillData.Discard)
		if act == nil {
			glog.Errorf("newDisCardAction error skillID=%d %v", skillData.ID, skillData.Discard)
		} else {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.SwitchCard) > 0 {
		act := newSwitchHandCardAction(skillData.SwitchCard)
		if act == nil {
			glog.Errorf("newSwitchHandCardAction error skillID=%d %v", skillData.ID, skillData.SwitchCard)
		} else {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	for _, othActionInfo := range skillData.ActionOth {
		act, forbidAddSkillID := newOthAction(sb, skillData, othActionInfo)
		if act != nil {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
			if forbidAddSkillID > 0 {
				sb.forbidAddSkillIDs = append(sb.forbidAddSkillIDs, forbidAddSkillID)
			}
		}
	}

	if len(skillData.PointChange) > 0 {
		actions := newPointChangeActions(skillData)
		for _, act := range actions {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	for _, moveInfo := range skillData.Move {
		act := newMoveAction(skillData, moveInfo)
		if act == nil {
			glog.Errorf("newMoveAction error skillID=%d %v", skillData.ID, moveInfo)
		} else {
			sb.hasMoveAction = true
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.Destroy) > 0 {
		destroyActions, reEnterBattleActions := newDestroyActions(skillData)
		for _, act := range destroyActions {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
		for _, act := range reEnterBattleActions {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.DrawCard) > 0 {
		act := newDrawCardAction(skillData, skillData.DrawCard)
		if act == nil {
			glog.Errorf("newDrawCardAction error skillID=%d %v", skillData.ID, skillData.DrawCard)
		} else {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if skillData.Return > 0 {
		act := newReturnAction(skillData)
		sb.actions = append(sb.actions, act)
		sb.addUsefulTargetIDs(act.getTargetIDs())
	}

	for _, summonInfo := range skillData.Summon {
		act := newSummonAction(skillData, summonInfo)
		sb.actions = append(sb.actions, act)
		sb.addUsefulTargetIDs(act.getTargetIDs())
	}

	if len(skillData.Attack) > 0 {
		actions := newAttackActions(skillData)
		for _, act := range actions {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	for _, copyInfo := range skillData.Copy {
		act := newCopyAction(skillData, copyInfo)
		if act != nil {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	for _, goldGobInfo := range skillData.GoldRob {
		act := newGoldGobAction(skillData, goldGobInfo)
		if act != nil {
			sb.actions = append(sb.actions, act)
			sb.addTargetIDs(act.getTargetIDs())
		}
	}

	if skillData.ChangeBoutTime != 0 {
		act := newChangeBoutTimeAction(skillData)
		if act != nil {
			sb.actions = append(sb.actions, act)
			sb.addUsefulTargetIDs(act.getTargetIDs())
		}
	}

	if len(skillData.RemoveEquip) > 0 {
		act := newRemoveEquipAction(skillData.RemoveEquip)
		sb.actions = append(sb.actions, act)
		sb.addUsefulTargetIDs(act.getTargetIDs())
	}

	return sb
}

func doInitSkill(gdata gamedata.IGameData) {
	skillGameData := gdata.(*gamedata.SkillGameData)
	skillBehaviors := map[int32]*skillBehavior{}
	for _, skillData := range skillGameData.GetAllSkill() {
		skillBehaviors[skillData.ID] = newSkillBehavior(skillData)
	}
	allSkillBehaviors = skillBehaviors

	drawnGameAction = newDrawCardAction(nil, [][]string{{"27", "1", "1013"}})
}

func initSkill() {
	gdata := gamedata.GetGameData(consts.Skill)
	gdata.AddReloadCallback(doInitSkill)
	doInitSkill(gdata)
}

func (sb *skillBehavior) addTargetIDs(targetIDs []int) {
L:
	for _, targetID := range targetIDs {
		for _, id := range sb.targetIDs {
			if id == targetID {
				continue L
			}
		}
		sb.targetIDs = append(sb.targetIDs, targetID)
	}
}

func (sb *skillBehavior) addUsefulTargetIDs(targetIDs []int) {
	sb.addTargetIDs(targetIDs)
L:
	for _, targetID := range targetIDs {
		for _, id := range sb.usefulTargetIDs {
			if id == targetID {
				continue L
			}
		}
		sb.usefulTargetIDs = append(sb.usefulTargetIDs, targetID)
	}
}

func (sb *skillBehavior) getTargetIDs(situation *battleSituation) []int {
	if situation.isAiThinking() {
		return sb.usefulTargetIDs
	} else {
		return sb.targetIDs
	}
}

// 过虑不再合法的目标
func (sb *skillBehavior) filterTarget(targetID int, sk *skill, skillOwner iTarget, actionTargets map[int][]iTarget,
	triggerCxt *triggerContext, situation *battleSituation, cacheTargets map[int][]iTarget) []iTarget {

	targets := actionTargets[targetID]
	if len(targets) <= 0 {
		return targets
	}

	var targets2 []iTarget
	filter := allTargetFilters[targetID]
	if filter == nil {
		return targets2
	}

	for _, t := range targets {
		if filter.isTarget(sk, skillOwner, t, triggerCxt, situation, cacheTargets) {
			targets2 = append(targets2, t)
		}
	}

	actionTargets[targetID] = targets2
	return targets
}

func (sb *skillBehavior) invoke(sk *skill, triggerType int, triggerTargets []iTarget, actionTargets map[int][]iTarget,
	variableAmount int, triggerCxt *triggerContext, result *triggerResult, situation *battleSituation) (*pb.SkillAct,
	[]*clientAction, []iTarget) {

	skillOwner := sk.getOwner()
	var actions []*clientAction
	skillAct := &pb.SkillAct{Owner: int32(skillOwner.getObjID()), SkillID: sk.getID(), IsEquip: sk.isEquip}
	var skillActs []*clientAction
	var acts []*clientAction
	var moveActs []*pb.MoveAct
	var afterMoveActs []*clientAction

	for _, action := range sb.actions {
		var targets []iTarget
		var needFilterTarget bool
		switch action.(type) {
		case *valueAction:
		case *textMovieAction:
			if situation.isAiThinking() {
				continue
			}
		case *mcMovieAction:
			if situation.isAiThinking() {
				continue
			}
		case *goldGobAction:
			if situation.isAiThinking() {
				continue
			}
		default:
			needFilterTarget = true
		}

		var _addTargets = func(targetID int, ts []iTarget) {
			triggerCxt.addActionTargets(targetID, ts)
			if len(targets) <= 0 {
				targets = append(targets, ts...)
			} else {
				for _, t := range ts {
					var isExist bool
					for _, t2 := range targets {
						if t == t2 {
							isExist = true
							break
						}
					}
					if !isExist {
						targets = append(targets, t)
					}
				}
			}
		}

		triggerCxt.delActionTargets()
		if !needFilterTarget {
			for _, targetID := range action.getTargetIDs() {
				if ts, ok := actionTargets[targetID]; ok {
					_addTargets(targetID, ts)
				}
			}
		} else {
			cacheTargets := map[int][]iTarget{}
			for _, targetID := range action.getTargetIDs() {
				ts := sb.filterTarget(targetID, sk, skillOwner, actionTargets, triggerCxt, situation, cacheTargets)
				_addTargets(targetID, ts)
			}
		}

		if len(targets) > 0 || (sk.fogTargets != nil && sk.fogTargets.Size() > 0) {
			skillActs, acts, moveActs, afterMoveActs, triggerTargets = action.invoke(sk, skillOwner, triggerType, triggerTargets,
				targets, variableAmount, triggerCxt, result, situation)

			for _, cliAct := range skillActs {
				actData := cliAct.packMsg()
				skillAct.Actions = append(skillAct.Actions, actData)
			}
			skillAct.MoveActs = append(skillAct.MoveActs, moveActs...)
			for _, cliAct := range afterMoveActs {
				actData := cliAct.packMsg()
				skillAct.AfterMoveActs = append(skillAct.AfterMoveActs, actData)
			}
			actions = append(actions, acts...)
		}
	}

	return skillAct, actions, triggerTargets
}
