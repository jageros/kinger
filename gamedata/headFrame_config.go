package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type HeadFrame struct {
	ID         string    `json:"__id__"`
	Name int `json:"name"`
}

func (f *HeadFrame) GetName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(f.Name)
}

type HeadFrameGameData struct {
	baseGameData
	HeadFrames []*HeadFrame
	ID2HeadFrame map[string]*HeadFrame
}

func newHeadFrameGameData() *HeadFrameGameData {
	r := &HeadFrameGameData{}
	r.i = r
	return r
}

func (gd *HeadFrameGameData) name() string {
	return consts.HeadFrame
}

func (gd *HeadFrameGameData) init(d []byte) error {
	var l []*HeadFrame
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.HeadFrames = l
	gd.ID2HeadFrame = map[string]*HeadFrame{}
	for _, h := range l {
		gd.ID2HeadFrame[h.ID] = h
	}

	return nil
}
