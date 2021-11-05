package huodong

import (
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
	"kinger/gopuppy/network"
	"kinger/apps/game/module"
)

func rpc_G2G_OnHuodongEvent(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.HuodongEvent)
	mod.handlerHuodongEvent(arg2)
	return nil, nil
}

func registerRpc() {
	if module.Service.GetAppID() != 1 {
		logic.RegisterRpcHandler(pb.MessageID_G2G_ON_HUODONG_EVENT, rpc_G2G_OnHuodongEvent)
	}
}
