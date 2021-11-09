package main

import (
	"math/rand"

	"kinger/gopuppy/common"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/db"
	_ "kinger/gopuppy/meta"
	"kinger/gopuppy/network"
	//"kinger/gopuppy/network/snet"
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/proto/pb"
	"time"
)

var (
	cService *centerService
	peer     *network.Peer
)

type centerService struct {
	appID                  uint32
	region                 uint32
	gates                  map[uint32]*network.Session
	games                  []*network.Session
	logics                 map[string]*logicSessionMgr
	clientDispatchInfos    map[common.UUid]*clientDispatchInfo
	appClients             map[string]map[uint32]common.UInt64Set
	playerRpcHandlerApps   map[int32]string
	noPlayerRpcHandlerApps map[int32][]string
}

func (cs *centerService) Start(appID uint16) {
	cs.appID = uint32(appID)
	cs.gates = make(map[uint32]*network.Session)
	cs.logics = make(map[string]*logicSessionMgr)
	cs.clientDispatchInfos = make(map[common.UUid]*clientDispatchInfo)
	cs.playerRpcHandlerApps = make(map[int32]string)
	cs.noPlayerRpcHandlerApps = make(map[int32][]string)
	cs.appClients = make(map[string]map[uint32]common.UInt64Set)
	cService = cs
	timer.StartTicks(500 * time.Millisecond)

	cfg := config.GetConfig()
	centerCfg := cfg.GetCenterConfig(appID)
	cs.region = centerCfg.Region
	peer = network.NewPeer(&network.PeerConfig{
		ReadTimeout:   20 * time.Second,
		WriteTimeout:  0,
		MaxPacketSize: 1024 * 1024 * 10,
		ReadBufSize:   16384,
		WriteBufSize:  16384,
	})

	registerRpc(peer)
	handlerEvent()
	mq.InitServer(peer)

	peer.ListenTcp(centerCfg.Listen.BindIP, centerCfg.Listen.Port, nil, nil)
}

func (cs *centerService) Stop() {
	if peer != nil {
		peer.Close()
	}

	db.Shutdown()
	evq.Stop()
	glog.Close()
}

func (cs *centerService) onGateRegister(appID uint32, ses *network.Session) {
	if oldSes, ok := cs.gates[appID]; ok {
		glog.Warnf("repeat gateID %d", appID)
		oldSes.SetProp("appName", nil)
		oldSes.SetProp("appID", nil)
		oldSes.Close()
	}

	ses.SetProp("appName", consts.AppGate)
	ses.SetProp("appID", appID)
	cs.gates[appID] = ses
}

func (cs *centerService) setAppClient(appName string, appID uint32, uid common.UUid) {
	if id2clients, ok := cs.appClients[appName]; ok {
		if clients, ok := id2clients[appID]; ok {
			clients.Add(uint64(uid))
		}
	}
}

func (cs *centerService) delAppClient(appName string, appID uint32, uid common.UUid) {
	if id2clients, ok := cs.appClients[appName]; ok {
		if clients, ok := id2clients[appID]; ok {
			clients.Remove(uint64(uid))
		}
	}
}

func (cs *centerService) getAppClients(appName string, appID uint32) []*pb.PlayerClient {
	if id2clients, ok := cs.appClients[appName]; ok {
		if clients, ok := id2clients[appID]; ok {
			var clis []*pb.PlayerClient
			clients.ForEach(func(uid uint64) bool {
				cpi := cs.getClientDispatchInfo(common.UUid(uid))
				if cpi != nil && !cpi.isKickout && cpi.clientID > 0 && cpi.gateID > 0 && cpi.getDispatchApp(appName) == appID {
					clis = append(clis, cpi.packMsg())
				}
				return true
			})
			return clis
		} else {
			id2clients[appID] = common.UInt64Set{}
		}
	} else {
		cs.appClients[appName] = map[uint32]common.UInt64Set{
			appID: common.UInt64Set{},
		}
	}
	return []*pb.PlayerClient{}
}

func (cs *centerService) onGameRegister(appID, region uint32, ses *network.Session, isReconnect bool) {
	for i, gses := range cs.games {
		gameID := gses.GetProp("appID")
		if gameID == nil || gameID.(uint32) == appID {
			glog.Warnf("repeat gameID %d", appID)
			gses.SetProp("appName", nil)
			gses.SetProp("appID", nil)
			gses.Close()
			cs.games = append(cs.games[:i], cs.games[i+1:]...)
			break
		}
	}

	ok := cs.onAppRegister(appID, consts.AppGame, region, ses, isReconnect)
	if ok {
		cs.games = append(cs.games, ses)
	}
}

func (cs *centerService) onAppRegister(appID uint32, appName string, region uint32, ses *network.Session,
	isReconnect bool) bool {

	lsm, ok := cs.logics[appName]
	if !ok {
		lsm = newLogicSessionMgr(appName)
		cs.logics[appName] = lsm
	}

	lsm.onAppRegister(appID, region, ses, isReconnect)
	return true
}

func (cs *centerService) onGateDisconnect(appID uint32) {
	if _, ok := cs.gates[appID]; !ok {
		return
	}

	clis := cs.getAppClients(consts.AppGate, appID)
	for _, c := range clis {
		cpi := cs.getClientDispatchInfo(common.UUid(c.Uid))
		if cpi != nil {
			cpi.notifyClose()
			//cs.delClientDispatchInfo(common.UUid(c.Uid))
			cpi.clearDispatchInfo()
		}
	}
	delete(cs.gates, appID)
}

func (cs *centerService) onGameDisconnect(appID uint32) {
	for i := 0; i < len(cs.games); {
		ses := cs.games[i]
		id := ses.GetProp("appID").(uint32)
		if id == appID {
			cs.games = append(cs.games[:i], cs.games[i+1:]...)
			break
		} else {
			i++
		}
	}

	cs.onLogicDisconnect(appID, consts.AppGame)
}

func (cs *centerService) onLogicDisconnect(appID uint32, appName string) {
	if lsm, ok := cs.logics[appName]; ok {
		lsm.onLogicDisconnect(appID)
	}
	glog.Infof("onLogicDisconnect %s %d", appName, appID)
}

func (cs *centerService) getClientDispatchInfo(uid common.UUid) *clientDispatchInfo {
	return cs.clientDispatchInfos[uid]
}

func (cs *centerService) delClientDispatchInfo(uid common.UUid) {
	if cpi, ok := cs.clientDispatchInfos[uid]; ok {
		delete(cs.clientDispatchInfos, uid)
		cpi.onDel()
	}
}

func (cs *centerService) getGateSession(gateID uint32) *network.Session {
	return cs.gates[gateID]
}

func (cs *centerService) getAppSession(appName string, appID uint32) *logicSession {
	if lsm, ok := cs.logics[appName]; ok {
		return lsm.getAppSession(appID)
	} else {
		return nil
	}
}

func (cs *centerService) chooseApp(appName string, region uint32) *logicSession {
	if lsm, ok := cs.logics[appName]; ok {
		return lsm.chooseApp(region)
	} else {
		return nil
	}
}

func (cs *centerService) getAppSessions(appName string) []*logicSession {
	if lsm, ok := cs.logics[appName]; ok {
		return lsm.getAppSessions()
	} else {
		return []*logicSession{}
	}
}

func (cs *centerService) getPlayerRpcHandlerApp(msgID int32) string {
	if appName, ok := cs.playerRpcHandlerApps[msgID]; ok {
		return appName
	} else {
		return ""
	}
}

func (cs *centerService) randomNoPlayerRpcHandlerApp(msgID int32) string {
	if appNames, ok := cs.noPlayerRpcHandlerApps[msgID]; ok && len(appNames) > 0 {
		return appNames[rand.Intn(len(appNames))]
	} else {
		return ""
	}
}

func (cs *centerService) getNoPlayerRpcHandlerApps(msgID int32) []string {
	return cs.noPlayerRpcHandlerApps[msgID]
}

func (cs *centerService) appRegisterPlayerRpc(appName string, msgID int32) bool {
	if appName2, ok := cs.playerRpcHandlerApps[msgID]; ok {
		if appName2 != appName {
			glog.Errorf("appRegisterPlayerRpc %s and %s handler same rpc %d", appName2, appName, msgID)
			return false
		}
	} else {
		cs.playerRpcHandlerApps[msgID] = appName
	}
	return true
}

func (cs *centerService) appRegisterNoPlayerRpc(appName string, msgID int32) bool {
	if appNames, ok := cs.noPlayerRpcHandlerApps[msgID]; ok {
		for _, name := range appNames {
			if name == appName {
				//glog.Errorf("centerService.appRegisterNoPlayerRpc %s is already handler rpc %d", appName, msgID)
				return true
			}
		}

		appNames = append(appNames, appName)
		cs.noPlayerRpcHandlerApps[msgID] = appNames
		glog.Warnf("appRegisterNoPlayerRpc %v handler same rpc %d", appNames, msgID)
	} else {
		cs.noPlayerRpcHandlerApps[msgID] = []string{appName}
	}
	return true
}

func (cs *centerService) newClientDispatchInfo(client *pb.PlayerClient, gameID uint32) *clientDispatchInfo {
	cpi := newClientDispatchInfo(common.UUid(client.ClientID), common.UUid(client.Uid), client.GateID, gameID, client.Region)
	cs.clientDispatchInfos[common.UUid(client.Uid)] = cpi
	return cpi
}

func main() {
	app.NewApplication(consts.AppCenter, &centerService{}).Run()
}
