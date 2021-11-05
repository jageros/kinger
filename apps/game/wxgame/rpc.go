package wxgame

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/network"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"time"
	"kinger/gopuppy/common/glog"
	"kinger/common/utils"
)

func onBeginBattle(args ...interface{}) {
	uid := args[0].(common.UUid)
	player := module.Player.GetPlayer(uid)
	if player != nil {
		player.GetComponent(consts.WxgameCpt).(*wxgameComponent).endShareBattleLose()
	}
}

func onLogout(args ...interface{}) {
	agent := args[0].(*logic.PlayerAgent)
	mod.CancelInviteBattle(agent.GetUid())
}

func rpc_C2S_WxInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, network.InternalErr
	}

	if !module.Pvp.CanPvpMatch(player) {
		return nil, gamedata.GameError(1)
	}
	mod.newInviteBattleRoom(uid)
	return nil, nil
}

func rpc_C2S_WxCancelInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	mod.CancelInviteBattle(uid)
	return nil, nil
}

func rpc_C2S_WxReplyInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil || player.IsInBattle() {
		return nil, gamedata.InternalErr
	}

	if !module.Pvp.CanPvpMatch(player) {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.ReplyWxInviteBattleArg)
	targetPlayer := module.Player.GetPlayer(common.UUid(arg2.Uid))
	if targetPlayer != nil {
		ok := mod.replyInviteBattle(common.UUid(arg2.Uid), player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData())
		if !ok {
			return nil, gamedata.InternalErr
		} else {
			module.Mission.OnInviteBattle(player)
			return nil, nil
		}
	} else {
		targetAgent := logic.NewPlayerAgent(&gpb.PlayerClient{
			Uid: arg2.Uid,
			Region: logic.GetAgentRegion(common.UUid(arg2.Uid)),
		})
		_, err := targetAgent.CallBackend(pb.MessageID_G2G_WX_REPLY_INVITE_BATTLE, &pb.G2GReplyWxInviteBattleArg{
			InviteUid: arg2.Uid,
			BeInviter: player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData(),
		})
		if err == nil {
			module.Mission.OnInviteBattle(player)
		}

		return nil, err
	}
}

func rpc_G2G_WxReplyInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil || player.IsInBattle() {
		return nil, gamedata.InternalErr
	}

	if !module.Pvp.CanPvpMatch(player) {
		return nil, gamedata.GameError(1)
	}

	ok := mod.replyInviteBattle(uid, arg.(*pb.G2GReplyWxInviteBattleArg).BeInviter)
	if !ok {
		return nil, gamedata.InternalErr
	} else {
		module.Mission.OnInviteBattle(player)
		return nil, nil
	}
}

func rpc_C2S_GetShareTreasureReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	wxCpt := player.GetComponent(consts.WxgameCpt).(*wxgameComponent)
	treasureReward := wxCpt.getShareTreasureReward(int(arg.(*pb.GetShareTreasureArg).Hid))
	if !treasureReward.OK {
		return nil, gamedata.GameError(1)
	}

	return treasureReward, nil
}

func rpc_C2S_ShareTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.ShareTreasureArg)
	wxcfg := config.GetConfig().Wxgame
	glog.Infof("rpc_C2S_ShareTreasure uid=%d, IsExamine=%v, WxGroupID=%s", uid, wxcfg.IsExamined2, arg2.WxGroupID)
	if wxcfg.IsExamined2 && arg2.WxGroupID == "" {
		return nil, nil
	}

	err := player.GetComponent(consts.WxgameCpt).(*wxgameComponent).shareTreasure(arg2.TreasureID, arg2.WxGroupID)
	if err != nil {
		return nil, err
	}
	glog.Infof("rpc_C2S_ShareTreasure 222222222222 uid=%d, IsExamine=%v, WxGroupID=%s", uid, wxcfg.IsExamined2, arg2.WxGroupID)

	if wxcfg.IsExamined2 {
		timer.AfterFunc(time.Duration(wxcfg.DelayRewardTime)*time.Second, func() {
			helpShareTreasure(player, arg2.TreasureID)
		})
	}

	return nil, nil
}

func helpShareTreasure(player types.IPlayer, treasureID uint32) {
	glog.Infof("helpShareTreasure uid=%d", player.GetUid())
	treasureCpt := player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent)
	if !treasureCpt.IsDailyTreasure(treasureID) {
		resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		if resCpt.GetResource(consts.AccTreasureCnt) <= 0 {
			return
		}
	}

	treasureCpt.HelpOpenTreasure(treasureID)
}

func rpc_C2S_HelpShareTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	if config.GetConfig().Wxgame.IsExamined2 {
		return nil, nil
	}

	arg2 := arg.(*pb.HelpShareTreasureArg)
	shareUid := common.UUid(arg2.ShareUid)
	if shareUid == agent.GetUid() {
		return nil, nil
	}

	sharePlayer := module.Player.GetPlayer(shareUid)
	if sharePlayer == nil {
		return nil, nil
	}

	helpShareTreasure(sharePlayer, arg2.TreasureID)
	return nil, nil
}

func rpc_C2S_ShareBattleLose(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	err := player.GetComponent(consts.WxgameCpt).(*wxgameComponent).shareBattleLose(arg.(*pb.ShareBattleLoseArg).WxGroupID)
	if err != nil {
		return nil, err
	}

	wxconfig := config.GetConfig().Wxgame
	if wxconfig.IsExamined2 {
		timer.AfterFunc(time.Duration(wxconfig.DelayRewardTime)*time.Second, func() {
			player.GetComponent(consts.WxgameCpt).(*wxgameComponent).helpShareBattleLose()
		})
	}

	return nil, nil
}

func rpc_C2S_EndShareBattleLose(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player.GetComponent(consts.WxgameCpt).(*wxgameComponent).endShareBattleLose()
	return nil, nil
}

func rpc_C2S_HelpShareBattleLose(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	if config.GetConfig().Wxgame.IsExamined2 {
		return nil, nil
	}

	arg2 := arg.(*pb.HelpShareBattleLoseArg)
	shareUid := common.UUid(arg2.ShareUid)
	if shareUid == agent.GetUid() {
		return nil, nil
	}

	sharePlayer := module.Player.GetPlayer(shareUid)
	if sharePlayer == nil {
		return nil, nil
	}

	sharePlayer.GetComponent(consts.WxgameCpt).(*wxgameComponent).helpShareBattleLose()
	return nil, nil
}

// -------------  广告 -----------------------------------

func rpc_DailyTreasureReadAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).DailyTreasureBeDobule(
		arg.(*pb.DailyTreasureReadAdsArg).IsConsumeJade)
	return nil, nil
}

func rpc_TreasureReadAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if !player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).HasResource(consts.AccTreasureCnt, 1) {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.TreasureReadAdsArg)
	remainTime, err := player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).AccTreasure(arg2.TreasureID,
		arg2.IsConsumeJade)
	if err != nil {
		return nil, err
	}

	return &pb.TreasureReadAdsReply{
		RemainTime: remainTime,
	}, nil
}

func rpc_BattleLoseReadAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	resCpt := player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.NotSubStarCnt, 1) {
		return nil, gamedata.GameError(1)
	}

	resCpt.BatchModifyResource(map[int]int{
		consts.Score:         1,
		consts.NotSubStarCnt: -1,
	})
	return &pb.BattleLoseReadAdsReply{
		AddStar: 1,
	}, nil
}

func rpc_C2S_WatchUpTreasureRareAds(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).UpTreasureRare(
		arg.(*pb.WatchUpTreasureRareAdsArg).IsConsumeJade)
}

func rpc_C2S_WxgameShare(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.WxgameShareArg)
	glog.Infof("wxshare uid=%d, shareID=%d, shareType=%d", uid, arg2.ShareID, arg2.ShareType)
	module.Mission.OnWxShare(player, 0)
	eventhub.Publish(consts.EvShare, player)
	return nil, nil

	var err error = nil
	switch arg2.ShareType {
	case stShopGoldAds:
		fallthrough
	case stShopJadeAds:
		fallthrough
	case stShopTreasureAds:
		fallthrough
	case stUpTreasureRareAds:
		fallthrough
	case stTreasureAddCardAds:
		err = player.GetComponent(consts.WxgameCpt).(*wxgameComponent).onShare(int(arg2.ShareType), arg2.WxGroupID, arg2.Data)
	}

	return nil, err
}

func rpc_C2S_ClickWxgameShare(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.ClickWxgameShareArg)
	glog.Infof("click wxshare uid=%d, shareUid=%d, shareID=%d, shareType=%d, isNewbie=%v", uid, arg2.ShareUid,
		arg2.ShareID, arg2.ShareType, player.GetComponent(consts.TutorialCpt).(types.ITutorialComponent).GetCampID() == 0)

	data := arg2.Data
	if arg2.ShareType == stDailyShare {

		msg := &pb.GWxDailyShare{
			Uid: uint64(uid),
			HeadImg: player.GetHeadImgUrl(),
			Name: player.GetName(),
		}
		data, _ = msg.Marshal()

	} else if arg2.ShareType == stDailyTreasureDouble {

		shareArg := &pb.TargetTreasure{}
		shareArg.Unmarshal(data)

		msg := &pb.GWxDailyTreasureShare{
			HelperUid: uint64(uid),
			HelperHeadImg: player.GetHeadImgUrl(),
			HelperName:  player.GetName(),
			TreasureID: shareArg.TreasureID,
		}
		data, _ = msg.Marshal()
	}

	sharePlayer := module.Player.GetPlayer(common.UUid(arg2.ShareUid))
	if sharePlayer != nil {
		sharePlayer.GetComponent(consts.WxgameCpt).(*wxgameComponent).OnShareBeHelp(int(arg2.ShareType),
			int64(arg2.ShareTime), data)
	} else {
		utils.PlayerMqPublish(common.UUid(arg2.ShareUid), pb.RmqType_WxShareBeHelp, &pb.RmqWxShareBeHelp{
			ShareType: arg2.ShareType,
			ShareTime: int64(arg2.ShareTime),
			Data: data,
		})
	}

	return nil, nil
}

func rpc_C2S_WatchAdsBegin(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	glog.Infof("watch ads begin, uid=%d", uid)
	return nil, nil
}

func rpc_C2S_WatchAdsEnd(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	glog.Infof("watch ads end, uid=%d", uid)
	return nil, nil
}

func rpc_C2S_CancelWxShare(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CancelWxShareArg)
	if arg2.ShareType == stTreasureAddCardAds {
		player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).CancelTreasureAddCardAds()
	}
	return nil, nil
}

func rpc_C2S_IosShare(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	jade, bowlder, ticket, err := player.GetComponent(consts.WxgameCpt).(*wxgameComponent).getDailyShareReward()
	if err != nil {
		return nil, err
	}

	eventhub.Publish(consts.EvShare, player)
	return &pb.DailyShareReward{
		Jade: int32(jade),
		Bowlder: int32(bowlder),
		Ticket: int32(ticket),
	}, nil
}

func rpc_C2S_FetchDailyShareInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.WxgameCpt).(*wxgameComponent).packDailyShareMsg(), nil
}

func rpc_C2S_GetDailyShareReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	wxCpt := player.GetComponent(consts.WxgameCpt).(*wxgameComponent)
	jade, bowlder, ticket, err := wxCpt.getDailyShareReward()
	if err != nil {
		return nil, err
	}

	helperUid := wxCpt.getDailyShareHelperUid()
	helper := module.Player.GetPlayer(helperUid)
	if helper != nil {
		mod.ReturnDailyShareReward(helper, player.GetName())
	} else {
		utils.PlayerMqPublish(helperUid, pb.RmqType_ReturnDailyShareReward, &pb.WxDailyShareArg{
			PlayerName: player.GetName(),
		})
	}

	return &pb.DailyShareReward{
		Jade: int32(jade),
		Bowlder: int32(bowlder),
		Ticket: int32(ticket),
	}, nil
}

func registerRpc() {
	eventhub.Subscribe(consts.EvBeginBattle, onBeginBattle)
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onLogout)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, onLogout)

	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WX_INVITE_BATTLE, rpc_C2S_WxInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WX_CANCEL_INVITE_BATTLE, rpc_C2S_WxCancelInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WX_REPLY_INVITE_BATTLE, rpc_C2S_WxReplyInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_SHARE_TREASURE_REWARD, rpc_C2S_GetShareTreasureReward)
	//logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SHARE_TREASURE, rpc_C2S_ShareTreasure)
	//logic.RegisterAgentRpcHandler(pb.MessageID_C2S_HELP_SHARE_TREASURE, rpc_C2S_HelpShareTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SHARE_BATTLE_LOSE, rpc_C2S_ShareBattleLose)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_END_SHARE_BATTLE_LOSE, rpc_C2S_EndShareBattleLose)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_HELP_SHARE_BATTLE_LOSE, rpc_C2S_HelpShareBattleLose)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DAILY_TREASURE_READ_ADS, rpc_DailyTreasureReadAds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_TREASURE_READ_ADS, rpc_TreasureReadAds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BATTLE_LOSE_READ_ADS, rpc_BattleLoseReadAds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_UP_TREASURE_RARE_ADS, rpc_C2S_WatchUpTreasureRareAds)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WXGAME_SHARE, rpc_C2S_WxgameShare)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CLICK_WXGAME_SHARE, rpc_C2S_ClickWxgameShare)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_ADS_BEGIN, rpc_C2S_WatchAdsBegin)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_ADS_END, rpc_C2S_WatchAdsEnd)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_WX_SHARE, rpc_C2S_CancelWxShare)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_IOS_SHARE, rpc_C2S_IosShare)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_DAILY_SHARE_INFO, rpc_C2S_FetchDailyShareInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_DAILY_SHARE_REWARD, rpc_C2S_GetDailyShareReward)

	logic.RegisterAgentRpcHandler(pb.MessageID_G2G_WX_REPLY_INVITE_BATTLE, rpc_G2G_WxReplyInviteBattle)
}
