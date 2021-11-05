package outstatus

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"strings"
	"time"
)

type iOutStatus interface {
	types.IOutStatus
	onInit()
	onLogin()
	onLogout()
	onHeartBeat()
	onAdd(args ...interface{})
	onDel()
	getAttr() *attribute.MapAttr
	setAttr(attr *attribute.MapAttr)
	isTimeout() bool
	setImp(imp iOutStatus)
	setPlayer(player types.IPlayer)
	setID(statusID string)
}

// 会通知客户端的状态
type iClientOutStatus interface {
	iOutStatus
	syncClientAdd()
	syncClientDel()
}

type baseStatus struct {
	i iOutStatus
	player types.IPlayer
	id string
	attr *attribute.MapAttr
}

func (st *baseStatus) GetID() string {
	return st.id
}

func (st *baseStatus) setID(statusID string) {
	st.id = statusID
}

func (st *baseStatus) onInit() {

}

func (st *baseStatus) onLogin() {

}

func (st *baseStatus) onLogout() {

}

func (st *baseStatus) onHeartBeat() {

}

func (st *baseStatus) onAdd(args ...interface{}) {

}

func (st *baseStatus) onDel() {
	eventhub.Publish(consts.EvDelOutStatus, st.player, st.i)
}

func (st *baseStatus) GetRemainTime() int {
	timeout := st.attr.GetInt64("timeout")
	if timeout < 0 {
		return -1
	}
	remainTime := timeout - time.Now().Unix()
	if remainTime < 0 {
		remainTime = 0
	}
	return int(remainTime)
}

func (st *baseStatus) getAttr() *attribute.MapAttr {
	return st.attr
}

func (st *baseStatus) setAttr(attr *attribute.MapAttr) {
	st.attr = attr
}

func (st *baseStatus) setImp(imp iOutStatus) {
	st.i = imp
}

func (st *baseStatus) setPlayer(player types.IPlayer) {
	st.player = player
}

func (st *baseStatus) isTimeout() bool {
	return st.GetRemainTime() == 0
}

func (st *baseStatus) Over(leftTime int, args ...interface{}) {
	timeout := st.attr.GetInt64("timeout")
	if timeout < 0 {
		return
	}
	st.attr.SetInt64("timeout", timeout + int64(leftTime))
}

func (st *baseStatus) GetLeftTime() int {
	timeout := st.attr.GetInt64("timeout")
	if timeout < 0 {
		return -1
	}
	return int(timeout)
}

func (st *baseStatus) SetLeftTime(leftTime int) {
	st.attr.SetInt64("timeout", int64(leftTime))
}

func (st *baseStatus) PackMsg() *pb.OutStatus {
	return &pb.OutStatus{
		StatusID: st.GetID(),
		RemainTime: int32(st.GetRemainTime()),
		BuffLevel: int32(module.OutStatus.GetBuffLevel(st.player)),
	}
}

// 会通知客户端的状态
type clientStatus struct {
	baseStatus
}

func (st *clientStatus) onAdd(args ...interface{}) {
	st.baseStatus.onAdd(args...)
	st.syncClientAdd()
}

func (st *clientStatus) syncClientAdd() {
	agent := st.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_ADD_OUT_STATUS, st.PackMsg())
	}
}

func (st *clientStatus) syncClientDel() {
	agent := st.player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_DEL_OUT_STATUS, &pb.TargetOutStatus{
			StatusID: st.GetID(),
		})
	}
}

func (st *clientStatus) onDel() {
	st.baseStatus.onDel()
	st.syncClientDel()
}

func (st *clientStatus) Over(leftTime int, args ...interface{}) {
	st.baseStatus.Over(leftTime, args...)
	st.syncClientAdd()
}

func newStatus(player types.IPlayer, statusID string, leftTime int) iOutStatus {
	attr := attribute.NewMapAttr()
	var timeout int64 = -1
	if leftTime >= 0 {
		timeout = time.Now().Unix() + int64(leftTime)
	}
	attr.SetInt64("timeout", timeout)
	return newStatusByAttr(player, statusID, attr)
}

func newStatusByAttr(player types.IPlayer, statusID string, attr *attribute.MapAttr) iOutStatus {
	var st iOutStatus = nil
	sids := strings.Split(statusID, "_")
	sid := fmt.Sprintf("%s_", sids[0])
	switch {
	case statusID == consts.OtVipCard:
		st = &vipCardSt{}
	case statusID == consts.OtMinVipCard:
		st = &clientStatus{}
	case statusID == consts.OtFatigue:
		st = &clientStatus{}
	case statusID == consts.OtAdvertProtecter:
		st = &baseStatus{}
	case sid == consts.OtBuffPrefix:
		st = newBuff(statusID)
	case sid == consts.OtForbid:
		st = newForbid(statusID)
	default:
	}

	if st == nil {
		return nil
	}

	st.setImp(st)
	st.setPlayer(player)
	st.setAttr(attr)
	st.setID(statusID)
	return st
}
