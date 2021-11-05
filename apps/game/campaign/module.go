package campaign

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/apps/game/module"
	"kinger/proto/pb"
	"kinger/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"time"
	"kinger/common/config"
)

const limitPvpLevel = 11

var mod *campaignModule

type campaignModule struct {
	info *pb.GCampaignInfo
}

func (m *campaignModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("campaign")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("campaign", attr)
	}
	return &campaignComponent{attr: attr}
}

func (m *campaignModule) OnBattleEnd(player types.IPlayer, isWin bool) {
	agent := player.GetAgent()
	if agent == nil {
		return
	}
	agent.PushBackend(pb.MessageID_G2CA_CAMPAIGN_ON_BATTLE_END, &pb.CampaignBattleEnd{
		IsWin: isWin,
	})

	var winUid uint64 = 1
	if isWin {
		winUid = uint64(player.GetUid())
	}
	agent.PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
		WinUid: winUid,
	})
}

func (m *campaignModule) GetPlayerInfo(player types.IPlayer) *pb.GCampaignPlayerInfo {
	return player.GetComponent(consts.CampaignCpt).(*campaignComponent).loadInfo()
}

func (m *campaignModule) GetPlayerOfflineInfo(player types.IPlayer) *pb.GCampaignPlayerInfo {
	return player.GetComponent(consts.CampaignCpt).(*campaignComponent).getInfo()
}

func (m *campaignModule) updateCampaignInfo(info *pb.GCampaignInfo)  {
	m.info = info
	glog.Infof("updateCampaignInfo state=%s", pb.CampaignState_StateEnum(info.CampaignState))
}

func (m *campaignModule) getCampaignInfo() *pb.GCampaignInfo {
	if config.GetConfig().IsMultiLan {
		return nil
	}
	
	if m.info == nil {
		reply, err := logic.CallBackend("", 0, pb.MessageID_G2CA_GET_CAMPAIGN_INFO, nil)
		if err != nil {
			glog.Errorf("getCampaignInfo err %s", err)
		} else {
			m.info = reply.(*pb.GCampaignInfo)
		}
	}
	return m.info
}

func (m *campaignModule) IsInCampaignMatch(player types.IPlayer) bool {
	info := m.getCampaignInfo()
	if info == nil {
		return false
	}

	state := pb.CampaignState_StateEnum(info.CampaignState)
	if state != pb.CampaignState_InWar {
		return false
	}

	if player.GetMaxPvpLevel() < limitPvpLevel {
		return false
	}

	_, err := logic.CallBackend("", 0, pb.MessageID_G2CA_IS_IN_CAMPAIGN_MATCH, &pb.TargetPlayer{
		Uid: uint64(player.GetUid()),
	})
	if err != nil {
		return false
	} else {
		return true
	}
}

func (m *campaignModule) IsInWar() bool {
	info := m.getCampaignInfo()
	if info == nil {
		return false
	}
	state := pb.CampaignState_StateEnum(info.CampaignState)
	return state == pb.CampaignState_InWar || state == pb.CampaignState_ReadyWar
}

func (m *campaignModule) ModifyContribution(player types.IPlayer, amount int) bool {
	return player.GetComponent(consts.CampaignCpt).(*campaignComponent).modifyContribution(amount)
}

func (m *campaignModule) OnUnifiedReward(player types.IPlayer, rank, contribution int, yourMajestyName, countryName string) {
	sender := module.Mail.NewMailSender(player.GetUid())
	sender.SetTypeAndArgs(pb.MailTypeEnum_CampaignUnified, yourMajestyName, countryName)
	mailReward := sender.GetRewardObj()
	mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, "10", 1)
	mailReward.AddAmountByType(pb.MailRewardType_MrtContribution, contribution)
	sender.Send()
}

func Initialize() {
	initializeGoods()
	mod = &campaignModule{}
	module.Campaign = mod
	registerRpc()

	timer.AfterFunc(2 * time.Second, func() {
		mod.getCampaignInfo()
	})
}
