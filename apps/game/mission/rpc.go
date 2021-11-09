package mission

import (
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/eventhub"
	"kinger/proto/pb"
)

func rpc_C2S_FetchMissionInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.MissionCpt).(*missionComponent).packMsg(), nil
}

func rpc_C2S_RefreshMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	mo, err := player.GetComponent(consts.MissionCpt).(*missionComponent).refreshMission(
		int(arg.(*pb.TargetMission).ID), false)
	if err != nil {
		return nil, err
	}
	return mo.packMsg(), nil
}

func rpc_C2S_GetMissionReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	jade, gold, bowlder, nextMission, err := player.GetComponent(consts.MissionCpt).(*missionComponent).getMissionReward(
		int(arg.(*pb.TargetMission).ID))
	if err != nil {
		return nil, err
	}
	return &pb.MissionReward{
		Jade:        int32(jade),
		Gold:        int32(gold),
		Bowlder:     int32(bowlder),
		NextMission: nextMission,
	}, nil
}

func rpc_C2S_OpenMissionTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reward, treasureObj, err := player.GetComponent(consts.MissionCpt).(*missionComponent).openTreasure()
	if err != nil {
		return nil, err
	}
	return &pb.OpenMissionTreasureReply{
		TreasureReward: reward,
		NextTreasure:   treasureObj.packMsg(),
	}, nil
}

func rpc_C2S_WatchOutVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	mod.OnWatchVideo(player)
	eventhub.Publish(consts.EVWatchBattleReport, player)
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_MISSION_INFO, rpc_C2S_FetchMissionInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REFRESH_MISSION, rpc_C2S_RefreshMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_MISSION_REWARD, rpc_C2S_GetMissionReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_OPEN_MISSION_TREASURE, rpc_C2S_OpenMissionTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_OUT_VIDEO, rpc_C2S_WatchOutVideo)
}
