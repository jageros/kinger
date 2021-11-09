package gamedata

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"kinger/common/consts"
	"math/rand"
)

type IFreeShopAds interface {
	GetID() int
	GetTime() int
}

type FreeGoldAds struct {
	ID   int `json:"__id__"`
	Gold int `json:"soldGold"`
	Time int `json:"time"`
	Team int `json:"team"`
}

func (fa *FreeGoldAds) GetID() int {
	return fa.ID
}

func (fa *FreeGoldAds) GetTime() int {
	return fa.Time
}

type FreeGoldAdsGameData struct {
	baseGameData
	rawData  []byte
	Team2Ads map[int][]*FreeGoldAds
	ID2Ads   map[int]*FreeGoldAds
}

func newFreeGoldAdsGameData() *FreeGoldAdsGameData {
	r := &FreeGoldAdsGameData{}
	r.i = r
	return r
}

func (gd *FreeGoldAdsGameData) name() string {
	return consts.FreeGoldAds
}

func (gd *FreeGoldAdsGameData) init(d []byte) error {
	if bytes.Equal(gd.rawData, d) {
		return errors.New("no update")
	}

	var l []*FreeGoldAds
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.rawData = d
	gd.Team2Ads = map[int][]*FreeGoldAds{}
	gd.ID2Ads = map[int]*FreeGoldAds{}
	for _, a := range l {
		tl := gd.Team2Ads[a.Team]
		gd.Team2Ads[a.Team] = append(tl, a)
		gd.ID2Ads[a.ID] = a
	}

	return nil
}

func (gd *FreeGoldAdsGameData) RandomAdsByTeam(team int) IFreeShopAds {
	tl := gd.Team2Ads[team]
	if len(tl) <= 0 {
		return nil
	}

	return tl[rand.Intn(len(tl))]
}
