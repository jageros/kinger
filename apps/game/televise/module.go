package televise

import (
	"kinger/apps/game/module"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"time"
)

var mod *televiseModule

type televiseModule struct {
}

func newTeleviseModule() *televiseModule {
	b := &televiseModule{}
	return b
}

func (tm *televiseModule) SendNotice(televiseType pb.TeleviseEnum, args ...interface{}) {
	msg := &pb.Televise{
		TeleviseType: televiseType,
		Arg:          tm.encodeArgs(televiseType, args),
	}

	api.BroadcastClient(pb.MessageID_S2C_PUSH_TELEVISE, msg, nil)
}

func (tm *televiseModule) encodeArgs(televiseType pb.TeleviseEnum, args []interface{}) []byte {
	var payload []byte
	if len(args) <= 0 {
		return payload
	}

	if televiseType == pb.TeleviseEnum_CardLevelupPurple || televiseType == pb.TeleviseEnum_CardLevelupOrange ||
		televiseType == pb.TeleviseEnum_RecruitGetCard || televiseType == pb.TeleviseEnum_GetSpecialCard ||
		televiseType == pb.TeleviseEnum_NewbieGiftGetCard || televiseType == pb.TeleviseEnum_LimitGiftGetCard {

		arg, _ := (&pb.TeleviseCardArg{
			PlayerName: args[0].(string),
			CardID:     args[1].(uint32),
		}).Marshal()
		return arg

	} else if televiseType == pb.TeleviseEnum_ClearanceLevel || televiseType == pb.TeleviseEnum_BuyVip ||
		televiseType == pb.TeleviseEnum_BuyGrowPlan {
		arg, _ := (&pb.TeleviseUidArg{
			PlayerName: args[0].(string),
		}).Marshal()
		return arg

	} else if televiseType == pb.TeleviseEnum_RankPre3Login {
		arg, _ := (&pb.TeleviseRankLoginArg{
			PlayerName: args[0].(string),
			Ranking:    args[1].(uint32),
		}).Marshal()
		return arg

	} else {
		return payload
	}
}

func onLogin(ev evq.IEvent) {
	uid := ev.(*evq.CommonEvent).GetData()[0].(common.UUid)
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return
	}

	if ranking, ok := module.Rank.GetRanking(player); ok {
		timer.AfterFunc(250*time.Millisecond, func() {
			module.Televise.SendNotice(pb.TeleviseEnum_RankPre3Login, player.GetName(), ranking)
		})
	}
}

func Initialize() {
	mod = newTeleviseModule()
	module.Televise = mod

	//evq.HandleEvent(consts.EvLogin, onLogin)
}
