package gamedata

import (
	"time"
	"kinger/common/consts"
	"encoding/json"
	"kinger/gopuppy/common/utils"
	"kinger/gopuppy/common/glog"
	"github.com/pkg/errors"
	"kinger/common/config"
)

var _ ISeasonPvp = &SeasonPvp{}
var _ ISeasonPvp = &MultiLanSeasonPvp{}

type ISeasonPvp interface {
	GetID() int
	GetStartTime() time.Time
	GetStopTime() time.Time
	GetLimitPvpTeam() int
	GetHandCardType() []int
	GetDefenseSkills1() []int32
	GetDefenseSkills2() []int32
	GetFortifications() [][]int
	GetChangeHandType() []int
}

type ISeasonPvpGameData interface {
	IGameData
	GetSeasonData(area int) ISeasonPvp
	IsSeasonEqual(id1, id2 int) bool
}

type SeasonPvp struct {
	ID             int     `json:"__id__"`
	BeginTime      string  `json:"beginTime"`
	EndTime        string  `json:"endTime"`
	LimitPvpTeam   int     `json:"team"`
	HandCardType   []int   `json:"handCardType"`
	DefenseSkills1 []int32 `json:"player1Buff"`
	DefenseSkills2 []int32 `json:"player2Buff"`
	Fortifications [][]int `json:"battleCard"`
	ChangeHandType []int   `json:"again"`
	Areas          [][]int `json:"areas"`

	StartTime time.Time
	StopTime time.Time
	areaLimit *AreaLimitConfig
	equalSeason []int
	notEqualSeason []int
}

func checkSeasonPvpData(session int, handCardType, changeHandType []int, fortifications [][]int) (err error) {
	switch len(handCardType) {
	case 0:
	case 1:
		if handCardType[0] != 1 {
			err = errors.Errorf("SeasonPvp init error, session=%d, HandCardType=%v, err=%s", session, handCardType, err)
		}
	case 2:
		if !(handCardType[0] == 3 || handCardType[0] == 4) || len(changeHandType) != 2 {
			err = errors.Errorf("SeasonPvp init error, session=%d, HandCardType=%v, err=%s", session, handCardType, err)
		}
	case 3:
		if handCardType[0] != 2 {
			err = errors.Errorf("SeasonPvp init error, session=%d, HandCardType=%v, err=%s", session, handCardType, err)
		}
	default:
		err = errors.Errorf("SeasonPvp init error, session=%d, HandCardType=%v, err=%s", session, handCardType, err)
	}
	if err != nil {
		glog.Errorf(err.Error())
		return err
	}

	for _, fort := range fortifications {
		if len(fort) != 2 {
			err = errors.Errorf("SeasonPvp init error, session=%d, Fortifications=%v, err=%s", session, fortifications, err)
			glog.Errorf(err.Error())
			return err
		}
	}
	return err
}

func (s *SeasonPvp) init() error {
	s.areaLimit = newAreaLimitConfig(s.Areas)
	var err error
	s.StartTime, err = utils.StringToTime(s.BeginTime, utils.TimeFormat2)
	if err != nil {
		glog.Errorf("SeasonPvp init error, session=%d, beginTime=%s, err=%s", s.ID, s.BeginTime, err)
		return err
	}

	s.StopTime, err = utils.StringToTime(s.EndTime, utils.TimeFormat2)
	if err != nil {
		glog.Errorf("SeasonPvp init error, session=%d, endTime=%s, err=%s", s.ID, s.EndTime, err)
		return err
	}

	return checkSeasonPvpData(s.ID, s.HandCardType, s.ChangeHandType, s.Fortifications)
}

func (s *SeasonPvp) isEqual(oth *SeasonPvp) bool {
	for _, id := range s.equalSeason {
		if id == oth.ID {
			return true
		}
	}

	for _, id := range s.notEqualSeason {
		if id == oth.ID {
			return false
		}
	}

	if len(s.HandCardType) != len(oth.HandCardType) {
		s.notEqualSeason = append(s.notEqualSeason, oth.ID)
		return false
	}
	for i, v := range s.HandCardType {
		if v != oth.HandCardType[i] {
			s.notEqualSeason = append(s.notEqualSeason, oth.ID)
			return false
		}
	}

	if len(s.DefenseSkills1) != len(oth.DefenseSkills1) {
		s.notEqualSeason = append(s.notEqualSeason, oth.ID)
		return false
	}
	for i, v := range s.DefenseSkills1 {
		if v != oth.DefenseSkills1[i] {
			s.notEqualSeason = append(s.notEqualSeason, oth.ID)
			return false
		}
	}

	if len(s.DefenseSkills2) != len(oth.DefenseSkills2) {
		s.notEqualSeason = append(s.notEqualSeason, oth.ID)
		return false
	}
	for i, v := range s.DefenseSkills2 {
		if v != oth.DefenseSkills2[i] {
			s.notEqualSeason = append(s.notEqualSeason, oth.ID)
			return false
		}
	}

	if len(s.Fortifications) != len(oth.Fortifications) {
		s.notEqualSeason = append(s.notEqualSeason, oth.ID)
		return false
	}

	for i, info := range s.Fortifications {
		othInfo := oth.Fortifications[i]
		if len(info) != len(othInfo) {
			s.notEqualSeason = append(s.notEqualSeason, oth.ID)
			return false
		}

		for i, v := range info {
			if v != othInfo[i] {
				s.notEqualSeason = append(s.notEqualSeason, oth.ID)
				return false
			}
		}
	}

	if len(s.ChangeHandType) != len(oth.ChangeHandType) {
		s.notEqualSeason = append(s.notEqualSeason, oth.ID)
		return false
	}
	for i, v := range s.ChangeHandType {
		if v != oth.ChangeHandType[i] {
			s.notEqualSeason = append(s.notEqualSeason, oth.ID)
			return false
		}
	}

	s.equalSeason = append(s.equalSeason, oth.ID)
	return true
}

func (s *SeasonPvp) GetID() int {
	return s.ID
}

func (s *SeasonPvp) GetStartTime() time.Time {
	return s.StartTime
}

func (s *SeasonPvp) GetStopTime() time.Time {
	return s.StopTime
}

func (s *SeasonPvp) GetLimitPvpTeam() int {
	return s.LimitPvpTeam
}

func (s *SeasonPvp) GetHandCardType() []int {
	return s.HandCardType
}

func (s *SeasonPvp) GetDefenseSkills1() []int32 {
	return s.DefenseSkills1
}

func (s *SeasonPvp) GetDefenseSkills2() []int32 {
	return s.DefenseSkills2
}

func (s *SeasonPvp) GetFortifications() [][]int {
	return s.Fortifications
}

func (s *SeasonPvp) GetChangeHandType() []int {
	return s.ChangeHandType
}

type SeasonPvpGameData struct {
	baseGameData
	ID2Season   map[int]*SeasonPvp
	Seasons     []*SeasonPvp
	area2Season map[int]*SeasonPvp
}

func newSeasonPvpGameData() *SeasonPvpGameData {
	gd := &SeasonPvpGameData{}
	gd.i = gd
	return gd
}

func (gd *SeasonPvpGameData) name() string {
	return consts.SeasonPvp
}

func (gd *SeasonPvpGameData) init(d []byte) error {
	var l []*SeasonPvp

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.ID2Season = map[int]*SeasonPvp{}
	gd.Seasons = []*SeasonPvp{}
	gd.area2Season = map[int]*SeasonPvp{}
	for _, s := range l {
		err := s.init()
		if err != nil {
			continue
		}
		gd.ID2Season[s.ID] = s
		gd.Seasons = append(gd.Seasons, s)

		s.areaLimit.forEachArea(func(area int) {
			gd.area2Season[area] = s
		})
	}

	return nil
}

func (gd *SeasonPvpGameData) GetNextSeason(session int) ISeasonPvp {
	now := time.Now()
	for _, s := range gd.Seasons {
		if s.ID <= session {
			continue
		}
		if s.StopTime.After(now) {
			return s
		}
	}
	return nil
}

func (gd *SeasonPvpGameData) GetCurSeason(session int) ISeasonPvp {
	if data, ok := gd.ID2Season[session]; ok {
		return data
	} else {
		return nil
	}
}

func (gd *SeasonPvpGameData) GetSeasonData(area int) ISeasonPvp {
	if data, ok := gd.area2Season[area]; ok {
		return data
	} else {
		return nil
	}
}

func (gd *SeasonPvpGameData) IsSeasonEqual(id1, id2 int) bool {
	if id1 == id2 {
		return true
	}

	data1 := gd.ID2Season[id1]
	if data1 == nil {
		return false
	}

	data2 := gd.ID2Season[id2]
	if data2 == nil {
		return false
	}

	return data1.isEqual(data2)
}

type MultiLanSeasonPvp struct {
	ID             int     `json:"__id__"`
	BeginDay       int     `json:"beginDay"`
	EndDay         int     `json:"endDay"`
	LimitPvpTeam   int     `json:"team"`
	HandCardType   []int   `json:"handCardType"`
	DefenseSkills1 []int32 `json:"player1Buff"`
	DefenseSkills2 []int32 `json:"player2Buff"`
	Fortifications [][]int `json:"battleCard"`
	ChangeHandType []int   `json:"again"`
}

func (s *MultiLanSeasonPvp) GetSession() int {
	return s.ID
}

func (s *MultiLanSeasonPvp) GetStartTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), s.BeginDay, 0, 0, 0, 0, now.Location())
}

func (s *MultiLanSeasonPvp) GetStopTime() time.Time {
	now := time.Now()
	month := now.Month()
	if s.EndDay <= s.BeginDay {
		if month == time.December {
			month = time.January
		} else {
			month++
		}
	}

	date := time.Date(now.Year(), month, s.EndDay, 0, 0, 0, 0, now.Location())
	return time.Unix(date.Unix() - 1, 0)
}

func (s *MultiLanSeasonPvp) GetLimitPvpTeam() int {
	return s.LimitPvpTeam
}

func (s *MultiLanSeasonPvp) GetHandCardType() []int {
	return s.HandCardType
}

func (s *MultiLanSeasonPvp) GetDefenseSkills1() []int32 {
	return s.DefenseSkills1
}

func (s *MultiLanSeasonPvp) GetDefenseSkills2() []int32 {
	return s.DefenseSkills2
}

func (s *MultiLanSeasonPvp) GetFortifications() [][]int {
	return s.Fortifications
}

func (s *MultiLanSeasonPvp) GetChangeHandType() []int {
	return s.ChangeHandType
}

func (s *MultiLanSeasonPvp) GetID() int {
	return s.ID
}

type MultiLanSeasonPvpGameData struct {
	baseGameData
	Data *MultiLanSeasonPvp
}

func newMultiLanSeasonPvpGameData() *MultiLanSeasonPvpGameData {
	gd := &MultiLanSeasonPvpGameData{}
	gd.i = gd
	return gd
}

func (gd *MultiLanSeasonPvpGameData) name() string {
	return consts.SeasonPvpHandjoy
}

func (gd *MultiLanSeasonPvpGameData) init(d []byte) error {
	var l []*MultiLanSeasonPvp

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	data := l[0]
	if checkSeasonPvpData(data.ID, data.HandCardType, data.ChangeHandType, data.Fortifications) != nil {
		return err
	}
	gd.Data = data

	return nil
}

func (gd *MultiLanSeasonPvpGameData) GetNextSeason(session int) ISeasonPvp {
	gd.Data.ID = session + 1
	return gd.Data
}

func (gd *MultiLanSeasonPvpGameData) GetCurSeason(session int) ISeasonPvp {
	gd.Data.ID = session
	return gd.Data
}

func (gd *MultiLanSeasonPvpGameData) GetSeasonData(area int) ISeasonPvp {
	return gd.Data
}

func (gd *MultiLanSeasonPvpGameData) IsSeasonEqual(id1, id2 int) bool {
	return true
}

func GetSeasonPvpGameData() ISeasonPvpGameData {
	if config.GetConfig().IsMultiLan {
		return GetGameData(consts.SeasonPvpHandjoy).(*MultiLanSeasonPvpGameData)
	} else {
		return GetGameData(consts.SeasonPvp).(*SeasonPvpGameData)
	}
}
