package gamedata

import (
	"encoding/json"

	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Diy struct {
	ID        int     `json:"__id__"`
	Skill     []int32 `json:"skill"`
	Ban       []int   `json:"ban"`
	Level     int     `json:"level"`
	MinSingle int     `json:"min_single"`
	MaxSingle int     `json:"max_single"`
	MinTotal  int     `json:"min_total"`
	MaxTotal  int     `json:"max_total"`
	Wine      int     `json:"wine"`
	Book      int     `json:"book"`
}

type DiyGameData struct {
	baseGameData
	diyMap map[int]*Diy // map[id]*Diy
}

func newDiyGameData() *DiyGameData {
	d := &DiyGameData{}
	d.i = d
	return d
}

func (dg *DiyGameData) name() string {
	return consts.Diy
}

func (dg *DiyGameData) init(d []byte) error {
	var _list []*Diy
	err := json.Unmarshal(d, &_list)
	if err != nil {
		return err
	}

	dg.diyMap = make(map[int]*Diy)
	for _, d := range _list {
		dg.diyMap[d.ID] = d
	}

	//glog.Infof("diyMap = %s", dg.diyMap)

	return nil
}

func (dg *DiyGameData) GetDiyData(id int) *Diy {
	return dg.diyMap[id]
}
