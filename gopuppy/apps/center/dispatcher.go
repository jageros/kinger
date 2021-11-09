package main

import (
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/gopuppy/proto/pb"
)

type clientDispatchInfo struct {
	clientID     common.UUid
	uid          common.UUid
	gateID       uint32
	region       uint32
	dispatchApps map[string]uint32
	isKickout    bool
}

func newClientDispatchInfo(clientID, uid common.UUid, gateID, gameID, region uint32) *clientDispatchInfo {
	cpi := &clientDispatchInfo{
		clientID: clientID,
		uid:      uid,
		gateID:   gateID,
		region:   region,
		dispatchApps: map[string]uint32{
			consts.AppGame: gameID,
		},
	}
	cService.setAppClient(consts.AppGame, gameID, uid)
	if gateID > 0 {
		cService.setAppClient(consts.AppGate, gateID, uid)
	}
	return cpi
}

func (cpi *clientDispatchInfo) packMsg() *pb.PlayerClient {
	return &pb.PlayerClient{
		ClientID: uint64(cpi.clientID),
		Uid:      uint64(cpi.uid),
		GateID:   cpi.gateID,
		Region:   cpi.region,
	}
}

func (cpi *clientDispatchInfo) setDispatchApp(appName string, appID uint32) {
	oldAppID, ok := cpi.dispatchApps[appName]
	if ok {
		cService.delAppClient(appName, oldAppID, cpi.uid)
	}

	if appID == 0 {
		delete(cpi.dispatchApps, appName)
	} else {
		cpi.dispatchApps[appName] = appID
		cService.setAppClient(appName, appID, cpi.uid)
	}
}

func (cpi *clientDispatchInfo) getDispatchApp(appName string) uint32 {
	return cpi.dispatchApps[appName]
}

func (cpi *clientDispatchInfo) onDel() {
	for appName, appID := range cpi.dispatchApps {
		cService.delAppClient(appName, appID, cpi.uid)
	}
	if cpi.gateID > 0 {
		cService.delAppClient(consts.AppGate, cpi.gateID, cpi.uid)
	}
}

func (cpi *clientDispatchInfo) clearDispatchApp(appName string) {
	if appID, ok := cpi.dispatchApps[appName]; ok {
		delete(cpi.dispatchApps, appName)
		cService.delAppClient(appName, appID, cpi.uid)
	}
}

func (cpi *clientDispatchInfo) clearDispatchInfo() {
	for appName, appID := range cpi.dispatchApps {
		if appName != consts.AppGame {
			delete(cpi.dispatchApps, appName)
			cService.delAppClient(appName, appID, cpi.uid)
		}
	}
}

func (cpi *clientDispatchInfo) getBackendSes(msgID int32) *logicSession {
	appName := cService.getPlayerRpcHandlerApp(msgID)
	if appName == "" {
		glog.Errorf("no app can handler %d", msgID)
		return nil
	}

	if appID, ok := cpi.dispatchApps[appName]; ok {
		return cService.getAppSession(appName, appID)
	} else {
		ses := cService.chooseApp(appName, cpi.region)
		if ses == nil {
			glog.Errorf("no app %s", appName)
			return nil
		}
		appID := ses.GetProp("appID").(uint32)
		cpi.setDispatchApp(appName, appID)
		return ses
	}
}

func (cpi *clientDispatchInfo) callBackend(arg *pb.RpcCallArg) (interface{}, error) {
	ses := cpi.getBackendSes(arg.MsgID)
	if ses == nil {
		return nil, network.InternalErr
	}

	if arg.Client.ClientID <= 0 {
		arg.Client.ClientID = uint64(cpi.clientID)
		arg.Client.GateID = cpi.gateID
		arg.Client.Region = cpi.region
	}

	return ses.Call(pb.MessageID_GT2C_CLIENT_RPC_CALL, arg)
}

func (cpi *clientDispatchInfo) pushBackend(arg *pb.RpcCallArg) {
	ses := cpi.getBackendSes(arg.MsgID)
	if ses == nil {
		return
	}
	ses.Push(pb.MessageID_GT2C_CLIENT_RPC_CALL, arg)
}

func (cpi *clientDispatchInfo) notifyKickout() {
	cpi.isKickout = true
	var cs []chan *network.RpcResult
	arg := &pb.PlayerClient{
		ClientID: uint64(cpi.clientID),
		Uid:      uint64(cpi.uid),
		GateID:   cpi.gateID,
		Region:   cpi.region,
	}

	gateSes := cService.getGateSession(cpi.gateID)
	if gateSes != nil {
		cs = append(cs, gateSes.CallAsync(pb.MessageID_C2L_KICK_OUT_PLAYER, arg))
	}

	for appName, appID := range cpi.dispatchApps {
		ses := cService.getAppSession(appName, appID)
		if ses != nil {
			cs = append(cs, ses.CallAsync(pb.MessageID_C2L_KICK_OUT_PLAYER, arg))
		}
	}

	evq.Await(func() {
		for _, c := range cs {
			<-c
		}
	})
}

func (cpi *clientDispatchInfo) notifyClose() {
	arg := &pb.PlayerClient{
		ClientID: uint64(cpi.clientID),
		Uid:      uint64(cpi.uid),
		GateID:   cpi.gateID,
		Region:   cpi.region,
	}

	for appName, appID := range cpi.dispatchApps {
		ses := cService.getAppSession(appName, appID)
		if ses != nil {
			ses.Push(pb.MessageID_GT2C_ON_CLIENT_CLOSE, arg)
		}
	}

	cpi.gateID = 0
	cpi.clientID = 0
}

func (cpi *clientDispatchInfo) notifySnetEvent(eventID int) {
	arg := &pb.PlayerClient{
		ClientID: uint64(cpi.clientID),
		Uid:      uint64(cpi.uid),
		GateID:   cpi.gateID,
		Region:   cpi.region,
	}

	appID, ok := cpi.dispatchApps[consts.AppGame]
	if ok {
		ses := cService.getAppSession(consts.AppGame, appID)
		if ses != nil {
			if eventID == consts.SNET_ON_DISCONNECT {
				ses.Push(pb.MessageID_GT2C_ON_SNET_DISCONNECT, arg)
			} else if eventID == consts.SNET_ON_RECONNECT {
				ses.Push(pb.MessageID_GT2C_ON_SNET_RECONNECT, arg)
			}
		}
	}
}
