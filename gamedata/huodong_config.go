package gamedata

import (
	"encoding/json"

	"github.com/pkg/errors"
	"kinger/common/consts"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"time"
)

type Huodong struct {
	ID              pb.HuodongTypeEnum `json:"__id__"`
	BeginTime       string             `json:"beginTime"`
	EndTime         string             `json:"endTime"`
	ExchangeEndTime string             `json:"exchangeEndTime"`
	Resource        string             `json:"resource"`
	Reward          string             `json:"reward"`

	StartTime        time.Time
	StopTime         time.Time
	ExchangeStopTime time.Time
}

func (h *Huodong) init() error {
	var err error
	if h.BeginTime == "" {
		return errors.Errorf("huodong %d no begin time", h.ID)
	}

	h.StartTime, err = utils.StringToTime(h.BeginTime, utils.TimeFormat2)
	if err != nil {
		return err
	}

	if h.EndTime == "" {
		return errors.Errorf("huodong %d no end time", h.ID)
	}

	h.StopTime, err = utils.StringToTime(h.EndTime, utils.TimeFormat2)
	if err != nil {
		return err
	}

	if h.ExchangeEndTime == "" {
		return errors.Errorf("huodong %d no exchangeEndTime time", h.ID)
	}

	h.ExchangeStopTime, err = utils.StringToTime(h.ExchangeEndTime, utils.TimeFormat2)
	return err
}

type HuodongGameData struct {
	baseGameData
	ID2Huodong map[pb.HuodongTypeEnum]*Huodong
}

func newHuodongGameData() *HuodongGameData {
	c := &HuodongGameData{}
	c.i = c
	return c
}

func (cg *HuodongGameData) name() string {
	return consts.HuodongConfig
}

func (cg *HuodongGameData) init(d []byte) error {
	var l []*Huodong
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	cg.ID2Huodong = map[pb.HuodongTypeEnum]*Huodong{}
	for _, c := range l {
		if err := c.init(); err != nil {
			return err
		}
		cg.ID2Huodong[c.ID] = c
	}

	return nil
}
