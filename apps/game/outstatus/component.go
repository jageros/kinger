package outstatus

import (
	"fmt"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
)

var _ types.IPlayerComponent = &outstatusComponent{}

type outstatusComponent struct {
	attr      *attribute.MapAttr
	player    types.IPlayer
	allStatus map[string]iOutStatus
}

func (oc *outstatusComponent) ComponentID() string {
	return consts.OutStatusCpt
}

func (oc *outstatusComponent) GetPlayer() types.IPlayer {
	return oc.player
}

func (oc *outstatusComponent) OnInit(player types.IPlayer) {
	oc.player = player
	oc.allStatus = map[string]iOutStatus{}
	oc.attr.ForEachKey(func(statusID string) {
		st := newStatusByAttr(player, statusID, oc.attr.GetMapAttr(statusID))
		if st != nil {
			st.onInit()
		}
		oc.allStatus[statusID] = st
	})
}

func (oc *outstatusComponent) OnLogin(isRelogin, isRestore bool) {
	for statusID, _ := range oc.allStatus {
		st := oc.getStatus(statusID)
		if st != nil {
			st.onLogin()
		}
	}
}

func (oc *outstatusComponent) OnLogout() {
	for statusID, _ := range oc.allStatus {
		st := oc.getStatus(statusID)
		if st != nil {
			st.onLogout()
		}
	}
}

func (oc *outstatusComponent) OnHeartBeat() {
	for statusID, _ := range oc.allStatus {
		st := oc.getStatus(statusID)
		if st != nil {
			st.onHeartBeat()
		}
	}
}

func (oc *outstatusComponent) delStatus(statusID string) {
	st, ok := oc.allStatus[statusID]
	if !ok {
		return
	}

	delete(oc.allStatus, statusID)
	oc.attr.Del(statusID)
	st.onDel()
}

func (oc *outstatusComponent) getStatus(statusID string) iOutStatus {
	if st, ok := oc.allStatus[statusID]; ok {
		if st.isTimeout() {
			oc.delStatus(statusID)
			return nil
		}
		return st
	} else {
		return nil
	}
}

func (oc *outstatusComponent) addStatus(statusID string, leftTime int, args ...interface{}) types.IOutStatus {
	st := oc.getStatus(statusID)
	if st != nil {
		return st
	}
	st = newStatus(oc.player, statusID, leftTime)
	if st == nil {
		return nil
	}

	oc.allStatus[statusID] = st
	oc.attr.SetMapAttr(statusID, st.getAttr())
	st.onAdd(args...)
	return st
}

func (oc *outstatusComponent) forEachBuff(callback func(b iBuff)) {
	for statusID, _ := range oc.allStatus {
		st := oc.getStatus(statusID)
		if st == nil {
			continue
		}
		if b, ok := st.(iBuff); ok {
			callback(b)
		}
	}
}

func (oc *outstatusComponent) forEachClientStatus(callback func(st iClientOutStatus)) {
	for statusID, _ := range oc.allStatus {
		st := oc.getStatus(statusID)
		if st == nil {
			continue
		}
		if b, ok := st.(iClientOutStatus); ok {
			callback(b)
		}
	}
}

func (oc *outstatusComponent) getBuffLevel() int {
	var count, lvl int
	privGameData := gamedata.GetGameData(consts.PrivConfig).(*gamedata.PrivilegeGameData)
	lv2Num := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).PrivLevel2Num
	if privGameData == nil {
		return 1
	}
	vipBuffId12 := fmt.Sprintf("%s%d", consts.OtBuffPrefix, consts.PrivDoubleRewardOfVip)
	vipBuffId13 := fmt.Sprintf("%s%d", consts.OtBuffPrefix, consts.PrivAddCardOfVip)
	vipBuffId14 := fmt.Sprintf("%s%d", consts.OtBuffPrefix, consts.PrivAutoOpenTreasureOfVip)

	for _, priv := range privGameData.Privileges {
		ot := mod.GetBuff(oc.player, priv.ID)
		if ot != nil && ot.GetID() != vipBuffId12 && ot.GetID() != vipBuffId13 && ot.GetID() != vipBuffId14 {
			count++
		}
	}
	for lv, num := range lv2Num {
		if count >= num {
			if lv > lvl {
				lvl = lv
			}
		}
	}
	return lvl
}

func (oc *outstatusComponent) hasAllPriv() bool {
	privGameData := gamedata.GetGameData(consts.PrivConfig).(*gamedata.PrivilegeGameData)
	lv2Num := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).PrivLevel2Num
	var count int
	for _, priv := range privGameData.Privileges {
		if priv.ID != consts.PrivTreasureCnt && priv.ID != consts.PrivAddPvpGold && priv.ID != consts.PrivDoubleRewardOfVip &&
			priv.ID != consts.PrivAddCardOfVip && priv.ID != consts.PrivAutoOpenTreasureOfVip {
			count++
		}
	}
	cur_num := lv2Num[oc.getBuffLevel()]
	return cur_num == count
}

func (oc *outstatusComponent) getBuffEffect(buffID int) []int {
	privGameData := gamedata.GetGameData(consts.PrivConfig).(*gamedata.PrivilegeGameData)
	if privGameData == nil {
		return nil
	}
	if p := privGameData.GetPrivilegeByID(buffID); p != nil {
		buffLevel := oc.getBuffLevel()
		switch buffLevel {
		case 4:
			return p.Level4Buff
		case 3:
			return p.Level3Buff
		case 2:
			return p.Level2Buff
		default:
			return p.Level1Buff
		}
	}
	return nil
}
