package main

import (
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
)

func onLogout(args ...interface{}) {
	agent := args[0].(*logic.PlayerAgent)
	uid := agent.GetUid()
	chatMgr.unsubscribeChat(pb.ChatChannel_World, uid, true)
	chatMgr.unsubscribeChat(pb.ChatChannel_CampaignCountry, uid, true)
}

func rpc_L2CA_SubscribeChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.SubscribeChatArg)
	chatlets := chatMgr.subscribeChat(arg2.Channel, agent.GetUid(), int(arg2.Area), arg2.CountryID)
	if chatlets == nil {
		return nil, gamedata.GameError(200)
	}
	return chatlets, nil
}

func rpc_C2S_UnsubscribeChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.TargetChatChannel)
	chatMgr.unsubscribeChat(arg2.Channel, uid, false)
	return nil, nil
}

func rpc_L2CA_SendChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.CSendChatArg)
	chatMgr.broadcastChat(agent.GetUid(), arg2, arg2.Channel, arg2.Msg)
	return nil, nil
}

func rpc_L2CA_ForbidChat(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.ForbidChatArg)
	if arg2.IsForbid {
		chatMgr.addForbidChatPlayer(arg2.Uid)
	} else {
		chatMgr.delForbidChatPlayer(arg2.Uid)
	}
	return nil, nil
}

func registerRpc() {
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onLogout)
	logic.RegisterAgentRpcHandler(pb.MessageID_L2CA_SUBSCRIBE_CHAT, rpc_L2CA_SubscribeChat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UNSUBSCRIBE_CHAT, rpc_C2S_UnsubscribeChat)
	logic.RegisterAgentRpcHandler(pb.MessageID_L2CA_SEND_CHAT, rpc_L2CA_SendChat)
	logic.RegisterRpcHandler(pb.MessageID_L2CA_FORBID_CHAT, rpc_L2CA_ForbidChat)
}
