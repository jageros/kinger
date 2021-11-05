package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type Privilege struct {
	ID         int   `json:"__id__"`
	Level1Buff []int `json:"level1Buff"`
	Level2Buff []int `json:"level2Buff"`
	Level3Buff []int `json:"level3Buff"`
	Level4Buff []int `json:"level4Buff"`
}

type PrivilegeGameData struct {
	baseGameData
	Privileges []*Privilege
}

func newPrivilegeGameData() *PrivilegeGameData {
	gd := &PrivilegeGameData{}
	gd.i = gd
	return gd
}

func (gd *PrivilegeGameData) name() string {
	return consts.PrivConfig
}

func (gd *PrivilegeGameData) init(d []byte) error {
	var l []*Privilege

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	gd.Privileges = l

	return nil
}

func (gd *PrivilegeGameData) GetPrivilegeByID(id int) *Privilege {
	for _, p := range gd.Privileges {
		if p.ID == id {
			return p
		}
	}
	return nil
}
