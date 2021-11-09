package event

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"time"
)

func reloadHuodongConfig(gdata gamedata.IGameData) {
	data := gdata.(*gamedata.HuodongGameData)
	areaGameData := gamedata.GetGameData(consts.AreaConfig).(*gamedata.AreaConfigGameData)
	for htype, huodongData := range data.ID2Huodong {

		var arg *pb.HuodongEvent
		for _, areaCfg := range areaGameData.Areas {

			area := areaCfg.Area
			needRefresh := false
			hd := htypes.Mod.GetHuodong(area, htype)
			if hd == nil {
				hd = htypes.Mod.NewHuodong(area, htype)
				if hd != nil {
					htypes.Mod.AddHuodong(area, hd)
					hd.Save()
					needRefresh = true
				}
			} else {
				needRefresh = hd.Refresh(huodongData)
			}

			if needRefresh {
				if arg == nil {
					arg = &pb.HuodongEvent{HdType: htype, Event: pb.HuodongEventType_HetRefresh}
				}
				arg.Areas = append(arg.Areas, int32(area))
			}
		}

		if arg != nil {
			logic.BroadcastBackend(pb.MessageID_G2G_ON_HUODONG_EVENT, arg)
		}

	}
}

func Initialize() {
	registerRpc()

	if module.Service.GetAppID() == 1 {
		huodongGameData := gamedata.GetGameData(consts.HuodongConfig).(*gamedata.HuodongGameData)
		huodongGameData.AddReloadCallback(reloadHuodongConfig)
	}
}

func GetHuodongItemType(player types.IPlayer, htype pb.HuodongTypeEnum) int {
	hd := htypes.Mod.GetHuodong(player.GetArea(), htype)
	if hd == nil || !hd.IsOpen() {
		return 0
	}

	hd2, ok := hd.(IEventHuodong)
	if !ok {
		return 0
	}

	if time.Now().After(hd2.GetThrowItemEndTime()) {
		return 0
	}

	return hd2.GetEventItemType()
}

func NewEventHdByAttr(area int, htype pb.HuodongTypeEnum, hd IEventHuodong, attr *attribute.AttrMgr) htypes.IHuodong {
	hd.SetImpl(hd)
	hd.SetHtype(htype)
	hd.SetAttr(attr)
	hd.SetTime()
	hd.SetArea(area)
	return hd
}

func NewEventHd(area int, htype pb.HuodongTypeEnum, hd IEventHuodong, attr *attribute.AttrMgr, gdata interface{}) htypes.IHuodong {
	if attr == nil || gdata == nil {
		return nil
	}

	data, ok := gdata.(*gamedata.Huodong)
	if !ok || data == nil {
		return nil
	}

	hd.SetImpl(hd)
	hd.SetHtype(htype)
	hd.SetAttr(attr)
	hd.SetBeginTime(data.StartTime)
	hd.SetEndTime(data.ExchangeStopTime)
	hd.SetThrowItemEndTime(data.StopTime)
	hd.SetVersion(1)
	hd.SetArea(area)
	if !time.Now().Before(hd.GetEndTime()) {
		hd.SetClose(true)
	}
	glog.Infof("NewEventHd %s", hd)
	return hd
}
