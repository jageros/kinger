package gamedata

import (
	"encoding/json"
	"math/rand"
	"kinger/common/consts"
	"bytes"
	"errors"
)

type FreeJadeAds struct {
	ID         int    `json:"__id__"`
	Jade        int    `json:"soldJade"`
	Time int `json:"time"`
	Team     int    `json:"team"`
}

func (fa *FreeJadeAds) GetID() int {
	return fa.ID
}

func (fa *FreeJadeAds) GetTime() int {
	return fa.Time
}

type FreeJadeAdsGameData struct {
	baseGameData
	rawData []byte
	Team2Ads map[int][]*FreeJadeAds
	ID2Ads map[int]*FreeJadeAds
}

func newFreeJadeAdsGameData() *FreeJadeAdsGameData {
	r := &FreeJadeAdsGameData{}
	r.i = r
	return r
}

func (gd *FreeJadeAdsGameData) name() string {
	return consts.FreeJddeAds
}

func (gd *FreeJadeAdsGameData) init(d []byte) error {
	if bytes.Equal(gd.rawData, d) {
		return errors.New("no update")
	}

	var l []*FreeJadeAds
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.rawData = d
	gd.Team2Ads = map[int][]*FreeJadeAds{}
	gd.ID2Ads = map[int]*FreeJadeAds{}
	for _, a := range l {
		tl := gd.Team2Ads[a.Team]
		gd.Team2Ads[a.Team] = append(tl, a)
		gd.ID2Ads[a.ID] = a
	}

	return nil
}

func (gd *FreeJadeAdsGameData) RandomAdsByTeam(team int) IFreeShopAds {
	tl := gd.Team2Ads[team]
	if len(tl) <= 0 {
		return nil
	}

	return tl[rand.Intn(len(tl))]
}
