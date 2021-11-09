package main

import (
	"fmt"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"math/rand"
)

type fighter struct {
	baseTarget
	agent        *logic.PlayerAgent
	hand         []int           // []objID
	initialHand  []*pb.SkinGCard // []cardID
	isFirstHand  bool
	uid          common.UUid
	battleID     common.UUid
	isRobot      bool
	sideboard    []uint32
	boutTimeout  int
	name         string
	nameText     int
	pvpScore     int
	mmr          int
	robotID      common.UUid
	waiting      bool
	guanxing     bool
	useCards     []uint32
	headImgUrl   string
	headFrame    string
	countryFlag  string
	casterSkills []int32
	pvpLevel     int
	winRate      int
	positive     int
	negative     int
	area         int
	drawCardPool *drawCardPoolSt
}

func newFighter(objID int, fighterData *pb.FighterData, battleID common.UUid, hand []*fightCard, sideboard []uint32,
	situation *battleSituation) *fighter {

	f := &fighter{
		waiting:      true,
		uid:          common.UUid(fighterData.Uid),
		battleID:     battleID,
		isRobot:      fighterData.IsRobot,
		boutTimeout:  boutTimeOut,
		name:         fighterData.Name,
		pvpScore:     int(fighterData.PvpScore),
		robotID:      common.UUid(fighterData.RobotID),
		sideboard:    sideboard,
		mmr:          int(fighterData.Mmr),
		headImgUrl:   fighterData.HeadImgUrl,
		headFrame:    fighterData.HeadFrame,
		countryFlag:  fighterData.CountryFlag,
		casterSkills: fighterData.CasterSkills,
		nameText:     int(fighterData.NameText),
		winRate:      int(fighterData.WinRate),
		positive:     -1,
		negative:     -1,
		drawCardPool: &drawCardPoolSt{},
		area:         int(fighterData.Area),
	}
	f.camp = int(fighterData.Camp)
	f.objID = objID
	f.targetType = stPlayer
	f.gridID = -1
	f.situation = situation

	f.agent = logic.NewPlayerAgent(&gpb.PlayerClient{
		Uid:      fighterData.Uid,
		ClientID: fighterData.ClientID,
		GateID:   fighterData.GateID,
		Region:   fighterData.Region,
	})

	for _, card := range hand {
		card.setController(f.uid)
		card.setOwner(f.uid)
		card.setTargetType(stHand)
		f.hand = append(f.hand, card.getObjID())
	}

	return f
}

func (f *fighter) String() string {
	return fmt.Sprintf("[fighter %d]", f.getUid())
}

func (f *fighter) copy(situation *battleSituation) iTarget {
	c := *f
	cpy := &c
	cpy.situation = situation
	cpy.effects = []*mcMovieEffect{}

	cpy.hand = make([]int, len(f.hand))
	copy(cpy.hand, f.hand)
	cpy.initialHand = make([]*pb.SkinGCard, len(f.initialHand))
	copy(cpy.initialHand, f.initialHand)
	cpy.sideboard = make([]uint32, len(f.sideboard))
	copy(cpy.sideboard, f.sideboard)
	cpy.useCards = make([]uint32, len(f.useCards))
	copy(cpy.useCards, f.useCards)
	cpy.drawCardPool = f.drawCardPool.copy()

	return cpy
}

func (f *fighter) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetInt("attrType", attrFighter)
	attr.SetMapAttr("baseTarget", f.baseTarget.packAttr())
	attr.SetUInt64("uid", uint64(f.uid))
	attr.SetBool("isRobot", f.isRobot)
	attr.SetStr("name", f.name)
	attr.SetInt("pvpScore", f.pvpScore)
	attr.SetBool("guanxing", f.guanxing)
	attr.SetBool("isFirstHand", f.isFirstHand)
	attr.SetInt("mmr", f.mmr)
	attr.SetUInt64("robotID", uint64(f.robotID))
	attr.SetUInt64("battleID", uint64(f.battleID))
	attr.SetStr("headImgUrl", f.headImgUrl)
	attr.SetStr("headFrame", f.headFrame)
	attr.SetStr("countryFlag", f.countryFlag)
	attr.SetInt("nameText", f.nameText)
	attr.SetInt("winRate", f.winRate)
	attr.SetInt("area", f.area)

	handAttr := attribute.NewListAttr()
	for _, objID := range f.hand {
		handAttr.AppendInt(objID)
	}
	attr.SetListAttr("hand", handAttr)

	casterSkillsAttr := attribute.NewListAttr()
	for _, skillID := range f.casterSkills {
		casterSkillsAttr.AppendInt32(skillID)
	}
	attr.SetListAttr("casterSkills", casterSkillsAttr)

	initialHandAttr := attribute.NewListAttr()
	for _, c := range f.initialHand {
		cAttr := attribute.NewMapAttr()
		cAttr.SetUInt32("gcardID", c.GCardID)
		cAttr.SetStr("skin", c.Skin)
		cAttr.SetStr("equip", c.Equip)
		initialHandAttr.AppendMapAttr(cAttr)
	}
	attr.SetListAttr("initialHand", initialHandAttr)

	sideboardAttr := attribute.NewListAttr()
	for _, cardID := range f.sideboard {
		sideboardAttr.AppendUInt32(cardID)
	}
	attr.SetListAttr("sideboard", sideboardAttr)

	//fightCardsAttr := attribute.NewListAttr()
	//for _, card := range self.fightCards {
	//	fightCardsAttr.AppendInt(card.GetObjectID())
	//}
	//attr.SetListAttr("fightCards", fightCardsAttr)

	drawCardsAttr := f.drawCardPool.packAttr()
	attr.SetMapAttr("drawCards", drawCardsAttr)

	useCardsAttr := attribute.NewListAttr()
	for _, gcardID := range f.useCards {
		useCardsAttr.AppendUInt32(gcardID)
	}
	attr.SetListAttr("useCards", useCardsAttr)
	return attr
}

func (f *fighter) restoredFromAttr(attr *attribute.MapAttr, situation *battleSituation) {
	(&f.baseTarget).restoredFromAttr(attr.GetMapAttr("baseTarget"), situation)
	f.uid = common.UUid(attr.GetUInt64("uid"))
	f.isRobot = attr.GetBool("isRobot")
	f.name = attr.GetStr("name")
	f.pvpScore = attr.GetInt("pvpScore")
	f.guanxing = attr.GetBool("guanxing")
	f.isFirstHand = attr.GetBool("isFirstHand")
	f.mmr = attr.GetInt("mmr")
	f.robotID = common.UUid(attr.GetUInt64("robotID"))
	f.battleID = common.UUid(attr.GetUInt64("battleID"))
	f.headImgUrl = attr.GetStr("headImgUrl")
	f.headFrame = attr.GetStr("headFrame")
	f.countryFlag = attr.GetStr("countryFlag")
	f.nameText = attr.GetInt("nameText")
	f.winRate = attr.GetInt("winRate")
	f.area = attr.GetInt("area")
	f.positive = -1
	f.negative = -1

	handAttr := attr.GetListAttr("hand")
	handAttr.ForEachIndex(func(index int) bool {
		f.hand = append(f.hand, handAttr.GetInt(index))
		return true
	})

	casterSkillsAttr := attr.GetListAttr("casterSkills")
	casterSkillsAttr.ForEachIndex(func(index int) bool {
		f.casterSkills = append(f.casterSkills, casterSkillsAttr.GetInt32(index))
		return true
	})

	initialHandAttr := attr.GetListAttr("initialHand")
	initialHandAttr.ForEachIndex(func(index int) bool {
		cAttr := initialHandAttr.GetMapAttr(index)
		c := &pb.SkinGCard{
			GCardID: cAttr.GetUInt32("gcardID"),
			Skin:    cAttr.GetStr("skin"),
			Equip:   cAttr.GetStr("equip"),
		}
		f.initialHand = append(f.initialHand, c)
		return true
	})

	sideboardAttr := attr.GetListAttr("sideboard")
	sideboardAttr.ForEachIndex(func(index int) bool {
		f.sideboard = append(f.sideboard, sideboardAttr.GetUInt32(index))
		return true
	})

	//fightCardsAttr := attr.GetListAttr("fightCards")
	//fightCardsAttr.ForEachIndex(func(index int) bool {
	//	card := desk.GetFightSituation().GetCard(fightCardsAttr.GetInt(index))
	//	if card != nil {
	//		// TODO
	//		f.fightCards[card.CardId()] = card
	//	}
	//	return true
	//})

	drawCardsAttr := attr.GetMapAttr("drawCards")
	f.drawCardPool = &drawCardPoolSt{}
	f.drawCardPool.restoredFromAttr(drawCardsAttr, f, situation)

	useCardsAttr := attr.GetListAttr("useCards")
	useCardsAttr.ForEachIndex(func(index int) bool {
		f.useCards = append(f.useCards, useCardsAttr.GetUInt32(index))
		return true
	})

	if f.isRobot {
		f.agent = logic.NewPlayerAgent(&gpb.PlayerClient{
			ClientID: 0,
			Uid:      1,
			GateID:   0,
		})
	}
}

func (f *fighter) getAiIQ(battleType int) (positive, negative int) {
	if f.positive < 0 || f.negative < 0 {
		/*
			aiData := gamedata.GetGameData(consts.AiMatch).(*gamedata.AiMatchGameData).GetAiIQ(f.winRate)
			if aiData != nil {
				f.positive, f.negative = aiData.PositiveIQ, aiData.NegativeIQ
			} else {
				f.positive, f.negative = 0, 0
			}
		*/
		if f.winRate >= 100 && battleType == consts.BtGuide {
			f.positive = 1
		} else {
			rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
			levelData := rankGameData.Ranks[f.pvpLevel]
			if levelData == nil || (levelData.PositiveIQ == 0 && levelData.NegativeIQ == 0) {
				f.positive = 1
			} else {
				f.positive, f.negative = levelData.PositiveIQ, levelData.NegativeIQ
			}
		}
	}
	return f.positive, f.negative
}

func (f *fighter) getPvpLevel() int {
	if f.pvpLevel <= 0 {
		rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
		f.pvpLevel = rankGameData.GetPvpLevelByStar(f.pvpScore)
	}
	return f.pvpLevel
}

func (f *fighter) getCopyTarget() iTarget {
	return f
}

func (f *fighter) onRestored(agent *logic.PlayerAgent) {
	if f.uid == agent.GetUid() {
		f.agent = agent
		f.wait()
	}
}

func (f *fighter) setSit(sit int) {
	f.sit = sit
	f.initSit = sit
}

func (f *fighter) initDrawCard(battleType int, fighterData *pb.FighterData) {
	f.drawCardPool.init(battleType, f, fighterData)
}

func (f *fighter) getBattle() iBattle {
	return mgr.getBattle(f.battleID)
}

func (f *fighter) getBattleID() common.UUid {
	return f.battleID
}

func (f *fighter) getUid() common.UUid {
	return f.uid
}

func (f *fighter) setInitialHand() {
	for _, objID := range f.hand {
		card := f.situation.getTargetMgr().getTarget(objID).(*fightCard)
		var equip string
		if card.equip != nil {
			equip = card.equip.data.ID
		}
		f.initialHand = append(f.initialHand, &pb.SkinGCard{
			GCardID: card.gcardID,
			Skin:    card.skin,
			Equip:   equip,
		})
	}
}

func (f *fighter) ready() {
	f.waiting = false
}

func (f *fighter) wait() {
	f.waiting = true
}

func (f *fighter) isReady() bool {
	return !f.waiting && f.agent != nil
}

func (f *fighter) setBoutTimeout(t int) {
	f.boutTimeout = t
}

func (f *fighter) getBoutTimeout() int {
	return f.boutTimeout
}

func (f *fighter) getHandAmount() int {
	return len(f.hand)
}

func (f *fighter) randomHandCard() int {
	handAmount := f.getHandAmount()
	if handAmount == 0 {
		return 0
	}

	if f.getPvpLevel() < 5 {
		return f.hand[0]
	}
	return f.hand[rand.Intn(handAmount)]
}

func (f *fighter) getHandCard(cardObjID int) *fightCard {
	for _, objID := range f.hand {
		if objID == cardObjID {
			return f.situation.getTargetMgr().getTargetCard(cardObjID)
		}
	}
	return nil
}

func (f *fighter) delHandCard(cardObjID int) {
	for i, objID := range f.hand {
		if objID == cardObjID {
			f.hand = append(f.hand[:i], f.hand[i+1:]...)
			return
		}
	}
}

func (f *fighter) addHandCard(card *fightCard) {
	f.hand = append(f.hand, card.getObjID())
	card.setTargetType(stHand)
	uid := f.getUid()
	card.setController(uid)
	card.setOwner(uid)
	f.situation.getTargetMgr().addTarget(card)
	// TODO fight card
}

func (f *fighter) drawSideboardCard() *clientAction {
	if len(f.sideboard) <= 0 {
		return nil
	}

	canDrawCardAmount := f.situation.getMaxHandCardAmount() - f.getHandAmount()
	if canDrawCardAmount <= 0 {
		return nil
	}

	var cardMsg []*pb.Card
	for canDrawCardAmount > 0 && len(f.sideboard) > 0 {
		gcardID := f.sideboard[0]
		f.sideboard = f.sideboard[1:]
		canDrawCardAmount--
		maxObjId := f.situation.genObjID()
		card := newFightCard(maxObjId, gcardID, "", "", f.situation)
		f.addHandCard(card)
		cardMsg = append(cardMsg, card.packMsg())
	}

	if len(cardMsg) > 0 {
		return &clientAction{
			actID: pb.ClientAction_DrawCard,
			actMsg: &pb.DrawCardAct{
				Items: []*pb.DrawCardItem{
					&pb.DrawCardItem{
						Uid:   uint64(f.uid),
						Cards: cardMsg,
					},
				},
			},
		}
	} else {
		return nil
	}
}

func (f *fighter) isGuanxing() bool {
	return f.guanxing
}

func (f *fighter) setGuangXing() {
	f.guanxing = true
}

func (f *fighter) packDrawCardShadow() []*pb.Card {
	return f.drawCardPool.packDrawCardShadow(f)
}

func (f *fighter) addUseCard(gcardID uint32) {
	f.useCards = append(f.useCards, gcardID)
}

func (f *fighter) disCard() int {
	return f.drawCardPool.disCard(f)
}

func (f *fighter) popDrawCard(casterCard *fightCard) *fightCard {
	return f.drawCardPool.popDrawCard(f.situation, casterCard)
}

func (f *fighter) popHeroesDrawCard(casterCard *fightCard) *fightCard {
	return f.drawCardPool.popHeroesDrawCard(f.situation, casterCard)
}

func (f *fighter) switchHandCard(casterCard *fightCard) (disCardObjID int, drawCard *fightCard) {
	return f.drawCardPool.switchHandCard(f, casterCard)
}

func (f *fighter) boutEnd() []*clientAction {
	var actions []*clientAction
	curBout := f.situation.getCurBout()
	for _, objID := range f.hand {
		card := f.situation.getTargetMgr().getTargetCard(objID)
		if card == nil {
			continue
		}

		card.forEachSkill(func(sk *skill) {
			isOldLost := sk.isLostByBout(curBout - 1)
			if isOldLost && !sk.isLostByBout(curBout) {
				actions = append(actions, &clientAction{
					actID:  pb.ClientAction_AddSkill,
					actMsg: &pb.ModifySkillAct{CardObjID: int32(card.getObjID()), SkillID: sk.getID()},
				})

				if sk.isAddInHand {
					actions = append(actions, sk.addStatusMcMovie()...)
				}
			}
		})
	}
	return actions
}

func (f *fighter) packMsg() *pb.Fighter {
	msg := &pb.Fighter{
		Uid:           uint64(f.uid),
		ObjId:         int32(f.objID),
		CasterSkills:  f.casterSkills,
		Camp:          int32(f.camp),
		Name:          f.name,
		PvpScore:      int32(f.pvpScore),
		MaxHandAmount: int32(f.situation.getMaxHandCardAmount()),
		HeadImgUrl:    f.headImgUrl,
		HeadFrame:     f.headFrame,
		CountryFlag:   f.countryFlag,
		NameText:      int32(f.nameText),
	}

	if f.isRobot {
		battleObj := f.getBattle()
		if battleObj != nil {
			if _, ok := battleObj.(*levelBattle); ok {
				msg.Camp = consts.Wu
			}
		}
	}

	for _, objID := range f.hand {
		c := f.getHandCard(objID)
		if c != nil {
			msg.Hand = append(msg.Hand, c.packMsg())
		}
	}
	return msg
}

type endFighterData struct {
	uid           common.UUid
	isRobot       bool
	isSurrender   bool
	camp          int
	useCards      []uint32
	initHandCards []*pb.SkinGCard
	isFighter1    bool
	isFirstHand   bool
	mmr           int
	pvpScore      int
	name          string
	nameText      int
	area          int
	indexDiff     int
}

func newEndFighterData(f *fighter, isSurrender bool, indexDiff int) *endFighterData {
	return &endFighterData{
		uid:           f.getUid(),
		isRobot:       f.isRobot,
		isSurrender:   isSurrender,
		camp:          f.getCamp(),
		useCards:      f.useCards,
		initHandCards: f.initialHand,
		isFighter1:    f.situation.getFighter1() == f,
		isFirstHand:   f.isFirstHand,
		mmr:           f.mmr,
		pvpScore:      f.pvpScore,
		name:          f.name,
		nameText:      f.nameText,
		area:          f.area,
		indexDiff:     indexDiff,
	}
}

func (ef *endFighterData) packVideoFighterData() *pb.VideoFighterData {
	return &pb.VideoFighterData{
		Uid:       uint64(ef.uid),
		Name:      ef.name,
		PvpScore:  int32(ef.pvpScore),
		Camp:      int32(ef.camp),
		HandCards: ef.initHandCards,
		IsRobot:   ef.isRobot,
		NameText:  int32(ef.nameText),
	}
}

func (ef *endFighterData) packEndFighterMsg() *pb.EndFighterData {
	return &pb.EndFighterData{
		Uid:           uint64(ef.uid),
		IsRobot:       ef.isRobot,
		IsSurrender:   ef.isSurrender,
		Camp:          int32(ef.camp),
		InitHandCards: ef.initHandCards,
		IsFighter1:    ef.isFighter1,
		IsFirstHand:   ef.isFirstHand,
		Mmr:           int32(ef.mmr),
		UseCards:      ef.useCards,
		Area:          int32(ef.area),
		IndexDiff:     int32(ef.indexDiff),
	}
}
