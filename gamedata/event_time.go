package gamedata

import (
	"encoding/json"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"kinger/common/consts"
	"strconv"
	"time"
)

type ActivityTime struct {
	ID            int    `json:"__id__"`
	TimeType      int    `json:"type"`
	Start         string `json:"star"`
	End           string `json:"end"`
	DurationDay   int    `json:"length"`
	OpenDayOfWeek []int  `json:"week"`
	ItemStopTime  string `json:"itemEndTime"`

	StartTime           time.Time
	EndTime             time.Time
	ItemEndTime         time.Time
	RegisterFirstFewDay int
}

func (at *ActivityTime) init() error {
	var err error
	switch at.TimeType {
	case consts.CreateDurationDay:
		if at.Start == "" {
			glog.Errorf("Activity time condition no start day, timeID: ", at.ID)
			break
		}
		at.RegisterFirstFewDay, err = strconv.Atoi(at.Start)
		if err != nil {
			return err
		}
	case consts.TimeToTime:
		if at.Start == "" {
			glog.Errorf("Activity time condition no start day, timeID: ", at.ID)
			break
		}
		at.StartTime, err = utils.StringToTime(at.Start, utils.TimeFormat2)
		if err != nil {
			glog.Errorf("StringToTime return error")
			return err
		}

		if at.End == "" {
			if at.DurationDay != 0 {
				at.EndTime = at.StartTime.AddDate(0, 0, at.DurationDay)
			}
		} else {
			at.EndTime, err = utils.StringToTime(at.End, utils.TimeFormat2)
			if err != nil {
				glog.Errorf("StringToTime return error")
				return err
			}
		}

		if at.ItemStopTime != "" {
			at.ItemEndTime, err = utils.StringToTime(at.End, utils.TimeFormat2)
			if err != nil {
				glog.Errorf("StringToTime return error")
				return err
			}
		}

	case consts.DayOfWeek:
	}
	return nil
}

type ActivityTimeGameData struct {
	baseGameData
	ActivityTimeMap map[int]*ActivityTime
}

func newActivityTimeGameData() *ActivityTimeGameData {
	c := &ActivityTimeGameData{}
	c.i = c
	return c
}

func (atd *ActivityTimeGameData) name() string {
	return consts.ActivityTime
}

func (atd *ActivityTimeGameData) init(d []byte) error {
	var l []*ActivityTime
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	atd.ActivityTimeMap = map[int]*ActivityTime{}
	for _, c := range l {
		if err := c.init(); err != nil {
			return err
		}
		atd.ActivityTimeMap[c.ID] = c
	}
	return nil
}
