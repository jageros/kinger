package event

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
)

func rpc_C2S_FetchHuodongDetail(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetHuodong)
	hd := htypes.Mod.GetHuodong(player.GetArea(), arg2.Type)
	if hd == nil || !hd.IsOpen() {
		return nil, gamedata.InternalErr
	}

	evHd, ok := hd.(IEventHuodong)
	if !ok {
		return nil, gamedata.InternalErr
	}

	pdata := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(arg2.Type)
	if pdata == nil {
		return nil, gamedata.InternalErr
	}

	pdata2, ok := pdata.(IEventHdPlayerData)
	if !ok {
		return nil, gamedata.InternalErr
	}

	reply := &pb.HuodongDetail{}
	detail := evHd.PackEventDetailMsg(pdata2)
	if detail != nil {
		reply.Data, _ = detail.Marshal()
	}
	return reply, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_HUODONG_DETAIL, rpc_C2S_FetchHuodongDetail)
}
