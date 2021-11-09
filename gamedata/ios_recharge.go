package gamedata

import (
	"bytes"
	"encoding/json"
	"errors"
	"kinger/common/consts"
)

type IRechargeGameData interface {
	IGameData
	GetJadeGoods(goodsID string) *Recharge
	GetAllGoods(area int) []*Recharge
}

type Recharge struct {
	GoodsID      string  `json:"__id__"`
	Price        int     `json:"price"`
	JadeCnt      int     `json:"jadeCnt"`
	JadePrice    int     `json:"jadePrice"`
	Areas        [][]int `json:"areas"`
	FirstJadeCnt int     `json:"firstJadeCnt"`
	FirstJadrVer int     `json:"firstJaceVersion"`
	areaLimit    *AreaLimitConfig
}

func (r *Recharge) init() {
	r.areaLimit = newAreaLimitConfig(r.Areas)
}

type RechargeGameData struct {
	baseGameData
	areaVersion     int
	rawData         []byte
	goodsList       []*Recharge
	areaToGoodsList map[int][]*Recharge
}

func newIosRechargeGameData() *RechargeGameData {
	rg := &RechargeGameData{}
	rg.i = rg
	return rg
}

func (rg *RechargeGameData) name() string {
	return consts.IosRecharge
}

func (rg *RechargeGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if rg.areaVersion == areaVersion && bytes.Equal(rg.rawData, d) {
		return errors.New("no update")
	}

	var l []*Recharge
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	rg.areaVersion = areaVersion
	rg.rawData = d
	rg.areaToGoodsList = map[int][]*Recharge{}
	rg.goodsList = l
	for _, r := range l {
		r.init()
		r.areaLimit.forEachArea(func(area int) {
			goodsList := rg.areaToGoodsList[area]
			rg.areaToGoodsList[area] = append(goodsList, r)
		})
	}

	return nil
}

func (rg *RechargeGameData) GetJadeGoods(goodsID string) *Recharge {
	for _, goods := range rg.goodsList {
		if goods.GoodsID == goodsID {
			return goods
		}
	}

	return nil
}

func (rg *RechargeGameData) GetAllGoods(area int) []*Recharge {
	return rg.areaToGoodsList[area]
}

type WxRechargeGameData struct {
	RechargeGameData
}

func newWxRechargeGameData() *WxRechargeGameData {
	rg := &WxRechargeGameData{}
	rg.i = rg
	return rg
}

func (rg *WxRechargeGameData) name() string {
	return consts.WxRecharge
}

type AndroidRechargeGameData struct {
	RechargeGameData
}

func newAndroidRechargeGameData() *AndroidRechargeGameData {
	rg := &AndroidRechargeGameData{}
	rg.i = rg
	return rg
}

func (rg *AndroidRechargeGameData) name() string {
	return consts.AndroidRecharge
}
