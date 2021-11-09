package gamedata

import (
	"encoding/json"
	"strconv"
	"strings"
)

type ActivityFirstRechargeReward struct {
	ID      int      `json:"__id__"`
	GoodsID string   `json:"recharge"`
	Reward  []string `json:"reward"`
}

type ActivityFirstRechargeRewardGameData struct {
	baseGameData
	name_                          string
	ID2ActivityFirstRechargeReward map[int]*ActivityFirstRechargeReward
	// 这样不太好，首冲不一定像现在一样解锁卡，可能只是得到卡
	UnlockCards []uint32
}

func newActivityFirstRechargeRewardGameData(name_ string) *ActivityFirstRechargeRewardGameData {
	c := &ActivityFirstRechargeRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityFirstRechargeRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityFirstRechargeRewardGameData) init(d []byte) error {
	var l []*ActivityFirstRechargeReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ID2ActivityFirstRechargeReward = map[int]*ActivityFirstRechargeReward{}
	ar.UnlockCards = []uint32{}
	for _, c := range l {
		ar.ID2ActivityFirstRechargeReward[c.ID] = c
		for _, reward := range c.Reward {
			rewardInfo := strings.Split(reward, ":")
			if len(rewardInfo) != 2 {
				continue
			}

			cardID, err := strconv.Atoi(rewardInfo[0])
			if err == nil {
				ar.UnlockCards = append(ar.UnlockCards, uint32(cardID))
			}
		}
	}
	return nil
}

func (ar *ActivityFirstRechargeRewardGameData) IsRechargeID(goodsID string) bool {
	if ar.ID2ActivityFirstRechargeReward == nil {
		return false
	}
	for _, data := range ar.ID2ActivityFirstRechargeReward {
		if data.GoodsID == goodsID {
			return true
		}
	}
	return false
}
