package event

import (
	htypes "kinger/apps/game/huodong/types"
)

type IEventHdPlayerData interface {
	htypes.IHdPlayerData
	GetVersion() int
	Reset(version int)
}

type EventPlayerData struct {
	htypes.BaseHdPlayerData
}

func (hpd *EventPlayerData) GetVersion() int {
	return hpd.Attr.GetInt("version")
}

func (hpd *EventPlayerData) Reset(version int)  {
	hpd.Attr.SetInt("version", version)
}
