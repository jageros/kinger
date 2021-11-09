package gamedata

import (
	"encoding/json"

	"kinger/common/consts"
)

type IHuodongGoodsGameData interface {
	IGameData
	GetGoods() map[int]*HuodongGoods
}

type HuodongGoods struct {
	ID          int    `json:"__id__"`
	Type        string `json:"type"`
	RewardID    string `json:"rewardId"`
	Amount      int    `json:"cnt"`
	ExchangeCnt int    `json:"exchangeCnt"`
	Price       int    `json:"price"`
}

type HuodongGoodsGameData struct {
	baseGameData
	ID2Goods map[int]*HuodongGoods
}

func newHuodongRewardGameData() *HuodongGoodsGameData {
	c := &HuodongGoodsGameData{}
	c.i = c
	return c
}

func (cg *HuodongGoodsGameData) name() string {
	return consts.HuodongReward
}

func (cg *HuodongGoodsGameData) init(d []byte) error {
	var l []*HuodongGoods
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	cg.ID2Goods = map[int]*HuodongGoods{}
	for _, c := range l {
		cg.ID2Goods[c.ID] = c
	}

	return nil
}

func (cg *HuodongGoodsGameData) GetGoods() map[int]*HuodongGoods {
	return cg.ID2Goods
}
