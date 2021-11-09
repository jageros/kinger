package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type CampaignParam struct {
	Key       string    `json:"__id__"`
	ParaValue []float64 `json:"para_value"`
}

type CampaignParamGameData struct {
	baseGameData
	SingleDamage         float64 // 单次攻城伤害
	DefenseRevise        float64
	SingleMerit          float64 // 单次攻城功绩
	SingleEncounterVic   float64 // 单次遭遇战胜利获得功绩
	SingleAttackVic      float64 // 单次攻城战胜利获得功绩
	SingleLoseVic        float64 // 单次攻城、守城战失败获得功绩
	KingPer              float64 // 主公功绩抽成／势力
	JunshiPer            float64 // 军师功绩抽成／势力
	ZhonglangjiangPer    float64 // 中郎将功绩抽成／势力
	TaishouPer           float64 // 太守功绩抽成／城池
	DuweiPer             float64 // 都尉功绩抽成／城池
	XiaoweiPer           float64 // 校尉功绩抽成／城池
	KingSalary           float64 // 主公俸禄／势力收益
	JunshiSalary         float64 // 军师俸禄／势力收益
	ZhonglangjiangSalary float64 // 中郎将俸禄／势力收益
	TaishouSalary        float64 // 太守俸禄／城池收益
	DuweiSalary          float64 // 都尉俸禄／城池收益
	XiaoweiSalary        float64 // 校尉俸禄／城池收益
	IrrigationTime       float64 // 灌溉耗时系数
	TradeTime            float64 // 贸易耗时系数
	BuildTime            float64 // 修筑耗时系数
	TaskTargetIrrigation float64 // 单次灌溉目标
	TaskTargetTrade      float64 // 单次贸易目标
	TaskTargetBuild      float64 // 单次修筑目标
	MarchSpeed           int     // 行军速度
	GoldConversion       float64 // 金币转化比
	ForageConversion     float64 // 粮草转化比
	InitialDefense       float64 // 初始城防比
	BundleForage         int     // 单捆粮草数量
	TaskTransportForage  int     // 单次运多少粮
	TaskTransportGold    int     // 单次运多少金
	TransportTime        float64 // 运输耗时系数
	TransferCost         float64 // 迁移花费系数
	HonorRevise          float64 // 势力荣誉修正
	IrrigationVic        float64 // 单次灌溉功绩
	TradeVic             float64 // 单次贸易功绩
	BuildVic             float64 // 单次修筑功绩
	TransportVic         float64 // 单位运输路程功绩
}

func newCampaignParamGameData() *CampaignParamGameData {
	r := &CampaignParamGameData{}
	r.i = r
	return r
}

func (gd *CampaignParamGameData) name() string {
	return consts.CampaignParam
}

func (gd *CampaignParamGameData) init(d []byte) error {
	var l []*CampaignParam
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	for _, p := range l {
		switch p.Key {
		case "single_damage":
			gd.SingleDamage = p.ParaValue[0]
		case "single_merit":
			gd.SingleMerit = p.ParaValue[0]
		case "single_encounter_vic":
			gd.SingleEncounterVic = p.ParaValue[0]
		case "single_attack_vic":
			gd.SingleAttackVic = p.ParaValue[0]
		case "king_per":
			gd.KingPer = p.ParaValue[0]
		case "junshi_per":
			gd.JunshiPer = p.ParaValue[0]
		case "zhonglangjiang_per":
			gd.ZhonglangjiangPer = p.ParaValue[0]
		case "taishou_per":
			gd.TaishouPer = p.ParaValue[0]
		case "duwei_per":
			gd.DuweiPer = p.ParaValue[0]
		case "xiaowei_per":
			gd.XiaoweiPer = p.ParaValue[0]
		case "king_salary":
			gd.KingSalary = p.ParaValue[0]
		case "junshi_salary":
			gd.JunshiSalary = p.ParaValue[0]
		case "zhonglangjiang_salary":
			gd.ZhonglangjiangSalary = p.ParaValue[0]
		case "taishou_salary":
			gd.TaishouSalary = p.ParaValue[0]
		case "duwei_salary":
			gd.DuweiSalary = p.ParaValue[0]
		case "xiaowei_salary":
			gd.XiaoweiSalary = p.ParaValue[0]
		case "irrigation_time":
			gd.IrrigationTime = p.ParaValue[0]
		case "trade_time":
			gd.TradeTime = p.ParaValue[0]
		case "build_time":
			gd.BuildTime = p.ParaValue[0]
		case "task_target_irrigation":
			gd.TaskTargetIrrigation = p.ParaValue[0]
		case "task_target_trade":
			gd.TaskTargetTrade = p.ParaValue[0]
		case "task_target_build":
			gd.TaskTargetBuild = p.ParaValue[0]
		case "march_speed":
			gd.MarchSpeed = int(p.ParaValue[0])
		case "gold_conversion":
			gd.GoldConversion = p.ParaValue[0]
		case "forage_conversion":
			gd.ForageConversion = p.ParaValue[0]
		case "initial_defense":
			gd.InitialDefense = p.ParaValue[0]
		case "bundle_forage":
			gd.BundleForage = int(p.ParaValue[0])
		case "task_target_transport":
			gd.TaskTransportGold = int(p.ParaValue[0])
			gd.TaskTransportForage = int(p.ParaValue[1])
		case "transport_time":
			gd.TransportTime = p.ParaValue[0]
		case "transfer_cost":
			gd.TransferCost = p.ParaValue[0]
		case "single_lose_vic":
			gd.SingleLoseVic = p.ParaValue[0]
		case "defense_revise":
			gd.DefenseRevise = p.ParaValue[0]
		case "honor_revise":
			gd.HonorRevise = p.ParaValue[0]
		case "irrigation_vic":
			gd.IrrigationVic = p.ParaValue[0]
		case "trade_vic":
			gd.TradeVic = p.ParaValue[0]
		case "build_vic":
			gd.BuildVic = p.ParaValue[0]
		case "transport_vic":
			gd.TransportVic = p.ParaValue[0]
		}
	}

	return nil
}
