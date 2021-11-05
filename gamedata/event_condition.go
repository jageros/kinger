package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

//type IActivityOpenConditionGameData interface {
//	IGameData
//	GetConditionByID(conditionID int) *ActivityOpenCondition
//}

type ActivityOpenCondition struct {
	ID              int      `json:"__id__"`
	RankLevel       int      `json:"rankLevel"`       //段位
	VipOrNot        int      `json:"vip"`             //是否士族
	InitialCamp     int      `json:"initialCamp"`     //初始国家
	AllCardCnt      int      `json:"allCardCnt"`      //总卡牌数量
	CardLevel       []string `json:"cardLevel"`       //指定卡牌等级
	PlayDay         int      `json:"playDay"`         //建号天数
	CreatTimeBefore string   `json:"creatTimeBefore"` //指定时间之前建号
	FightCnt        int      `json:"fightCnt"`        //战斗场次
	PassLevel       int      `json:"level"`           //通过关卡数
	OffensiveRate   int      `json:"offensiveRate"`   //先手胜率
	DefensiveRate   int      `json:"defensiveRate"`   //后手胜率
	Areas           [][]int  `json:"areas"`           //可开启活动分区
	Platform        []string `json:"platform"`        //可开启的

	IsVip     bool
	AreaLimit *AreaLimitConfig
}

func (ac *ActivityOpenCondition) init() {
	ac.AreaLimit = newAreaLimitConfig(ac.Areas)
	if ac.VipOrNot == 1 {
		ac.IsVip = true
	} else {
		ac.IsVip = false
	}
}

type ActivityOpenConditionGameData struct {
	baseGameData
	ActivityOpenConditionMap map[int]*ActivityOpenCondition
}

func newActivityOpenConditionGameData() *ActivityOpenConditionGameData {
	c := &ActivityOpenConditionGameData{}
	c.i = c
	return c
}

func (ac *ActivityOpenConditionGameData) name() string {
	return consts.ActivityOpenCondition
}

func (ac *ActivityOpenConditionGameData) init(d []byte) error {
	var l []*ActivityOpenCondition
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}
	ac.ActivityOpenConditionMap = map[int]*ActivityOpenCondition{}
	for _, c := range l {
		c.init()
		ac.ActivityOpenConditionMap[c.ID] = c
	}
	return nil
}
