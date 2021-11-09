package social

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/wordfilter"
	"kinger/gopuppy/network"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"math"
	"sort"
	"time"
)

func onLogout(args ...interface{}) {
	agent := args[0].(*logic.PlayerAgent)
	mod.CancelInviteBattle(agent.GetUid())
}

func rpc_C2S_SendChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	//if player.IsForbidChat() {
	//	return nil, gamedata.GameError(1)
	//}

	arg2 := arg.(*pb.SendChatArg)

	msg, hasDirty, dirtyWs, wTy := wordfilter.ContainsDirtyWords(arg2.Msg, true)
	glog.JsonInfo("chat", glog.String("name", player.GetName()), glog.Uint64("uid", uint64(player.GetUid())),
		glog.String("chatChanel", "世界"), glog.String("ObjectName", ""), glog.Uint64("ObjectUid", 0),
		glog.String("content", arg2.Msg), glog.String("dirtyWords", dirtyWs), glog.Int("area", player.GetArea()))
	if hasDirty {
		if wTy == gconsts.AccurateWords {
			player.Forbid(consts.ForbidAccount, true, -1, arg2.Msg, true)
			return nil, nil
		}
		if wTy == gconsts.FuzzyWords {
			player.Forbid(consts.ForbidMonitor, true, -1, arg2.Msg, true)
		}
	}

	if player.GetComponent(consts.SocialCpt).(*socialComponent).getAdvertProtecter().onChat(msg) {
		return nil, nil
	}

	garg := &pb.CSendChatArg{
		Msg:         msg,
		Channel:     arg2.Channel,
		Name:        player.GetName(),
		HeadImgUrl:  player.GetHeadImgUrl(),
		PvpLevel:    int32(player.GetPvpLevel()),
		Country:     player.GetCountry(),
		HeadFrame:   player.GetHeadFrame(),
		ChatPop:     player.GetChatPop(),
		CountryFlag: player.GetCountryFlag(),
	}

	if arg2.Channel == pb.ChatChannel_CampaignCountry {
		campaignPlayerInfo := module.Campaign.GetPlayerInfo(player)
		if campaignPlayerInfo == nil || campaignPlayerInfo.CountryID <= 0 {
			return nil, gamedata.GameError(2)
		}
		garg.CountryID = campaignPlayerInfo.CountryID
		garg.CityID = campaignPlayerInfo.CityID
		garg.CityJob = campaignPlayerInfo.CityJob
		garg.CountryJob = campaignPlayerInfo.CountryJob
	}

	return agent.CallBackend(pb.MessageID_L2CA_SEND_CHAT, garg)
}

func rpc_C2S_FetchFriendList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	socialCpt := player.GetComponent(consts.SocialCpt).(*socialComponent)
	reply := &pb.FriendList{}
	reply.Friends, reply.LastOpponent = socialCpt.getFriendList()
	return reply, nil
}

func rpc_C2S_FetchPlayerInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetPlayer)
	targetUid := common.UUid(arg2.Uid)
	splayer := module.Player.GetSimplePlayerInfo(targetUid)
	if splayer == nil {
		return nil, gamedata.GameError(1)
	}
	firstHandAmount := splayer.GetFirstHandAmount()
	backHandAmount := splayer.GetBackHandAmount()
	firstHandWinAmount := splayer.GetFirstHandWinAmount()
	backHandWinAmount := splayer.GetBackHandWinAmount()
	f := player.GetComponent(consts.SocialCpt).(*socialComponent).getFriend(common.UUid(arg2.Uid))

	reply := &pb.PlayerInfo{
		Name:            splayer.GetName(),
		PvpScore:        int32(splayer.GetPvpScore()),
		HeadImgUrl:      splayer.GetHeadImgUrl(),
		BattleAmount:    int32(firstHandAmount + backHandAmount),
		BattleWinAmount: int32(firstHandWinAmount + backHandWinAmount),
		RankScore:       int32(splayer.GetRankScore()),
		FavoriteCards:   splayer.GetFavoriteCards(),
		FightCards:      splayer.FightCards,
		IsFriend:        f != nil,
		IsWechatFriend:  f != nil && f.isWechatFriend(),
		CanInviteBattle: splayer.IsOnline && !splayer.IsInBattle,
		Country:         splayer.Country,
		HeadFrame:       splayer.HeadFrame,
		StatusIDs:       splayer.StatusIDs,
		CrossAreaHonor:  splayer.CrossAreaHonor,
		CountryFlag:     splayer.CountryFlag,
	}

	if firstHandAmount <= 0 || firstHandWinAmount <= 0 {
		reply.FirstHandWinRate = 0
	} else {
		reply.FirstHandWinRate = int32(math.Round(float64(firstHandWinAmount) / float64(firstHandAmount) * 100))
	}

	if backHandAmount <= 0 || backHandWinAmount <= 0 {
		reply.BackHandWinRate = 0
	} else {
		reply.BackHandWinRate = int32(math.Round(float64(backHandWinAmount) / float64(backHandAmount) * 100))
	}

	targetPlayer := module.Player.GetPlayer(targetUid)
	var campaignInfo2 *pb.GCampaignPlayerInfo
	if targetPlayer != nil {
		campaignInfo2 = module.Campaign.GetPlayerInfo(targetPlayer)
	} else {
		campaignInfo, err := logic.CallBackend("", 0, pb.MessageID_G2CA_FETCH_CAMPAIGN_TARGET_PLAYER_INFO, arg)
		if err == nil {
			campaignInfo2 = campaignInfo.(*pb.GCampaignPlayerInfo)
		}
	}

	if campaignInfo2 != nil {
		reply.CityID = campaignInfo2.CityID
		reply.CityJob = campaignInfo2.CityJob
		reply.CountryJob = campaignInfo2.CountryJob
		reply.CampaignCountry = campaignInfo2.CountryName
	}

	return reply, nil
}

func rpc_C2S_AddFriendApply(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetPlayer)
	return nil, mod.AddFriendApply(player, common.UUid(arg2.Uid), false)
}

func rpc_C2S_FetchFriendApplyList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	applyList := player.GetComponent(consts.SocialCpt).(*socialComponent).getFriendApplyList()
	return &pb.FriendApplyList{
		FriendApplys: applyList,
	}, nil
}

func rpc_C2S_ReplyFriendApply(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.ReplyFriendApplyArg)
	player.GetComponent(consts.SocialCpt).(*socialComponent).replyFirendApply(common.UUid(arg2.Uid), arg2.IsAgree)
	return nil, nil
}

func rpc_C2S_DelFriend(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	arg2 := arg.(*pb.TargetPlayer)
	player.GetComponent(consts.SocialCpt).(*socialComponent).DelFriend(common.UUid(arg2.Uid))

	targetPlayer := module.Player.GetPlayer(common.UUid(arg2.Uid))
	if targetPlayer != nil {
		targetPlayer.GetComponent(consts.SocialCpt).(*socialComponent).DelFriend(uid)
		targetPlayer.GetAgent().PushClient(pb.MessageID_S2C_BE_DEL_FRIEND, &pb.TargetPlayer{
			Uid: uint64(uid),
		})
	} else {
		utils.PlayerMqPublish(common.UUid(arg2.Uid), pb.RmqType_BeDelFriend, &pb.RmqBeDelFriend{
			FromUid: uint64(uid),
		})
	}

	return nil, nil
}

func rpc_C2S_FetchPrivateChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.FetchPrivateChatArg)
	return &pb.PrivateChatList{
		PrivateChatItems: player.GetComponent(consts.SocialCpt).(*socialComponent).getAndDelPrivateChat(int(arg2.MaxID)),
	}, nil
}

func rpc_C2S_SendPrivateChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.SendPrivateChatArg)
	toUid := common.UUid(arg2.ToUid)
	toName := module.Player.GetSimplePlayerInfo(toUid).GetName()
	msg, hasDirty, dirtyWs, wTy := wordfilter.ContainsDirtyWords(arg2.Msg, true)
	glog.JsonInfo("chat", glog.String("name", player.GetName()), glog.Uint64("uid", uint64(player.GetUid())),
		glog.String("chatChanel", "私聊"), glog.String("ObjectName", toName), glog.Uint64("ObjectUid", arg2.ToUid),
		glog.String("content", arg2.Msg), glog.String("dirtyWords", dirtyWs), glog.Int("area", player.GetArea()))
	if hasDirty {
		if wTy == gconsts.AccurateWords {
			player.Forbid(consts.ForbidAccount, true, -1, arg2.Msg, true)
			return nil, nil
		}
		if wTy == gconsts.FuzzyWords {
			player.Forbid(consts.ForbidMonitor, true, -1, arg2.Msg, true)
		}
	}

	if player.GetComponent(consts.SocialCpt).(*socialComponent).getFriend(toUid) == nil {
		return nil, gamedata.GameError(1)
	}

	toPlayer := module.Player.GetPlayer(toUid)
	if toPlayer != nil {
		toPlayer.GetComponent(consts.SocialCpt).(*socialComponent).OnReceivePrivateChat(uid, player.GetName(),
			player.GetHeadImgUrl(), player.GetPvpLevel(), arg2.Msg, int(time.Now().Unix()), player.GetCountry(),
			player.GetHeadFrame(), player.GetChatPop(), player.GetCountryFlag())
	} else {
		utils.PlayerMqPublish(toUid, pb.RmqType_PrivateChat, &pb.RmqPrivateChat{
			FromUid:         uint64(uid),
			FromName:        player.GetName(),
			FromHeadImgUrl:  player.GetHeadImgUrl(),
			Msg:             msg,
			Time:            int32(time.Now().Unix()),
			FromPvpLevel:    int32(player.GetPvpLevel()),
			FromCountry:     player.GetCountry(),
			FromHeadFrame:   player.GetHeadFrame(),
			ChatPop:         player.GetChatPop(),
			FromCountryFlag: player.GetCountryFlag(),
		})
	}

	return nil, nil
}

func rpc_C2S_InviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if !module.Pvp.CanPvpMatch(player) {
		return nil, gamedata.GameError(4)
	}

	arg2 := arg.(*pb.TargetPlayer)
	if common.UUid(arg2.Uid) == uid {
		return nil, gamedata.GameError(1)
	}

	targetPlayer := module.Player.GetPlayer(common.UUid(arg2.Uid))
	if targetPlayer != nil && !targetPlayer.IsOnline() {
		return nil, gamedata.GameError(2)
	}
	if targetPlayer != nil && targetPlayer.IsInBattle() {
		return nil, gamedata.GameError(3)
	}

	if targetPlayer != nil {
		targetPlayer.GetAgent().PushClient(pb.MessageID_S2C_BE_INVITE_BATTLE, &pb.BeInviteBattleArg{
			Uid:  uint64(uid),
			Name: player.GetName(),
		})
	} else {
		// TODO maybe wrong
		agent := logic.GetPlayerAgent(common.UUid(arg2.Uid))
		if agent == nil {
			agent = logic.NewPlayerAgent(&gpb.PlayerClient{
				Uid:    arg2.Uid,
				Region: logic.GetAgentRegion(common.UUid(arg2.Uid)),
			})
		}
		if _, err := agent.CallBackend(pb.MessageID_G2G_INVITE_BATTLE, &pb.BeInviteBattleArg{
			Uid:  uint64(uid),
			Name: player.GetName(),
		}); err != nil {
			return nil, err
		}
	}

	mod.newInviteBattleRoom(uid, common.UUid(arg2.Uid))
	return nil, nil
}

func rpc_G2G_InviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	evq.CallLater(func() {
		player.GetAgent().PushClient(pb.MessageID_S2C_BE_INVITE_BATTLE, arg)
	})

	return nil, nil
}

func rpc_C2S_CancelInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	mod.CancelInviteBattle(uid)
	return nil, nil
}

func rpc_C2S_ReplyInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if !module.Pvp.CanPvpMatch(player) {
		return nil, gamedata.GameError(1)
	}

	arg2 := arg.(*pb.ReplyInviteBattleArg)
	inviterUid := common.UUid(arg2.Uid)
	inviter := module.Player.GetPlayer(inviterUid)
	var fighterData *pb.FighterData
	if arg2.IsAgree {
		fighterData = player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData()
	}

	if inviter != nil {
		if ok := mod.replyInviteBattle(fighterData, uid, inviterUid, arg2.IsAgree); ok {
			module.Mission.OnInviteBattle(player)
			eventhub.Publish(consts.EvCombat, player, inviter)
			return nil, nil
		} else {
			return nil, gamedata.InternalErr
		}
	} else {
		agent := logic.GetPlayerAgent(inviterUid)
		if agent == nil {
			agent = logic.NewPlayerAgent(&gpb.PlayerClient{Uid: arg2.Uid, Region: logic.GetAgentRegion(common.UUid(arg2.Uid))})
		}
		_, err := agent.CallBackend(pb.MessageID_G2G_REPLY_INVITE_BATTLE, &pb.G2GReplyInviteBattleArg{
			BeInviter: fighterData,
			IsAgree:   arg2.IsAgree,
			TargetUid: uint64(uid),
		})

		if err == nil {
			module.Mission.OnInviteBattle(player)
			eventhub.Publish(consts.EvCombat, player)
		}

		return nil, err
	}
}

func rpc_C2S_FetchWxInviteFriends(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.SocialCpt).(*socialComponent).packWxInviteFriendMsg(), nil
}

func rpc_C2S_GetWxInviteReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	rewardID := int(arg.(*pb.GetWxInviteRewardArg).ID)
	return player.GetComponent(consts.SocialCpt).(*socialComponent).getWxInviteFriendReward(rewardID)
}

func rpc_G2G_ReplyInviteBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.G2GReplyInviteBattleArg)
	if ok := mod.replyInviteBattle(arg2.BeInviter, common.UUid(arg2.TargetUid), uid, arg2.IsAgree); ok {
		return nil, nil
	} else {
		return nil, gamedata.InternalErr
	}
}

func rpc_C2S_SubscribeChat(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	player := module.Player.GetPlayer(agent.GetUid())
	arg2 := arg.(*pb.TargetChatChannel)
	garg := &pb.SubscribeChatArg{
		Channel: arg2.Channel,
		Area:    int32(player.GetArea()),
	}

	if arg2.Channel == pb.ChatChannel_CampaignCountry {
		if player == nil {
			return nil, gamedata.InternalErr
		}

		campaignPlayerInfo := module.Campaign.GetPlayerInfo(player)
		if campaignPlayerInfo == nil || campaignPlayerInfo.CountryID <= 0 {
			return nil, gamedata.GameError(1)
		}
		garg.CountryID = campaignPlayerInfo.CountryID
	}

	reply, err := agent.CallBackend(pb.MessageID_L2CA_SUBSCRIBE_CHAT, garg)
	if err != nil {
		return nil, err
	}

	reply2 := reply.(*pb.ChatItemList)
	var chatlets []interface {
		GetTime() int32
		Marshal() ([]byte, error)
	}
	for _, item := range reply2.ChatItems {
		chatlets = append(chatlets, item)
	}

	if arg2.Channel == pb.ChatChannel_CampaignCountry {
		campaignPlayerInfo := module.Campaign.GetPlayerInfo(player)
		for _, n := range campaignPlayerInfo.Notices {
			chatlets = append(chatlets, n)
		}

		sort.Slice(chatlets, func(i, j int) bool {
			return chatlets[i].GetTime() <= chatlets[j].GetTime()
		})
		if len(chatlets) > 40 {
			chatlets = chatlets[len(chatlets)-40:]
		}
	}

	resp := &pb.ChatList{}
	for _, item := range chatlets {
		clet := &pb.Chatlet{}
		switch item.(type) {
		case *pb.ChatItem:
			clet.Type = pb.Chatlet_Normal
		case *pb.CampaignNotice:
			clet.Type = pb.Chatlet_CampaignNotice
		default:
			continue
		}

		clet.Data, _ = item.Marshal()
		resp.Chatlets = append(resp.Chatlets, clet)
	}
	return resp, nil
}

func rpc_G2G_OnSendAdvertChat(_ *network.Session, arg interface{}) (interface{}, error) {
	mod.lastAdvertChat = arg.(*pb.OnSendAdvertChatArg).Msg
	return nil, nil
}

func rpc_G2G_OnForbidOrUnblockIpAddr(_ *network.Session, arg interface{}) (interface{}, error) {
	forbidInfo := arg.(*pb.IpAddrArg)
	utils.ForbidIpAddr(forbidInfo.IpAddr, forbidInfo.IsForbid)
	return nil, nil
}

func rpc_G2G_OnImportDirtyWords(_ *network.Session, arg interface{}) (interface{}, error) {
	words := arg.(*pb.ImportWordArg)
	utils.AddDirtyWords(words.AddWordsStr, words.DelWordsStr, words.IsAccurate)
	return nil, nil
}

func registerRpc() {
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onLogout)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, onLogout)

	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SEND_CHAT, rpc_C2S_SendChat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_FRIEND_LIST, rpc_C2S_FetchFriendList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_PLAYER_INFO, rpc_C2S_FetchPlayerInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ADD_FRIEND_APPLY, rpc_C2S_AddFriendApply)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_FRIEND_APPLY_LIST, rpc_C2S_FetchFriendApplyList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REPLY_AFRIEND_APPLY, rpc_C2S_ReplyFriendApply)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DEL_FRIEND, rpc_C2S_DelFriend)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_PRIVATE_CHAT, rpc_C2S_FetchPrivateChat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SEND_PRIVATE_CHAT, rpc_C2S_SendPrivateChat)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_INVITE_BATTLE, rpc_C2S_InviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2G_INVITE_BATTLE, rpc_G2G_InviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CANCEL_INVITE_BATTLE, rpc_C2S_CancelInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REPLY_INVITE_BATTLE, rpc_C2S_ReplyInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_WX_INVITE_FRIENDS, rpc_C2S_FetchWxInviteFriends)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_WX_INVITE_REWARD, rpc_C2S_GetWxInviteReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2G_REPLY_INVITE_BATTLE, rpc_G2G_ReplyInviteBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SUBSCRIBE_CHAT, rpc_C2S_SubscribeChat)

	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_SEND_ADVERT_CHAT, rpc_G2G_OnSendAdvertChat)
	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_FORBID_OR_UNBLOCK_IP, rpc_G2G_OnForbidOrUnblockIpAddr)
	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_IMPORT_DIRTY_WORDS, rpc_G2G_OnImportDirtyWords)
}
