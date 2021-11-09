package main

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"math/rand"
)

type drawCardGenerator func(dc *drawCard, casterCard *fightCard, situation *battleSituation) *fightCard

func normalDrawCardGenerator(dc *drawCard, casterCard *fightCard, situation *battleSituation) *fightCard {
	return newFightCard(situation.genObjID(), dc.gcardID, dc.skin, dc.equip, situation)
}

func trainDrawCardGenerator(dc *drawCard, casterCard *fightCard, situation *battleSituation) *fightCard {
	if casterCard == nil {
		return normalDrawCardGenerator(dc, casterCard, situation)
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	cardData := poolGameData.GetCardByGid(dc.gcardID)
	if cardData == nil {
		return normalDrawCardGenerator(dc, casterCard, situation)
	}

	cardData2 := poolGameData.GetCard(cardData.CardID, casterCard.getLevel())
	if cardData2 == nil {
		return normalDrawCardGenerator(dc, casterCard, situation)
	}
	return newFightCard(situation.genObjID(), cardData2.GCardID, dc.skin, dc.equip, situation)
}

func getDrawCardGenerator(battleType int, f *fighter) drawCardGenerator {
	if !f.isRobot {
		return normalDrawCardGenerator
	}
	if battleType == consts.BtTraining {
		return trainDrawCardGenerator
	}
	return normalDrawCardGenerator
}

type drawCard struct {
	gcardID uint32
	skin    string
	equip   string
}

func newDrawCard(gcardID uint32, skin, equip string) *drawCard {
	return &drawCard{
		gcardID: gcardID,
		skin:    skin,
		equip:   equip,
	}
}

func (dc *drawCard) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	attr.SetUInt32("gcardID", dc.gcardID)
	attr.SetStr("skin", dc.skin)
	attr.SetStr("equip", dc.equip)
	return attr
}

func (dc *drawCard) restoredFromAttr(attr *attribute.MapAttr) {
	dc.gcardID = attr.GetUInt32("gcardID")
	dc.skin = attr.GetStr("skin")
	dc.equip = attr.GetStr("equip")
}

// 补牌池
type drawCardPoolSt struct {
	drawCardPool     []*fightCard // 被观星的，待抓的牌
	drawCards        []*drawCard  // 可以抓哪些牌（当前阵营）
	drawHeroesCards  []*drawCard  // 可以抓哪些群雄牌
	othDrawCards     []*drawCard  // 其他两个国家的牌
	initHandMinLevel int
	cardGenerator    drawCardGenerator

	disCards       []*drawCard // 被弃的牌（当前阵营）
	disHeroesCards []*drawCard // 被弃的群雄牌
	othDisCards    []*drawCard // 被弃的其他两个国家的牌
}

func (dcp *drawCardPoolSt) copy() *drawCardPoolSt {
	cpy := &drawCardPoolSt{initHandMinLevel: dcp.initHandMinLevel}
	cpy.cardGenerator = dcp.cardGenerator
	cpy.drawCardPool = make([]*fightCard, len(dcp.drawCardPool))
	copy(cpy.drawCardPool, dcp.drawCardPool)

	cpy.drawCards = make([]*drawCard, len(dcp.drawCards))
	copy(cpy.drawCards, dcp.drawCards)

	cpy.drawHeroesCards = make([]*drawCard, len(dcp.drawHeroesCards))
	copy(cpy.drawHeroesCards, dcp.drawHeroesCards)

	cpy.othDrawCards = make([]*drawCard, len(dcp.othDrawCards))
	copy(cpy.othDrawCards, dcp.othDrawCards)

	cpy.disCards = make([]*drawCard, len(dcp.disCards))
	copy(cpy.disCards, dcp.disCards)

	cpy.disHeroesCards = make([]*drawCard, len(dcp.disHeroesCards))
	copy(cpy.disHeroesCards, dcp.disHeroesCards)

	cpy.othDisCards = make([]*drawCard, len(dcp.othDisCards))
	copy(cpy.othDisCards, dcp.othDisCards)

	return cpy
}

func (dcp *drawCardPoolSt) packDrawCardAttr(cards []*drawCard) *attribute.ListAttr {
	attr := attribute.NewListAttr()
	for _, c := range cards {
		attr.AppendMapAttr(c.packAttr())
	}
	return attr
}

func (dcp *drawCardPoolSt) packAttr() *attribute.MapAttr {
	attr := attribute.NewMapAttr()
	drawCardPoolAttr := attribute.NewListAttr()
	for _, c := range dcp.drawCardPool {
		drawCardPoolAttr.AppendMapAttr(c.packAttr())
	}
	attr.SetListAttr("drawCardPool", drawCardPoolAttr)

	attr.SetListAttr("drawCards", dcp.packDrawCardAttr(dcp.drawCards))
	attr.SetListAttr("drawHeroesCards", dcp.packDrawCardAttr(dcp.drawHeroesCards))
	attr.SetListAttr("othDrawCards", dcp.packDrawCardAttr(dcp.othDrawCards))
	attr.SetListAttr("disCards", dcp.packDrawCardAttr(dcp.disCards))
	attr.SetListAttr("disHeroesCards", dcp.packDrawCardAttr(dcp.disHeroesCards))
	attr.SetListAttr("othDisCards", dcp.packDrawCardAttr(dcp.othDisCards))
	attr.SetInt("initHandMinLevel", dcp.initHandMinLevel)
	return attr
}

func (dcp *drawCardPoolSt) restoredDrawCardsFromAttr(attr *attribute.ListAttr) []*drawCard {
	var cards []*drawCard
	attr.ForEachIndex(func(index int) bool {
		c := &drawCard{}
		c.restoredFromAttr(attr.GetMapAttr(index))
		cards = append(cards, c)
		return true
	})
	return cards
}

func (dcp *drawCardPoolSt) restoredFromAttr(attr *attribute.MapAttr, f *fighter, situation *battleSituation) {
	dcp.cardGenerator = getDrawCardGenerator(situation.battleType, f)
	drawCardPoolAttr := attr.GetListAttr("drawCardPool")
	drawCardPoolAttr.ForEachIndex(func(index int) bool {
		c := &fightCard{}
		c.restoredFromAttr(drawCardPoolAttr.GetMapAttr(index), situation)
		dcp.drawCardPool = append(dcp.drawCardPool, c)
		return true
	})

	dcp.drawCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("drawCards"))
	dcp.drawHeroesCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("drawHeroesCards"))
	dcp.othDrawCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("othDrawCards"))
	dcp.disCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("disCards"))
	dcp.disHeroesCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("disHeroesCards"))
	dcp.othDisCards = dcp.restoredDrawCardsFromAttr(attr.GetListAttr("othDisCards"))
	dcp.initHandMinLevel = attr.GetInt("initHandMinLevel")
}

func (dcp *drawCardPoolSt) init(battleType int, f *fighter, fighterData *pb.FighterData) {
	dcp.cardGenerator = getDrawCardGenerator(battleType, f)
	if dcp.initHandMinLevel <= 0 {
		dcp.initHandMinLevel = 100
	}
	if len(dcp.drawCards) > 0 || len(dcp.drawHeroesCards) > 0 {
		return
	}
	if f.camp != consts.Wei && f.camp != consts.Shu && f.camp != consts.Wu {
		return
	}

	//glog.Infof("initDrawCard uid=%d, isRobot=%v, drawGCardIDs=%v, DrawCardPool=%v", f.getUid(),
	//	f.drawGCardIDs, fighterData.DrawCardPool)

	var handCards []*fightCard
	for _, objID := range f.hand {
		c := f.situation.getTargetMgr().getTargetCard(objID)
		if c == nil {
			continue
		}
		handCards = append(handCards, c)
		if f.isRobot && c.getLevel() < dcp.initHandMinLevel {
			dcp.initHandMinLevel = c.getLevel()
		}
	}

	var addDrawGCardID = func(cardData *gamedata.Card, skin, equip string) {
		if cardData == nil {
			return
		}
		if cardData.GetLevel() > dcp.initHandMinLevel {
			return
		}
		if cardData.IsSystemCard() {
			return
		}
		for _, c := range handCards {
			if c.cardID == cardData.CardID {
				return
			}
		}

		if cardData.GetCamp() == consts.Heroes {
			dcp.drawHeroesCards = append(dcp.drawHeroesCards, newDrawCard(cardData.GetGCardID(), skin, equip))
		} else if cardData.GetCamp() == f.camp {
			dcp.drawCards = append(dcp.drawCards, newDrawCard(cardData.GetGCardID(), skin, equip))
		} else {
			dcp.othDrawCards = append(dcp.othDrawCards, newDrawCard(cardData.GetGCardID(), skin, equip))
		}
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	if f.isRobot {
		allCardDatas := poolGameData.GetCardsByLevel(dcp.initHandMinLevel)
		for _, cardData := range allCardDatas {
			if !cardData.IsSpCard() {
				addDrawGCardID(cardData, "", "")
			}
		}
	} else {
		for _, c := range fighterData.DrawCardPool {
			addDrawGCardID(poolGameData.GetCardByGid(c.GCardID), c.Skin, c.Equip)
		}
	}

	for i := range dcp.drawCards {
		j := rand.Intn(i + 1)
		dcp.drawCards[i], dcp.drawCards[j] = dcp.drawCards[j], dcp.drawCards[i]
	}
	for i := range dcp.drawHeroesCards {
		j := rand.Intn(i + 1)
		dcp.drawHeroesCards[i], dcp.drawHeroesCards[j] = dcp.drawHeroesCards[j], dcp.drawHeroesCards[i]
	}
}

func (dcp *drawCardPoolSt) packDrawCardShadow(f *fighter) []*pb.Card {
	needCardShadowAmount := f.situation.getMaxHandCardAmount() - f.getHandAmount()
	needGenCardAmount := needCardShadowAmount - len(dcp.drawCardPool)
	for i := needGenCardAmount; i > 0; i-- {
		if len(dcp.drawCards) <= 0 {
			break
		}
		dc := dcp.drawCards[0]
		dcp.drawCards = dcp.drawCards[1:]
		card := newFightCard(f.situation.genObjID(), dc.gcardID, dc.skin, dc.equip, f.situation)
		if card == nil {
			continue
		}
		dcp.drawCardPool = append(dcp.drawCardPool, card)
	}

	if needGenCardAmount > 0 && len(dcp.drawCards) <= 0 {
		dcp.drawCards = dcp.disCards
		dcp.disCards = []*drawCard{}
	}

	var shadows []*pb.Card
	for _, card := range dcp.drawCardPool {
		shadows = append(shadows, card.packMsg())
	}
	return shadows
}

func (dcp *drawCardPoolSt) disCard(f *fighter) int {
	handCardAmount := f.getHandAmount()
	if handCardAmount == 0 {
		return 0
	}

	i := rand.Intn(handCardAmount)
	objID := f.hand[i]
	card := f.situation.getTargetMgr().getTargetCard(objID)
	f.delHandCard(objID)
	f.situation.getTargetMgr().delTarget(objID)

	if card != nil {
		camp := card.getCamp()
		if camp == consts.Heroes {
			dcp.disHeroesCards = append(dcp.disHeroesCards, card.packDrawCard())
		} else if camp == f.getCamp() {
			dcp.disCards = append(dcp.disCards, card.packDrawCard())
		} else {
			dcp.othDisCards = append(dcp.othDisCards, card.packDrawCard())
		}
	}

	return objID
}

func (dcp *drawCardPoolSt) doDrawCard(drawCards, disCards []*drawCard, drawIdx int, casterCard *fightCard,
	situation *battleSituation) (newDrawCards, newDisCards []*drawCard, card *fightCard) {

	if drawIdx < 0 {
		drawIdx = 0
	} else if drawIdx >= len(drawCards) {
		drawIdx = len(drawCards) - 1
	}

	dc := drawCards[drawIdx]
	newDrawCards = append(drawCards[:drawIdx], drawCards[drawIdx+1:]...)
	newDisCards = disCards
	if len(newDrawCards) <= 0 {
		newDrawCards = disCards
		newDisCards = []*drawCard{}
	}
	card = dcp.cardGenerator(dc, casterCard, situation)
	return
}

func (dcp *drawCardPoolSt) popDrawCard(situation *battleSituation, casterCard *fightCard) *fightCard {
	if len(dcp.drawCardPool) > 0 {
		card := dcp.drawCardPool[0]
		dcp.drawCardPool = dcp.drawCardPool[1:]
		return card
	} else if len(dcp.drawCards) > 0 {
		var card *fightCard
		dcp.drawCards, dcp.disCards, card = dcp.doDrawCard(dcp.drawCards, dcp.disCards, 0, casterCard, situation)
		return card
	}
	return nil
}

func (dcp *drawCardPoolSt) popHeroesDrawCard(situation *battleSituation, casterCard *fightCard) *fightCard {
	if len(dcp.drawHeroesCards) > 0 {
		dc := dcp.drawHeroesCards[0]
		dcp.drawHeroesCards = dcp.drawHeroesCards[1:]
		if len(dcp.drawHeroesCards) <= 0 {
			dcp.drawHeroesCards = dcp.disHeroesCards
			dcp.disHeroesCards = []*drawCard{}
		}
		return dcp.cardGenerator(dc, casterCard, situation)
	}
	return nil
}

func (dcp *drawCardPoolSt) switchHandCard(f *fighter, casterCard *fightCard) (disCardObjID int, drawCard *fightCard) {
	handAmount := f.getHandAmount()
	if handAmount <= 0 {
		return
	}

	//drawCardPoolAmount := len(dcp.drawCardPool)
	drawCardsAmount := len(dcp.drawCards)
	drawHeroesCardsAmount := len(dcp.drawHeroesCards)
	othDrawCardsAmount := len(dcp.othDrawCards)
	allCardAmount := drawCardsAmount + drawHeroesCardsAmount + othDrawCardsAmount
	if allCardAmount <= 0 {
		return
	}

	disCardObjID = dcp.disCard(f)
	i := rand.Intn(allCardAmount)
	if i < drawCardsAmount {
		dcp.drawCards, dcp.disCards, drawCard = dcp.doDrawCard(dcp.drawCards, dcp.disCards, i, casterCard, f.situation)
		return
	}

	i -= drawCardsAmount
	if i < drawHeroesCardsAmount {
		dcp.drawHeroesCards, dcp.disHeroesCards, drawCard = dcp.doDrawCard(dcp.drawHeroesCards, dcp.disHeroesCards, i,
			casterCard, f.situation)
		return
	}

	i -= drawHeroesCardsAmount
	if i < othDrawCardsAmount {
		dcp.othDrawCards, dcp.othDisCards, drawCard = dcp.doDrawCard(dcp.othDrawCards, dcp.othDisCards, i, casterCard,
			f.situation)
	}
	return
}
