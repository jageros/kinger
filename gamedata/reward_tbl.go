package gamedata

import (
	"encoding/json"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type RewardItem struct {
	ID int `json:"__id__"`
	Type string `json:"type"`
	ItemID string `json:"itemID"`
	Amount int `json:"amount"`
	Pro int `json:"pro"`
	Team int `json:"team"`
	Version int `json:"version"`
	Areas          [][]int    `json:"areas"`
	AreaLimit *AreaLimitConfig
}

func (r *RewardItem) init() {
	r.AreaLimit = newAreaLimitConfig(r.Areas)
}

type RewardItemList struct {
	Rewards []*RewardItem
	TotalPro int
}

type RewardTblGameData struct {
	baseGameData
	name_        string
	Team2Rewards map[int]*RewardItemList
	Area2MaxVer  map[int]int
	Type         string
}

func newRewardTblGameData(name string) *RewardTblGameData {
	r := &RewardTblGameData{name_: name}
	r.i = r
	return r
}

func (gd *RewardTblGameData) name() string {
	return gd.name_
}

func (gd *RewardTblGameData) init(d []byte) error {
	var l []*RewardItem
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Team2Rewards = map[int]*RewardItemList{}
	var ver = map[int]int{}
	rwType := l[0].Type
	for _, r := range l {
		if r.Pro <= 0 {
			continue
		}
		r.init()
		rlist, ok := gd.Team2Rewards[r.Team]
		if !ok {
			rlist = &RewardItemList{}
			gd.Team2Rewards[r.Team] = rlist
		}
		rlist.Rewards = append(rlist.Rewards, r)
		rlist.TotalPro += r.Pro
		area := GetGameData(consts.AreaConfig).(*AreaConfigGameData)
		area.ForEachOpenedArea(func(config *AreaConfig){
			if v, ok := ver[config.Area]; ok {
				if v < r.Version {
					ver[config.Area] = r.Version
				}
			}else {
				ver[config.Area] = r.Version
			}
		})
		if rwType != r.Type {
			rwType = "uncertain"
		}
	}
	gd.Type = rwType
	gd.Area2MaxVer = ver
	return nil
}

func getAllRewardTblName() []string {
	var names []string
	attr := attribute.NewAttrMgr("gamedata", "__reward_tbl__", true)
	err := attr.Load()
	if err != nil {
		glog.Errorf("getAllRewardTblName error %s", err)
		return names
	}

	namesAttr := attr.GetListAttr("data")
	namesAttr.ForEachIndex(func(index int) bool {
		names = append(names, namesAttr.GetStr(index))
		return true
	})
	return names
}
