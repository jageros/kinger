package main

import (
	"math/rand"
	"kinger/gamedata"
	"kinger/proto/pb"
	//"kinger/gopuppy/common/glog"
	"kinger/gopuppy/attribute"
	"strconv"
	"fmt"
	"kinger/gopuppy/common"
)

type skill struct {
	situation *battleSituation
	behavior       *skillBehavior
	ownerObjID int
	playCardIdx int

	// 开始生效的那个回合
	addBout     int
	// 生效后多少个回合后失去
	boutTimeout int
	// 失去技能，直到lostUntilBout那个回合为止，小于0时永远失去
	lostUntilBout int
	// 是否由于owner被翻面而失去了
	turnDel     bool
	// 生效时所属阵营
	initSit     int
	// 总触发了多少次
	triggerTotalTimes int
	targetTriggerTimes int
	statusMcMovie map[int]string
	fogTargets common.IntSet   // 由于这个技能处于大雾中的target
	isAddInHand bool   // 是否卡在手上时加的技能
	isEquip bool
}

func newSkill(skillID int32, situation *battleSituation, owner iCaster) *skill {
	sb, ok := allSkillBehaviors[skillID]
	if !ok {
		return nil
	}

	return &skill{
		situation: situation,
		behavior: sb,
		ownerObjID: owner.getObjID(),
		addBout: -1,
	}
}

func (s *skill) String() string {
	return fmt.Sprintf("[skillID=%d, owner=%s, addBout=%d, boutTimeout=%d, curBout=%d]", s.getID(), s.getOwner(), s.addBout,
		s.boutTimeout, s.situation.getCurBout())
}

func (s *skill) copy(situation *battleSituation) *skill {
	sk := *s
	cpy := &sk
	cpy.situation = situation
	cpy.statusMcMovie = nil
	cpy.fogTargets = cpy.fogTargets.Copy()
	return cpy
}

func (s *skill) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt32("id", s.getID())
	attr.SetInt("ownerObjID", s.ownerObjID)
	attr.SetInt("playCardIdx", s.playCardIdx)
	attr.SetInt("addBout", s.addBout)
	attr.SetInt("boutTimeout", s.boutTimeout)
	attr.SetInt("lostUntilBout", s.lostUntilBout)
	attr.SetBool("turnDel", s.turnDel)
	attr.SetInt("initSit", s.initSit)
	attr.SetInt("triggerTotalTimes", s.triggerTotalTimes)
	attr.SetInt("targetTriggerTimes", s.targetTriggerTimes)
	attr.SetBool("isAddInHand", s.isAddInHand)
	attr.SetBool("isEquip", s.isEquip)

	if s.statusMcMovie != nil {
		statusMcMovieAttr := attribute.NewMapAttr()
		attr.SetMapAttr("statusMcMovie", statusMcMovieAttr)
		for objID, movieID := range s.statusMcMovie {
			statusMcMovieAttr.SetStr(strconv.Itoa(objID), movieID)
		}
	}

	if s.fogTargets != nil {
		fogTargetsAttr := attribute.NewListAttr()
		attr.SetListAttr("fogTargets", fogTargetsAttr)
		for objID, _ := range s.fogTargets {
			fogTargetsAttr.AppendInt(objID)
		}
	}

	return attr
}

func (s *skill) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	s.situation = situation
	id := attr.GetInt32("id")
	s.behavior = allSkillBehaviors[id]
	s.ownerObjID = attr.GetInt("ownerObjID")
	s.playCardIdx = attr.GetInt("playCardIdx")
	s.addBout = attr.GetInt("addBout")
	s.boutTimeout = attr.GetInt("boutTimeout")
	s.lostUntilBout = attr.GetInt("lostUntilBout")
	s.turnDel = attr.GetBool("turnDel")
	s.initSit = attr.GetInt("initSit")
	s.triggerTotalTimes = attr.GetInt("triggerTotalTimes")
	s.targetTriggerTimes = attr.GetInt("targetTriggerTimes")
	s.isAddInHand = attr.GetBool("isAddInHand")
	s.isEquip = attr.GetBool("isEquip")

	statusMcMovieAttr := attr.GetMapAttr("statusMcMovie")
	if statusMcMovieAttr != nil {
		s.statusMcMovie = map[int]string{}
		statusMcMovieAttr.ForEachKey(func(key string) {
			objID, _ := strconv.Atoi(key)
			s.statusMcMovie[objID] = statusMcMovieAttr.GetStr(key)
		})
	}

	fogTargetsAttr := attr.GetListAttr("fogTargets")
	if fogTargetsAttr != nil {
		s.fogTargets = common.IntSet{}
		fogTargetsAttr.ForEachIndex(func(index int) bool {
			objID := fogTargetsAttr.GetInt(index)
			s.fogTargets.Add(objID)
			return true
		})
	}
}

func (s *skill) getData() *gamedata.Skill {
	return s.behavior.data
}

func (s *skill) isShownSkill() bool {
	return s.behavior.data.SkillGroup != "" || s.behavior.data.Name > 0
}

func (s *skill) getID() int32 {
	return s.behavior.data.ID
}

func (s *skill) getOwner() iCaster {
	t := s.situation.getTargetMgr().getTarget(s.ownerObjID)
	if c, ok := t.(iCaster); ok {
		return c
	} else {
		return nil
	}
}

func (s *skill) getPlayCardIdx() int {
	return s.playCardIdx
}

func (s *skill) setPlayCardIdx(idx int) {
	s.playCardIdx = idx
}

func (s *skill) getRemainBout() int {
	if s.boutTimeout > 0 {
		remain := s.boutTimeout - s.situation.getCurBout() - s.addBout
		if remain < 0 {
			return 0
		} else {
			return remain
		}
	} else {
		return -1
	}
}

func (s *skill) setBoutTimeout(boutTimeout int) {
	s.addBout = s.situation.getCurBout()
	s.boutTimeout = boutTimeout
}

func (s *skill) onAdd(initSit int) []*clientAction {
	s.initSit = initSit
	if s.addBout < 0 {
		s.addBout = s.situation.getCurBout()
		s.boutTimeout = s.behavior.round
	}
	if s.isLost() {
		return []*clientAction{}
	} else {
		return s.onEffective()
	}
}

func (s *skill) onEffective() []*clientAction {
	var acts []*clientAction
	c, ok := s.getOwner().(*fightCard)
	if ok {
		c.onSkillEffective(s)
	}

	if s.fogTargets != nil {
		targetMgr := s.situation.getTargetMgr()
		if s.behavior.isAwaysTriggerFogSkill {
			s.fogTargets = nil
			targetIDs := s.behavior.getTargetIDs(s.situation)
			for _, targetID := range targetIDs {
				targets := targetMgr.findTarget(s, s.getOwner(), targetID, &triggerContext{}, nil)
				for _, t := range targets {
					c, ok := t.(*fightCard)
					if !ok {
						continue
					}

					if s.fogTargets == nil {
						s.fogTargets = common.IntSet{}
					}
					s.fogTargets.Add(c.getObjID())
				}
			}
		}

		if s.fogTargets != nil {
			s.fogTargets.ForEach(func(objID int) {
				c := targetMgr.getTargetCard(objID)
				if c == nil || c.isDestroy {
					return
				}
				isOldInFog := c.isInFog()
				c.onEnterFog()
				if !isOldInFog && c.isInFog() {
					acts = append(acts, &clientAction{
						actID: pb.ClientAction_EnterFog,
						actMsg: &pb.EnterFogAct{
							Target:  int32(c.getObjID()),
							IsPublicEnemy: c.isPublicEnemy(),
						},
					})
				}
			})
		}
	}

	acts = append(acts, s.addStatusMcMovie()...)
	return acts
}

func (s *skill) onInvalid() []*clientAction {
	var acts []*clientAction
	owner := s.getOwner()
	if owner != nil {
		c, ok := owner.(*fightCard)
		if ok {
			c.onSkillInvalid(s)
		}
	}

	if s.fogTargets != nil {
		targetMgr := s.situation.getTargetMgr()
		var leaveFogAct *pb.LeaveFogAct
		s.fogTargets.ForEach(func(objID int) {
			c := targetMgr.getTargetCard(objID)
			if c != nil {
				isOldInFog := c.isInFog()
				c.onLeaveFog()
				if isOldInFog && !c.isInFog() {
					if leaveFogAct == nil {
						leaveFogAct = &pb.LeaveFogAct{}
					}
					leaveFogAct.Targets = append(leaveFogAct.Targets, int32(objID))
				}
			}
		})

		if leaveFogAct != nil {
			acts = append(acts, &clientAction{
				actID: pb.ClientAction_LeaveFog,
				actMsg: leaveFogAct,
			})
		}
	}

	acts = append(acts, s.delStatusMcMovie()...)
	return acts
}

func (s *skill) addStatusMcMovie() []*clientAction {
	var acts []*clientAction
	if s.statusMcMovie != nil {
		return acts
	}
	if s.situation.isAiThinking() {
		return acts
	}
	data := s.behavior.data
	if data.TriggerTimes > 0 && s.targetTriggerTimes + 1 != data.TriggerTimes {
		return acts
	}
	statusEffect := data.GetStatusEffect()
	if len(statusEffect) <= 0 {
		return acts
	}

	s.statusMcMovie = map[int]string{}
	cacheTargets := map[int][]iTarget{}
	owner := s.getOwner()
	if owner == nil {
		return acts
	}
	for _, info := range statusEffect {
		targetIDs := info[0].([]int)
		movieID := info[1].(string)
		playType := info[2].(int)
		var targets []iTarget
		for _, targetID := range targetIDs {
			ts := s.situation.getTargetMgr().findTarget(s, owner, targetID, &triggerContext{}, cacheTargets)
		L:for _, t := range ts {
			for _, t2 := range targets {
				if t == t2 {
					continue L
				}
			}
			targets = append(targets, t)
		}
		}

		for _, t := range targets {
			if t == nil {
				continue
			}
			act := t.addEffect(s.ownerObjID, movieID, playType, -1)
			act.actID = pb.ClientAction_SkillStatusMovie
			acts = append(acts, act)
			s.statusMcMovie[t.getObjID()] = movieID
		}
	}
	return acts
}

func (s *skill) delStatusMcMovie() []*clientAction {
	var acts []*clientAction
	if s.statusMcMovie == nil {
		return acts
	}
	if s.situation.isAiThinking() {
		return acts
	}

	for objID, movieID := range s.statusMcMovie {
		t := s.situation.getTargetMgr().getTarget(objID)
		if t != nil {
			act := t.delEffect(movieID)
			if act != nil {
				act.actID = pb.ClientAction_SkillStatusMovie
				acts = append(acts, act)
			}
		}
	}
	s.statusMcMovie = nil
	return acts
}

func (s *skill) onEnterFog(targets []iTarget) []*clientAction {
	var acts []*clientAction
	var leaveFogAct *pb.LeaveFogAct
	if s.fogTargets != nil {
	L1: for objID, _ := range s.fogTargets {
			for _, t := range targets {
				if t.getObjID() == objID {
					continue L1
				}
			}
			c := s.situation.getTargetMgr().getTargetCard(objID)
			if c != nil {
				isOldInFog := c.isInFog()
				c.onLeaveFog()
				if isOldInFog && !c.isInFog() {
					// 离开了大雾
					if leaveFogAct == nil {
						leaveFogAct = &pb.LeaveFogAct{}
					}
					leaveFogAct.Targets = append(leaveFogAct.Targets, int32(objID))
				}
			}
			s.fogTargets.Remove(objID)
		}
	}

	for _, t := range targets {
		c, ok := t.(*fightCard)
		if !ok {
			continue
		}

		objID := c.getObjID()
		if s.fogTargets != nil && s.fogTargets.Contains(objID) {
			continue
		}

		if s.fogTargets == nil {
			s.fogTargets = common.IntSet{}
		}
		s.fogTargets.Add(objID)
		isOldInFog := c.isInFog()
		c.onEnterFog()
		if !isOldInFog && c.isInFog() {
			// 进入大雾
			acts = append(acts, &clientAction{
				actID: pb.ClientAction_EnterFog,
				actMsg: &pb.EnterFogAct{
					Target: int32(objID),
					IsPublicEnemy: c.isPublicEnemy(),
				},
			})
		}
	}

	if leaveFogAct != nil {
		acts = append(acts, &clientAction{
			actID: pb.ClientAction_LeaveFog,
			actMsg: leaveFogAct,
		})
	}

	return acts
}

func (s *skill) isEffective() bool {
	curBout := s.situation.getCurBout()
	return s.isEffectiveIgnoreLostByBout(curBout) && !s.isLostByBout(curBout)
}

func (s *skill) isEffectiveByBout(bout int) bool {
	return s.isEffectiveIgnoreLostByBout(bout) && !s.isLostByBout(bout)
}

func (s *skill) isEffectiveIgnoreLost() bool {
	return s.isEffectiveIgnoreLostByBout(s.situation.getCurBout())
}

func (s *skill) isEffectiveIgnoreLostByBout(bout int) bool {
	if s.turnDel {
		return false
	}
	if s.boutTimeout > 0 && bout >= s.addBout + s.boutTimeout {
		return false
	}
	if s.behavior.totalTimes > 0 && s.triggerTotalTimes >= s.behavior.totalTimes {
		return false
	}
	return true
}

func (s *skill) isLost() bool {
	return s.isLostByBout(s.situation.getCurBout())
}

func (s *skill) isLostByBout(bout int) bool {
	return s.lostUntilBout < 0 || (s.lostUntilBout > 0 && s.lostUntilBout > bout)
}

func (s *skill) lost(untilBout int, ignoreCantDel bool) ([]*clientAction, bool) {
	if s.behavior.cantDel && !ignoreCantDel {
		return []*clientAction{}, false
	}

	isOldLost := s.isLost()
	if s.lostUntilBout == 0 {
		s.lostUntilBout = untilBout
		return s.onInvalid(), true
	} else if s.lostUntilBout < 0 {
		return []*clientAction{}, false
	} else if untilBout < 0 {
		s.lostUntilBout = untilBout
	} else if untilBout > s.lostUntilBout {
		s.lostUntilBout = untilBout
	}

	if isOldLost {
		return []*clientAction{}, false
	} else {
		return s.onInvalid(), true
	}
}

func (s *skill) onOwnerTurnOver(curSit int) []*clientAction {
	var actions []*clientAction
	if s.turnDel {
		// 已经翻面后被删除
		if s.behavior.isTurnRecover && s.initSit == curSit {
			// 翻回来恢复
			s.turnDel = false
			skillOwner := s.getOwner()
			if skillOwner != nil {
				if s.isEffective() {
					actions = append(actions, s.onEffective()...)
					actions = append(actions, &clientAction{
						actID:  pb.ClientAction_AddSkill,
						actMsg: &pb.ModifySkillAct{CardObjID: int32(skillOwner.getObjID()), SkillID: s.getID()},
					})
				}
			}
		}
	} else if s.behavior.isTurnDel && s.initSit != curSit {
		// 翻面后删除
		s.turnDel = true
		actions = append(actions, s.onInvalid()...)
		skillOwner := s.getOwner()
		if skillOwner != nil {
			actions = append(actions, &clientAction{
				actID: pb.ClientAction_DelSkill,
				actMsg: &pb.ModifySkillAct{CardObjID: int32(skillOwner.getObjID()), SkillID: s.getID(), IsEquip: s.isEquip},
			})
			//for _, skillID := range s.behavior.data.SkillCom {
				// 后续技能
			//	actions = append(actions, skillOwner.addSkill(skillID, 0)...)
			//}
			return actions
		}
	}
	return actions
}

func (s *skill) boutEnd() []*clientAction {
	var actions []*clientAction
	curBout := s.situation.getCurBout()

	if s.boutTimeout > 0 && curBout >= s.addBout + s.boutTimeout {
		// 时间到了
		s.situation.delTaunt(s.ownerObjID, s.getID())
		owner := s.getOwner()
		if owner != nil {
			actions = owner.delSkill(s)
		}

		if s.isEffectiveIgnoreLostByBout(curBout - 1) && !s.isLost() {
			// 上回合有效，现在失效了
			if owner != nil {
				for _, skillID := range s.behavior.data.SkillCom {
					actions = append(actions, owner.addSkill(skillID, 0)...)
				}
			}
		}

	} else if !s.isEffectiveByBout(curBout - 1) && s.isEffective() {
		// 上回合没效，现在有效了
		if s.behavior.data.TriggerTimes <= 0 || s.targetTriggerTimes < s.behavior.data.TriggerTimes {
			owner := s.getOwner()
			if owner == nil {
				return actions
			}

			actions = s.onEffective()
			actions = append(actions, &clientAction{
				actID: pb.ClientAction_AddSkill,
				actMsg: &pb.ModifySkillAct{CardObjID: int32(owner.getObjID()), SkillID: s.getID()},
			})
		}
	}

	return actions
}

// 是否是该技能的触发对象
func (s *skill) isTargetTriggerObj(triggerType int, triggerTargets []iTarget, triggerCxt *triggerContext,
	cacheTargets map[int][]iTarget) bool {

	data := s.behavior.data
	if data.TriggerObj == 0 {
		return true
	}

	filter := allTargetFilters[data.TriggerObj]
	if filter == nil {
		return false
	}

	owner := s.getOwner()
	if cacheTargets == nil {
		cacheTargets = map[int][]iTarget{}
	}
	for _, triggerTarget := range triggerTargets {
		if filter.isTarget(s, owner, triggerTarget, triggerCxt, s.situation, cacheTargets) {
			return true
		}
	}

	return false
}

var awaysTriggerIgnore = map[int]struct{} {
	afterBoutEndTrigger:   struct{}{},
	preAddSkillTrigger:    struct{}{},
	loseSkillTrigger:      struct{}{},
	beforeDrawCardTrigger: struct{}{},
	afterDrawCardTrigger:  struct{}{},
	preGameEndTrigger:     struct{}{},
	preEnterBattleTrigger: struct{}{},
	surrenderTrigger:      struct{}{},
	preAddValueTrigger:    struct{}{},
	preSubValueTrigger:    struct{}{},
	afterAddValueTrigger:  struct{}{},
	afterSubValueTrigger:  struct{}{},
}
func (s *skill) canTrigger(triggerType int, triggerTargets []iTarget, triggerCxt *triggerContext,
	cacheTargets map[int][]iTarget) (bool, bool, []*clientAction) {

	//glog.Infof("canTrigger 1111111111111 %s", s)
	data := s.behavior.data
	var actions []*clientAction
	var isTargetTriggerTimesLimit bool

	if data.TriggerOpp == awaysTrigger {
		if s.behavior.isAwaysTriggerFogSkill && triggerType == preEnterBattleTrigger {
			// 大雾特殊处理，在进场前就触发大雾
		} else if _, ok := awaysTriggerIgnore[triggerType]; ok {
			// 防止死循环
			return false, isTargetTriggerTimesLimit, actions
		}
	} else if data.TriggerOpp != triggerType {
		return false, isTargetTriggerTimesLimit, actions
	}

	if s.behavior.effectiveBout == 1 {
		// 我方回合有效
		if s.situation.getCurBoutFighter().getSit() != s.getOwner().getSit() {
			return false, isTargetTriggerTimesLimit, actions
		}
	} else if s.behavior.effectiveBout == 2 {
		// 敌方回合有效
		if s.situation.getCurBoutFighter().getSit() == s.getOwner().getSit() {
			return false, isTargetTriggerTimesLimit, actions
		}
	}

	//glog.Infof("canTrigger 2222222222222 %s", s)

	if data.TriggerOpp == preAddSkillTrigger {
		// 特殊处理 preAddSkillTrigger，防止禁止获得技能的技能经常闪
		can := false
	L:	for _, skID := range triggerCxt.preAddSkillIDs {
			for _, skillID := range s.behavior.forbidAddSkillIDs {
				if skillID == skID{
					can = true
					break L
				}
			}
		}
		if !can {
			return false, isTargetTriggerTimesLimit, actions
		}
	}

	if s.behavior.totalTimes > 0 && s.triggerTotalTimes >= s.behavior.totalTimes {
		// 总触发次数
		return false, isTargetTriggerTimesLimit, actions
	}
	if data.TriggerTimes > 0 && s.targetTriggerTimes >= data.TriggerTimes {
		// emmm
		return false, isTargetTriggerTimesLimit, actions
	}

	//glog.Infof("canTrigger 3333333333333333 %s", s)

	isTarget := s.isTargetTriggerObj(triggerType, triggerTargets, triggerCxt, cacheTargets)
	if !isTarget {
		return false, isTargetTriggerTimesLimit, actions
	}

	//glog.Infof("canTrigger 4444444444444 %s", s)
	isEffective := s.isEffective()
	if data.TriggerOpp != awaysTrigger && data.TriggerObj > 0 && data.TriggerTimes > 0 {
		s.targetTriggerTimes ++
		// emmm
		if data.TriggerTimes - 1 == s.targetTriggerTimes {
			owner := s.getOwner()
			if owner == nil || !isEffective {
				return false, isTargetTriggerTimesLimit, actions
			}
			return false, isTargetTriggerTimesLimit, s.onEffective()
		} else if data.TriggerTimes == s.targetTriggerTimes {
			isTargetTriggerTimesLimit = true
		} else if data.TriggerTimes != s.targetTriggerTimes {
			return false, isTargetTriggerTimesLimit, actions
		}
	}

	return isEffective, isTargetTriggerTimesLimit, actions
}

// @return
// *pb.SkillAct  这个技能的动作
// []*clientAction  结算这个技能期间又触发的动作
// []iTarget  可能会改变触发对象
// bool  是否要说话  shit
func (s *skill) trigger(triggerType int, triggerTargets []iTarget, triggerCxt *triggerContext, result *triggerResult) (
	*pb.SkillAct, []*clientAction, []iTarget, bool) {

	//glog.Infof("skill trigger 1111111111111111 sk=%s", s)
	data := s.behavior.data
	var actions []*clientAction
	owner := s.getOwner()
	// 判断是否已失去该技能
	if !s.isEffective() {
		return nil, actions, triggerTargets, false
	}
	if !owner.hasSkill(s) {
		return nil, actions, triggerTargets, false
	}
	if s.behavior.totalTimes > 0 && s.triggerTotalTimes >= s.behavior.totalTimes {
		return nil, actions, triggerTargets, false
	}
	if data.ActRate > 0 && rand.Intn(100) >= data.ActRate {
		return nil, actions, triggerTargets, false
	}

	//glog.Infof("skill trigger 22222222222222 sk=%s", s)

// --------------------------- 判断条件 ----------------------------------
	cacheTargets := map[int][]iTarget{}
	targetCounts := []int{0, 0, 0}
	if len(s.behavior.conditions) > 0 {
		if len(data.Variable) > 3 {
			return nil, actions, triggerTargets, false
		}
		for i, variableTargetID := range data.Variable {
			if targets, ok := cacheTargets[variableTargetID]; ok {
				targetCounts[i] = len(targets)
			} else {
				targets = s.situation.getTargetMgr().findTarget(s, owner, variableTargetID, triggerCxt, cacheTargets)
				targetCounts[i] = len(targets)
			}
		}

		for _, c := range s.behavior.conditions {
			if !c.check(targetCounts[0], targetCounts[1], targetCounts[2]) {
				return nil, actions, triggerTargets, false
			}
		}
	}
// --------------------------- 判断条件 end ----------------------------------

	//glog.Infof("skill trigger 3333333333333 sk=%s", s)

	owner.triggerSkill(s.getID())
// ------------------- 这尼玛，进场时机的各种特殊情况 ---------------------------
	triggerTargetsAmount := len(triggerTargets)
	needTalk := false  // 我没话说
	if triggerType == enterBattleTrigger {
		if card, ok := owner.(*fightCard); ok && triggerCxt.isEnterBattleCard(card) {
			needTalk = true
			if card.hasTurnOverCauseByOth {
				// 如果进场曾经被别人翻了，不触发技能
				if data.TriggerObj == 1 {
					// 如果是自己的进场技，删
					actions = append(actions, owner.delSkill(s)...)
				}
				return nil, actions, triggerTargets, false
			}
		} else {
			for _, triggerTarget := range triggerTargets {
				if card, ok := triggerTarget.(*fightCard); ok && triggerCxt.isEnterBattleCard(card) {
					if card.hasTurnOverCauseByOth {
						// 如果进场曾经被别人翻了，不触发技能
						triggerTargetsAmount--
					//} else if card.getSit() != card.getInitSit() {
					} else if card.hasTurnOver {
						// 如果进场被自己翻了，不触发有sit限制的技能
						if data.TriggerObj != 1 {
							filter := allTargetFilters[data.TriggerObj]
							if filter != nil && filter.data.Side != 0 {
								triggerTargetsAmount--
							}
						}
					}
				}
			}

			if triggerTargetsAmount <= 0 && data.TriggerObj != 0 {
				return nil, actions, triggerTargets, false
			}
		}
	}
// -------------------------------- 进场几时特殊情况 end ------------------------------------------

	//glog.Infof("skill trigger 44444444444444 sk=%s", s)

// --------------------------------- 找技能目标 ----------------------------------------
	targetMgr := s.situation.getTargetMgr()
	actionTargets := map[int][]iTarget{}
	targetIDs := s.behavior.getTargetIDs(s.situation)
	for _, targetID := range targetIDs {
		targets, ok := cacheTargets[targetID]
		if !ok {
			targets = targetMgr.findTarget(s, owner, targetID, triggerCxt, cacheTargets)
		}

		for i := 0; i < len(targets); {
			t := targets[i]
			if c, ok := t.(*fightCard); ok && c.isDestroy {
				targets = append(targets[:i], targets[i+1:]...)
			} else {
				triggerCxt.addActionTargetSit(t)
				i++
			}
		}

		if len(targets) <= 0 {
			continue
		}
		actionTargets[targetID] = targets
	}

	//glog.Infof("skill trigger action targets=%v", actionTargets)

	if len(actionTargets) <= 0 && (s.fogTargets == nil || s.fogTargets.Size() <= 0) {
		return nil, actions, triggerTargets, false
	}
// --------------------------------- 找技能目标 end ----------------------------------------

	//glog.Infof("skill trigger 55555555555555 sk=%s", s)

// --------------------------------- 对象计数 ----------------------------------------
	variableAmount := 0
	if len(data.Variable) > 0 {
		if targets, ok := actionTargets[data.Variable[0]]; ok {
			variableAmount = len(targets)
		} else if targets, ok = cacheTargets[data.Variable[0]]; ok {
			variableAmount = len(targets)
		} else {
			targets = targetMgr.findTarget(s, owner, data.Variable[0], triggerCxt, cacheTargets)
			variableAmount = len(targets)
		}
	}
// --------------------------------- 对象计数 end ----------------------------------------

	//glog.Infof("skill trigger, skid=%d, owner=%d, target=%v", s.getID(), s.ownerObjID, actionTargets)

	s.triggerTotalTimes++
	triggerCxt.setTriggerTargetSit(s.situation)

	var skillAct *pb.SkillAct
	skillAct, actions, triggerTargets = s.behavior.invoke(s, triggerType, triggerTargets, actionTargets, variableAmount,
		triggerCxt, result, s.situation)
	//glog.Infof("skill trigger 66666666666 sk=%s, actions=%s", s, actions)

	if s.behavior.totalTimes > 0 && s.triggerTotalTimes >= s.behavior.totalTimes {
		owner = s.getOwner()
		if owner != nil {
			actions = append(actions, owner.delSkill(s)...)
		}
	}

	if data.TriggerOpp != enterBattleTrigger {
		needTalk = false
	}

	return skillAct, actions, triggerTargets, needTalk
}
