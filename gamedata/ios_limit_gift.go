package gamedata

import (
	"kinger/common/consts"
	"encoding/json"
	"strconv"
	"bytes"
	"errors"
)

const (
	limitNewbieGiftWeiPrefix = "giftwei"
	limitNewbieGiftShuPrefix = "giftshu"
	limitNewbieGiftWuPrefix = "giftwu"

	limitCommonGiftWeiPrefix = "commonWei"
	limitCommonGiftShuPrefix = "commonShu"
	limitCommonGiftWuPrefix = "commonwu"
)

type ILimitGiftGameData interface {
	IGameData
	GetGiftByID(giftID string) *LimitGift
	GetAllLimitGifts(area int) []*LimitGift
	GetVipCard(area int, isFirst bool) *LimitGift
	GetAreaGift(area int, giftID string) *LimitGift
}

type LimitGift struct {
	GiftID         string     `json:"__id__"`
	Price          int        `json:"price"`
	Reward         string     `json:"reward"`
	BuyConditions  [][]string `json:"teamCondition"`
	ContinueHour   int64      `json:"continueTime"`
	RefreshHour    int64      `json:"refreashTime"`
	JadePrice      int        `json:"jadePrice"`
	ShowConditions [][]string `json:"showTrue"`
	HideConditions [][]string  `json:"showFalse"`
	HeadFrame      string     `json:"headFrame"`
	Version        int        `json:"version"`
	BowlderPrice   int        `json:"bowlderPrice"`
	BeginTime      string     `json:"beginTime"`
	EndTime        string     `json:"endTime"`
	BuyLimitCnt    int        `json:"buyLimit"`
	RewardTbl      string     `json:"reward_tbl"`
	Areas          [][]int    `json:"areas"`
	Visiable       int        `json:"visiable"`

	MaxTeam int
	Visible bool
	RefreshTime int64
	ContinueTime int64
	areaLimit *AreaLimitConfig

	GiftIDPrefix string
	NewbieGiftIDPrefix string
	CommonGiftIDPrefix string
}

func (l *LimitGift) init() {
	l.RefreshTime = l.RefreshHour * 3600
	l.ContinueTime = l.ContinueHour * 3600
	l.areaLimit = newAreaLimitConfig(l.Areas)
	if l.Visiable != 0 {
		l.Visible = true
	}

	var rs []rune
	for _, r := range l.GiftID {
		if _, err := strconv.Atoi(string(r)); err == nil {
			break
		}
		rs = append(rs, r)
	}
	l.GiftIDPrefix = string(rs)

	switch l.GiftIDPrefix {
	case limitNewbieGiftWeiPrefix:
		l.NewbieGiftIDPrefix = limitNewbieGiftWeiPrefix
	case limitNewbieGiftShuPrefix:
		l.NewbieGiftIDPrefix = limitNewbieGiftShuPrefix
	case limitNewbieGiftWuPrefix:
		l.NewbieGiftIDPrefix = limitNewbieGiftWuPrefix
	case limitCommonGiftWeiPrefix:
		l.CommonGiftIDPrefix = limitCommonGiftWeiPrefix
	case limitCommonGiftShuPrefix:
		l.CommonGiftIDPrefix = limitCommonGiftShuPrefix
	case limitCommonGiftWuPrefix:
		l.CommonGiftIDPrefix = limitCommonGiftWuPrefix
	}
}

func (l *LimitGift) IsFirstVip() bool {
	return l.GiftID == "gift30"
}

func (l *LimitGift) IsVip() bool {
	return l.GiftID == "advip"
}

type LimitGiftGameData struct {
	baseGameData
	rawData []byte
	areaVersion int
	idToGift map[string]*LimitGift
	areaToLimitGifts map[int][]*LimitGift
	areaToVipCard    map[int]*LimitGift
	areaToFirstVipCard map[int]*LimitGift
}

func newLimitGiftGameData() *LimitGiftGameData {
	rg := &LimitGiftGameData{}
	rg.i = rg
	return rg
}

func (rg *LimitGiftGameData) name() string {
	return consts.IosLimitGift
}

func (rg *LimitGiftGameData) init(d []byte) error {
	areaVersion := GetGameData(consts.AreaConfig).(*AreaConfigGameData).Version
	if rg.areaVersion == areaVersion && bytes.Equal(rg.rawData, d) {
		return errors.New("no update")
	}

	var l []*LimitGift
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	rg.areaVersion = areaVersion
	rg.rawData = d
	rg.areaToLimitGifts = map[int][]*LimitGift{}
	rg.areaToVipCard =    map[int]*LimitGift{}
	rg.areaToFirstVipCard = map[int]*LimitGift{}
	rg.idToGift = map[string]*LimitGift{}
	for _, g := range l {
		g.init()
		rg.idToGift[g.GiftID] = g
		isVip := g.IsVip()
		isFirstVip := g.IsFirstVip()

		g.areaLimit.forEachArea(func(area int) {
			if isVip {
				rg.areaToVipCard[area] = g
			} else if isFirstVip {
				rg.areaToFirstVipCard[area] = g
			} else {
				limitGifts := rg.areaToLimitGifts[area]
				rg.areaToLimitGifts[area] = append(limitGifts, g)
			}
		})

		for _, conditions := range g.ShowConditions {
			if len(conditions) < 2 {
				continue
			}
			if conditions[0] != "maxTeam" {
				continue
			}
			g.MaxTeam, _ = strconv.Atoi(conditions[1])
			break
		}
	}

	return nil
}

func (rg *LimitGiftGameData) GetVipCard(area int, isFirst bool) *LimitGift {
	if isFirst {
		return rg.areaToFirstVipCard[area]
	}
	return rg.areaToVipCard[area]
}

func (rg *LimitGiftGameData) GetGiftByID(giftID string) *LimitGift {
	return rg.idToGift[giftID]
}

func (rg *LimitGiftGameData) GetAreaGift(area int, giftID string) *LimitGift {
	if g, ok := rg.idToGift[giftID]; ok && g.areaLimit.IsEffective(area) {
		return g
	} else {
		return nil
	}
}

func (rg *LimitGiftGameData) GetAllLimitGifts(area int) []*LimitGift {
	return rg.areaToLimitGifts[area]
}

type WxLimitGiftGameData struct {
	LimitGiftGameData
}

func newWxLimitGiftGameData() *WxLimitGiftGameData {
	rg := &WxLimitGiftGameData{}
	rg.i = rg
	return rg
}

func (rg *WxLimitGiftGameData) name() string {
	return consts.WxLimitGift
}

type AndroidLimitGiftGameData struct {
	LimitGiftGameData
}

func newAndroidLimitGiftGameData() *AndroidLimitGiftGameData {
	rg := &AndroidLimitGiftGameData{}
	rg.i = rg
	return rg
}

func (rg *AndroidLimitGiftGameData) name() string {
	return consts.AndroidLimitGift
}

type AndroidHandJoyLimitGiftGameData struct {
	LimitGiftGameData
}

func newAndroidHandJoyLimitGiftGameData() *AndroidHandJoyLimitGiftGameData {
	rg := &AndroidHandJoyLimitGiftGameData{}
	rg.i = rg
	return rg
}

func (rg *AndroidHandJoyLimitGiftGameData) name() string {
	return consts.AndroidHandjoyLimitGift
}

type IosHandJoyLimitGiftGameData struct {
	LimitGiftGameData
}

func newIosHandJoyLimitGiftGameData() *IosHandJoyLimitGiftGameData {
	rg := &IosHandJoyLimitGiftGameData{}
	rg.i = rg
	return rg
}

func (rg *IosHandJoyLimitGiftGameData) name() string {
	return consts.IosHandjoyLimitGift
}
