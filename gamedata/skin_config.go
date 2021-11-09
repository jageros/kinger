package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type CardSkin struct {
	ID     string `json:"__id__"`
	Name   int    `json:"name"`
	Bind   string `json:"bind"`
	CardID uint32 `json:"general"`
}

func (ck *CardSkin) GetName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(ck.Name)
}

type CardSkinGameData struct {
	baseGameData
	CardSkins   []*CardSkin
	ID2CardSkin map[string]*CardSkin
}

func newCardSkinGameData() *CardSkinGameData {
	r := &CardSkinGameData{}
	r.i = r
	return r
}

func (gd *CardSkinGameData) name() string {
	return consts.CardSkin
}

func (gd *CardSkinGameData) init(d []byte) error {
	var l []*CardSkin
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.CardSkins = l
	gd.ID2CardSkin = map[string]*CardSkin{}
	for _, s := range l {
		gd.ID2CardSkin[s.ID] = s
	}

	return nil
}
