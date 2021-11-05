package player

import (
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
)

func init() {
	mq.RegisterRmqMessage(pb.RmqType_BattleBegin, func() mq.IRmqMessge {
		return &pb.RmqBattleBegin{}
	})

	mq.RegisterRmqMessage(pb.RmqType_BattleEnd, func() mq.IRmqMessge {
		return &pb.RmqBattleEnd{}
	})

	mq.RegisterRmqMessage(pb.RmqType_Bonus, func() mq.IRmqMessge {
		return &pb.RmqBonus{}
	})

	mq.RegisterRmqMessage(pb.RmqType_AddFriendApply, func() mq.IRmqMessge {
		return &pb.RmqAddFriendApply{}
	})

	mq.RegisterRmqMessage(pb.RmqType_ReplyFriendApply, func() mq.IRmqMessge {
		return &pb.RmqReplyFriendApply{}
	})

	mq.RegisterRmqMessage(pb.RmqType_BeDelFriend, func() mq.IRmqMessge {
		return &pb.RmqBeDelFriend{}
	})

	mq.RegisterRmqMessage(pb.RmqType_PrivateChat, func() mq.IRmqMessge {
		return &pb.RmqPrivateChat{}
	})

	mq.RegisterRmqMessage(pb.RmqType_HelpLevel, func() mq.IRmqMessge {
		return &pb.RmqHelpLevel{}
	})

	mq.RegisterRmqMessage(pb.RmqType_WxInviteFriendTp, func() mq.IRmqMessge {
		return &pb.RmqWxInviteFriend{}
	})

	mq.RegisterRmqMessage(pb.RmqType_ForbidLogin, func() mq.IRmqMessge {
		return &pb.RmqForbidLogin{}
	})
	mq.RegisterRmqMessage(pb.RmqType_MonitorAccount, func() mq.IRmqMessge {
		return &pb.RmqMonitorAccount{}
	})

	mq.RegisterRmqMessage(pb.RmqType_ForbidChat, func() mq.IRmqMessge {
		return &pb.RmqForbidChat{}
	})

	mq.RegisterRmqMessage(pb.RmqType_WxShareBeHelp, func() mq.IRmqMessge {
		return &pb.RmqWxShareBeHelp{}
	})

	mq.RegisterRmqMessage(pb.RmqType_SendMail, func() mq.IRmqMessge {
		return &pb.Mail{}
	})

	mq.RegisterRmqMessage(pb.RmqType_SdkRecharge, func() mq.IRmqMessge {
		return &pb.RmqSdkRecharge{}
	})

	mq.RegisterRmqMessage(pb.RmqType_CampaignMissionDone, func() mq.IRmqMessge {
		return &pb.RmqCampaignMissionDone{}
	})

	mq.RegisterRmqMessage(pb.RmqType_CampaignAcceptMission, func() mq.IRmqMessge {
		return &pb.RmqCampaignAcceptMission{}
	})

	mq.RegisterRmqMessage(pb.RmqType_UnifiedReward, func() mq.IRmqMessge {
		return &pb.RmqUnifiedReward{}
	})

	mq.RegisterRmqMessage(pb.RmqType_ReturnDailyShareReward, func() mq.IRmqMessge {
		return &pb.WxDailyShareArg{}
	})

	mq.RegisterRmqMessage(pb.RmqType_CompensateRecharge, func() mq.IRmqMessge {
		return &pb.RmqCompensateRecharge{}
	})
}

type mqConsumer struct {
	player *Player
}

func (mc *mqConsumer) onBonusReward(msg interface{}) {
	changeRes := msg.(*pb.RmqBonus).ChangeRes
	resCpt := mc.player.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	modifyRes := map[int]int{}
	for _, res := range changeRes {
		modifyRes[int(res.Type)] = int(res.Amount)
	}
	resCpt.BatchModifyResource(modifyRes, consts.RmrBonus)
}

func (mc *mqConsumer) onReplyFriendApply(msg interface{}) {
	msg2 := msg.(*pb.RmqReplyFriendApply)
	mc.player.GetAgent().PushClient(pb.MessageID_S2C_FRIEND_APPLY_RESULT, &pb.FriendApplyResult{
		Name:    msg2.FromName,
		IsAgree: msg2.IsAgree,
	})

	if msg2.IsAgree {
		mc.player.GetComponent(consts.SocialCpt).(types.ISocialComponent).AddFriend(common.UUid(msg2.FromUid))
	}
}

func (mc *mqConsumer) onBeDelFriend(msg interface{}) {
	msg2 := msg.(*pb.RmqBeDelFriend)
	socialCpt := mc.player.GetComponent(consts.SocialCpt).(types.ISocialComponent)
	socialCpt.DelFriend(common.UUid(msg2.FromUid))
	mc.player.GetAgent().PushClient(pb.MessageID_S2C_BE_DEL_FRIEND, &pb.TargetPlayer{
		Uid: msg2.FromUid,
	})
}

func (mc *mqConsumer) onReceivePrivateChat(msg interface{}) {
	msg2 := msg.(*pb.RmqPrivateChat)
	mc.player.GetComponent(consts.SocialCpt).(types.ISocialComponent).OnReceivePrivateChat(common.UUid(msg2.FromUid),
		msg2.FromName, msg2.FromHeadImgUrl, int(msg2.FromPvpLevel), msg2.Msg, int(msg2.Time), msg2.FromCountry,
		msg2.FromHeadFrame, msg2.ChatPop, msg2.FromCountryFlag)
}

func (mc *mqConsumer) Consume(type_ int32, msg mq.IRmqMessge) {

	switch type_ {
	case pb.RmqType_BattleBegin.ID():
		msg2 := msg.(*pb.RmqBattleBegin)
		mc.player.OnBeginBattle(common.UUid(msg2.BattleID), int(msg2.BattleType), uint32(msg2.AppID))

	case pb.RmqType_BattleEnd.ID():
		msg2 := msg.(*pb.RmqBattleEnd)
		mc.player.OnBattleEnd(common.UUid(msg2.BattleID), int(msg2.BattleType), msg2.Winner, msg2.Loser, int(msg2.LevelID),
			msg2.IsWonderful)

	case pb.RmqType_Bonus.ID():
		mc.onBonusReward(msg)

	case pb.RmqType_AddFriendApply.ID():
		msg2 := msg.(*pb.RmqAddFriendApply)
		mc.player.GetComponent(consts.SocialCpt).(types.ISocialComponent).AddFriendApply(common.UUid(
			msg2.FromUid), msg2.FromName, msg2.IsInvite)

	case pb.RmqType_ReplyFriendApply.ID():
		mc.onReplyFriendApply(msg)

	case pb.RmqType_BeDelFriend.ID():
		mc.onBeDelFriend(msg)

	case pb.RmqType_PrivateChat.ID():
		mc.onReceivePrivateChat(msg)

	case pb.RmqType_HelpLevel.ID():
		msg2 := msg.(*pb.RmqHelpLevel)
		mc.player.GetComponent(consts.LevelCpt).(types.ILevelComponent).OnBeHelpBattle(common.UUid(msg2.HelperUid),
			msg2.HelperName, int(msg2.LevelID), common.UUid(msg2.BattleID))

	case pb.RmqType_WxInviteFriendTp.ID():
		msg2 := msg.(*pb.RmqWxInviteFriend)
		mc.player.GetComponent(consts.SocialCpt).(types.ISocialComponent).OnWxInviteFriendUpdate(common.UUid(
			msg2.Uid), msg2.HeadImgUrl, int(msg2.MaxPvpLevel))

	case pb.RmqType_ForbidLogin.ID():
		msg2 := msg.(*pb.RmqForbidLogin)
		mc.player.Forbid(consts.ForbidAccount, msg2.IsForbid, msg2.OverTime, "", false)

	case pb.RmqType_ForbidChat.ID():
		msg2 := msg.(*pb.RmqForbidChat)
		mc.player.Forbid(consts.ForbidChat, msg2.IsForbid, msg2.OverTime, "",  false)

	case pb.RmqType_MonitorAccount.ID():
		msg2 := msg.(*pb.RmqMonitorAccount)
		mc.player.Forbid(consts.ForbidAccount, msg2.IsForbid, msg2.OverTime, "", false)

	case pb.RmqType_WxShareBeHelp.ID():
		msg2 := msg.(*pb.RmqWxShareBeHelp)
		mc.player.GetComponent(consts.WxgameCpt).(types.IWxgameComponent).OnShareBeHelp(int(msg2.ShareType),
			int64(msg2.ShareTime), msg2.Data)

	case pb.RmqType_SendMail.ID():
		msg2 := msg.(*pb.Mail)
		mc.player.GetComponent(consts.MailCpt).(types.IMailComponent).SendMail(msg2.Title, msg2.Content, msg2.SenderName,
			int(msg2.Time), module.Mail.NewMailRewardByMsg(msg2.Rewards), msg2.MailType, msg2.Arg)

	case pb.RmqType_SdkRecharge.ID():
		msg2 := msg.(*pb.RmqSdkRecharge)
		mc.player.GetComponent(consts.ShopCpt).(types.IShopComponent).OnSdkRecharge(msg2.ChannelUid, msg2.CpOrderID,
			msg2.ChannelOrderID, int(msg2.PaymentAmount), msg2.NeedCheckMoney)

	case pb.RmqType_CampaignMissionDone.ID():
		msg2 := msg.(*pb.RmqCampaignMissionDone)
		module.Card.OnCampaignMissionDone(mc.player, msg2.Cards)

	case pb.RmqType_CampaignAcceptMission.ID():
		msg2 := msg.(*pb.RmqCampaignAcceptMission)
		module.Card.OnAcceptCampaignMission(mc.player, msg2.Cards)

	case pb.RmqType_UnifiedReward.ID():
		msg2 := msg.(*pb.RmqUnifiedReward)
		module.Campaign.OnUnifiedReward(mc.player, int(msg2.Rank), int(msg2.Contribution), msg2.YourMajestyName, msg2.CountryName)

	case pb.RmqType_ReturnDailyShareReward.ID():
		msg2 := msg.(*pb.WxDailyShareArg)
		module.WxGame.ReturnDailyShareReward(mc.player, msg2.PlayerName)

	case pb.RmqType_CompensateRecharge.ID():
		msg2 := msg.(*pb.RmqCompensateRecharge)
		mc.player.GetComponent(consts.ShopCpt).(types.IShopComponent).CompensateRecharge(msg2.CpOrderID,
			msg2.ChannelOrderID, msg2.GoodsID)

	default:
		glog.Errorf("mqConsumer err type = %s", type_)
	}
}
