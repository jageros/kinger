package main

import (
	"kinger/proto/pb"
	"kinger/common/consts"
	"kinger/gopuppy/common"
	"math/rand"
	"kinger/gopuppy/common/utils"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/apps/logic"
	//"kinger/gopuppy/common/glog"
	"kinger/gamedata"
)

type playInHandCard struct {
	objID int
	sit int
}

func (c *playInHandCard) copy() *playInHandCard {
	c2 := *c
	return &c2
}

func (c *playInHandCard) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("objID", c.objID)
	attr.SetInt("sit", c.sit)
	return attr
}

func (c *playInHandCard) restoredFromAttr(attr *attribute.MapAttr) {
	c.objID = attr.GetInt("objID")
	c.sit = attr.GetInt("sit")
}

type tauntBuff struct {
	skillID int32
	tauntFighterObjID int
	ownerObjID        int
	grids             []int
}

func (t *tauntBuff) copy() *tauntBuff {
	t2 := *t
	return &t2
}

func (t *tauntBuff) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt32("skillID", t.skillID)
	attr.SetInt("tauntFighterObjID", t.tauntFighterObjID)
	attr.SetInt("ownerObjID", t.ownerObjID)
	gridsAttr := attribute.NewListAttr()
	attr.SetListAttr("grids", gridsAttr)
	for _, gridID := range t.grids {
		gridsAttr.AppendInt(gridID)
	}
	return attr
}

func (t *tauntBuff) restoredFromAttr(attr *attribute.MapAttr) {
	t.skillID = attr.GetInt32("skillID")
	t.tauntFighterObjID = attr.GetInt("tauntFighterObjID")
	t.ownerObjID = attr.GetInt("ownerObjID")
	gridsAttr := attr.GetListAttr("grids")
	gridsAttr.ForEachIndex(func(index int) bool {
		t.grids = append(t.grids, gridsAttr.GetInt(index))
		return true
	})
}

func (t *tauntBuff) canPlayCard(f *fighter, gridID int, situation *battleSituation) bool {
	if f.getObjID() != t.tauntFighterObjID {
		return true
	}

	skOwner := situation.getTargetMgr().getTarget(t.ownerObjID)
	if skOwner == nil {
		return true
	}

	c, ok := skOwner.(iCaster)
	if !ok {
		return true
	}

	if !c.hasSkillByID(t.skillID) {
		return true
	}

	var tauntGrids []int
	for _, gid := range t.grids {
		t := situation.getTargetInGrid(gid)
		if t != nil && t.getType() == stEmptyGrid {
			tauntGrids = append(tauntGrids, gid)
		}
	}

	if len(tauntGrids) > 0 {
		for _, gid := range tauntGrids {
			if gid == gridID {
				return true
			}
		}
		return false
	} else {
		return true
	}
}

type battleSituation struct {
	maxObjID int
	state int
	grids []int  // []objID
	curBout int
	scale int
	battleRes int
	maxHandCardAmount int
	gridColumn int
	gridsAmount int
	aiThinking bool
	playCardQueue []*playInHandCard  // 出牌队列
	taunts []*tauntBuff
	bonusObj *bonus
	isWonderful1 bool
	isWonderful2 bool
	battleType int

	fighter1       *fighter
	fighter2       *fighter
	curBoutFighter *fighter
	nextBoutFighter    *fighter
	doActionFighter *fighter

	fort1 *fortCaster
	fort2 *fortCaster

	targetMgr *skillTargetMgr
	triggerMgr *skillTriggerMgr
}

func newBattleSituation(fighterData1, fighterData2 *pb.FighterData, battleID common.UUid, upperType, bonusType, scale,
	battleRes int, needFortifications bool, battleType int, seasonData *gamedata.SeasonPvp) *battleSituation {

	situation := &battleSituation{
		battleRes: battleRes,
		state: bsCreate,
		battleType: battleType,
	}

	situation.bonusObj = newBonus(bonusType, situation)
	situation.setBattleScale(scale)
	situation.targetMgr = newSkillTargetMgr(situation)
	situation.triggerMgr = newSkillTriggerMgr(situation)
	situation.createGrids()
	hand1, hand2, sideboard1, sideboard2 := situation.createHands(fighterData1, fighterData2, seasonData)
	situation.createFighter(battleID, fighterData1, fighterData2, hand1, hand2, sideboard1, sideboard2)
	situation.chooseFirstHand(upperType)
	situation.createGridCards(fighterData1, fighterData2, needFortifications, seasonData)
	situation.createFort(fighterData1, fighterData2, seasonData)

	return situation
}

func (bs *battleSituation) createGrids() {
	// 棋盘9宫格
	if bs.scale == consts.BtScale33 {
		bs.gridsAmount = 9
	} else if bs.scale == consts.BtScale43 {
		bs.gridsAmount = 12
	} else {
		bs.gridsAmount = 15
	}
	bs.grids = make([]int, bs.gridsAmount, bs.gridsAmount)
	for i := 0; i < bs.gridsAmount; i++ {
		grid := newGrid(bs.genObjID(), i, bs)
		bs.targetMgr.addTarget(grid)
		bs.grids[i] = grid.getObjID()
	}
}

func (bs *battleSituation) createHands(fighterData1, fighterData2 *pb.FighterData, seasonData *gamedata.SeasonPvp) (
	hand1, hand2 []*fightCard, sideboard1, sideboard2 []uint32) {

	// TODO
	seasonData = nil
	// 双方手牌
	hands := [][]*fightCard{hand1, hand2}
	sideboards := [][]uint32{sideboard1, sideboard2}
	fighterDatas := []*pb.FighterData{fighterData1, fighterData2}
	handType := pb.BattleHandType_UnknowType
	if seasonData != nil {
		handType = pb.BattleHandType(seasonData.HandCardType[0])
	}

	if handType == pb.BattleHandType_UnknowType || handType == pb.BattleHandType_Default {
		for i, fighterData := range fighterDatas {
			handData := fighterData.GetHandCards()
			hand := hands[i]
			sideboard := sideboards[i]
			for _, c := range handData {
				if len(hand) >= bs.maxHandCardAmount {
					sideboard = append(sideboard, c.GCardID)
				} else {
					c := newFightCard(bs.genObjID(), c.GCardID, c.Skin, c.Equip, bs)
					bs.targetMgr.addTarget(c)
					hand = append(hand, c)
				}
			}
			hands[i] = hand
			utils.ShuffleUInt32(sideboard)
			sideboards[i] = sideboard
		}

	}

	return hands[0], hands[1], sideboards[0], sideboards[1]
}

func (bs *battleSituation) createFighter(battleID common.UUid, fighterData1, fighterData2 *pb.FighterData, hand1,
	hand2 []*fightCard, sideboard1, sideboard2 []uint32) {

	// 2个玩家
	hands := [][]*fightCard{hand1, hand2}
	fighterDatas := []*pb.FighterData{fighterData1, fighterData2}
	sideboards := [][]uint32{sideboard1, sideboard2}

	for i, fighterData := range fighterDatas {
		hand := hands[i]
		f := newFighter(bs.genObjID(), fighterData, battleID, hand, sideboards[i], bs)
		bs.targetMgr.addTarget(f)
		if i == 0 {
			f.setSit(consts.SitOne)
			bs.fighter1 = f
		} else {
			bs.fighter2 = f
			f.setSit(consts.SitTwo)
		}
	}
}

func (bs *battleSituation) genFortifications(seasonData *gamedata.SeasonPvp) []*pb.InGridCard {
	var cardInfos [][]int
	if seasonData == nil || len(seasonData.Fortifications) == 0 {
		cardInfos = [][]int{ []int{1, 293}, []int{2, 294}, []int{3, 295} }
	} else {
		cardInfos = seasonData.Fortifications
	}

	r := rand.Intn(len(cardInfos))
	gcardID := uint32(cardInfos[r][1])
	amount := cardInfos[r][0]

	gridAmount := len(bs.grids)
	gridIDs := make([]int, gridAmount, gridAmount)
	for i := 0; i < gridAmount; i++ {
		gridIDs[i] = i
	}
	utils.ShuffleInt(gridIDs)

	if amount >= gridAmount {
		amount = gridAmount - 1
	}
	gridIDs = gridIDs[:amount]
	gridCards := make([]*pb.InGridCard, amount, amount)
	for i, gridID := range gridIDs {
		gridCards[i] = &pb.InGridCard{
			GCardID: gcardID,
			GridID:  int32(gridID),
		}
	}
	return gridCards
}

func (bs *battleSituation) createGridCards(fighterData1, fighterData2 *pb.FighterData, needFortifications bool,
	seasonData *gamedata.SeasonPvp) {
	// 双方在场上的牌
	fighterDatas := []*pb.FighterData{fighterData1, fighterData2}
	fighters := []*fighter{bs.fighter1, bs.fighter2}

	for i, fighterData := range fighterDatas {
		fighter := fighters[i]
		gridCards := fighterData.GetGridCards()
		if len(gridCards) == 0 && needFortifications && !fighter.isFirstHand {
			gridCards = bs.genFortifications(seasonData)
		}

		for _, cardData := range gridCards {
			c := newFightCard(bs.genObjID(), cardData.GCardID, "", "", bs)
			c.setTargetType(stInDesk)
			uid := fighter.getUid()
			c.setController(uid)
			c.setOwner(uid)
			bs.targetMgr.addTarget(c)

			objId := bs.grids[cardData.GridID]
			t := bs.targetMgr.getTarget(objId)
			gridObj := t.(*deskGrid)
			bs.targetMgr.delSkillTarget(objId)

			bs.grids[cardData.GridID] = c.getObjID()
			c.setGrid(gridObj)
			bs.triggerMgr.casterEnterBattle(c)
		}
	}

}

func (bs *battleSituation) createFort(fighterData1, fighterData2 *pb.FighterData, seasonData *gamedata.SeasonPvp)  {
	// 双方城防
	fighterDatas := []*pb.FighterData{fighterData1, fighterData2}
	if seasonData != nil {
		fighterData1.CasterSkills = append(fighterData1.CasterSkills, seasonData.DefenseSkills1...)
		fighterData2.CasterSkills = append(fighterData2.CasterSkills, seasonData.DefenseSkills2...)
	}
	for i, fighterData := range fighterDatas {
		fc := newFortCaster(bs.genObjID(), common.UUid(fighterData.Uid), fighterData.CasterSkills, bs)
		bs.targetMgr.addTarget(fc)
		bs.triggerMgr.casterEnterBattle(fc)
		if i == 0 {
			bs.fort1 = fc
		} else {
			bs.fort2 = fc
		}
	}
}

func (bs *battleSituation) chooseFirstHand(upperType int) {
	if upperType == 3 {
		// 随机先手
		fs := []interface{}{bs.fighter1, bs.fighter2}
		utils.Shuffle(fs)
		r := rand.Intn(100)
		if r < 50 {
			bs.curBoutFighter = fs[0].(*fighter)
		} else {
			bs.curBoutFighter = fs[1].(*fighter)
		}
	} else if upperType == 1 {
		bs.curBoutFighter = bs.fighter1
	} else {
		bs.curBoutFighter = bs.fighter2
	}
	bs.curBoutFighter.isFirstHand = true
}

func (bs *battleSituation) getTargetMgr() *skillTargetMgr {
	return bs.targetMgr
}

func (bs *battleSituation) getTriggerMgr() *skillTriggerMgr {
	return bs.triggerMgr
}

func (bs *battleSituation) genObjID() int {
	bs.maxObjID++
	return bs.maxObjID
}

func (bs *battleSituation) getFighter(uid common.UUid) *fighter {
	if bs.fighter1.getUid() == uid {
		return bs.fighter1
	} else {
		return bs.fighter2
	}
}

func (bs *battleSituation) getFighterBySit(sit int) *fighter {
	if sit == consts.SitOne {
		return bs.fighter1
	} else {
		return bs.fighter2
	}
}

func (bs *battleSituation) getFighter1() *fighter {
	return bs.fighter1
}

func (bs *battleSituation) getFighter2() *fighter {
	return bs.fighter2
}

func (bs *battleSituation) getCurBoutFighter() *fighter {
	return bs.curBoutFighter
}

func (bs *battleSituation) getDoActionFighter() *fighter {
	return bs.doActionFighter
}

func (bs *battleSituation) getEnemyFighter(f *fighter) *fighter {
	if bs.fighter1.getUid() == f.getUid() {
		return bs.fighter2
	} else {
		return bs.fighter1
	}
}

func (bs *battleSituation) getCurBout() int {
	return bs.curBout
}

func (bs *battleSituation) setState(state int) {
	bs.state = state
}

func (bs *battleSituation) getState() int {
	return bs.state
}

func (bs *battleSituation) getBattleScale() int {
	return bs.scale
}

func (bs *battleSituation) setBattleScale(scale int) {
	switch scale {
	case consts.BtScale33:
		bs.scale = scale
		bs.maxHandCardAmount = 5
		bs.gridColumn = 3
	case consts.BtScale43:
		bs.scale = scale
		bs.maxHandCardAmount = 7
		bs.gridColumn = 4
	case consts.BtScale53:
		bs.scale = scale
		bs.maxHandCardAmount = 8
		bs.gridColumn = 5
	default:
		bs.scale = consts.BtScale33
		bs.maxHandCardAmount = 5
		bs.gridColumn = 3
	}
}

func (bs *battleSituation) getMaxHandCardAmount() int {
	return bs.maxHandCardAmount
}

func (bs *battleSituation) getGridColumn() int {
	return bs.gridColumn
}

func (bs *battleSituation) copy() *battleSituation {
	b := *bs
	cpy := &b
	cpy.bonusObj = nil
	cpy.aiThinking = true
	cpy.grids = make([]int, bs.gridsAmount)
	copy(cpy.grids, bs.grids)

	cpy.playCardQueue = make([]*playInHandCard, len(bs.playCardQueue))
	for i, c := range bs.playCardQueue {
		cpy.playCardQueue[i] = c.copy()
	}

	cpy.taunts = make([]*tauntBuff, len(bs.taunts))
	for i, t := range bs.taunts {
		cpy.taunts[i] = t.copy()
	}

	cpy.targetMgr = bs.targetMgr.copy(cpy)
	cpy.triggerMgr = bs.triggerMgr.copy(cpy)

	cpy.fighter1 = cpy.targetMgr.getTarget(bs.fighter1.getObjID()).(*fighter)
	cpy.fighter2 = cpy.targetMgr.getTarget(bs.fighter2.getObjID()).(*fighter)
	if bs.curBoutFighter.getUid() == cpy.fighter1.getUid() {
		cpy.curBoutFighter = cpy.fighter1
	} else {
		cpy.curBoutFighter = cpy.fighter2
	}
	if bs.nextBoutFighter != nil {
		if bs.nextBoutFighter.getUid() == cpy.fighter1.getUid() {
			cpy.nextBoutFighter = cpy.fighter1
		} else {
			cpy.nextBoutFighter = cpy.fighter2
		}
	}

	cpy.fort1 = cpy.targetMgr.getTarget(bs.fort1.getObjID()).(*fortCaster)
	cpy.fort2 = cpy.targetMgr.getTarget(bs.fort2.getObjID()).(*fortCaster)

	return cpy
}

func (bs *battleSituation) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("maxObjID", bs.maxObjID)
	attr.SetInt("state", bs.state)
	attr.SetInt("curBout", bs.curBout)
	attr.SetInt("scale", bs.scale)
	attr.SetInt("battleRes", bs.battleRes)
	attr.SetInt("maxHandCardAmount", bs.maxHandCardAmount)
	attr.SetInt("gridColumn", bs.gridColumn)
	attr.SetInt("gridsAmount", bs.gridsAmount)
	attr.SetInt("f1ObjID", bs.fighter1.getObjID())
	attr.SetInt("f2ObjID", bs.fighter2.getObjID())
	attr.SetInt("curFighterObjID", bs.curBoutFighter.getObjID())
	attr.SetBool("isWonderful1", bs.isWonderful1)
	attr.SetBool("isWonderful2", bs.isWonderful2)
	attr.SetInt("fr1ObjID", bs.fort1.getObjID())
	attr.SetInt("fr2ObjID", bs.fort2.getObjID())
	if bs.nextBoutFighter != nil {
		attr.SetInt("nextFighterObjID", bs.nextBoutFighter.getObjID())
	}

	gridsAttr := attribute.NewListAttr()
	attr.SetListAttr("grids", gridsAttr)
	for _, objID := range bs.grids {
		gridsAttr.AppendInt(objID)
	}

	playCardQueueAttr := attribute.NewListAttr()
	attr.SetListAttr("playCardQueue", playCardQueueAttr)
	for _, c := range bs.playCardQueue {
		playCardQueueAttr.AppendMapAttr(c.packAttr())
	}

	tauntsAttr := attribute.NewListAttr()
	attr.SetListAttr("taunts", tauntsAttr)
	for _, t := range bs.taunts {
		tauntsAttr.AppendMapAttr(t.packAttr())
	}

	if bs.bonusObj != nil {
		attr.SetInt("bonusType", bs.bonusObj.type_)
	}

	attr.SetMapAttr("targetMgr", bs.targetMgr.packAttr())
	attr.SetListAttr("triggerMgr", bs.triggerMgr.packAttr())

	return attr
}

func (bs *battleSituation) restoredFromAttr(attr *attribute.MapAttr, agent *logic.PlayerAgent, battleType int) {
	bs.battleType = battleType
	bs.maxObjID = attr.GetInt("maxObjID")
	bs.state = attr.GetInt("state")
	bs.curBout = attr.GetInt("curBout")
	bs.scale = attr.GetInt("scale")
	bs.battleRes = attr.GetInt("battleRes")
	bs.maxHandCardAmount = attr.GetInt("maxHandCardAmount")
	bs.gridColumn = attr.GetInt("gridColumn")
	bs.gridsAmount = attr.GetInt("gridsAmount")
	bs.isWonderful1 = attr.GetBool("isWonderful1")
	bs.isWonderful2 = attr.GetBool("isWonderful2")

	gridsAttr := attr.GetListAttr("grids")
	gridsAttr.ForEachIndex(func(index int) bool {
		bs.grids = append(bs.grids, gridsAttr.GetInt(index))
		return true
	})

	playCardQueueAttr := attr.GetListAttr("playCardQueue")
	playCardQueueAttr.ForEachIndex(func(index int) bool {
		cAttr := playCardQueueAttr.GetMapAttr(index)
		c := &playInHandCard{}
		c.restoredFromAttr(cAttr)
		bs.playCardQueue = append(bs.playCardQueue, c)
		return true
	})

	tauntsAttr := attr.GetListAttr("taunts")
	tauntsAttr.ForEachIndex(func(index int) bool {
		tAttr := tauntsAttr.GetMapAttr(index)
		t := &tauntBuff{}
		t.restoredFromAttr(tAttr)
		bs.taunts = append(bs.taunts, t)
		return true
	})

	bs.bonusObj = newBonus(attr.GetInt("bonusType"), bs)
	bs.targetMgr = &skillTargetMgr{}
	bs.targetMgr.restoredFromAttr(attr.GetMapAttr("targetMgr"), bs)
	bs.triggerMgr = &skillTriggerMgr{}
	bs.triggerMgr.restoredFromAttr(attr.GetListAttr("triggerMgr"), bs)

	bs.fighter1 = bs.targetMgr.getTarget(attr.GetInt("f1ObjID")).(*fighter)
	bs.fighter2 = bs.targetMgr.getTarget(attr.GetInt("f2ObjID")).(*fighter)
	bs.curBoutFighter = bs.targetMgr.getTarget(attr.GetInt("curFighterObjID")).(*fighter)
	nextBoutFighter := bs.targetMgr.getTarget(attr.GetInt("nextFighterObjID"))
	if nextBoutFighter != nil {
		bs.nextBoutFighter = nextBoutFighter.(*fighter)
	}
	bs.fighter1.onRestored(agent)
	bs.fighter2.onRestored(agent)

	bs.fort1 = bs.targetMgr.getTarget(attr.GetInt("fr1ObjID")).(*fortCaster)
	bs.fort2 = bs.targetMgr.getTarget(attr.GetInt("fr2ObjID")).(*fortCaster)
}

func (bs *battleSituation) boutBegin() []*clientAction {
	bs.state = bsInBout
	bs.doActionFighter = nil
	if bs.nextBoutFighter != nil {
		// 某个玩家额外下一张牌，不会触发回合开始的种种东西
		bs.nextBoutFighter = nil
		return []*clientAction{}
	}

	actions := bs.targetMgr.boutBegin()

	acts, _, _ := bs.triggerMgr.trigger(map[int][]iTarget{boutBeginTrigger: []iTarget{bs.curBoutFighter}}, &triggerContext{
		triggerType: boutBeginTrigger,
	})
	actions = append(actions, acts...)

	acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{afterBoutBeginTrigger: []iTarget{bs.curBoutFighter}}, &triggerContext{
		triggerType: afterBoutBeginTrigger,
	})
	actions = append(actions, acts...)

	return actions
}

func (bs *battleSituation) boutEnd() {
	bs.state = bsWaitClient
	if bs.nextBoutFighter != nil {
		bs.curBoutFighter = bs.nextBoutFighter
	} else if bs.curBoutFighter.getUid() == bs.fighter1.getUid() {
		bs.curBoutFighter = bs.fighter2
	} else {
		bs.curBoutFighter = bs.fighter1
	}
}

func (bs *battleSituation) getGridsTarget() []iTarget {
	tars := make([]iTarget, bs.gridsAmount, bs.gridsAmount)
	for i, objID := range bs.grids {
		t := bs.targetMgr.getTarget(objID)
		tars[i] = t
	}
	return tars
}

func (bs *battleSituation) getTargetInGrid(gridID int) iTarget {
	if gridID < 0 || gridID >= bs.gridsAmount {
		return nil
	}
	return bs.targetMgr.getTarget(bs.grids[gridID])
}

// 在pos方向上，距离gridID  n格
func (bs *battleSituation) getPosTargetByGrid(gridID, pos, n int) iTarget {
	gridID2 := -1
	switch pos {
	case consts.UP:
		gridID2 = gridID - n * bs.gridColumn
		if gridID2 < 0 || gridID2 >= bs.gridsAmount {
			gridID2 = -1
		}
	case consts.DOWN:
		gridID2 = gridID + n * bs.gridColumn
		if gridID2 < 0 || gridID2 >= bs.gridsAmount {
			gridID2 = -1
		}
	case consts.LEFT:
		gridID2 = gridID - n
		if !(gridID2 >= 0 && gridID2 < bs.gridsAmount && (gridID2/bs.gridColumn == gridID/bs.gridColumn)) {
			gridID2 = -1
		}
	case consts.RIGHT:
		gridID2 = gridID + n
		if !(gridID2 >= 0 && gridID2 < bs.gridsAmount && (gridID2/bs.gridColumn == gridID/bs.gridColumn)) {
			gridID2 = -1
		}
	default:
		if gridID >= 0 && gridID < bs.gridsAmount {
			gridID2 = gridID
		}
	}

	if gridID2 >= 0 {
		return bs.targetMgr.getTarget(bs.grids[gridID2])
	}

	return nil
}

func (bs *battleSituation) randomUnUseGrid(f *fighter) int {
	var grids []int
	for gridID, objID := range bs.grids {
		t := bs.targetMgr.getTarget(objID)
		if t != nil && t.getType() == stEmptyGrid && bs.canPlayCard(f, gridID) {
			grids = append(grids, t.getGrid())
		}
	}

	amount := len(grids)
	if amount > 0 {
		if f.getPvpLevel() < 5 {
			return grids[0]
		}
		return grids[rand.Intn(amount)]
	} else {
		return -1
	}
}

func (bs *battleSituation) checkResult() (common.UUid, bool) {
	var p1 int
	var p2 int
	isEnd := bs.fighter1.getHandAmount() <= 0 && bs.fighter2.getHandAmount() <= 0
	for _, objID := range bs.grids {
		t := bs.targetMgr.getTarget(objID)
		if t == nil {
			continue
		}
		if t.getType() == stEmptyGrid {
			if !isEnd {
				return 0, false
			} else {
				continue
			}
		} else {
			if t.getSit() == bs.fighter1.getSit() {
				p1++
			} else {
				p2++
			}
		}
	}
	if p1 > p2 {
		return bs.fighter1.getUid(), false
	} else {
		return bs.fighter2.getUid(), p1 == p2
	}
}

func (bs *battleSituation) isAiThinking() bool {
	return bs.aiThinking || (bs.fighter1.isRobot && bs.fighter2.isRobot)
}

// 平局补牌
func (bs *battleSituation) drawnGame() []*clientAction {
	var acts []*clientAction
	triggerCxt := &triggerContext{}
	targets := []iTarget{bs.fighter1, bs.fighter2}
	triggerCxt.addActionTargets(27, targets)
	skillActs1, acts1, _, _, _ := drawnGameAction.invoke(nil, bs.fort1, preGameEndTrigger, []iTarget{},
		targets, 1, triggerCxt, &triggerResult{}, bs)

	skillAct1 := &pb.SkillAct{Owner: int32(bs.fort1.getObjID())}
	for _, cliAct := range skillActs1 {
		actData := cliAct.packMsg()
		skillAct1.Actions = append(skillAct1.Actions, actData)
	}

	acts = append(acts, &clientAction{
		actID: pb.ClientAction_Skill,
		actMsg: skillAct1,
	})
	acts = append(acts, acts1...)

	/*
	skillActs2, acts2, _, _, _ := drawnGameAction.invoke(nil, bs.fort2, preGameEndTrigger, []iTarget{},
		[]iTarget{bs.fighter2}, 1, &triggerContext{}, &triggerResult{}, bs)

	skillAct2 := &pb.SkillAct{Owner: int32(bs.fort2.getObjID())}
	for _, cliAct := range skillActs2 {
		actData := cliAct.packMsg()
		skillAct2.Actions = append(skillAct2.Actions, actData)
	}

	acts = append(acts, &clientAction{
		actID: pb.ClientAction_Skill,
		actMsg: skillAct2,
	})
	acts = append(acts, acts2...)
	*/

	return acts
}

func (bs *battleSituation) doAction(f *fighter, useCardObjID, gridID int) (*pb.FightBoutResult, *fightCard, int32) {
	if bs.state != bsInBout {
		return nil, nil, 1
	}
	var actions []*clientAction
	var card *fightCard
	var errcode int32
	var needTalk bool
	var isInFog bool
	var isPublicEnemy bool
	if useCardObjID > 0 {
		actions, card, needTalk, errcode, isInFog, isPublicEnemy = bs.tryUseCard(f, useCardObjID, gridID)
		if card == nil {
			return nil, nil, errcode
		}
	}

	if bs.nextBoutFighter == nil {
		// 没有人可以多下一张牌，回合结束
		acts, _, _ := bs.triggerMgr.trigger(map[int][]iTarget{boutEndTrigger: []iTarget{bs.curBoutFighter}}, &triggerContext{})
		actions = append(actions, acts...)

		acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{afterBoutEndTrigger: []iTarget{bs.curBoutFighter}}, &triggerContext{})
		actions = append(actions, acts...)

		bs.curBout++
		actions = append(actions, bs.triggerMgr.boutEnd()...)
		actions = append(actions, bs.fighter1.boutEnd()...)
		actions = append(actions, bs.fighter2.boutEnd()...)
	}

	winUid, isDrawn := bs.checkResult()
	reply := &pb.FightBoutResult{
		BoutUid: uint64(f.getUid()),
		UseCardObjID: int32(useCardObjID),
		TargetGridId: int32(gridID),
		CardNeedTalk: needTalk,
		WinUid: uint64(winUid),
		IsUseCardInFog: isInFog,
		IsUseCardPublicEnemy: isPublicEnemy,
		//BattleID: uint64(f.getBattle().getBattleID()),
	}

	if !bs.isAiThinking() {
		reply.BattleID = uint64(f.getBattleID())
	}

	if reply.WinUid != 0 {
		//acts, _, _ := bs.triggerMgr.trigger(map[int][]iTarget{preGameEndTrigger: []iTarget{}}, &triggerContext{})
		if isDrawn {
			actions = append(actions, bs.drawnGame()...)
			reply.WinUid = 0
		} else {

			acts, _, _ := bs.triggerMgr.trigger(map[int][]iTarget{preGameEndTrigger: []iTarget{}}, &triggerContext{})
			actions = append(actions, acts...)
			winUid, _ = bs.checkResult()
			reply.WinUid = uint64(winUid)
		}
	}

	if reply.WinUid == 0 {
		act := bs.fighter1.drawSideboardCard()
		if act != nil {
			actions = append(actions, act)
		}
		act = bs.fighter2.drawSideboardCard()
		if act != nil {
			actions = append(actions, act)
		}
	}

	if !bs.isAiThinking() {
		// --------- 观星
		var guanxingUids []uint64
		if bs.fighter1.isGuanxing() {
			guanxingUids = append(guanxingUids, uint64(bs.fighter1.getUid()))
		}
		if bs.fighter2.isGuanxing() {
			guanxingUids = append(guanxingUids, uint64(bs.fighter2.getUid()))
		}
		if len(guanxingUids) > 0 {
			actions = append(actions, &clientAction{
				actID: pb.ClientAction_Guanxing,
				actMsg: &pb.GuanxingAct{
					Uids:            guanxingUids,
					SitOneDrawCards: bs.fighter1.packDrawCardShadow(),
					SitTwoDrawCards: bs.fighter2.packDrawCardShadow(),
				},
			})
		}
	}

	for _, act := range actions {
		reply.Actions = append(reply.Actions, act.packMsg())
	}

	bs.doActionFighter = f
	return reply, card, 0
}

func (bs *battleSituation) canPlayCard(f *fighter, gridID int) bool {
	// 嘲讽
	for _, t := range bs.taunts {
		if !t.canPlayCard(f, gridID, bs) {
			return false
		}
	}
	return true
}

func (bs *battleSituation) tryUseCard(f *fighter, useCardObjID, gridID int) (acts []*clientAction, card *fightCard,
	needTalk bool, errcode int32, isInFog, isPublicEnemy bool) {

	uid := bs.curBoutFighter.getUid()
	if f.getUid() != uid {
		errcode = 1
		return
	}

	card = f.getHandCard(useCardObjID)
	if card == nil {
		errcode = 2
		return
	}

	var gridObj *deskGrid
	if gridID >= 0 && gridID < bs.gridsAmount {
		objID := bs.grids[gridID]
		t := bs.targetMgr.getTarget(objID)
		if t != nil && t.getType() == stEmptyGrid {
			gridObj = t.(*deskGrid)
		}
	}
	if gridObj == nil {
		card = nil
		errcode = 3
		return
	}

	// 嘲讽
	if !bs.canPlayCard(f, gridID) {
		card = nil
		errcode = 100
		return
	}

	f.delHandCard(useCardObjID)
	card.isPlayInHand = true
	acts = bs.preCardEnterBattle(card, gridObj, f, nil)
	var acts2 []*clientAction
	acts2, needTalk, isInFog, isPublicEnemy = bs.cardEnterBattle(card, gridObj, f, nil)
	acts = append(acts, acts2...)
	return
}

// if oldCard != nil   reEnterDesk
func (bs *battleSituation) addPlayInHandCard(card *fightCard, oldCard *fightCard, f *fighter) {
	ic := &playInHandCard {
		objID:       card.getObjID(),
	}

	if oldCard != nil {
		for i, c := range bs.playCardQueue {
			if c.objID == oldCard.getObjID() {
				bs.playCardQueue[i] = ic
				ic.sit = c.sit
				break
			}
		}
	} else {
		if f != nil {
			ic.sit = f.getSit()
		}
		bs.playCardQueue = append(bs.playCardQueue, ic)
	}

	idx := len(bs.playCardQueue)
	card.setPlayCardIdx(idx)
}

func (bs *battleSituation) getPlayCardQueueIdx() int {
	return len(bs.playCardQueue)
}

func (bs *battleSituation) getPrePlayCard(relativeCard iCaster, triggerType, side, n int) int {
	if n <= 0 {
		return 0
	}
	if len(bs.playCardQueue) <= 0 {
		return 0
	}

	i := len(bs.playCardQueue) - 1
	if triggerType == enterBattleTrigger || triggerType == preEnterBattleTrigger {
		if bs.playCardQueue[i].objID == relativeCard.getObjID() {
			i--
		}
	}

	for ; i >= 0; i-- {
		c := bs.playCardQueue[i]
		if side != 0 {
			if side == sOwn && c.sit != relativeCard.getSit() {
				continue
			}
			if side == sEnemy && c.sit == relativeCard.getSit() {
				continue
			}
			if side == sInitOwn1 && c.sit != relativeCard.getInitSit() {
				continue
			}
			if side == sInitEnemy1 && c.sit == relativeCard.getInitSit() {
				continue
			}
		}
		n--
		if n == 0 {
			return c.objID
		}
	}

	return 0
}

func (bs *battleSituation) getNextPlayCard(relativeCard iCaster, idx, side, n int) int {
	if n <= 0 {
		return 0
	}
	if len(bs.playCardQueue) <= idx {
		return 0
	}

	i := idx
	for ; i < len(bs.playCardQueue); i++ {
		c := bs.playCardQueue[i]
		if side != 0 {
			if side == sOwn && c.sit != relativeCard.getSit() {
				continue
			}
			if side == sEnemy && c.sit == relativeCard.getSit() {
				continue
			}
			if side == sInitOwn1 && c.sit != relativeCard.getInitSit() {
				continue
			}
			if side == sInitEnemy1 && c.sit == relativeCard.getInitSit() {
				continue
			}
		}
		n--
		if n == 0 {
			return c.objID
		}
	}

	return 0
}

func (bs *battleSituation) setGrid(gridID int, t iTarget) {
	if gridID >= 0 && gridID < bs.gridsAmount {
		bs.grids[gridID] = t.getObjID()
	}
}

// if oldCard != nil   reEnterBattle
func (bs *battleSituation) preCardEnterBattle(card *fightCard, grid *deskGrid, f *fighter, oldCard *fightCard) []*clientAction {
	if card.isPlayInHand {
		bs.addPlayInHandCard(card, oldCard, f)
	}

	bs.targetMgr.delSkillTarget(grid.getObjID())
	bs.grids[grid.getGrid()] = card.getObjID()
	card.setTargetType(stInDesk)
	card.setGrid(grid)
	return bs.triggerMgr.casterEnterBattle(card)
}

// if oldCard != nil   reEnterBattle
func (bs *battleSituation) cardEnterBattle(card *fightCard, grid *deskGrid, f *fighter, oldCard *fightCard) (
	[]*clientAction, bool, bool, bool) {

	var actions []*clientAction
	acts, _, needTalk1 := bs.triggerMgr.trigger(map[int][]iTarget{preEnterBattleTrigger: []iTarget{card}}, &triggerContext{
		triggerType:     preEnterBattleTrigger,
		enterBattleCard: card,
	})
	actions = append(actions, acts...)
	isInFog := card.isInFog()

	card = card.getCopyTarget().(*fightCard)
	if !card.isInBattle() {
		return actions, needTalk1, card.isInFog(), card.isPublicEnemy()
	}

	acts2, _, needTalk2 := bs.triggerMgr.trigger(map[int][]iTarget{enterBattleTrigger: []iTarget{card}}, &triggerContext{
		triggerType:     enterBattleTrigger,
		enterBattleCard: card,
	})
	actions = append(actions, acts2...)

	isPublicEnemy := card.isPublicEnemy()
	card = card.getCopyTarget().(*fightCard)
	if !card.isInBattle() {
		return actions, needTalk1 || needTalk2, isInFog, isPublicEnemy
	}

	acts3, needTalk3 := bs.attack(card, nil, false)
	actions = append(actions, acts3...)
	return actions, needTalk1 || needTalk2 || needTalk3, isInFog, isPublicEnemy
}

func (bs *battleSituation) attack(card *fightCard, attackCxt *attackContext, isCauseBySkill bool) ([]*clientAction, bool) {
	var actions []*clientAction
	//glog.Infof("attack 111111111111111 card=%s", card)
	if !card.isInBattle() {
		// 可能被消灭或回手了
		return actions, false
	}

	if attackCxt == nil {
		attackCxt = &attackContext{}
	}
// ------------------ 找攻击目标 ------------------
	//var triggerRs *triggerResult
	actions, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{findAttackTargetTrigger: []iTarget{card}}, &triggerContext{
		triggerType: findAttackTargetTrigger,
		attackCxt: attackCxt,
		attackCard: card,
	})

	//glog.Infof("attack 2222222222222222222 card=%s", card)

	//if !isCauseBySkill && ((card.hasTurnOver && card.getSit() != card.getInitSit()) || card.hasTurnOverCauseByOth) {
	if !isCauseBySkill && card.hasTurnOver {
		// 如果进场被翻了，不攻击
		return actions, false
	}

	beAttackCards := attackCxt.findAttackTarget(bs, card)
	triggerCxt := &triggerContext{
		attackCard: card,
		beAttackCards: beAttackCards,
		attackCxt: attackCxt,
	}

	//glog.Infof("attack 33333333333333333 card=%s, beAttackCards=%v", card, beAttackCards)

// ----------------- 攻击前 ------------------
	triggerCxt.triggerType = attackTrigger
	acts, _, _ := bs.triggerMgr.trigger(map[int][]iTarget{attackTrigger: []iTarget{card}}, triggerCxt)
	actions = append(actions, acts...)

	if len(beAttackCards) > 0 {
		var ts []iTarget
		for _, c := range beAttackCards {
			ts = append(ts, c)
		}
		triggerCxt.triggerType = preBeAttackTrigger
		acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{preBeAttackTrigger: ts}, triggerCxt)
		actions = append(actions, acts...)
	}

	if !card.isInBattle() {
		return actions, false
	}

	attackAct := &pb.AttackAct{
		Attacker: int32(card.getObjID()),
	}
	act := &clientAction {
		actID: pb.ClientAction_Attack,
		actMsg: attackAct,
	}
	actions = append(actions, act)

	if _, ok := attackCxt.getAttackFinder(card.getObjID()).(*arrowAttackTargetFinderSt); ok {
		attackAct.IsArrow = true
	}

// --------------------- 比点 ------------------
	var beAttackTargets []iTarget
	for _, c := range beAttackCards {
		if !c.isInBattle() {
			continue
		}
		beAttackTargets = append(beAttackTargets, c)

		atkPos, defPos, b := attackCxt.getAttacker(card.getObjID()).attack(card, c, attackCxt)
		// 记录比点结果
		triggerCxt.setAttackResult(card, c, b)
		switch b {
		case bLt:
			triggerCxt.setAttackResult(c, card, bGt)
			triggerCxt.setAttackOutcome(c, card, bWin)
			triggerCxt.setAttackOutcome(card, c, bLose)
		case bGt:
			triggerCxt.setAttackResult(c, card, bLt)
			triggerCxt.setAttackOutcome(c, card, bLose)
			triggerCxt.setAttackOutcome(card, c, bWin)
		default:
			triggerCxt.setAttackResult(c, card, bEq)
			triggerCxt.setAttackOutcome(c, card, bEq)
			triggerCxt.setAttackOutcome(card, c, bEq)
		}

		attackAct.WinActs = append(attackAct.WinActs, &pb.AttackWinAct{
			BeAttacker: int32(c.getObjID()),
			WinPos: int32(atkPos),
			LosePos: int32(defPos),
		})
	}
	triggerCxt.attackCxt = nil

	preTurnSit := map[int]int{}
	preTurnSit[card.getObjID()] = card.getSit()
	for _, t := range beAttackTargets {
		preTurnSit[t.getObjID()] = t.getSit()
	}

// -------------- 比点后 ------------------
	triggerCxt.triggerType = afterAttackTrigger
	acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{afterAttackTrigger: []iTarget{card},
		beAttackTrigger: beAttackTargets}, triggerCxt)
	actions = append(actions, acts...)

	// 为了客户端攻击和移动能同时播放
	for _, act := range actions {
		if act.actID == pb.ClientAction_Skill {
			skAct := act.actMsg.(*pb.SkillAct)
			if skAct.Owner == int32(card.getObjID()) {
				if skAct.MoveActs != nil && len(skAct.MoveActs) > 0 {
					attackAct.MoveActs = append(attackAct.MoveActs, skAct.MoveActs...)
					skAct.MoveActs = nil
				}
				if skAct.AfterMoveActs != nil && len(skAct.AfterMoveActs) > 0 {
					attackAct.AfterMoveActs = append(attackAct.AfterMoveActs, skAct.AfterMoveActs...)
					skAct.AfterMoveActs = nil
				}
			}
		}
	}

	if len(beAttackTargets) <= 0 {
		return actions, len(attackAct.WinActs) > 0
	}

// ----------------------- 翻面前 ------------------
	var preTurnTargets []iTarget
	for _, t := range beAttackTargets {
		c := t.(*fightCard)
		if !c.isInBattle() {
			continue
		}

		batOutcome := triggerCxt.getAttackOutcome(card, c)
		batResult := triggerCxt.getAttackResult(card, c)
		if batOutcome != bWin && batResult != bGt && !card.isForceAttack() && !triggerCxt.isOnceAttackWin(card, c) {
			for i, act := range attackAct.WinActs {
				if act.BeAttacker == int32(c.getObjID()) {
					attackAct.WinActs = append(attackAct.WinActs[:i], attackAct.WinActs[i+1:]...)
					break
				}
			}
		}

		if batOutcome == bWin && triggerCxt.isCanTurn(c) {
			preTurnTargets = append(preTurnTargets, t)
		}
	}

	var beTurnTargets []iTarget
	if len(preTurnTargets) > 0 {
		triggerCxt.setTurner(card, card.getSit())
		triggerCxt.triggerType = preBeTurnTrigger
		for i := 0; i < len(preTurnTargets); {
			t := preTurnTargets[i]
			c := t.(*fightCard)
			if !c.isInBattle() {
				preTurnTargets = append(preTurnTargets[:i], preTurnTargets[i+1:]...)
				continue
			}
			i++
			triggerCxt.addBeTurners(c, c.getSit())
		}

		if len(preTurnTargets) > 0 {
			acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{preBeTurnTrigger: preTurnTargets}, triggerCxt)
			actions = append(actions, acts...)
		}
	}

// ----------------------- 翻面 ------------------
	turnOverAct := &pb.TurnOverAct{}
	actions = append(actions, &clientAction{
		actID: pb.ClientAction_TurnOver,
		actMsg: turnOverAct,
	})
	for _, t := range preTurnTargets {
		c := t.(*fightCard)
		if !c.isInBattle() {
			continue
		}
		if triggerCxt.isCanTurn(c) {
			beTurnTargets = append(beTurnTargets, t)
		}
	}

	triggerCxt.triggerType = turnTrigger
	triggerCxt.beTurners = nil
	triggerCxt.turner = nil
	for _, t := range beTurnTargets {
		c := t.(*fightCard)
		triggerCxt.addBeTurners(c, t.getSit())
	}

	for i := 0; i < len(beTurnTargets); {
		t := beTurnTargets[i]
		c := t.(*fightCard)
		if sit, ok := preTurnSit[c.getObjID()]; ok && sit == c.getSit() {
			// 上面几个时机在玩家眼中是同时的，不能在这些时机里连续翻2次
			actions = append(actions, c.turnOver(card)...)
			turnOverAct.BeTurners = append(turnOverAct.BeTurners, int32(c.getObjID()))
			i++
			continue
		}
		beTurnTargets = append(beTurnTargets[:i], beTurnTargets[i+1:]...)
	}

	var turnTriggerObj []iTarget
	if len(beTurnTargets) > 0 {
		triggerCxt.setTurner(card, preTurnSit[card.getObjID()])
		turnTriggerObj = []iTarget{card}
	}

	// 被翻面时机，尝试触发技能
	acts, _, _ = bs.triggerMgr.trigger(map[int][]iTarget{turnTrigger:turnTriggerObj, beTurnTrigger:beTurnTargets}, triggerCxt)
	actions = append(actions, acts...)

	return actions, len(attackAct.WinActs) > 0
}

func (bs *battleSituation) addTaunt(sk *skill, caster iCaster, f *fighter, grids []int) {
	t := &tauntBuff{
		skillID: sk.getID(),
		tauntFighterObjID: f.getObjID(),
		ownerObjID: caster.getObjID(),
		grids: grids,
	}
	bs.taunts = append(bs.taunts , t)
}

func (bs *battleSituation) delTaunt(casterObjID int, skillID int32) {
	for i, t := range bs.taunts {
		if t.ownerObjID == casterObjID && t.skillID == skillID {
			bs.taunts = append(bs.taunts[:i], bs.taunts[i+1:]...)
			return
		}
	}
}

func (bs *battleSituation) getLeft(grid int) (*fightCard, bool) {
	if grid == 0 || grid == bs.gridColumn || grid == 2*bs.gridColumn {
		return nil, false
	} else {
		c := bs.targetMgr.getTargetCard(bs.grids[grid - 1])
		if c != nil {
			return c, false
		} else {
			return nil, true
		}
	}
}

func (bs *battleSituation) getRight(grid int) (*fightCard, bool) {
	if grid == bs.gridColumn-1 || grid == 2*bs.gridColumn-1 || grid == 3*bs.gridColumn-1 {
		return nil, false
	} else {
		c := bs.targetMgr.getTargetCard(bs.grids[grid+1])
		if c != nil {
			return c, false
		} else {
			return nil, true
		}
	}
}

func (bs *battleSituation) getUp(grid int) (*fightCard, bool) {
	if grid >= 0 && grid < bs.gridColumn {
		return nil, false
	} else {
		c := bs.targetMgr.getTargetCard(bs.grids[grid - bs.gridColumn])
		if c != nil {
			return c, false
		} else {
			return nil, true
		}
	}
}

func (bs *battleSituation) getDown(grid int) (*fightCard, bool) {
	if grid >= 2*bs.gridColumn && grid < 3*bs.gridColumn {
		return nil, false
	} else {
		c := bs.targetMgr.getTargetCard(bs.grids[grid + bs.gridColumn])
		if c != nil {
			return c, false
		} else {
			return nil, true
		}
	}
}

func (bs *battleSituation) evaluateAiValue() float32 {
	curUid := bs.curBoutFighter.getUid()
	var curValue float32
	var oppValue float32
	var p1 int
	var p2 int

	for i, objID := range bs.grids {
		c := bs.targetMgr.getTargetCard(objID)
		if c == nil {
			continue
		}

		value := curValue
		if c.getControllerUid() != curUid {
			value = oppValue
			p2 += 1
		} else {
			p1 += 1
		}

		value += c.cardValue

		upCard, isOut := bs.getUp(i)
		if isOut {
			value += float32(c.up) * c.upValueRate
		} else if upCard != nil && upCard.getControllerUid() == c.getControllerUid() {
			value += upCard.adjFValue
		}

		downCard, isOut := bs.getDown(i)
		if isOut {
			value += float32(c.down) * c.downValueRate
		} else if downCard != nil && downCard.getControllerUid() == c.getControllerUid() {
			value += downCard.adjFValue
		}

		leftCard, isOut := bs.getLeft(i)
		if isOut {
			value += float32(c.left) * c.leftValueRate
		} else if leftCard != nil && leftCard.getControllerUid() == c.getControllerUid() {
			value += leftCard.adjFValue
		}

		rightCard, isOut := bs.getRight(i)
		if isOut {
			value += float32(c.right) * c.rightValueRate
		} else if rightCard != nil && rightCard.getControllerUid() == c.getControllerUid() {
			value += rightCard.adjFValue
		}

		if c.getControllerUid() != curUid {
			oppValue = value
		} else {
			curValue = value
		}

		//glog.Debugf("Evaluate id=%d, cur=%f, opp=%f, cardvalue=%f", c.gid, curValue, oppValue, c.GetCardValue())
	}

	if p1+p2 >= len(bs.grids) {
		if p1 > p2 {
			return infinityValue
		} else {
			return - infinityValue
		}
	}

	return curValue - oppValue
}

func (bs *battleSituation) genAllActions() []*battleAction {
	var acts []*battleAction

	for _, cardObjID := range bs.curBoutFighter.hand {
		for gridID, objID := range bs.grids {
			t := bs.targetMgr.getTargetGrid(objID)
			if t == nil {
				continue
			}
			if bs.canPlayCard(bs.curBoutFighter, gridID) {
				acts = append(acts, &battleAction{
					cardObjID: cardObjID,
					gridID: gridID,
				})
			}
		}
	}

	return acts
}

func (bs *battleSituation) bonusBoutEnd(winUid common.UUid) []*clientAction {
	if bs.bonusObj != nil {
		acts, isWonderful1, isWonderful2 := bs.bonusObj.boutEnd(winUid)
		if isWonderful1 {
			bs.isWonderful1 = true
		}
		if isWonderful2 {
			bs.isWonderful2 = true
		}
		return acts
	} else {
		return []*clientAction{}
	}
}
