package gamedata

import (
	"encoding/json"
)

//type IActivityLoginRewardGameData interface {
//	IGameData
//	GetLoginRewardByID(rewardID int) *ActivityLoginReward
//}

type ActivityLoginRechargeReward struct {
	ID         int      `json:"__id__"`
	LoginDay   int      `json:"loginDay"`
	RechargeID string   `json:"recharge"`
	Reward     []string `json:"reward"`
}

type ActivityLoginRechargeRewardGameData struct {
	baseGameData
	name_                          string
	ActivityLoginRechargeRewardMap map[int]*ActivityLoginRechargeReward
}

func newActivityLoginRechargeRewardGameData(name_ string) *ActivityLoginRechargeRewardGameData {
	c := &ActivityLoginRechargeRewardGameData{name_: name_}
	c.i = c
	return c
}

func (ar *ActivityLoginRechargeRewardGameData) name() string {
	return ar.name_
}

func (ar *ActivityLoginRechargeRewardGameData) init(d []byte) error {
	var l []*ActivityLoginRechargeReward
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ar.ActivityLoginRechargeRewardMap = map[int]*ActivityLoginRechargeReward{}
	for _, c := range l {
		ar.ActivityLoginRechargeRewardMap[c.ID] = c
	}
	return nil
}

func (ar *ActivityLoginRechargeRewardGameData) IsRechargeID(goodsID string) bool {
	if ar.ActivityLoginRechargeRewardMap == nil {
		return false
	}
	for _, data := range ar.ActivityLoginRechargeRewardMap {
		if data.RechargeID == goodsID {
			return true
		}
	}
	return false
}
