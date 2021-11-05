package gamedata

import (
	"encoding/json"
	"math/rand"
	"kinger/common/consts"
	"bytes"
	"errors"
)

type FreeTreasureAds struct {
	ID         int    `json:"__id__"`
	TreasureModelID        string    `json:"treasureId"`
	Time int `json:"time"`
	Team     int    `json:"team"`
}

func (fa *FreeTreasureAds) GetID() int {
	return fa.ID
}

func (fa *FreeTreasureAds) GetTime() int {
	return fa.Time
}

type FreeTreasureAdsGameData struct {
	baseGameData
	rawData []byte
	Team2Ads map[int][]*FreeTreasureAds
	ID2Ads map[int]*FreeTreasureAds
}

func newFreeTreasureAdsGameData() *FreeTreasureAdsGameData {
	r := &FreeTreasureAdsGameData{}
	r.i = r
	return r
}

func (gd *FreeTreasureAdsGameData) name() string {
	return consts.FreeTreasureAds
}

func (gd *FreeTreasureAdsGameData) init(d []byte) error {
	if bytes.Equal(gd.rawData, d) {
		return errors.New("no update")
	}

	var l []*FreeTreasureAds
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.rawData = d
	gd.Team2Ads = map[int][]*FreeTreasureAds{}
	gd.ID2Ads = map[int]*FreeTreasureAds{}
	for _, a := range l {
		tl := gd.Team2Ads[a.Team]
		gd.Team2Ads[a.Team] = append(tl, a)
		gd.ID2Ads[a.ID] = a
	}

	return nil
}

func (gd *FreeTreasureAdsGameData) RandomAdsByTeam(team int) IFreeShopAds {
	tl := gd.Team2Ads[team]
	if len(tl) <= 0 {
		return nil
	}

	return tl[rand.Intn(len(tl))]
}

type FreeGoodTreasureAdsGameData struct {
	FreeTreasureAdsGameData
}

func newFreeGoodTreasureAdsGameData() *FreeGoodTreasureAdsGameData {
	r := &FreeGoodTreasureAdsGameData{}
	r.i = r
	return r
}

func (gd *FreeGoodTreasureAdsGameData) name() string {
	return consts.FreeGoodTreasureAds
}