package gamedata

import (
	"encoding/json"
	"fmt"
	"kinger/common/config"
	"kinger/common/consts"
	"math/rand"
)

type iNameGameData interface {
	IGameData
	randomName() string
}

var (
	name1   = newNameGameData1()
	name2   = newNameGameData2()
	name3   = newNameGameData3()
	name4   = newNameGameData4()
	name5   = newNameGameData5()
	nameEn1 = newNameEnGameData1()
	nameEn2 = newNameEnGameData2()
	names1  = []iNameGameData{name1, name2, name3}
	names2  = []iNameGameData{name3, name4, name5}
	namesEn = []iNameGameData{nameEn1, nameEn2}
)

type Name struct {
	ID    int    `json:"__id__"`
	Value string `json:"value"`
}

type NameGameData1 struct {
	baseGameData
	Names []*Name
}

func newNameGameData1() *NameGameData1 {
	r := &NameGameData1{}
	r.i = r
	return r
}

func (nd *NameGameData1) name() string {
	return consts.Name1
}

func (nd *NameGameData1) init(d []byte) error {
	var l []*Name
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	nd.Names = l

	return nil
}

func (nd *NameGameData1) randomName() string {
	return nd.Names[rand.Intn(len(nd.Names))].Value
}

type NameGameData2 struct {
	NameGameData1
}

func newNameGameData2() *NameGameData2 {
	r := &NameGameData2{}
	r.i = r
	return r
}

func (nd *NameGameData2) name() string {
	return consts.Name2
}

type NameGameData3 struct {
	NameGameData1
}

func newNameGameData3() *NameGameData3 {
	r := &NameGameData3{}
	r.i = r
	return r
}

func (nd *NameGameData3) name() string {
	return consts.Name3
}

type NameGameData4 struct {
	NameGameData1
}

func newNameGameData4() *NameGameData4 {
	r := &NameGameData4{}
	r.i = r
	return r
}

func (nd *NameGameData4) name() string {
	return consts.Name4
}

type NameGameData5 struct {
	NameGameData1
}

func newNameGameData5() *NameGameData5 {
	r := &NameGameData5{}
	r.i = r
	return r
}

func (nd *NameGameData5) name() string {
	return consts.Name5
}

type NameEnGameData1 struct {
	NameGameData1
}

func newNameEnGameData1() *NameEnGameData1 {
	r := &NameEnGameData1{}
	r.i = r
	return r
}

func (nd *NameEnGameData1) name() string {
	return consts.NameEn1
}

type NameEnGameData2 struct {
	NameGameData1
}

func newNameEnGameData2() *NameEnGameData2 {
	r := &NameEnGameData2{}
	r.i = r
	return r
}

func (nd *NameEnGameData2) name() string {
	return consts.NameEn2
}

func RandomRobotName(pvpLevel int) string {
	if pvpLevel >= 5 {
		return RandomHighLevelRobotName()
	}
	if config.GetConfig().IsMultiLan {
		return RandomEnRobotName()
	}
	var names []iNameGameData
	if rand.Intn(2) == 0 {
		names = names1
	} else {
		names = names2
	}
	name := ""
	for _, nameData := range names {
		name += nameData.randomName()
	}
	return name
}

func RandomEnRobotName() string {
	return fmt.Sprintf("%s %s", nameEn1.randomName(), nameEn2.randomName())
}
