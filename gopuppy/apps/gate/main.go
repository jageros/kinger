package main

import (
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/app"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/opmon"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/db"
	_ "kinger/gopuppy/meta"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/snet"
	"kinger/gopuppy/proto/pb"
	"sync"
	"time"
)

var (
	peer     *network.Peer
	gService *gateService
)

type gateService struct {
	appID         uint32
	region        uint32
	clientProxies sync.Map
	filterTrees   map[string]*filterTree
}

func (gs *gateService) Start(appID uint16) {
	gs.appID = uint32(appID)
	gService = gs
	gs.filterTrees = make(map[string]*filterTree)
	timer.StartTicks(500 * time.Millisecond)
	common.InitUUidGenerator("clientID")
	cfg := config.GetConfig()
	opmon.Initialize(consts.AppGate, gs.appID, cfg.Opmon.DumpInterval, cfg.Opmon.FalconAgentPort)
	gateCfg := cfg.GetGateConfig(appID)
	gs.region = gateCfg.Region

	peer = network.NewPeer(&network.PeerConfig{
		ReadTimeout:   time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:  0,
		MaxPacketSize: cfg.MaxPacketSize,
	})

	api.Initialize(appID, consts.AppGate, gs.region, nil)
	registerRpc(peer)
	handlerEvent()

	for _, lnCfg := range gateCfg.Listens {
		peer.ListenWs(lnCfg.BindIP, lnCfg.Port, lnCfg.Certfile, lnCfg.Keyfile, nil, &snet.Config{
			EnableCrypt:        false,
			HandshakeTimeout:   10 * time.Second,
			RewriterBufferSize: 100 * 1024,
			ReconnWaitTimeout:  2 * time.Minute,
		})
	}
}

func (gs *gateService) Stop() {
	if peer != nil {
		peer.Close()
	}

	api.OnAppClose()
	db.Shutdown()
	evq.Stop()
	glog.Close()
}

func (gs *gateService) newClientProxy(ses *network.Session) {
	cp := newClientProxy(ses, common.GenUUid("clientID"))
	gs.clientProxies.Store(cp.clientID, cp)
	//api.SelectCenterByUUid(cp.clientID).Push(pb.MessageID_GT2C_ON_CLIENT_ACCEPT, &pb.ClientProxy{
	//	GateID:   gs.appID,
	//	ClientID: uint64(cp.clientID),
	//})
}

func (gs *gateService) getClientProxy(clientID common.UUid) *clientProxy {
	if cp, ok := gs.clientProxies.Load(clientID); ok {
		return cp.(*clientProxy)
	} else {
		return nil
	}
}

func (gs *gateService) notifyClientClose(uid, clientID common.UUid) {
	if uid > 0 {
		api.SelectCenterByUUid(uid, gs.region).Push(pb.MessageID_GT2C_ON_CLIENT_CLOSE, &pb.PlayerClient{
			GateID:   gs.appID,
			ClientID: uint64(clientID),
			Uid:      uint64(uid),
			Region:   gs.region,
		})
	}
}

func (gs *gateService) notifySnetDisconnect(uid, clientID common.UUid) {
	if uid > 0 {
		api.SelectCenterByUUid(uid, gs.region).Push(pb.MessageID_GT2C_ON_SNET_DISCONNECT, &pb.PlayerClient{
			GateID:   gs.appID,
			ClientID: uint64(clientID),
			Uid:      uint64(uid),
			Region:   gs.region,
		})
	}
}

func (gs *gateService) notifySnetReconnect(uid, clientID common.UUid) {
	if uid > 0 {
		api.SelectCenterByUUid(uid, gs.region).Push(pb.MessageID_GT2C_ON_SNET_RECONNECT, &pb.PlayerClient{
			GateID:   gs.appID,
			ClientID: uint64(clientID),
			Uid:      uint64(uid),
			Region:   gs.region,
		})
	}
}

func (gs *gateService) onClientProxyClose(ses *network.Session) {
	clientID2 := ses.GetProp("clientID")
	if clientID2 == nil {
		return
	}
	clientID := clientID2.(common.UUid)
	if cp2, ok := gs.clientProxies.Load(clientID); ok {
		gs.clientProxies.Delete(clientID)
		cp := cp2.(*clientProxy)

		for key, val := range cp.filterProps {
			ft := gs.filterTrees[key]
			if ft != nil {
				glog.Debugf("DROP CLIENT %s FILTER PROP: %s = %s", cp, key, val)
				ft.Remove(cp, val)
			}
		}

		if !cp.isKickout {
			gs.notifyClientClose(cp.uid, cp.clientID)
		}
	}
}

func (gs *gateService) onSnetEvent(ses *network.Session, eventID int) {
	clientID2 := ses.GetProp("clientID")
	if clientID2 == nil {
		return
	}
	clientID := clientID2.(common.UUid)
	if cp2, ok := gs.clientProxies.Load(clientID); ok {
		cp := cp2.(*clientProxy)
		if !cp.isKickout {
			if eventID == consts.SNET_ON_DISCONNECT {
				gs.notifySnetDisconnect(cp.uid, cp.clientID)
			} else if eventID == consts.SNET_ON_RECONNECT {
				gs.notifySnetReconnect(cp.uid, cp.clientID)
			}
		}
	}
}

func main() {
	app.NewApplication(consts.AppGate, &gateService{}).Run()
}
