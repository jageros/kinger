package treasure

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/proto/pb"
)

func rpc_C2S_GetTreasure(agent *logic.PlayerAgent, _ interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	treasureComponent := player.GetComponent(consts.TreasureCpt).(*treasureComponent)
	msg := &pb.GetTreasuresReply{}
	msg.Treasures, msg.DailyTreasure1 = treasureComponent.getTreasures()
	return msg, nil
}

func rpc_C2S_OpenTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg1 := arg.(*pb.OpenTreasureArg)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	treasureComponent := player.GetComponent(consts.TreasureCpt).(*treasureComponent)
	treasureReward, isDaily, modelID := treasureComponent.openTreasure(int(arg1.TreasureID))
	if treasureReward.OK && !isDaily {
		treasureReward.ShareHid = int32(player.GetComponent(consts.WxgameCpt).(types.IWxgameComponent).TriggerTreasureShareHD(modelID))
		treasureReward.CanWatchAddCardAds = treasureComponent.triggerAddCardAds(arg1.TreasureID, modelID, treasureReward.CardIDs)
	}
	return treasureReward, nil
}

func rpc_C2S_ActivateRewardTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg1 := arg.(*pb.ActivateRewardTreasureArg)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	treasureComponent := player.GetComponent(consts.TreasureCpt).(*treasureComponent)
	openTimeout, ok := treasureComponent.activateRewardTreasure(int(arg1.TreasureID))
	return &pb.ActivateRewardTreasureReply{OK: ok, OpenTimeout: openTimeout}, nil
}

func rpc_C2S_JadeAccTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	treasureComponent := player.GetComponent(consts.TreasureCpt).(*treasureComponent)
	return nil, treasureComponent.jadeAccTreasure(arg.(*pb.TargetTreasure).TreasureID)
}

func rpc_C2S_WatchTreasureAddCardAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.WatchTreasureAddCardAdsArg)
	return player.GetComponent(consts.TreasureCpt).(*treasureComponent).WatchTreasureAddCardAds(
		arg2.TreasureID, arg2.IsConsumeJade)
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_TREASURES, rpc_C2S_GetTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_OPEN_TREASURE, rpc_C2S_OpenTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ACTIVATE_REWARD_TREASURE, rpc_C2S_ActivateRewardTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_JADE_ACC_TREASURE, rpc_C2S_JadeAccTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_TREASURE_ADD_CARD_ADS, rpc_C2S_WatchTreasureAddCardAds)
}
