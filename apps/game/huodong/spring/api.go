package spring

import (
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/huodong/event"
)

func Initialize() {
	initAllGoods()
	registerRpc()
}

func NewSpringHdByAttr(area int, attr *attribute.AttrMgr) htypes.IHuodong {
	return event.NewEventHdByAttr(area, pb.HuodongTypeEnum_HSpringExchange, &springHd{}, attr)
}

func NewSpringHd(area int, attr *attribute.AttrMgr, gdata interface{}) htypes.IHuodong {
	return event.NewEventHd(area, pb.HuodongTypeEnum_HSpringExchange, &springHd{}, attr, gdata)
}
