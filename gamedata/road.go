package gamedata

import (
	"encoding/json"

	"kinger/common/consts"
	"github.com/pkg/errors"
)

type Road struct {
	ID          int     `json:"__id__"`
	City1       int     `json:"city1"`
	City2       int     `json:"city2"`
	Distance int `json:"distance"`
}

func (r *Road) GetOthCityID(cityID int) int {
	if r.City1 == cityID {
		return r.City2
	} else {
		return r.City1
	}
}

type RoadGameData struct {
	baseGameData
	Roads     []*Road
	ID2Road   map[int]*Road
	City2Road map[int]map[int]*Road
}

func newRoadGameData() *RoadGameData {
	r := &RoadGameData{}
	r.i = r
	return r
}

func (rd *RoadGameData) name() string {
	return consts.Road
}

func (rd *RoadGameData) init(d []byte) error {
	var l []*Road
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	rd.ID2Road = map[int]*Road{}
	rd.City2Road = map[int]map[int]*Road{}
	for _, r := range l {
		rd.ID2Road[r.ID] = r
		rs, ok := rd.City2Road[r.City1]
		if !ok {
			rs = map[int]*Road{}
			rd.City2Road[r.City1] = rs
		}
		rs2, ok := rd.City2Road[r.City2]
		if !ok {
			rs2 = map[int]*Road{}
			rd.City2Road[r.City2] = rs2
		}

		if _, ok := rs[r.City2]; ok {
			return errors.Errorf("fuck road %s", r)
		}
		if _, ok := rs2[r.City1]; ok {
			return errors.Errorf("fuck road %s", r)
		}

		rs[r.City2] = r
		rs2[r.City1] = r
	}

	rd.Roads = l
	return nil
}

func (rd *RoadGameData) GetRoad(city1, city2 int) *Road {
	if rs, ok := rd.City2Road[city1]; ok {
		return rs[city2]
	}
	return nil
}
