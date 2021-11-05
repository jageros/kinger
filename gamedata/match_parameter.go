package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
	"time"
)

type MatchParam struct {
	ID string `json:"__id__"`
	Value float64 `json:"para_value"`
}

type MatchParamGameData struct {
	baseGameData

	CardStreWeight        float64       // 卡等权重
	StarWeight            float64       // 星数权重
	WinRateWeight         float64       // 胜率权重
	RebornWeight          float64       // 下野次数权重
	EquipWeight           float64       // 宝物权重
	AreaRevise            float64       // 跨区修正
	IndexInterval         float64       // 匹配区间
	IndexRedLine          float64       // 同区指数红线
	TimeInterval          time.Duration // 时间间隔
	SameAreaReborn        int           // 同区下野跨度
	CrossAreaReborn       int           // 跨区下野跨度
	RecentlyOpponentIndex float64       // 上局对手提高的匹配指数
	WinningStreakIndex    float64       // 每连胜一场提高的匹配指数
	CrossAreaIndexRedLine float64       // 跨区指数红线
	RechargeRevise float64              // 充值修正比
	RechargeReviseLimit float64         // 充值修正负限
	WinRevise float64
	LoseRevise float64
	RechargeReviseRecovery int
}

func newMatchParamGameData() *MatchParamGameData {
	r := &MatchParamGameData{}
	r.i = r
	return r
}

func (md *MatchParamGameData) name() string {
	return consts.MatchParam
}

func (md *MatchParamGameData) init(d []byte) error {
	var l []*MatchParam
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	for _, m := range l {
		switch m.ID {
		case "cardlevel_weight":
			md.CardStreWeight = m.Value
		case "star_weight":
			md.StarWeight = m.Value
		case "win_weight":
			md.WinRateWeight = m.Value
		case "reborn_weight":
			md.RebornWeight = m.Value
		case "item_weight":
			md.EquipWeight = m.Value
		case "region_revise":
			md.AreaRevise = m.Value
		case "index_interval":
			md.IndexInterval = m.Value
		case "index_redline":
			md.IndexRedLine = m.Value
		case "time_interval":
			md.TimeInterval = time.Duration(m.Value) * time.Second
		case "same_region_reborn":
			md.SameAreaReborn = int(m.Value)
		case "dif_region_reborn":
			md.CrossAreaReborn = int(m.Value)
		case "recently_opponent":
			md.RecentlyOpponentIndex = m.Value
		case "winning_streak":
			md.WinningStreakIndex = m.Value
		case "dif_index_redline":
			md.CrossAreaIndexRedLine = m.Value
		case "recharge_revise":
			md.RechargeRevise = m.Value
		case "recharge_revise_limit":
			md.RechargeReviseLimit = - m.Value
		case "win_revise":
			md.WinRevise = m.Value
		case "lose_revise":
			md.LoseRevise = m.Value
		case "recharge_revise_recovery":
			md.RechargeReviseRecovery = int(m.Value)
		}
	}

	return nil
}
