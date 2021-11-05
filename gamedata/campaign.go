package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Campaign struct {
	ID               int      `json:"__id__"`
	Type             int      `json:"type"`
	Level            int      `json:"level"`
	Support          [][]int  `json:"support"`
	FieldTimes       int      `json:"fieldTimes"`
	Master           []int    `json:"master"`
	FieldMaster      []uint32 `json:"fieldMaster"`
	FieldSoldier     []uint32 `json:"fieldSoldier"`
	Guardian         []uint32 `json:"guardian"`
	GuardSide        [][]int  `json:"guard_side"`
	GuardDeck        []uint32 `json:"guard_deck"`
	Castle           []int32  `json:"castle"`
	RateAttack       float32  `json:"rate_attack"`
	SurrenderCon     float32  `json:"surrenderCon"`
	EnergyCon        float32  `json:"energyCon"`
	RewardExp        int      `json:"reward_exp"`
	FieldExp         int      `json:"field_exp"`
	DefendExp        int      `json:"defend_exp"`
	RewardWeapAttack [][]int  `json:"rewardWeap_attack"`
	RewardHorAttack  [][]int  `json:"rewardHor_attack"`
	RewardMatAttack  [][]int  `json:"rewardMat_attack"`
	RewardGoldAttack [][]int  `json:"rewardGold_attack"`
	RewardForAttack  [][]int  `json:"rewardFor_attack"`
	RewardMedAttack  [][]int  `json:"rewardMed_attack"`
	RewardBanAttack  [][]int  `json:"rewardBan_attack"`
	RewardWeapDefend [][]int  `json:"rewardWeap_defend"`
	RewardHorDefend  [][]int  `json:"rewardHor_defend"`
	RewardMatDefend  [][]int  `json:"rewardMat_defend"`
	RewardGoldDefend [][]int  `json:"rewardGold_defend"`
	RewardForDefend  [][]int  `json:"rewardFor_defend"`
	RewardMedDefend  [][]int  `json:"rewardMed_defend"`
	RewardBanDefend  [][]int  `json:"rewardBan_defend"`
}

type CampaignGameData struct {
	baseGameData
	campaignMap map[int]map[int]*Campaign // map[type]map[level]*Campaign
}

func newCampaignGameData() *CampaignGameData {
	c := &CampaignGameData{}
	c.i = c
	return c
}

func (cgd *CampaignGameData) name() string {
	return consts.Campaign
}

func (cgd *CampaignGameData) init(d []byte) error {
	cgd.campaignMap = make(map[int]map[int]*Campaign)

	var _list []*Campaign
	err := json.Unmarshal(d, &_list)
	if err != nil {
		return err
	}

	for _, cg := range _list {
		m, ok := cgd.campaignMap[cg.Type]
		if !ok {
			m = make(map[int]*Campaign)
			cgd.campaignMap[cg.Type] = m
		}
		m[cg.Level] = cg
	}
	//glog.Infof("campaignMap = %s", cgd.campaignMap)

	return nil
}

func (cgd *CampaignGameData) GetCampaignData(campaignType int, level int) *Campaign {
	if m, ok := cgd.campaignMap[campaignType]; ok {
		if cg, ok := m[level]; ok {
			return cg
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (cgd *CampaignGameData) GetAllCampaign() map[int]map[int]*Campaign {
	return cgd.campaignMap
}
