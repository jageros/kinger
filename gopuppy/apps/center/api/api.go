package api

import (
	"fmt"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/proto/pb"
	"math/rand"
	"time"
)

var (
	region2CenterSessions = map[uint32][]*centerSession{}
	peer                  *network.Peer
	isAppClose            = false
)

type ICenterClientDelegate interface {
	HandleCenterClientDisconnect(centerID uint16, ses *network.Session)
}

func RegisterCenterRpcHandler(msgID protoc.IMessageID, handler network.RpcHandler) {
	peer.RegisterRpcHandler(msgID, handler)
}

func Initialize(appID uint16, appName string, region uint32, centerClientDelegate ICenterClientDelegate) {
	cfg := config.GetConfig()
	if len(cfg.Centers) == 0 {
		panic("no center")
	}

	peer = network.NewPeer(&network.PeerConfig{
		ReadTimeout:   20 * time.Second,
		WriteTimeout:  0,
		MaxPacketSize: 1024 * 1024 * 10,
		ReadBufSize:   16384,
		WriteBufSize:  16384,
	})

	for _, centerCfg := range cfg.Centers {
		ses := newCenterSession(appID, appName, centerCfg, centerClientDelegate)
		ses.assureConnected(region, false)
		sess := region2CenterSessions[centerCfg.Region]
		region2CenterSessions[centerCfg.Region] = append(sess, ses)
	}

	for _, region := range cfg.AllRegions {
		if len(region2CenterSessions[region]) <= 0 {
			panic(fmt.Sprintf("region %d no center", region))
		}
	}

	pingCenterSessions()

	evq.HandleEvent(consts.SESSION_ON_CLOSE_EVENT, func(event evq.IEvent) {
		ses := event.(*evq.CommonEvent).GetData()[0].(*network.Session)
		if ses.FromPeer() != peer {
			return
		}

		glog.Errorf("center ses on close %s", ses)

		for _, sess := range region2CenterSessions {
			for _, cs := range sess {
				if cs.rawSes == ses {
					cs.rawSes = nil
					if cs.centerClientDelegate != nil {
						cs.centerClientDelegate.HandleCenterClientDisconnect(cs.cfg.ID, ses)
					}

					if !isAppClose {
						evq.Await(func() {
							cs.assureConnected(region, true)
						})
					}
					return
				}
			}
		}
	})
}

func pingCenterSessions() {
	timer.AddTicker(15*time.Second, func() {
		for _, sess := range region2CenterSessions {
			for _, ses := range sess {
				if ses.rawSes != nil {
					ses.rawSes.Ping()
				}
			}
		}
	})
}

func OnAppClose() {
	isAppClose = true
	peer.Close()
}

func SelectCenterByUUid(id common.UUid, region uint32) *network.Session {
	if region == 0 {
		region = 1
	}
	h := hashUUid(id)
	sess := region2CenterSessions[region]
	return sess[h%len(sess)].rawSes
}

func SelectCenterByAppID(appID uint32, region uint32) *network.Session {
	if region == 0 {
		region = 1
	}
	h := hashAppID(appID)
	sess := region2CenterSessions[region]
	return sess[h%len(sess)].rawSes
}

func SelectCenterByRandom(region uint32) *network.Session {
	if region == 0 {
		region = 1
	}
	sess := region2CenterSessions[region]
	return sess[rand.Intn(len(sess))].rawSes
}

func SelectCenterByString(s string, region uint32) *network.Session {
	if region == 0 {
		region = 1
	}
	h := hashString(s)
	sess := region2CenterSessions[region]
	return sess[h%len(sess)].rawSes
}

func BroadcastCenter(msgID protoc.IMessageID, arg interface{}) {
	for _, sess := range region2CenterSessions {
		for _, ses := range sess {
			ses.rawSes.Push(msgID, arg)
		}
	}
}

func CallAllCenter(msgID protoc.IMessageID, arg interface{}) {
	var cs []chan *network.RpcResult
	for _, sess := range region2CenterSessions {
		for _, ses := range sess {
			cs = append(cs, ses.rawSes.CallAsync(msgID, arg))
		}
	}

	evq.Await(func() {
		for _, c := range cs {
			if c != nil {
				<-c
			}
		}
	})
}

func BroadcastClient(msgID protoc.IMessageID, arg interface{}, filter *pb.BroadcastClientFilter) {
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("BroadcastClient not meta %s", msgID)
	}
	payload, _ := meta.EncodeArg(arg)
	SelectCenterByRandom(1).Push(pb.MessageID_C2GT_BROADCAST_CLIENT, &pb.BroadcastClientArg{
		MsgID:   msgID.ID(),
		Payload: payload,
		Filter:  filter,
	})
}

func NotifyPlayerBeginLogin(uid, clientID common.UUid, gateID, region uint32) {
	if region == 0 {
		region = 1
	}
	SelectCenterByUUid(uid, region).Call(pb.MessageID_G2C_PLAYER_LOGIN, &pb.PlayerClient{
		ClientID: uint64(clientID),
		Uid:      uint64(uid),
		GateID:   gateID,
		Region:   region,
	})
}

func NotifyPlayerLoginDone(uid, clientID common.UUid, gateID, region uint32, beMonitorArgs ...bool) {
	if region == 0 {
		region = 1
	}
	var beMonitor bool
	if len(beMonitorArgs) > 0 {
		beMonitor = beMonitorArgs[0]
	}

	SelectCenterByUUid(uid, region).Push(pb.MessageID_PLAYER_LOGIN_DONE, &pb.PlayerLoginDone{
		Client: &pb.PlayerClient{
			ClientID: uint64(clientID),
			Uid:      uint64(uid),
			GateID:   gateID,
			Region:   region,
		},
		BeMonitor: beMonitor,
	})
}

func EndMonitorPlayer(uid, clientID common.UUid, gateID, region uint32) {
	if region == 0 {
		region = 1
	}

	SelectCenterByUUid(uid, region).Push(pb.MessageID_END_MONITOR_PLAYER, &pb.PlayerClient{
		ClientID: uint64(clientID),
		Uid:      uint64(uid),
		GateID:   gateID,
		Region:   region,
	})
}

func DelClientDispatchInfo(uid common.UUid, region uint32) {
	SelectCenterByUUid(uid, region).Push(pb.MessageID_DEL_CLIENT_DISPATCH_INFO, &pb.TargetPlayer{
		Uid: uint64(uid),
	})
}
