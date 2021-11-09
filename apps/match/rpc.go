package main

import (
	"kinger/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/network"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
)

func onPlayerLogut(args ...interface{}) {
	pa := args[0].(*logic.PlayerAgent)
	gMatchMgr.stopMatch(pa.GetUid())
	pa.SetDispatchApp(consts.AppMatch, 0)
}

func onRestoreAgent(args ...interface{}) {
	clients := args[0].([]*gpb.PlayerClient)
	for _, cli := range clients {
		if cli.GateID > 0 && cli.ClientID > 0 {
			agent := logic.NewPlayerAgent(cli)
			agent.PushClient(pb.MessageID_S2C_MATCH_TIMEOUT, nil)
			agent.SetDispatchApp(consts.AppMatch, 0)
		}
	}
}

func rpc_G2M_BeginMatch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.BeginMatchArg)
	//agent.SetDispatchApp(consts.AppMatch, mService.AppID)
	return nil, gMatchMgr.beginMatch(agent, _arg)
	//glog.Infof("rpc_G2M_BeginMatch, uid=%d, camp=%d, pvpScore=%d", uid, _arg.Camp, _arg.PvpScore)
}

func rpc_G2M_BeginGuideMatch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	//agent.SetDispatchApp(consts.AppMatch, mService.AppID)
	_arg := arg.(*pb.BeginBattleArg)
	evq.CallLater(func() {
		r := newGuideMatchRoom(gMatchMgr.genRoomID(), agent, _arg)
		gMatchMgr.addRoom(r)
		r.syncMatchInfo()
	})
	//glog.Infof("rpc_G2M_BeginGuideMatch, uid=%d, camp=%d", agent.GetUid(), _arg.Fighter1.Camp)

	return nil, nil
}

func rpc_C2S_StopMatch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	agent.SetDispatchApp(consts.AppMatch, 0)
	uid := agent.GetUid()
	//glog.Infof("rpc_C2S_StopMatch, uid=%d", uid)

	return nil, gMatchMgr.stopMatch(uid)
}

func rpc_C2S_MatchReadyDone(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.MatchDoneArg)
	gMatchMgr.onMatchReadyDone(_arg.RoomId)
	return nil, nil
}

func rpc_G2M_BeginNewbiePvpMatch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	gMatchMgr.stopMatch(agent.GetUid())
	gMatchMgr.doMatchNewbiePvp(agent, arg.(*pb.BeginNewbiePvpMatchArg))
	//glog.Infof("rpc_G2M_BeginNewbiePvpMatch, uid=%d, camp=%d, enemyCamp=%d", agent.GetUid(), _arg.Camp, _arg.EnemyCamp)

	return nil, nil
}

func rpc_B2L_OnRobotBattleEnd(_ *network.Session, arg interface{}) (interface{}, error) {
	robotMgr.onRobotBattleEnd(arg.(*pb.OnRobotBattleEndArg))
	return nil, nil
}

func registerRpc() {
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onPlayerLogut)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, onPlayerLogut)
	eventhub.Subscribe(logic.RESTORE_AGENT_EV, onRestoreAgent)

	logic.RegisterAgentRpcHandler(pb.MessageID_G2M_BEGIN_MATCH, rpc_G2M_BeginMatch)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2M_BEGIN_GUIDE_MATCH, rpc_G2M_BeginGuideMatch)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2M_BEGIN_NEWBIE_PVP_MATCH, rpc_G2M_BeginNewbiePvpMatch)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_STOP_MATCH, rpc_C2S_StopMatch)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MATCH_READY_DONE, rpc_C2S_MatchReadyDone)

	logic.RegisterRpcHandler(pb.MessageID_B2L_ON_ROBOT_BATTLE_END, rpc_B2L_OnRobotBattleEnd)
}
