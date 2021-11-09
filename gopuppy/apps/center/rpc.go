package main

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/network"
	"kinger/gopuppy/proto/pb"
	"time"
)

func rpc_A2C_RegisterApp(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.AppInfo)
	if arg2.AppName == consts.AppGate {
		cService.onGateRegister(arg2.AppID, ses)
	} else if arg2.AppName == consts.AppGame {
		cService.onGameRegister(arg2.AppID, arg2.Region, ses, arg2.IsReconnect)
	} else if arg2.AppName == consts.AppCenter {
		glog.Errorf("rpc_A2C_RegisterApp error appname %s, appid = %d", arg2.AppName, arg2.AppID)
		return nil, network.InternalErr
	} else {
		cService.onAppRegister(arg2.AppID, arg2.AppName, arg2.Region, ses, arg2.IsReconnect)
	}
	glog.Infof("rpc_A2C_RegisterApp %s %d", arg2.AppName, arg2.AppID)
	return nil, nil
}

func rpc_GT2C_ClientRpcCall(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RpcCallArg)
	uid := common.UUid(arg2.Client.Uid)
	if uid <= 0 {
		return cService.chooseApp(consts.AppGame, arg2.Client.Region).Call(pb.MessageID_GT2C_CLIENT_RPC_CALL, arg2)
	} else {
		cpi := cService.getClientDispatchInfo(uid)
		if cpi == nil {
			glog.Errorf("rpc_GT2C_ClientRpcCall no DispatchInfo clientID=%d, uid=%d", arg2.Client.ClientID, uid)
			return nil, network.InternalErr
		}

		//if cpi.clientID != common.UUid(arg2.Client.ClientID) {
		//	return nil, network.InternalErr
		//}

		return cpi.callBackend(arg2)
	}
}

func rpc_GT2C_ClientRpcPush(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RpcCallArg)
	uid := common.UUid(arg2.Client.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	//if cpi.clientID != common.UUid(arg2.Client.ClientID) {
	//	return nil, network.InternalErr
	//}
	if cpi == nil {
		glog.Errorf("rpc_GT2C_ClientRpcPush no DispatchInfo, uid=%d, clientID=%d, gateID=%d, msgID=%d", uid,
			arg2.Client.ClientID, arg2.Client.GateID, arg2.MsgID)
		return nil, nil
	}
	cpi.pushBackend(arg2)
	return nil, nil
}

func rpc_G2C_PlayerLogin(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	cpi := cService.getClientDispatchInfo(common.UUid(arg2.Uid))
	if cpi == nil {
		return nil, nil
	}

	cpi.notifyKickout()
	evq.CallLater(func() {
		cpi2 := cService.getClientDispatchInfo(common.UUid(arg2.Uid))
		if cpi == cpi2 {
			cService.delClientDispatchInfo(common.UUid(arg2.Uid))
		}
	})
	return nil, nil
}

func rpc_PlayerLoginDone(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerLoginDone)
	uid := common.UUid(arg2.Client.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	for cpi != nil && !cpi.isKickout {
		cpi.notifyKickout()
		cpi2 := cService.getClientDispatchInfo(uid)
		if cpi == cpi2 {
			cService.delClientDispatchInfo(uid)
			break
		} else {
			cpi = cpi2
		}
	}

	gateSes := cService.getGateSession(arg2.Client.GateID)
	if gateSes == nil {
		glog.Errorf("rpc_PlayerLoginDone no gate %d, uid=%d, clientID=%d",
			arg2.Client.GateID, arg2.Client.Uid, arg2.Client.ClientID)
		return nil, nil
	}

	cService.newClientDispatchInfo(arg2.Client, ses.GetProp("appID").(uint32))
	gateSes.Push(pb.MessageID_PLAYER_LOGIN_DONE, arg2)
	return nil, nil
}

func rpc_EndMonitorPlayer(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	gateSes := cService.getGateSession(arg2.GateID)
	if gateSes == nil {
		return nil, nil
	}

	gateSes.Push(pb.MessageID_END_MONITOR_PLAYER, arg2)
	return nil, nil
}

func rpc_L2C_ReportRpc(ses *network.Session, arg interface{}) (interface{}, error) {
	appName, _ := ses.GetProp("appName").(string)
	appID, _ := ses.GetProp("appID").(uint32)
	glog.Infof("rpc_L2C_ReportRpc appName=%s appID=%d", appName, appID)

	arg2 := arg.(*pb.RpcHandlers)
	for _, h := range arg2.Handlers {
		if h.IsPlayer {
			ok := cService.appRegisterPlayerRpc(arg2.AppName, h.MsgID)
			if !ok {
				return nil, network.InternalErr
			}
		} else {
			ok := cService.appRegisterNoPlayerRpc(arg2.AppName, h.MsgID)
			if !ok {
				return nil, network.InternalErr
			}
		}
	}

	if appName == consts.AppGate || appName == consts.AppCenter {
		return nil, nil
	}

	var onLogicSesRestored = func() {
		glog.Infof("onLogicSesRestored appName=%s, appID=%d", appName, appID)
		lses := cService.getAppSession(appName, appID)
		if lses != nil && !lses.isAlive {
			lses.onRestored()
		}
	}

	t := timer.AfterFunc(10*time.Second, func() {
		onLogicSesRestored()
		glog.Infof("LogicSesRestored time out appName=%s, appID=%d", appName, appID)
	})

	clis := cService.getAppClients(appName, appID)
	if len(clis) > 0 {
		ses.Call(pb.MessageID_C2L_RESTORE_AGENT, &pb.RestoreAgentArg{
			Clients: clis,
		})
	}

	t.Cancel()
	onLogicSesRestored()
	return nil, nil
}

func rpc_GT2C_OnClientClose(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	clientID := common.UUid(arg2.ClientID)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil || cpi.clientID != clientID {
		return nil, nil
	}

	cpi.notifyClose()
	cpi.clearDispatchInfo()
	return nil, nil
}

func rpc_C2GT_PushClient(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RpcCallArg)
	uid := common.UUid(arg2.Client.Uid)
	clientID := common.UUid(arg2.Client.ClientID)
	cpi := cService.getClientDispatchInfo(uid)

	//if cpi == nil || cpi.gateID <= 0 || (arg2.Client.GateID > 0 && cpi.clientID != clientID) {
	if cpi == nil || cpi.gateID <= 0 {
		return nil, nil
	}

	gateSes := cService.getGateSession(cpi.gateID)
	if gateSes == nil {
		glog.Errorf("rpc_C2GT_PushClient no gate %d, uid=%d, clientID=%d", cpi.gateID, uid, clientID)
		return nil, nil
	}

	arg2.Client.GateID = cpi.gateID
	arg2.Client.ClientID = uint64(cpi.clientID)
	gateSes.Push(pb.MessageID_C2GT_PUSH_CLIENT, arg2)
	return nil, nil
}

func rpc_L2C_SetDispatch(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.SetDispatchArg)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil {
		glog.Errorf("rpc_L2C_SetDispatch no ClientDispatchInfo uid=%d", uid)
		return nil, network.InternalErr
	}

	cpi.setDispatchApp(arg2.AppName, arg2.AppID)
	return nil, nil
}

func rpc_L2L_NoPlayerRpcCall(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.NoPlayerRpcCallArg)
	var appSes *logicSession
	if arg2.AppName != "" && arg2.AppID > 0 {
		appSes = cService.getAppSession(arg2.AppName, arg2.AppID)
	} else {
		appName := cService.randomNoPlayerRpcHandlerApp(arg2.MsgID)
		if appName != "" {
			appSes = cService.chooseApp(appName, cService.region)
		}
	}

	if appSes == nil {
		glog.Errorf("rpc_L2L_NoPlayerRpcCall no appSes appName=%s, appID=%d, msgID=%d",
			arg2.AppName, arg2.AppID, arg2.MsgID)
		return nil, network.InternalErr
	}

	//glog.Infof("rpc_L2L_NoPlayerRpcCall %d", arg2.MsgID)

	return appSes.Call(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, arg2)
}

func rpc_L2L_NoPlayerRpcPush(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.NoPlayerRpcCallArg)
	var appSes *logicSession
	if arg2.AppName != "" && arg2.AppID > 0 {
		appSes = cService.getAppSession(arg2.AppName, arg2.AppID)
	} else {
		appName := cService.randomNoPlayerRpcHandlerApp(arg2.MsgID)
		if appName != "" {
			appSes = cService.chooseApp(appName, cService.region)
		}
	}

	if appSes == nil {
		glog.Errorf("rpc_L2L_NoPlayerRpcPush no appSes appName=%s, appID=%d, msgID=%d",
			arg2.AppName, arg2.AppID, arg2.MsgID)
		return nil, network.InternalErr
	}

	//glog.Infof("rpc_L2L_NoPlayerRpcPush %d", arg2.MsgID)

	appSes.Push(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, arg2)
	return nil, nil
}

func rpc_L2L_NoPlayerBroadcast(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.NoPlayerBroadcastArg)
	payload, _ := (&pb.NoPlayerRpcCallArg{
		MsgID:   arg2.MsgID,
		Payload: arg2.Payload,
	}).Marshal()

	appNames := cService.getNoPlayerRpcHandlerApps(arg2.MsgID)
	for _, appName := range appNames {
		sess := cService.getAppSessions(appName)
		for _, toSes := range sess {
			if toSes.Session != ses {
				toSes.Push(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, payload)
			}
		}
	}
	return nil, nil
}

func rpc_LoadPlayer(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetPlayer)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil {
		cpi = cService.newClientDispatchInfo(&pb.PlayerClient{
			Uid:    arg2.Uid,
			Region: cService.region,
		}, ses.GetProp("appID").(uint32))
	}

	appID := cpi.getDispatchApp(consts.AppGame)
	gameSes := cService.getAppSession(consts.AppGame, appID)
	if gameSes == nil {
		glog.Errorf("rpc_LoadPlayer no game %d", appID)
		return nil, network.InternalErr
	}

	return gameSes.Call(pb.MessageID_LOAD_PLAYER, arg)
}

func rpc_DelClientDispatchInfo(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetPlayer)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil || (cpi.gateID > 0 || cpi.clientID > 0) {
		return nil, nil
	}

	cService.delClientDispatchInfo(uid)
	return nil, nil
}

func rpc_C2GT_BroadcastClient(ses *network.Session, arg interface{}) (interface{}, error) {
	for _, gateSes := range cService.gates {
		gateSes.Push(pb.MessageID_C2GT_BROADCAST_CLIENT, arg)
	}
	return nil, nil
}

func rpc_C2GT_ClientSetFilter(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.ClientSetFilterArg)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil {
		return nil, nil
	}

	gateSes := cService.getGateSession(cpi.gateID)
	if gateSes == nil {
		glog.Errorf("rpc_C2GT_ClientSetFilter no gate %d, uid=%d, clientID=%d", cpi.gateID, uid, cpi.clientID)
		return nil, nil
	}

	gateSes.Push(pb.MessageID_C2GT_CLIENT_SET_FILTER, arg)
	return nil, nil
}

func rpc_C2GT_ClientClearFilter(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil {
		return nil, nil
	}

	gateSes := cService.getGateSession(cpi.gateID)
	if gateSes == nil {
		glog.Errorf("rpc_C2GT_ClientClearFilter no gate %d, uid=%d, clientID=%d", cpi.gateID, uid, cpi.clientID)
		return nil, nil
	}

	gateSes.Push(pb.MessageID_C2GT_CLIENT_CLEAR_FILTER, arg)
	return nil, nil
}

func rpc_GT2C_OnSnetDisconnect(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	clientID := common.UUid(arg2.ClientID)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil || cpi.clientID != clientID {
		return nil, nil
	}

	cpi.notifySnetEvent(consts.SNET_ON_DISCONNECT)
	return nil, nil
}

func rpc_GT2C_OnSnetReconnect(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	clientID := common.UUid(arg2.ClientID)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil || cpi.clientID != clientID {
		return nil, nil
	}

	cpi.notifySnetEvent(consts.SNET_ON_RECONNECT)
	return nil, nil
}

func rpc_L2C_BeginHotFix(ses *network.Session, arg interface{}) (interface{}, error) {
	appName := ses.GetProp("appName").(string)
	appID := ses.GetProp("appID").(uint32)
	lses := cService.getAppSession(appName, appID)
	if lses != nil {
		lses.beginHotFix()
		glog.Infof("lses.beginHotFix appName=%s appID=%d", appName, appID)
	}
	return nil, nil
}

func rpc_L2C_PlayerLogout(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	cpi := cService.getClientDispatchInfo(uid)
	if cpi == nil {
		return nil, nil
	}

	gateSes := cService.getGateSession(cpi.gateID)
	if gateSes == nil {
		glog.Errorf("rpc_L2C_PlayerLogout no gate %d, uid=%d, clientID=%d", cpi.gateID, uid, cpi.clientID)
		return nil, nil
	}

	gateSes.Push(pb.MessageID_L2C_PLAYER_LOGOUT, arg)
	return nil, nil
}

func registerRpc(peer *network.Peer) {
	peer.RegisterRpcHandler(pb.MessageID_A2C_REGISTER_APP, rpc_A2C_RegisterApp)
	//peer.RegisterRpcHandler(pb.MessageID_GT2C_ON_CLIENT_ACCEPT, rpc_GT2C_OnClientAccept)
	peer.RegisterRpcHandler(pb.MessageID_GT2C_CLIENT_RPC_CALL, rpc_GT2C_ClientRpcCall)
	peer.RegisterRpcHandler(pb.MessageID_GT2C_CLIENT_RPC_PUSH, rpc_GT2C_ClientRpcPush)
	peer.RegisterRpcHandler(pb.MessageID_G2C_PLAYER_LOGIN, rpc_G2C_PlayerLogin)
	peer.RegisterRpcHandler(pb.MessageID_PLAYER_LOGIN_DONE, rpc_PlayerLoginDone)
	peer.RegisterRpcHandler(pb.MessageID_END_MONITOR_PLAYER, rpc_EndMonitorPlayer)
	peer.RegisterRpcHandler(pb.MessageID_L2C_REPORT_RPC, rpc_L2C_ReportRpc)
	peer.RegisterRpcHandler(pb.MessageID_GT2C_ON_CLIENT_CLOSE, rpc_GT2C_OnClientClose)
	peer.RegisterRpcHandler(pb.MessageID_C2GT_PUSH_CLIENT, rpc_C2GT_PushClient)
	peer.RegisterRpcHandler(pb.MessageID_L2C_SET_DISPATCH, rpc_L2C_SetDispatch)
	peer.RegisterRpcHandler(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, rpc_L2L_NoPlayerRpcCall)
	peer.RegisterRpcHandler(pb.MessageID_L2L_NO_PLAYER_RPC_PUSH, rpc_L2L_NoPlayerRpcPush)
	peer.RegisterRpcHandler(pb.MessageID_L2L_NO_PLAYER_BROADCAST, rpc_L2L_NoPlayerBroadcast)
	peer.RegisterRpcHandler(pb.MessageID_LOAD_PLAYER, rpc_LoadPlayer)
	peer.RegisterRpcHandler(pb.MessageID_DEL_CLIENT_DISPATCH_INFO, rpc_DelClientDispatchInfo)
	peer.RegisterRpcHandler(pb.MessageID_C2GT_BROADCAST_CLIENT, rpc_C2GT_BroadcastClient)
	peer.RegisterRpcHandler(pb.MessageID_C2GT_CLIENT_SET_FILTER, rpc_C2GT_ClientSetFilter)
	peer.RegisterRpcHandler(pb.MessageID_C2GT_CLIENT_CLEAR_FILTER, rpc_C2GT_ClientClearFilter)
	peer.RegisterRpcHandler(pb.MessageID_GT2C_ON_SNET_DISCONNECT, rpc_GT2C_OnSnetDisconnect)
	peer.RegisterRpcHandler(pb.MessageID_GT2C_ON_SNET_RECONNECT, rpc_GT2C_OnSnetReconnect)
	peer.RegisterRpcHandler(pb.MessageID_L2C_BEGIN_HOT_FIX, rpc_L2C_BeginHotFix)
	peer.RegisterRpcHandler(pb.MessageID_L2C_PLAYER_LOGOUT, rpc_L2C_PlayerLogout)
}
