package outstatus

import (
	"kinger/common/consts"
	"kinger/apps/game/module"
	"kinger/gamedata"
)

type vipCardSt struct {
	clientStatus
}

func (st *vipCardSt) onAdd(args ...interface{}) {
	st.clientStatus.onAdd(args...)
	isNewbie := false
	argsNum := len(args)
	if argsNum > 0 {
		isNewbie = args[0].(bool)
	}

	needTicket := true
	if argsNum > 1 {
		needTicket = args[1].(bool)
	}

	if !isNewbie {
		st.attr.Del("isNewbie")
	} else {
		st.attr.SetBool("isNewbie", true)
	}

	/*
	agent := st.player.GetAgent()
	if agent != nil {
		msgID := pb.MessageID_S2C_ADD_VIP
		if isNewbie {
			msgID = pb.MessageID_S2C_ADD_NEWBIE_VIP
		}
		agent.PushClient(msgID, &pb.VipRemainTime{
			RemainTime: int32(st.GetRemainTime()),
		})
	}
	*/

	if needTicket {
		module.Player.ModifyResource(st.player, consts.AccTreasureCnt, gamedata.GetGameData(consts.FunctionPrice).(
			*gamedata.FunctionPriceGameData).VipAccTicket)
	}
	module.OutStatus.AddVipBuff(st.player)
}

func (st *vipCardSt) onDel() {
	if st.player.IsVip() {
		return
	}
	st.clientStatus.onDel()

	/*
	msgID := pb.MessageID_S2C_VIP_TIMEOUT
	if st.attr.GetBool("isNewbie") {
		msgID = pb.MessageID_S2C_NEWBIE_VIP_TIMEOUT
	}

	timer.AfterFunc(2 * time.Second, func() {
		agent := st.player.GetAgent()
		if agent != nil {
			agent.PushClient(msgID, nil)
		}
	})
	*/
}

func (st *vipCardSt) Over(leftTime int, args ...interface{}) {
	st.baseStatus.Over(leftTime, args...)
	st.onAdd()
}

func (b *vipCardSt) onLogin() {
	module.OutStatus.AddVipBuff(b.player)
}