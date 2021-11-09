package logic

import (
	"fmt"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/gopuppy/network/protoc"
	"kinger/gopuppy/proto/pb"
)

var (
	uid2Agent = make(map[common.UUid]*PlayerAgent)
)

func GetPlayerAgent(uid common.UUid) *PlayerAgent {
	if pa, ok := uid2Agent[uid]; ok {
		return pa
	} else {
		return nil
	}
}

func DelPlayerAgent(uid, clientID common.UUid) {
	//if pa, ok := uid2Agent[uid]; ok && pa.clientID == clientID {
	delete(uid2Agent, uid)
	//}
}

type PlayerAgent struct {
	clientID common.UUid
	uid      common.UUid
	gateID   uint32
	region   uint32 // 机房
	ip       string
}

func NewPlayerAgent(client *pb.PlayerClient) *PlayerAgent {
	if client.Uid > 0 && client.Region > 0 {
		cacheAgentRegion(common.UUid(client.Uid), client.Region)
	}

	return &PlayerAgent{
		clientID: common.UUid(client.ClientID),
		uid:      common.UUid(client.Uid),
		gateID:   client.GateID,
		region:   client.Region,
		ip:       client.IP,
	}
}

func NewRobotAgent() *PlayerAgent {
	return &PlayerAgent{
		uid: 1,
	}
}

func (pa *PlayerAgent) String() string {
	return fmt.Sprintf("clientID=%d uid=%d gateID=%d region=%d", pa.clientID, pa.uid, pa.gateID, pa.region)
}

func (pa *PlayerAgent) SetUid(uid common.UUid) {
	pa.uid = uid
	uid2Agent[uid] = pa

	if uid > 0 && pa.region > 0 {
		cacheAgentRegion(uid, pa.region)
	}
}

func (pa *PlayerAgent) GetIP() string {
	return pa.ip
}

func (pa *PlayerAgent) GetUid() common.UUid {
	return pa.uid
}

func (pa *PlayerAgent) GetClientID() common.UUid {
	return pa.clientID
}

func (pa *PlayerAgent) GetGateID() uint32 {
	return pa.gateID
}

func (pa *PlayerAgent) GetRegion() uint32 {
	return pa.region
}

func (pa *PlayerAgent) IsRobot() bool {
	return pa.uid <= 1
}

func (pa *PlayerAgent) CallBackend(msgID protoc.IMessageID, arg interface{}) (interface{}, error) {
	c := pa.CallBackendAsync(msgID, arg)
	var result *network.RpcResult
	evq.Await(func() {
		result = <-c
	})
	return result.Reply, result.Err
}

func (pa *PlayerAgent) CallBackendAsync(msgID protoc.IMessageID, arg interface{}) chan *network.RpcResult {
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("CallBackend no meta %s", msgID)
		c := make(chan *network.RpcResult, 1)
		c <- &network.RpcResult{
			Reply: nil,
			Err:   network.InternalErr,
		}
		return c
	}

	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("CallBackend no center")
		c := make(chan *network.RpcResult, 1)
		c <- &network.RpcResult{
			Reply: nil,
			Err:   network.InternalErr,
		}
		return c
	}

	payload, err := meta.EncodeArg(arg)
	if err != nil {
		glog.Errorf("CallBackend EncodeArg %d %s", msgID, err)
		c := make(chan *network.RpcResult, 1)
		c <- &network.RpcResult{
			Reply: nil,
			Err:   network.InternalErr,
		}
		return c
	}

	region := pa.region
	if region <= 0 {
		region = GetAgentRegion(pa.uid)
	}

	resultChan := centerSes.CallAsync(pb.MessageID_GT2C_CLIENT_RPC_CALL, &pb.RpcCallArg{
		Client: &pb.PlayerClient{
			ClientID: uint64(pa.clientID),
			Uid:      uint64(pa.uid),
			GateID:   pa.gateID,
			Region:   region,
		},
		MsgID:   msgID.ID(),
		Payload: payload,
	})

	c := make(chan *network.RpcResult, 1)
	go func() {
		result := <-resultChan
		if result.Err != nil {
			c <- result
			return
		}

		reply2 := result.Reply.(*pb.RpcCallReply)
		reply3, err := meta.DecodeReply(reply2.Payload)
		if err != nil {
			c <- &network.RpcResult{
				Reply: nil,
				Err:   network.InternalErr,
			}
		} else {
			c <- &network.RpcResult{
				Reply: reply3,
				Err:   nil,
			}
		}
	}()

	return c
}

func (pa *PlayerAgent) PushBackend(msgID protoc.IMessageID, arg interface{}) {
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("PushBackend no meta %s", msgID)
		return
	}

	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("PushBackend no center")
		return
	}

	payload, err := meta.EncodeArg(arg)
	if err != nil {
		glog.Errorf("PushBackend EncodeArg %d %s", msgID, err)
		return
	}

	region := pa.region
	if region <= 0 {
		region = GetAgentRegion(pa.uid)
	}

	centerSes.Push(pb.MessageID_GT2C_CLIENT_RPC_PUSH, &pb.RpcCallArg{
		Client: &pb.PlayerClient{
			ClientID: uint64(pa.clientID),
			Uid:      uint64(pa.uid),
			GateID:   pa.gateID,
			Region:   region,
		},
		MsgID:   msgID.ID(),
		Payload: payload,
	})
}

func (pa *PlayerAgent) PushClient(msgID protoc.IMessageID, arg interface{}) {
	meta := protoc.GetMeta(msgID.ID())
	if meta == nil {
		glog.Errorf("PushBackend no meta %s", msgID)
		return
	}

	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("PushBackend no center")
		return
	}

	payload, err := meta.EncodeArg(arg)
	if err != nil {
		glog.Errorf("PushBackend EncodeArg %d %s", msgID, err)
		return
	}

	region := pa.region
	if region <= 0 {
		region = GetAgentRegion(pa.uid)
	}

	centerSes.Push(pb.MessageID_C2GT_PUSH_CLIENT, &pb.RpcCallArg{
		Client: &pb.PlayerClient{
			ClientID: uint64(pa.clientID),
			Uid:      uint64(pa.uid),
			GateID:   pa.gateID,
			Region:   region,
		},
		MsgID:   msgID.ID(),
		Payload: payload,
	})
}

func (pa *PlayerAgent) SetDispatchApp(appName string, appID uint32) {
	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("SetDispatchApp no center")
		return
	}

	if appID > 0 {
		uid2Agent[pa.uid] = pa
	} else {
		delete(uid2Agent, pa.uid)
	}

	centerSes.Push(pb.MessageID_L2C_SET_DISPATCH, &pb.SetDispatchArg{
		Uid:     uint64(pa.uid),
		AppName: appName,
		AppID:   appID,
	})
}

func (pa *PlayerAgent) SetClientFilter(key, val string) {
	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("SetClientFilter no center")
		return
	}

	centerSes.Push(pb.MessageID_C2GT_CLIENT_SET_FILTER, &pb.ClientSetFilterArg{
		Uid:      uint64(pa.uid),
		ClientID: uint64(pa.clientID),
		Filter: &pb.BroadcastClientFilter{
			Key: key,
			Val: val,
		},
	})
}

func (pa *PlayerAgent) ClearClientFilter() {
	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("SetClientFilter no center")
		return
	}

	region := pa.region
	if region <= 0 {
		region = GetAgentRegion(pa.uid)
	}

	centerSes.Push(pb.MessageID_C2GT_CLIENT_CLEAR_FILTER, &pb.PlayerClient{
		ClientID: uint64(pa.clientID),
		GateID:   pa.gateID,
		Uid:      uint64(pa.uid),
		Region:   region,
	})
}

func (pa *PlayerAgent) packMsg() *pb.PlayerClient {
	return &pb.PlayerClient{
		ClientID: uint64(pa.clientID),
		Uid:      uint64(pa.uid),
		GateID:   pa.gateID,
		Region:   pa.GetRegion(),
	}
}

func (pa *PlayerAgent) Logout() {
	centerSes := api.SelectCenterByUUid(pa.uid, pa.region)
	if centerSes == nil {
		glog.Errorf("PlayerAgent.Logout no center")
		return
	}

	region := pa.region
	if region <= 0 {
		region = GetAgentRegion(pa.uid)
	}

	centerSes.Push(pb.MessageID_L2C_PLAYER_LOGOUT, pa.packMsg())
}
