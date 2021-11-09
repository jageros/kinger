package types

import (
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"time"
)

var Mod IMod

type IMod interface {
	GetHuodong(area int, htype pb.HuodongTypeEnum) IHuodong
	NewHuodong(area int, htype pb.HuodongTypeEnum) IHuodong
	AddHuodong(area int, hd IHuodong)
}

type IHuodong interface {
	IsOpen() bool
	IsClose() bool
	SetClose(closed bool)
	GetHtype() pb.HuodongTypeEnum
	GetVersion() int
	SetVersion(version int)
	OnStart()
	OnStop()
	Refresh(gdata interface{}) bool
	Save()
	HeartBeat(now time.Time) chan pb.HuodongEventType
	GetBeginTime() time.Time
	GetEndTime() time.Time
	NewPlayerData(player types.IPlayer) IHdPlayerData
	NewPlayerDataByAttr(player types.IPlayer, attr *attribute.MapAttr) IHdPlayerData
	OnPlayerLogin(player types.IPlayer, hdData IHdPlayerData)
}

type IEventHuodong interface {
	IHuodong
	PackEventMsg() *pb.HuodongData
}

type IHdPlayerData interface {
	GetAttr() *attribute.MapAttr
}

type IHuodongComponent interface {
	GetOrNewHdData(htype pb.HuodongTypeEnum) IHdPlayerData
}

type BaseHdPlayerData struct {
	Player types.IPlayer
	Attr   *attribute.MapAttr
}

func (hpd *BaseHdPlayerData) GetAttr() *attribute.MapAttr {
	return hpd.Attr
}

type BaseHuodong struct {
	I         IHuodong
	Attr      *attribute.AttrMgr
	htype     pb.HuodongTypeEnum
	beginTime time.Time
	endTime   time.Time
}

func (hd *BaseHuodong) IsOpen() bool {
	return hd.Attr.GetBool("opened") && !hd.IsClose()
}

func (hd *BaseHuodong) IsClose() bool {
	return hd.Attr.GetBool("closed")
}

func (hd *BaseHuodong) SetClose(closed bool) {
	hd.Attr.SetBool("closed", closed)
}

func (hd *BaseHuodong) setOpen(opened bool) {
	hd.Attr.SetBool("opened", opened)
}

func (hd *BaseHuodong) GetHtype() pb.HuodongTypeEnum {
	return hd.htype
}

func (hd *BaseHuodong) SetHtype(htype pb.HuodongTypeEnum) {
	hd.htype = htype
}

func (hd *BaseHuodong) GetVersion() int {
	return hd.Attr.GetInt("version")
}

func (hd *BaseHuodong) SetVersion(version int) {
	hd.Attr.SetInt("version", version)
}

func (hd *BaseHuodong) Save() {
	err := hd.Attr.Save(true)
	if err != nil {
		glog.Errorf("huodong Save error, htype=%d, err=%s", hd.htype, err)
	}
}

func (hd *BaseHuodong) SetBeginTime(t time.Time) {
	hd.Attr.SetInt64("beginTime", t.Unix())
	hd.beginTime = t
}

func (hd *BaseHuodong) SetEndTime(t time.Time) {
	hd.Attr.SetInt64("endTime", t.Unix())
	hd.endTime = t
}

func (hd *BaseHuodong) SetTime() {
	hd.beginTime = time.Unix(hd.Attr.GetInt64("beginTime"), 0)
	hd.endTime = time.Unix(hd.Attr.GetInt64("endTime"), 0)
}

func (hd *BaseHuodong) GetBeginTime() time.Time {
	return hd.beginTime
}

func (hd *BaseHuodong) GetEndTime() time.Time {
	return hd.endTime
}

func (hd *BaseHuodong) OnStart() {
	hd.Attr.SetBool("opened", true)
	hd.Attr.SetBool("closed", false)
	glog.Infof("huodong OnStart htype=%d, version=%d", hd.GetHtype(), hd.GetVersion())
}

func (hd *BaseHuodong) OnStop() {
	hd.Attr.SetBool("opened", false)
	hd.Attr.SetBool("closed", true)
	glog.Infof("huodong OnStop htype=%d, version=%d", hd.GetHtype(), hd.GetVersion())
}

func (hd *BaseHuodong) HeartBeat(now time.Time) chan pb.HuodongEventType {
	closed := hd.Attr.GetBool("closed")
	if closed {
		return nil
	}
	opened := hd.Attr.GetBool("opened")
	if !opened {
		if !now.Before(hd.beginTime) {
			c := make(chan pb.HuodongEventType, 1)
			evq.CallLater(func() {
				hd.I.OnStart()
				hd.Attr.Save(false)
				c <- pb.HuodongEventType_HetStart
			})
			return c
		}
	} else {
		if !now.Before(hd.endTime) {
			c := make(chan pb.HuodongEventType, 1)
			evq.CallLater(func() {
				hd.I.OnStop()
				hd.Attr.Save(false)
				c <- pb.HuodongEventType_HetStop
			})
			return c
		}
	}
	return nil
}
