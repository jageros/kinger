package campaign

import (
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
)

type campaignComponent struct {
	attr        *attribute.MapAttr
	player      types.IPlayer
	info        *pb.GCampaignPlayerInfo
	offlineInfo *pb.GCampaignPlayerInfo
}

func (cc *campaignComponent) ComponentID() string {
	return consts.CampaignCpt
}

func (cc *campaignComponent) GetPlayer() types.IPlayer {
	return cc.player
}

func (cc *campaignComponent) OnInit(player types.IPlayer) {
	cc.player = player

}

func (cc *campaignComponent) OnLogin(isRelogin, isRestore bool) {

}

func (cc *campaignComponent) OnLogout() {

}

func (cc *campaignComponent) loadInfo() *pb.GCampaignPlayerInfo {
	if config.GetConfig().IsMultiLan {
		return nil
	}

	var info *pb.GCampaignPlayerInfo
	if cc.info == nil {

		agent := cc.player.GetAgent()
		if agent != nil {
			reply, err := agent.CallBackend(pb.MessageID_G2CA_FETCH_CAMPAIGN_PLAYER_INFO, nil)
			if err == nil {
				cc.info = reply.(*pb.GCampaignPlayerInfo)
			}
			info = cc.info
		} else {
			reply, err := logic.CallBackend("", 0, pb.MessageID_G2CA_FETCH_CAMPAIGN_TARGET_PLAYER_INFO,
				&pb.TargetPlayer{
					Uid: uint64(cc.player.GetUid()),
				})
			if err == nil {
				cc.offlineInfo = reply.(*pb.GCampaignPlayerInfo)
			}
			info = cc.offlineInfo
		}

	} else {
		return cc.info
	}
	return info
}

func (cc *campaignComponent) getInfo() *pb.GCampaignPlayerInfo {
	if cc.info != nil {
		return cc.info
	} else if cc.offlineInfo != nil {
		return cc.offlineInfo
	} else {
		return cc.loadInfo()
	}
}

func (cc *campaignComponent) setInfo(info *pb.GCampaignPlayerInfo) {
	if cc.info != nil {
		countryID := cc.info.CountryID
		cc.info.CountryID = info.CountryID
		cc.info.CityID = info.CityID
		cc.info.CityJob = info.CityJob
		cc.info.CountryJob = info.CountryJob
		cc.info.CountryName = info.CountryName

		if countryID != cc.info.CountryID {
			cc.player.GetAgent().PushBackend(pb.MessageID_C2S_UNSUBSCRIBE_CHAT, &pb.TargetChatChannel{
				Channel: pb.ChatChannel_CampaignCountry,
			})

			cc.player.GetAgent().PushClient(pb.MessageID_S2C_ON_UNSUBSCRIBE_CHAT, &pb.TargetChatChannel{
				Channel: pb.ChatChannel_CampaignCountry,
			})
		}
	}
}

func (cc *campaignComponent) addNotice(notice *pb.CampaignNotice) {
	if cc.info != nil {
		cc.info.Notices = append(cc.info.Notices, notice)
		agent := cc.player.GetAgent()
		if agent == nil {
			return
		}

		clet := &pb.Chatlet{Type: pb.Chatlet_CampaignNotice}
		clet.Data, _ = notice.Marshal()
		agent.PushClient(pb.MessageID_S2C_CHAT_NOTIFY, &pb.ChatNotify{
			Channel: pb.ChatChannel_CampaignCountry,
			Chat:    clet,
		})
	}
}

func (cc *campaignComponent) modifyContribution(amount int) bool {
	_, err := logic.CallBackend("", 0, pb.MessageID_G2CA_MODIFY_CONTRIBUTION, &pb.ModifyContributionArg{
		Uid:    uint64(cc.player.GetUid()),
		Amount: int32(amount),
	})
	return err == nil
}
