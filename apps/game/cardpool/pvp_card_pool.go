package cardpool

import "kinger/gopuppy/attribute"

type pvpCardPool struct {
	attr *attribute.MapAttr
}

func (p *pvpCardPool) getAttr() *attribute.MapAttr {
	return p.attr
}

func (p *pvpCardPool) getPoolID() int {
	return p.attr.GetInt("poolID")
}

func (p *pvpCardPool) getCamp() int {
	return p.attr.GetInt("camp")
}

func (p *pvpCardPool) isFight() bool {
	return p.attr.GetBool("isFight")
}

func (p *pvpCardPool) getCards() []uint32 {
	var cards []uint32
	cardsAttr := p.attr.GetListAttr("cards")
	cardsAttr.ForEachIndex(func(index int) bool {
		cards = append(cards, cardsAttr.GetUInt32(index))
		return true
	})
	return cards
}

func (p *pvpCardPool) updateCard(cardID uint32, index int) {
	p.attr.GetListAttr("cards").SetUInt32(index, cardID)
}

func (p *pvpCardPool) delCard(cardID uint32) {
	cardsAttr := p.attr.GetListAttr("cards")
	cardsAttr.ForEachIndex(func(index int) bool {
		if cardsAttr.GetUInt32(index) == cardID {
			cardsAttr.SetUInt32(index, 0)
			return false
		}
		return true
	})
}

func (p *pvpCardPool) setCards(cardIDs []uint32) {
	cardsAttr := p.attr.GetListAttr("cards")
	for i, cardID := range cardIDs {
		cardsAttr.SetUInt32(i, cardID)
	}
}

func (p *pvpCardPool) setFight(isFight bool) {
	p.attr.SetBool("isFight", isFight)
}

func (p *pvpCardPool) reset() {
	cardsAttr := p.attr.GetListAttr("cards")
	for i := 0; i < cardsAttr.Size(); i++ {
		cardsAttr.SetUInt32(i, 0)
	}
}

func newPvpCardPoolByAttr(attr *attribute.MapAttr) *pvpCardPool {
	return &pvpCardPool{
		attr: attr,
	}
}

func newPvpCardPool(poolID, camp int, isFight bool, cards []uint32) *pvpCardPool {
	attr := attribute.NewMapAttr()
	attr.SetInt("poolID", poolID)
	attr.SetInt("camp", camp)
	attr.SetBool("isFight", isFight)
	cardsAttr := attribute.NewListAttr()
	for _, cardID := range cards {
		cardsAttr.Append(cardID)
	}
	attr.SetListAttr("cards", cardsAttr)
	return newPvpCardPoolByAttr(attr)
}
