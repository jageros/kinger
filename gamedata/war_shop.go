package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type WarShopCard struct {
	ID int `json:"__id__"`
	Type string `json:"type"`
	CardID uint32 `json:"cardId"`
	FightPrice int `json:"fightPrice"`
}

type WarShopCardGameData struct {
	baseGameData
	Goods []*WarShopCard
	CardID2Goods map[uint32]*WarShopCard
}

func newWarShopCardGameData() *WarShopCardGameData {
	gd := &WarShopCardGameData{}
	gd.i = gd
	return gd
}

func (gd *WarShopCardGameData) name() string {
	return consts.WarShopCard
}

func (gd *WarShopCardGameData) init(d []byte) error {
	var l []*WarShopCard
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Goods = l
	gd.CardID2Goods = map[uint32]*WarShopCard{}
	for _, goods := range l {
		gd.CardID2Goods[goods.CardID] = goods
	}

	return nil
}

type WarShopEquip struct {
	ID int `json:"__id__"`
	Type string `json:"type"`
	EquipID string `json:"equipId"`
	FightPrice int `json:"fightPrice"`
}

type WarShopEquipGameData struct {
	baseGameData
	Goods []*WarShopEquip
	ID2Goods map[string]*WarShopEquip
}

func newWarShopEquipGameData() *WarShopEquipGameData {
	gd := &WarShopEquipGameData{}
	gd.i = gd
	return gd
}

func (gd *WarShopEquipGameData) name() string {
	return consts.WarShopEquip
}

func (gd *WarShopEquipGameData) init(d []byte) error {
	var l []*WarShopEquip
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Goods = l
	gd.ID2Goods = map[string]*WarShopEquip{}
	for _, goods := range l {
		gd.ID2Goods[goods.EquipID] = goods
	}

	return nil
}

type WarShopSkin struct {
	ID int `json:"__id__"`
	Type string `json:"type"`
	SkinID string `json:"skinId"`
	FightPrice int `json:"fightPrice"`
}

type WarShopSkinGameData struct {
	baseGameData
	Goods []*WarShopSkin
}

func newWarShopSkinGameData() *WarShopSkinGameData {
	gd := &WarShopSkinGameData{}
	gd.i = gd
	return gd
}

func (gd *WarShopSkinGameData) name() string {
	return consts.WarShopSkin
}

func (gd *WarShopSkinGameData) init(d []byte) error {
	var l []*WarShopSkin
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Goods = l

	return nil
}

type WarShopRes struct {
	ID int `json:"__id__"`
	Type string `json:"type"`
	Amount int `json:"cnt"`
	FightPrice int `json:"fightPrice"`
}

type WarShopResGameData struct {
	baseGameData
	Goods []*WarShopRes
}

func newWarShopResGameData() *WarShopResGameData {
	gd := &WarShopResGameData{}
	gd.i = gd
	return gd
}

func (gd *WarShopResGameData) name() string {
	return consts.WarShopRes
}

func (gd *WarShopResGameData) init(d []byte) error {
	var l []*WarShopRes
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Goods = l

	return nil
}
