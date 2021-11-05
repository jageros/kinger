package forbidlist

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/proto/pb"
	"time"
)

const (
	//attr key
	mgrName_           = "forbid_list"
	listID_            = "forbid_info"
	forbidAccountKey_  = "forbid_account"
	forbidIPKey_       = "forbid_ip"
	monitorAccountKey_ = "monitor_account"
	uid_               = "uid"
	msg_               = "msg"
	opTime_            = "op_time"
)

var (
	forbidIPs       map[string]interface{}
)

func InitForbidIP() {
	forbidIPs = map[string]interface{}{}
	attr := attribute.NewAttrMgr(mgrName_, forbidIPKey_, true)
	aAttr := attr.GetListAttr(listID_)
	if aAttr != nil {
		aAttr.ForEachIndex(func(index int) bool {
			ipAddr := aAttr.GetStr(index)
			forbidIPs[ipAddr] = ipAddr
			return true
		})

	}
}

func AddForbidInfo(uid common.UUid, msg string, forbidType int) {
	var key string
	var flag int
	opTim := time.Now().Unix()
	switch forbidType {
	case consts.ForbidAccount:
		key = forbidAccountKey_
	case consts.ForbidMonitor:
		key = monitorAccountKey_
	}

	attr := attribute.NewAttrMgr(mgrName_, key, true)
	attr.Load()
	aAttr := attr.GetListAttr(listID_)
	if aAttr == nil {
		aAttr = attribute.NewListAttr()
		attr.SetListAttr(listID_, aAttr)
	} else if forbidType == consts.ForbidAccount {
		aAttr.ForEachIndex(func(index int) bool {
			mAttr := aAttr.GetMapAttr(index)
			auid := mAttr.GetUInt64(uid_)
			if auid == uint64(uid) {
				flag = 1
				return false
			}
			return true
		})
		if flag == 1 {
			return
		}
	}
	maAttr := attribute.NewMapAttr()
	maAttr.SetUInt64(uid_, uint64(uid))
	maAttr.SetStr(msg_, msg)
	maAttr.SetInt64(opTime_, opTim)
	aAttr.AppendMapAttr(maAttr)
	attr.Save(false)
}

func DelForbidInfo(uid common.UUid, forbidType int, opTime int64, isDelAll bool) {
	var key string
	switch forbidType {
	case consts.ForbidAccount:
		key = forbidAccountKey_
	case consts.ForbidMonitor:
		key = monitorAccountKey_
	}

	attr := attribute.NewAttrMgr(mgrName_, key, true)
	attr.Load()
	aAttr := attr.GetListAttr(listID_)
	if aAttr != nil {
		if !isDelAll {
			aAttr.ForEachIndex(func(index int) bool {
				maAttr := aAttr.GetMapAttr(index)
				muid := maAttr.GetUInt64(uid_)
				tim := maAttr.GetInt64(opTime_)
				if muid == uint64(uid) && (opTime == 0 || opTime == tim) {
					aAttr.DelByIndex(index)
					return false
				}
				return true
			})
		} else {
			for {
				flag := 0
				aAttr.ForEachIndex(func(index int) bool {
					maAttr := aAttr.GetMapAttr(index)
					muid := maAttr.GetUInt64(uid_)
					if muid == uint64(uid) {
						aAttr.DelByIndex(index)
						flag = 1
						return false
					}
					return true
				})
				if flag == 0 {
					break
				}
			}
		}
	}
	attr.Save(false)
}

func AddForbidIP(ipaddr string) {
	if _, ok := forbidIPs[ipaddr]; !ok {
		forbidIPs[ipaddr] = ipaddr

		if module.Service.GetAppID() != 1 {
			return
		}

		attr := attribute.NewAttrMgr(mgrName_, forbidIPKey_, true)
		attr.Load()
		aAttr := attr.GetListAttr(listID_)
		if aAttr == nil {
			aAttr = attribute.NewListAttr()
			attr.SetListAttr(listID_, aAttr)
		}
		aAttr.AppendStr(ipaddr)
		attr.Save(false)

		arg := &pb.IpAddrArg{
			IpAddr: ipaddr,
			IsForbid: true,
		}
		logic.BroadcastBackend(pb.MessageID_G2G_ON_FORBID_OR_UNBLOCK_IP, arg)
	}
}

func DelForbidIP(ipaddr string) {
	if _, ok := forbidIPs[ipaddr]; ok {
		delete(forbidIPs, ipaddr)
	}

	if module.Service.GetAppID() != 1 {
		return
	}

	attr := attribute.NewAttrMgr(mgrName_, forbidIPKey_, true)
	attr.Load()
	aAttr := attr.GetListAttr(listID_)
	if aAttr != nil {
		aAttr.ForEachIndex(func(index int) bool {
			if aAttr.GetStr(index) == ipaddr {
				aAttr.DelByIndex(index)
				return false
			}
			return true
		})
	}
	attr.Save(false)

	arg := &pb.IpAddrArg{
		IpAddr: ipaddr,
		IsForbid: false,
	}
	logic.BroadcastBackend(pb.MessageID_G2G_ON_FORBID_OR_UNBLOCK_IP, arg)
}

func IsForbidIP(ipaddr string) bool {
	if _, ok := forbidIPs[ipaddr]; ok {
		return true
	}
	return false
}
