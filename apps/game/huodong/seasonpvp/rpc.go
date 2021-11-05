package seasonpvp

import (
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
	"kinger/apps/game/module"
	"kinger/gamedata"
	"kinger/common/consts"
	htypes "kinger/apps/game/huodong/types"
)

func rpc_C2S_FetchSeasonPvpLimitTime(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	return &pb.SeasonPvpLimitTime{
		LimitTime: module.Huodong.GetSeasonPvpLimitTime(module.Player.GetPlayer(agent.GetUid())),
	}, nil
}

func rpc_C2S_FetchSeasonPvpInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.SeasonPvpInfo{
		LimitTime: module.Huodong.GetSeasonPvpLimitTime(player),
	}
	hdData := player.GetComponent(consts.HuodongCpt).(htypes.IHuodongComponent).GetOrNewHdData(pb.HuodongTypeEnum_HSeasonPvp)
	if hdData != nil {
		data := hdData.(*seasonPvpHdPlayerData)
		reply.FirstHandAmount = int32(data.getFirstHandAmount())
		reply.BackHandAmount = int32(data.getBackHandAmount())
		reply.FirstHandWinAmount = int32(data.getFirstHandWinAmount())
		reply.BackHandWinAmount = int32(data.getBackHandWinAmount())
	}
	return reply, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SEASON_PVP_LIMIT_TIME, rpc_C2S_FetchSeasonPvpLimitTime)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SEASON_PVP_INFO, rpc_C2S_FetchSeasonPvpInfo)
}
