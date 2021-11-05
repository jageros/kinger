package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type GiftCode struct {
	ID           int      `json:"__id__"`
	Cnt          int      `json:"cnt"`
	Reward       []string `json:"reward"`
	Repeat       int      `json:"repeat"`
	BeginTimeStr string   `json:"beginTime"`
	EndTimeStr   string   `json:"endTime"`
}

type GiftCodeGameData struct {
	baseGameData
	Type2Code map[int]*GiftCode
}

func newGiftCodeGameData() *GiftCodeGameData {
	r := &GiftCodeGameData{}
	r.i = r
	return r
}

func (gd *GiftCodeGameData) name() string {
	return consts.GiftCode
}

func (gd *GiftCodeGameData) init(d []byte) error {
	var l []*GiftCode
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("GiftCodeGameData = %v", l)

	gd.Type2Code = map[int]*GiftCode{}
	for _, c := range l {
		gd.Type2Code[c.ID] = c
	}

	return nil
}
