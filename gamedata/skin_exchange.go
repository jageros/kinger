package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
)

type PieceSkinGoods struct {
	ID int `json:"__id__"`
	SkinID string `json:"skinId"`
	Price int `json:"piece"`
	Areas [][]int `json:"areas"`

	AreaLimit *AreaLimitConfig
}

func (s *PieceSkinGoods) init() {
	s.AreaLimit = newAreaLimitConfig(s.Areas)
}

type PieceSkinGameData struct {
	baseGameData
	Goods []*PieceSkinGoods
}

func newPieceSkinGameData() *PieceSkinGameData {
	c := &PieceSkinGameData{}
	c.i = c
	return c
}

func (cg *PieceSkinGameData) name() string {
	return consts.PieceSkin
}

func (cg *PieceSkinGameData) init(d []byte) error {
	var l []*PieceSkinGoods
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	cg.Goods = l
	for _, s := range l {
		s.init()
	}

	return nil
}
