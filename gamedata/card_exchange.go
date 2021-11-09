package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type PieceCardGoods struct {
	ID     int     `json:"__id__"`
	CardID uint32  `json:"cardId"`
	Price  int     `json:"piece"`
	Areas  [][]int `json:"areas"`

	AreaLimit *AreaLimitConfig
}

func (c *PieceCardGoods) init() {
	c.AreaLimit = newAreaLimitConfig(c.Areas)
}

type PieceCardGameData struct {
	baseGameData
	Goods      []*PieceCardGoods
	area2Goods map[int]map[uint32]*PieceCardGoods // map[area]map[cardID]*PieceCardGoods
}

func newPieceCardGameData() *PieceCardGameData {
	c := &PieceCardGameData{}
	c.i = c
	return c
}

func (cg *PieceCardGameData) name() string {
	return consts.PieceCard
}

func (cg *PieceCardGameData) init(d []byte) error {
	var l []*PieceCardGoods
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	cg.Goods = l
	cg.area2Goods = map[int]map[uint32]*PieceCardGoods{}
	for _, g := range l {
		g.init()
		g.AreaLimit.forEachArea(func(area int) {
			cardID2Goods, ok := cg.area2Goods[area]
			if !ok {
				cardID2Goods = map[uint32]*PieceCardGoods{}
				cg.area2Goods[area] = cardID2Goods
			}

			cardID2Goods[g.CardID] = g
		})
	}

	return nil
}

func (cg *PieceCardGameData) GetCardID2Goods(area int) map[uint32]*PieceCardGoods {
	if cardID2Goods, ok := cg.area2Goods[area]; ok {
		return cardID2Goods
	} else {
		return map[uint32]*PieceCardGoods{}
	}
}
