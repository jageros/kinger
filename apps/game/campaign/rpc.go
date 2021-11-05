package campaign

import (
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/wordfilter"
	"kinger/gopuppy/network"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
)

func rpc_C2S_FetchCamoaignInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	if module.OutStatus.GetStatus(player, consts.OtFatigue) != nil {
		player.Tellme("由于您的实名认证未达要求，游戏时长已达3小时，现已进入#cb防沉迷#n状态，今天无法继续对战。", 0)
		return nil, gamedata.InternalErr
	}

	if player.GetMaxPvpLevel() < limitPvpLevel {
		return nil, gamedata.GameError(1)
	}

	if config.GetConfig().IsMultiLan {
		return nil, gamedata.GameError(2)
	}

	return agent.CallBackend(pb.MessageID_G2CA_FETCH_CAMPAIGN_INFO, &pb.CampaignSimplePlayer{
		Name: player.GetName(),
		HeadImg: player.GetHeadImgUrl(),
		HeadFrame: player.GetHeadFrame(),
		PvpScore: int32(module.Player.GetResource(player, consts.Score)),
	})
}

func rpc_C2S_SettleCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetCity)
	playerArg := &pb.CampaignSimplePlayer{
		Name: player.GetName(),
		HeadImg: player.GetHeadImgUrl(),
		HeadFrame: player.GetHeadFrame(),
		PvpScore: int32(module.Player.GetResource(player, consts.Score)),
	}
	return agent.CallBackend(pb.MessageID_G2CA_SETTLE_CITY, &pb.GSettleCityArg{
		CityID: arg2.CityID,
		Player: playerArg,
	})
}

func rpc_C2S_CreateCountry(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CreateCountryArg)
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if arg2.Gold <= 0 || !resCpt.HasResource(consts.Gold, int(arg2.Gold)) {
		return nil, gamedata.GameError(1)
	}

	resCpt.ModifyResource(consts.Gold, - int(arg2.Gold), consts.RmrCreateCountry)
	glog.Infof("rpc_C2S_CreateCountry, uid=%d, gold=%d", uid, arg2.Gold)

	reply, err := agent.CallBackend(pb.MessageID_G2CA_CREATE_COUNTRY, arg)
	if err != nil {
		resCpt.ModifyResource(consts.Gold, int(arg2.Gold), consts.RmrUnknownOutput)
		glog.Errorf("rpc_C2S_CreateCountry error, uid=%d, gold=%d, err=%s", uid, arg2.Gold, err)
		return nil, err
	}
	return reply, nil
}

func rpc_C2S_AcceptCampaignMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.AcceptCampaignMissionArg)
	var gcardIDs []uint32
	var cards []types.ICollectCard
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)

	for _, cardID := range arg2.Cards {
		card := cardCpt.GetCollectCard(cardID)
		if card == nil {
			return nil, gamedata.GameError(1)
		}

		state := card.GetState()
		if state == pb.CardState_InCampaignMs || state == pb.CardState_InSeasonPvp {
			return nil, gamedata.GameError(2)
		}

		gcardIDs = append(gcardIDs, card.GetCardGameData().GCardID)
		cards = append(cards, card)
	}
	arg2.Cards = gcardIDs

	if arg2.Type != pb.CampaignMsType_Dispatch && len(gcardIDs) < 5 {
		return nil, gamedata.GameError(3)
	}

	reply, err := agent.CallBackend(pb.MessageID_G2CA_ACCEPT_CAMPAIGN_MISSION, arg2)
	if err != nil {
		return nil, err
	}

	reply2 := reply.(*pb.GAcceptCampaignMissionReply)
	if reply2.RewardGold > 0 {
		module.Player.ModifyResource(player, consts.Gold, int(reply2.RewardGold), consts.RmrCampaignMission)
	}

	return reply2.AcceptReply, nil
}

func rpc_C2S_CancelCampaignMission(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return agent.CallBackend(pb.MessageID_G2CA_CANCEL_CAMPAIGN_MISSION, nil)
}

func rpc_C2S_GetCampaignMissionReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply, err := agent.CallBackend(pb.MessageID_G2CA_GET_CAMPAIGN_MISSION_REWARD, nil)
	if err != nil {
		return nil, err
	}

	reply2 := reply.(*pb.GGetCampaignMissionRewardReply)
	module.Player.ModifyResource(player, consts.Gold, int(reply2.Gold), consts.RmrCampaignMission)
	//module.Card.OnCampaignMissionDone(player, reply2.CardIDs)
	return nil, nil
}

func rpc_C2S_CityCapitalInjection(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CityCapitalInjectionArg)
	if arg2.Gold <= 0 {
		return nil, gamedata.GameError(1)
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Gold, int(arg2.Gold)) {
		return nil, gamedata.GameError(2)
	}
	resCpt.ModifyResource(consts.Gold, - int(arg2.Gold), consts.RmrCityCapitalInjection)

	reply, err := agent.CallBackend(pb.MessageID_G2CA_CITY_CAPITAL_INJECTION, arg)
	if err != nil {
		resCpt.ModifyResource(consts.Gold, int(arg2.Gold), consts.RmrUnknownOutput)
	}

	return reply, err
}

func rpc_C2S_MoveCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.MoveCityArg)
	if arg2.Gold < 0 {
		return nil, gamedata.GameError(1)
	}

	reply, err := agent.CallBackend(pb.MessageID_G2CA_GET_MY_COUNTRY, nil)
	if err != nil {
		return nil, err
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	needGold := reply.(*pb.GetMyCountryReply).CountryID > 0
	if needGold {
		if arg2.Gold == 0 {
			return nil, gamedata.GameError(2)
		}

		if !resCpt.HasResource(consts.Gold, int(arg2.Gold)) {
			return nil, gamedata.GameError(3)
		}
		resCpt.ModifyResource(consts.Gold, - int(arg2.Gold), consts.RmrMoveCity)
	}

	reply, err = agent.CallBackend(pb.MessageID_G2CA_MOVE_CITY, &pb.TargetCity{
		CityID: arg2.CityID,
	})
	if err != nil {
		if needGold {
			resCpt.ModifyResource(consts.Gold, int(arg2.Gold), consts.RmrUnknownOutput)
		}
		return nil, err
	}

	return nil, nil
}

func rpc_C2S_CountryModifyName(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CountryModifyNameArg)
	if _, hasDirty, _, wTy := wordfilter.ContainsDirtyWords(arg2.Name, false); hasDirty && wTy == gconsts.GeneralWords {
		return nil, gamedata.GameError(101)
	}

	jade := 100
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Jade, jade) {
		return nil, gamedata.GameError(1)
	}
	resCpt.ModifyResource(consts.Jade, - jade, consts.RmrCountryModifyName)

	_, err := agent.CallBackend(pb.MessageID_G2CA_COUNTRY_MODIFY_NAME, arg)
	if err != nil {
		resCpt.ModifyResource(consts.Jade, jade, consts.RmrUnknownOutput)
	}
	return nil, err
}

func rpc_C2S_DefCity(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	fighterData := player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData()
	if fighterData == nil || len(fighterData.HandCards) < consts.MaxHandCardAmount {
		return nil, gamedata.GameError(1)
	}

	return agent.CallBackend(pb.MessageID_G2CA_DEF_CITY, fighterData)
}

func rpc_CA2G_UpdateCampaignPlayerInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	player.GetComponent(consts.CampaignCpt).(*campaignComponent).setInfo(arg.(*pb.GCampaignPlayerInfo))
	return nil, nil
}

func rpc_CA2G_CampaignNoticeNotify(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	player.GetComponent(consts.CampaignCpt).(*campaignComponent).addNotice(arg.(*pb.CampaignNotice))
	return nil, nil
}

func rpc_C2S_AcceptMilitaryOrder(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	arg2 := arg.(*pb.AcceptMilitaryOrderArg)
	if len(arg2.CardIDs) < consts.MaxHandCardAmount {
		return nil, gamedata.GameError(1)
	}

	var camp int
	for _, cardID := range arg2.CardIDs {
		card := cardCpt.GetCollectCard(cardID)
		if card == nil || card.GetState() == pb.CardState_InCampaignMs {
			return nil, gamedata.GameError(2)
		}

		cardData := card.GetCardGameData()
		if camp == 0 {
			if cardData.Camp != consts.Heroes {
				camp = cardData.Camp
			}
		} else if cardData.Camp != consts.Heroes && cardData.Camp != camp {
			return nil, gamedata.GameError(3)
		}
	}

	return agent.CallBackend(pb.MessageID_G2CA_ACCEPT_MILITARY_ORDER, arg)
}

func rpc_C2S_EscapedFromJail(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	gold := 2000
	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Gold, gold) {
		return nil, gamedata.GameError(100)
	}

	resCpt.ModifyResource(consts.Gold, - gold, consts.RmrEscapedFromJail)
	reply, err := agent.CallBackend(pb.MessageID_G2CA_ESCAPED_FROM_JAIL, nil)
	if err != nil {
		resCpt.ModifyResource(consts.Gold, gold, consts.RmrUnknownOutput)
		return nil, err
	}
	return reply, nil
}

func rpc_C2S_CampaignBuyGoods(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CampaignBuyGoodsArg)
	goods := getGoods(arg2.Type, int(arg2.GoodsID))
	if goods == nil {
		return nil, gamedata.GameError(1)
	}

	if !goods.canBuy(player) {
		return nil, gamedata.GameError(2)
	}

	if !player.GetComponent(consts.CampaignCpt).(*campaignComponent).modifyContribution(- goods.getPrice()) {
		return nil, gamedata.GameError(3)
	}

	itemID, itemName := goods.buy(player)
	module.Shop.LogShopBuyItem(player, itemID, itemName, 1, "campaign", "contribution",
		"战功", goods.getPrice(), "")
	return nil, nil
}

func rpc_CA2G_UpdateCampaignInfo(_ *network.Session, arg interface{}) (interface{}, error) {
	mod.updateCampaignInfo(arg.(*pb.GCampaignInfo))
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CAMPAIGN_INFO, rpc_C2S_FetchCamoaignInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SETTLE_CITY, rpc_C2S_SettleCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CREATE_COUNTRY, rpc_C2S_CreateCountry)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ACCEPT_CAMPAIGN_MISSION, rpc_C2S_AcceptCampaignMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_CAMPAIGN_MISSION, rpc_C2S_CancelCampaignMission)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_CAMPAIGN_MISSION_REWARD, rpc_C2S_GetCampaignMissionReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CITY_CAPITAL_INJECTION, rpc_C2S_CityCapitalInjection)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MOVE_CITY, rpc_C2S_MoveCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_COUNTRY_MODIFY_NAME, rpc_C2S_CountryModifyName)
	//logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DEF_CITY, rpc_C2S_DefCity)
	logic.RegisterAgentRpcHandler(pb.MessageID_CA2G_UPDATE_CAMPAIGN_PLAYER_INFO, rpc_CA2G_UpdateCampaignPlayerInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_CA2G_CAMPAIGN_NOTICE_NOTIFY, rpc_CA2G_CampaignNoticeNotify)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ACCEPT_MILITARY_ORDER, rpc_C2S_AcceptMilitaryOrder)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ESCAPED_FROM_JAIL, rpc_C2S_EscapedFromJail)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CAMPAIGN_BUY_GOODS, rpc_C2S_CampaignBuyGoods)

	logic.RegisterRpcHandler(pb.MessageID_CA2G_UPDATE_CAMPAIGN_INFO, rpc_CA2G_UpdateCampaignInfo)
}
