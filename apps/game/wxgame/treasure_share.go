package wxgame

import (
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"math/rand"
	"strconv"
	"time"
)

var treasureShareHDs []*treasureShareHD

type treasureShareHD struct {
	id                   int
	beginTime            time.Time
	endTime              time.Time
	rewardTreasureID     string
	triggerTreasuresRate map[string]int
}

func newTreasureShareHD(data *gamedata.TreasureShare) *treasureShareHD {
	beginTime, err := utils.StringToTime(data.BeginTime, utils.TimeFormat2)
	if err != nil {
		glog.Errorf("newTreasureShareHD StringToTime error id=%d, err=%s", data.ID, err)
		return nil
	}
	endTime, err := utils.StringToTime(data.EndTime, utils.TimeFormat2)
	if err != nil {
		glog.Errorf("newTreasureShareHD StringToTime error id=%d, err=%s", data.ID, err)
		return nil
	}

	ts := &treasureShareHD{
		id:                   data.ID,
		beginTime:            beginTime,
		endTime:              endTime,
		rewardTreasureID:     data.Reward,
		triggerTreasuresRate: map[string]int{},
	}

	for _, triggerTreasure := range data.TreasureId {
		if len(triggerTreasure) < 2 {
			glog.Errorf("newTreasureShareHD triggerTreasure %v", triggerTreasure)
			return nil
		}

		rate, err := strconv.Atoi(triggerTreasure[1])
		if err != nil {
			glog.Errorf("newTreasureShareHD triggerTreasure %v", triggerTreasure)
			return nil
		}

		ts.triggerTreasuresRate[triggerTreasure[0]] = rate
	}

	return ts
}

func (ts *treasureShareHD) isOpen(now time.Time) bool {
	return !now.Before(ts.beginTime) && now.Before(ts.endTime)
}

func (ts *treasureShareHD) canTrigger(treasureID string) bool {
	if !ts.isOpen(time.Now()) {
		return false
	}

	if rate, ok := ts.triggerTreasuresRate[treasureID]; ok {
		return rand.Intn(10000) < rate
	}
	return false
}

func doParseTreasureShareHD(gdata gamedata.IGameData) {
	now := time.Now()
	treasureShareHDs = []*treasureShareHD{}
	treasureShareGameData := gdata.(*gamedata.TreasureShareGameData)
	for _, data := range treasureShareGameData.TreasureShares {
		hd := newTreasureShareHD(data)
		if hd != nil && hd.isOpen(now) {
			treasureShareHDs = append(treasureShareHDs, hd)
		}
	}
}

func parseTreasureShareHD() {
	gdata := gamedata.GetGameData(consts.TreasureShare)
	gdata.AddReloadCallback(doParseTreasureShareHD)
	doParseTreasureShareHD(gdata)
}
