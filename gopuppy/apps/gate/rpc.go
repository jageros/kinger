package main

import (
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/monitor"
	"kinger/gopuppy/common/opmon"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/proto/pb"
	"time"
)

func forwardClientPacket(ses *network.Session, msgID int32, msgType protoc.MessageType, pktSeq uint32, payload []byte) {
	op := opmon.StartOperation("forwardClient")
	clientID2 := ses.GetProp("clientID")
	if clientID2 == nil {
		if msgType == protoc.MsgReq {
			ses.Write(protoc.GetErrReplyPacket(pktSeq, network.InternalErr.Errcode()))
		}
		op.Finish(100 * time.Millisecond)
		return
	}
	clientID := clientID2.(common.UUid)
	cp := gService.getClientProxy(clientID)
	var centerSes *network.Session
	if cp == nil || cp.isKickout {
		if msgType == protoc.MsgReq {
			ses.Write(protoc.GetErrReplyPacket(pktSeq, network.InternalErr.Errcode()))
		}
		op.Finish(100 * time.Millisecond)
		return
	}

	// monitor
	var errcode int32
	var replyPayload []byte
	var isLogin, isRpcReply bool

	if cp.uid > 0 {
		centerSes = api.SelectCenterByUUid(cp.uid, gService.region)
	} else {
		isLogin = msgID == 100 // rpc login shit
		centerSes = api.SelectCenterByAppID(gService.appID, gService.region)
	}

	arg := &pb.RpcCallArg{
		Client: &pb.PlayerClient{
			ClientID: uint64(clientID),
			GateID:   gService.appID,
			Uid:      uint64(cp.uid),
			Region:   gService.region,
			IP:       ses.GetIP(),
		},
		MsgID:   msgID,
		Payload: payload,
	}

	if msgType == protoc.MsgReq {
		isRpcReply = true
		c := centerSes.CallAsync(pb.MessageID_GT2C_CLIENT_RPC_CALL, arg)
		result := <-c
		if result.Err != nil {
			errcode = result.Err.Errcode()
			evq.CallLater(func() {
				ses.Write(protoc.GetErrReplyPacket(pktSeq, errcode))
			})
		} else {
			rp := protoc.GetReplyPacket(msgID, pktSeq, nil)
			replyPayload = result.Reply.(*pb.RpcCallReply).Payload
			if replyPayload == nil {
				replyPayload = []byte{}
			}
			rp.SetPayload(replyPayload)
			evq.CallLater(func() {
				ses.Write(rp)
			})
		}
	} else if msgType == protoc.MsgPush {
		centerSes.Push(pb.MessageID_GT2C_CLIENT_RPC_PUSH, arg)
	} else {
		glog.Errorf("forwardClientPacket cant handler msgType %d, clientID=%d", msgType, clientID)
		op.Finish(100 * time.Millisecond)
		return
	}

	if isRpcReply {
		var monitorTask = func() {
			if isLogin {
				if errcode == 0 {
					monitor.Login(cp.uid, pktSeq, replyPayload)
				}
			} else {
				monitor.RpcReply(cp.uid, msgID, pktSeq, errcode, replyPayload)
			}
		}

		if cp.uid == 0 {
			if isLogin {
				cp.addMonitorTask(monitorTask)
			}
		} else if cp.beMonitor {
			monitorTask()
		}
	}
	op.Finish(100 * time.Millisecond)
}

func rpc_C2GT_PushClient(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RpcCallArg)
	cp := gService.getClientProxy(common.UUid(arg2.Client.ClientID))
	if cp == nil {
		glog.Errorf("rpc_C2GT_PushClient, no clientProxy %d", arg2.Client.ClientID)
		return nil, nil
	}

	cp.ses.Push(pb.MessageID(arg2.MsgID), arg2.Payload)
	if cp.uid > 0 {
		if cp.beMonitor {
			monitor.RpcPush(cp.uid, arg2.MsgID, arg2.Payload)
		}
	} else {
		cp.addMonitorTask(func() {
			monitor.RpcPush(cp.uid, arg2.MsgID, arg2.Payload)
		})
	}
	return nil, nil
}

func rpc_PlayerLoginDone(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerLoginDone)
	cp := gService.getClientProxy(common.UUid(arg2.Client.ClientID))
	if cp == nil {
		gService.notifyClientClose(common.UUid(arg2.Client.Uid), common.UUid(arg2.Client.ClientID))
		return nil, nil
	}

	cp.uid = common.UUid(arg2.Client.Uid)
	cp.beMonitor = arg2.BeMonitor
	cp.executeMonitorTask()
	return nil, nil
}

func rpc_EndMonitorPlayer(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	cp := gService.getClientProxy(common.UUid(arg2.ClientID))
	if cp != nil {
		cp.beMonitor = false
	}
	return nil, nil
}

func rpc_C2L_KickOutPlayer(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	cp := gService.getClientProxy(common.UUid(arg2.ClientID))
	if cp != nil {
		cp.isKickout = true
	}
	return nil, nil
}

func rpc_BroadcastClient(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.BroadcastClientArg)
	if arg2.Filter == nil {
		gService.clientProxies.Range(func(key, value interface{}) bool {
			cp := value.(*clientProxy)
			if !cp.isKickout {
				cp.ses.Push(pb.MessageID(arg2.MsgID), arg2.Payload)
			}
			return true
		})
		return nil, nil
	}

	ft := gService.filterTrees[arg2.Filter.Key]
	if ft != nil {
		ft.Visit(arg2.Filter.OP, arg2.Filter.Val, func(cp *clientProxy) {
			if !cp.isKickout {
				cp.ses.Push(pb.MessageID(arg2.MsgID), arg2.Payload)
				if cp.uid > 0 {
					if cp.beMonitor {
						monitor.RpcPush(cp.uid, arg2.MsgID, arg2.Payload)
					}
				} else {
					cp.addMonitorTask(func() {
						monitor.RpcPush(cp.uid, arg2.MsgID, arg2.Payload)
					})
				}
			}
		})
	}
	return nil, nil
}

func rpc_C2GT_ClientSetFilter(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.ClientSetFilterArg)
	cp := gService.getClientProxy(common.UUid(arg2.ClientID))
	if cp == nil || cp.isKickout {
		return nil, nil
	}

	ft, ok := gService.filterTrees[arg2.Filter.Key]
	if !ok {
		ft = newFilterTree()
		gService.filterTrees[arg2.Filter.Key] = ft
	}

	oldVal, ok := cp.filterProps[arg2.Filter.Key]
	if ok {
		ft.Remove(cp, oldVal)
	}
	cp.filterProps[arg2.Filter.Key] = arg2.Filter.Val
	ft.Insert(cp, arg2.Filter.Val)
	return nil, nil
}

func rpc_C2GT_ClientClearFilter(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	cp := gService.getClientProxy(common.UUid(arg2.ClientID))
	if cp == nil || cp.isKickout {
		return nil, nil
	}

	for key, val := range cp.filterProps {
		ft, ok := gService.filterTrees[key]
		if !ok {
			continue
		}
		ft.Remove(cp, val)
	}
	return nil, nil
}

func rpc_L2C_PlayerLogout(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	cp := gService.getClientProxy(common.UUid(arg2.ClientID))
	if cp == nil || cp.isKickout || cp.ses == nil {
		return nil, nil
	}

	cp.ses.Close()
	return nil, nil
}

func registerRpc(peer *network.Peer) {
	peer.SetRawPacketHandler(forwardClientPacket)
	api.RegisterCenterRpcHandler(pb.MessageID_C2GT_PUSH_CLIENT, rpc_C2GT_PushClient)
	api.RegisterCenterRpcHandler(pb.MessageID_PLAYER_LOGIN_DONE, rpc_PlayerLoginDone)
	api.RegisterCenterRpcHandler(pb.MessageID_END_MONITOR_PLAYER, rpc_EndMonitorPlayer)
	api.RegisterCenterRpcHandler(pb.MessageID_C2L_KICK_OUT_PLAYER, rpc_C2L_KickOutPlayer)
	api.RegisterCenterRpcHandler(pb.MessageID_C2GT_BROADCAST_CLIENT, rpc_BroadcastClient)
	api.RegisterCenterRpcHandler(pb.MessageID_C2GT_CLIENT_SET_FILTER, rpc_C2GT_ClientSetFilter)
	api.RegisterCenterRpcHandler(pb.MessageID_C2GT_CLIENT_CLEAR_FILTER, rpc_C2GT_ClientClearFilter)
	api.RegisterCenterRpcHandler(pb.MessageID_L2C_PLAYER_LOGOUT, rpc_L2C_PlayerLogout)
}
