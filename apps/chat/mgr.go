package main

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"fmt"
	"kinger/common/config"
)

var chatMgr *chatMgrSt

type chatMgrSt struct {
	uid2Channels map[common.UUid][]*chatChannel
	channels map[string]*chatChannel   // map[channelKey]*chatChannel
}

func newChatMgr() {
	chatMgr = &chatMgrSt{
		uid2Channels: map[common.UUid][]*chatChannel{},
		channels: map[string]*chatChannel{},
	}
}

func (cm *chatMgrSt) genBlackChannelKey(channelKey string) string {
	return channelKey + "_black"
}

func (cm *chatMgrSt) genChannelKey(channel pb.ChatChannel, area int, campaignCryID uint32) string {
	if channel == pb.ChatChannel_CampaignCountry {
		return fmt.Sprintf("%s%d", pb.ChatChannel_CampaignCountry, campaignCryID)
	} else if area == 1 && config.GetConfig().IsOldXfServer() {
		return fmt.Sprintf("%s%d", channel, area)
	} else {
		return channel.String()
	}
}

func (cm *chatMgrSt) addForbidChatPlayer(uid uint64) {
	if blackUidList.Contains(uid) {
		return
	}

	blackUidList.Add(uid)
	if cService.AppID == 1 {
		attr := attribute.NewAttrMgr("forbid_chat", int64(uid))
		attr.SetDirty(true)
		attr.Save(false)
	}

	agent := logic.GetPlayerAgent(common.UUid(uid))
	if agent == nil {
		return
	}

	for _, cc := range cm.uid2Channels[common.UUid(uid)] {
		agent.SetClientFilter(cm.genBlackChannelKey(cc.getKey()), "1")
	}
}

func (cm *chatMgrSt) delForbidChatPlayer(uid uint64) {
	blackUidList.Remove(uid)
	if cService.AppID == 1 {
		attribute.NewAttrMgr("forbid_chat", int64(uid)).Delete(false)
	}

	agent := logic.GetPlayerAgent(common.UUid(uid))
	if agent == nil {
		return
	}

	for _, cc := range cm.uid2Channels[common.UUid(uid)] {
		agent.SetClientFilter(cm.genBlackChannelKey(cc.getKey()), "0")
	}
}

func (cm *chatMgrSt) subscribeChat(channel pb.ChatChannel, uid common.UUid, area int, campaignCryID uint32) *pb.ChatItemList {
	if channel == pb.ChatChannel_CampaignCountry && campaignCryID <= 0 {
		return nil
	}
	channelKey := cm.genChannelKey(channel, area, campaignCryID)

	cc := cm.getPlayerSubscribeChannel(uid, channel)
	if cc != nil {
		return cc.subscribe(uid)
	}

	cc = cm.channels[channelKey]
	if cc == nil {
		cc = newChatChannel(channel, area, campaignCryID)
		cm.channels[channelKey] = cc
	}

	cm.uid2Channels[uid] = append(cm.uid2Channels[uid], cc)
	return cc.subscribe(uid)
}

func (cm *chatMgrSt) unsubscribeChat(channel pb.ChatChannel, uid common.UUid, isLogout bool) {
	channels := cm.uid2Channels[uid]
	for i, cc := range channels {
		if cc.getChannel() == channel {
			cc.unsubscribe(uid, isLogout)
			channels = append(channels[:i], channels[i+1:]...)
			if len(channels) <= 0 {
				delete(cm.uid2Channels, uid)
			} else {
				cm.uid2Channels[uid] = channels
			}
			return
		}
	}
}

func (cm *chatMgrSt) getPlayerSubscribeChannel(uid common.UUid, channel pb.ChatChannel) *chatChannel {
	channels := cm.uid2Channels[uid]
	for _, cc := range channels {
		if cc.getChannel() == channel {
			return cc
		}
	}
	return nil
}

func (cm *chatMgrSt) broadcastChat(senderUid common.UUid, sender *pb.CSendChatArg, channel pb.ChatChannel, msg string) {
	cc := cm.getPlayerSubscribeChannel(senderUid, channel)
	if cc == nil {
		return
	}

	if channel == pb.ChatChannel_CampaignCountry {
		if cc.getCampaignCryID() != sender.CountryID {
			cm.unsubscribeChat(channel, senderUid, true)
			cm.subscribeChat(channel, senderUid, cc.getArea(), sender.CountryID)

			cc = cm.getPlayerSubscribeChannel(senderUid, channel)
			if cc == nil {
				return
			}
		}
	}

	cc.broadcast(senderUid, sender, msg)
}
