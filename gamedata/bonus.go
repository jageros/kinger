package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Bonus struct {
	ID       int      `json:"__id__"`
	Priority int      `json:"priority"`
	Function []string `json:"function"`
	Total    []int    `json:"total"`
	Weapon   []int    `json:"weapon"`
	Horse    []int    `json:"horse"`
	Material []int    `json:"material"`
	Forage   []int    `json:"forage"`
	Medicine []int    `json:"medicine"`
	Bandage  []int    `json:"bandage"`
	Gold     int      `json:"gold"`
	Energy   float32  `json:"energy"`
	WineDuel []int    `json:"wine_duel"`
	BookDuel []int    `json:"book_duel"`
}

type BonusGameData struct {
	baseGameData
	bonusList []*Bonus
}

func newBonusGameData() *BonusGameData {
	b := &BonusGameData{}
	b.i = b
	return b
}

func (bg *BonusGameData) name() string {
	return consts.Bonus
}

func (bg *BonusGameData) init(d []byte) error {
	err := json.Unmarshal(d, &bg.bonusList)
	if err != nil {
		return err
	}

	//glog.Infof("bonusList = %s", bg.bonusList)

	return nil
}

func (bg *BonusGameData) GetAllBonusData() []*Bonus {
	return bg.bonusList
}
