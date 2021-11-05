package logic

import (
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/proto/pb"
	"github.com/pkg/errors"
)

var (
	agentRpcHandlers   = make(map[int32]AgentRpcHandler)
	noAgentRpcHandlers = make(map[int32]network.RpcHandler)
	loadPlayerHandler  func(uid common.UUid) ([]byte, error)
)

func CallBackend(appName string, appID uint32, msgID protoc.IMessageID, arg interface{}, regionArg ...uint32) (
	interface{}, error) {

	resultChan := CallBackendAsync(appName, appID, msgID, arg, regionArg...)
	var reply interface{}
	var err error
	evq.Await(func() {
		result := <-resultChan
		reply = result.Reply
		err = result.Err
	})
	return reply, err
}

func CallBackendAsync(appName string, appID uint32, msgID protoc.IMessageID, arg interface{}, regionArg ...uint32) (
	c chan *network.RpcResult) {
	meta := protoc.GetMeta(msgID.ID())
	c = make(chan *network.RpcResult, 1)
	if meta == nil {
		glog.Errorf("CallBackend no meta msgID=%s", msgID)
		c <- &network.RpcResult{
			Reply: nil,
			Err:   network.InternalErr,
		}
		return c
	}

	payload, _ := meta.EncodeArg(arg)
	region := lService.Region
	if len(regionArg) > 0 && regionArg[0] > 0 {
		region = regionArg[0]
	}

	resultChan := api.SelectCenterByAppID(lService.AppID, region).CallAsync(pb.MessageID_L2L_NO_PLAYER_RPC_CALL,
		&pb.NoPlayerRpcCallArg{
			AppName: appName,
			AppID:   appID,
			MsgID:   msgID.ID(),
			Payload: payload,
	})

	go func() {
		result := <-resultChan
		result2 := &network.RpcResult{}
		if result.Err != nil {
			result2.Err = result.Err
		} else {
			result2.Reply, _ = meta.DecodeReply(result.Reply.(*pb.RpcCallReply).Payload)
		}
		c <- result2
	}()

	return c
}

func PushBackend(appName string, appID uint32, msgID protoc.IMessageID, arg interface{}, regionArg ...uint32) {
	//glog.Infof("PushBackend %s", msgID)
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("PushBackend no meta msgID=%s", msgID)
		return
	}

	payload, _ := meta.EncodeArg(arg)
	region := lService.Region
	if len(regionArg) > 0 && regionArg[0] > 0 {
		region = regionArg[0]
	}

	api.SelectCenterByAppID(lService.AppID, region).Push(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, &pb.NoPlayerRpcCallArg{
		AppName: appName,
		AppID:   appID,
		MsgID:   msgID.ID(),
		Payload: payload,
	})
}

func BroadcastBackend(msgID protoc.IMessageID, arg interface{}) {
	//glog.Infof("PushBackend %s", msgID)
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("BroadcastBackend no meta msgID=%s", msgID)
		return
	}
	payload, _ := meta.EncodeArg(arg)
	api.SelectCenterByAppID(lService.AppID, lService.Region).Push(pb.MessageID_L2L_NO_PLAYER_BROADCAST, &pb.NoPlayerBroadcastArg{
		MsgID:   msgID.ID(),
		Payload: payload,
	})
}

func LoadPlayer(uid common.UUid) ([]byte, error) {
	region := GetAgentRegion(uid)
	if region <= 0 {
		return nil, errors.Errorf("LoadPlayer %d no region", uid)
	}

	reply, err := api.SelectCenterByUUid(uid, region).Call(pb.MessageID_LOAD_PLAYER, &pb.TargetPlayer{Uid: uint64(uid)})
	if err != nil {
		return nil, err
	} else {
		return reply.(*pb.RpcCallReply).Payload, nil
	}
}

func LoadPlayerAsync(uid common.UUid) chan []byte {
	c := make(chan []byte, 1)
	region := GetAgentRegion(uid)
	if region <= 0 {
		c <- nil
		return c
	}

	resultChan := api.SelectCenterByUUid(uid, region).CallAsync(pb.MessageID_LOAD_PLAYER, &pb.TargetPlayer{Uid: uint64(uid)})
	go func() {
		result := <-resultChan
		if result.Err != nil {
			c <- nil
		} else {
			c <- result.Reply.(*pb.RpcCallReply).Payload
		}
	}()
	return c
}

type AgentRpcHandler func(agent *PlayerAgent, arg interface{}) (interface{}, error)

func RegisterAgentRpcHandler(msgID protoc.IMessageID, handler AgentRpcHandler) {
	agentRpcHandlers[msgID.ID()] = handler
}

func RegisterRpcHandler(msgID protoc.IMessageID, handler network.RpcHandler) {
	noAgentRpcHandlers[msgID.ID()] = handler
	//api.RegisterCenterRpcHandler(msgID, handler)
}

func RegisterLoadPlayerHandler(handler func(uid common.UUid) ([]byte, error)) {
	loadPlayerHandler = handler
}

func onClientRpcCall(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RpcCallArg)
	clientID := common.UUid(arg2.Client.ClientID)
	uid := common.UUid(arg2.Client.Uid)
	agent := GetPlayerAgent(uid)
	if agent == nil {
		agent = NewPlayerAgent(arg2.Client)
		if agent.uid > 0 {
			uid2Agent[uid] = agent
		}
	} else if agent.clientID != clientID {
		agent.clientID = clientID
	}

	if arg2.Client.IP != "" {
		agent.ip = arg2.Client.IP
	}

	handler, ok := agentRpcHandlers[arg2.MsgID]
	if !ok {
		glog.Errorf("onClientRpcCall no handler %d", arg2.MsgID)
		return nil, network.InternalErr
	}

	meta := protoc.GetMeta(arg2.MsgID)
	if meta == nil {
		glog.Errorf("onClientRpcCall no meta %d", arg2.MsgID)
		return nil, network.InternalErr
	}

	arg3, err := meta.DecodeArg(arg2.Payload)
	if err != nil {
		glog.Errorf("onClientRpcCall DecodeArg err %s", err)
		return nil, err
	}

	glog.Debugf("onClientRpcCall, msgID=%d", arg2.MsgID)
	reply, err := handler(agent, arg3)
	if err != nil {
		return nil, err
	} else {
		payload2, _ := meta.EncodeReply(reply)
		return &pb.RpcCallReply{
			Payload: payload2,
		}, nil
	}
}

func onClientClose(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	pa := GetPlayerAgent(uid)
	//if pa != nil && pa.clientID == common.UUid(arg2.ClientID) {
	glog.Infof("onClientClose uid=%d pa=%s", uid, pa)
	if pa != nil {
		eventhub.Publish(CLIENT_CLOSE_EV, pa)
		DelPlayerAgent(uid, pa.clientID)
	}
	return nil, nil
}

func onPlayerKickout(_ *network.Session, arg interface{}) (reply interface{}, err error) {
	arg2 := arg.(*pb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	pa := GetPlayerAgent(uid)
	//if pa != nil && pa.clientID == common.UUid(arg2.ClientID) {
	glog.Infof("onPlayerKickout uid=%d pa=%s", uid, pa)
	if pa != nil {
		eventhub.Publish(PLAYER_KICK_OUT_EV, pa)
		DelPlayerAgent(uid, pa.clientID)
	}
	return nil, nil
}

func rpc_L2L_NoPlayerRpcCall(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.NoPlayerRpcCallArg)
	handler, ok := noAgentRpcHandlers[arg2.MsgID]
	if !ok {
		glog.Errorf("rpc_L2L_NoPlayerRpcCall no handler %d", arg2.MsgID)
		return nil, network.InternalErr
	}

	meta := protoc.GetMeta(arg2.MsgID)
	if meta == nil {
		glog.Errorf("rpc_L2L_NoPlayerRpcCall no meta %d", arg2.MsgID)
		return nil, network.InternalErr
	}

	arg3, err := meta.DecodeArg(arg2.Payload)
	if err != nil {
		glog.Errorf("rpc_L2L_NoPlayerRpcCall DecodeArg err %s", err)
		return nil, err
	}

	//glog.Infof("rpc_L2L_NoPlayerRpcCall %d", arg2.MsgID)

	reply, err := handler(ses, arg3)
	if err != nil {
		return nil, err
	} else {
		payload2, _ := meta.EncodeReply(reply)
		return &pb.RpcCallReply{
			Payload: payload2,
		}, nil
	}
}

func rpc_C2L_RestoreAgent(ses *network.Session, arg interface{}) (reply interface{}, err error) {
	eventhub.Publish(RESTORE_AGENT_EV, arg.(*pb.RestoreAgentArg).Clients)
	return nil, nil
}

func rpc_LoadPlayer(ses *network.Session, arg interface{}) (reply interface{}, err error) {
	if loadPlayerHandler == nil {
		return nil, network.InternalErr
	}

	uid := common.UUid(arg.(*pb.TargetPlayer).Uid)
	payload, err := loadPlayerHandler(uid)
	if err != nil {
		glog.Errorf("rpc_LoadPlayer %s", err)
		return nil, err
	}

	agent := GetPlayerAgent(uid)
	if agent == nil {
		agent = NewPlayerAgent(&pb.PlayerClient{
			Uid: uint64(uid),
			Region: GetAgentRegion(uid),
		})
		agent.SetUid(uid)
	}

	return &pb.RpcCallReply{
		Payload: payload,
	}, nil
}

func registerRpc() {
	api.RegisterCenterRpcHandler(pb.MessageID_GT2C_CLIENT_RPC_CALL, onClientRpcCall)
	api.RegisterCenterRpcHandler(pb.MessageID_GT2C_ON_CLIENT_CLOSE, onClientClose)
	api.RegisterCenterRpcHandler(pb.MessageID_C2L_KICK_OUT_PLAYER, onPlayerKickout)
	api.RegisterCenterRpcHandler(pb.MessageID_L2L_NO_PLAYER_RPC_CALL, rpc_L2L_NoPlayerRpcCall)
	api.RegisterCenterRpcHandler(pb.MessageID_C2L_RESTORE_AGENT, rpc_C2L_RestoreAgent)
	api.RegisterCenterRpcHandler(pb.MessageID_LOAD_PLAYER, rpc_LoadPlayer)
}
