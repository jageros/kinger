package event

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"time"
)

// 读event_config表的活动
type IEventHuodong interface {
	htypes.IHuodong
	HasEventItem(player types.IPlayer, amount int) bool
	SubEventItem(player types.IPlayer, amount int)
	GetEventItemType() int
	GetThrowItemEndTime() time.Time
	SetImpl(impl IEventHuodong)
	SetAttr(attr *attribute.AttrMgr)
	PackEventDetailMsg(data IEventHdPlayerData) proto.Marshaler
	PackEventMsg() *pb.HuodongData
	SetTime()
	SetHtype(htype pb.HuodongTypeEnum)
	SetBeginTime(t time.Time)
	SetEndTime(t time.Time)
	SetThrowItemEndTime(t time.Time)
	SetArea(area int)
}

type EventHd struct {
	area int
	htypes.BaseHuodong
	throwItemEndTime time.Time
}

func (hd *EventHd) SetImpl(impl IEventHuodong) {
	hd.I = impl
}

func (hd *EventHd) SetAttr(attr *attribute.AttrMgr) {
	hd.Attr = attr
}

func (hd *EventHd) PackEventMsg() *pb.HuodongData {
	msg := &pb.HuodongData{
		Type: hd.GetHtype(),
	}

	now := time.Now()
	var remainTime float64
	endTime := hd.GetThrowItemEndTime()
	if endTime.After(now) {
		remainTime = endTime.Sub(now).Seconds()
	}
	msg.RemainTime = int32(remainTime)

	remainTime = 0
	endTime = hd.GetEndTime()
	if endTime.After(now) {
		remainTime = endTime.Sub(now).Seconds()
	}
	msg.RemainExchangeTime = int32(remainTime)
	return msg
}

func (hd *EventHd) String() string {
	return fmt.Sprintf("[EventHd htype=%s, version=%d, beginTime=%s, endTime=%s, isOpen=%v, isClose=%v]",
		hd.GetHtype(), hd.GetVersion(), hd.GetBeginTime(), hd.GetEndTime(), hd.IsOpen(), hd.IsClose())
}

func (hd *EventHd) SetTime() {
	hd.BaseHuodong.SetTime()
	hd.throwItemEndTime = time.Unix(hd.Attr.GetInt64("throwItemEndTime"), 0)
}

func (hd *EventHd) GetThrowItemEndTime() time.Time {
	return hd.throwItemEndTime
}

func (hd *EventHd) SetThrowItemEndTime(t time.Time) {
	hd.throwItemEndTime = t
	hd.Attr.SetInt64("throwItemEndTime", t.Unix())
}

func (hd *EventHd) OnStart() {
	hd.BaseHuodong.OnStart()

	now := time.Now()
	remainTimef := hd.GetThrowItemEndTime().Sub(now).Seconds()
	if remainTimef < 0 {
		remainTimef = 0
	}

	exchangeRemainTimef := hd.GetEndTime().Sub(now).Seconds()
	if exchangeRemainTimef < 0 {
		exchangeRemainTimef = 0
	}

	beginArg := &pb.HuodongData{
		Type:               hd.GetHtype(),
		RemainTime:         int32(remainTimef),
		RemainExchangeTime: int32(exchangeRemainTimef),
	}

	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		if player.GetArea() != hd.area {
			return
		}

		hdCpt := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent)
		hdData := hdCpt.GetOrNewHdData(hd.GetHtype())
		if hdData == nil {
			return
		}
		hd.OnPlayerLogin(player, hdData)

		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_HUODONG_BEGIN, beginArg)
		}
	})
}

func (hd *EventHd) OnStop() {
	hd.BaseHuodong.OnStop()

	arg := &pb.TargetHuodong{
		Type: hd.GetHtype(),
	}

	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		if player.GetArea() != hd.area {
			return
		}

		agent := player.GetAgent()
		if agent != nil {
			agent.PushClient(pb.MessageID_S2C_HUODONG_END, arg)
		}
	})
}

func (hd *EventHd) Refresh(gdata interface{}) bool {
	if gdata == nil {
		return false
	}

	data, ok := gdata.(*gamedata.Huodong)
	if !ok || data == nil {
		return false
	}

	startTime := data.StartTime
	throwItemEndTime := data.StopTime
	stopTime := data.ExchangeStopTime
	if hd.GetBeginTime().Equal(startTime) && hd.GetEndTime().Equal(stopTime) && hd.throwItemEndTime.Equal(throwItemEndTime) {
		return false
	}

	if hd.IsClose() {
		hd.SetVersion(hd.GetVersion() + 1)
		hd.SetClose(false)
	}
	hd.SetBeginTime(startTime)
	hd.SetEndTime(stopTime)
	hd.SetThrowItemEndTime(throwItemEndTime)
	hd.Save()
	return true
}

func (hd *EventHd) OnPlayerLogin(player types.IPlayer, hdData htypes.IHdPlayerData) {
	if hdData == nil {
		return
	}

	spData, ok := hdData.(IEventHdPlayerData)
	if !ok {
		return
	}

	version := hd.GetVersion()
	if spData.GetVersion() != version {
		spData.Reset(version)
	}
}

func (hd *EventHd) SetArea(area int) {
	hd.area = area
}

func (hd *EventHd) HasEventItem(player types.IPlayer, amount int) bool {
	// TODO
	return module.Player.HasResource(player, consts.EventItem1, amount)
}

func (hd *EventHd) SubEventItem(player types.IPlayer, amount int) {
	module.Player.ModifyResource(player, consts.EventItem1, -amount)
}

func (hd *EventHd) GetEventItemType() int {
	return consts.EventItem1
}
