package main

import (
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"container/list"
	"kinger/gopuppy/apps/logic"
	"time"
	"kinger/gopuppy/apps/center/api"
	gpb "kinger/gopuppy/proto/pb"
)

type chatChannel struct {
	channel pb.ChatChannel
	area int                      // 区
	campaignCryID uint32          // 国战国家id
	subscribes common.UInt64Set   // 哪些玩家关注了
	chatLets *list.List           // 当前频道聊天记录
	key string
}

func newChatChannel(channel pb.ChatChannel, area int, campaignCryID uint32) *chatChannel {
	return &chatChannel{
		channel: channel,
		area: area,
		campaignCryID: campaignCryID,
		subscribes: common.UInt64Set{},
		chatLets: list.New(),
	}
}

func (cc *chatChannel) getChannel() pb.ChatChannel {
	return cc.channel
}

func (cc *chatChannel) getCampaignCryID() uint32 {
	return cc.campaignCryID
}

func (cc *chatChannel) getArea() int {
	return cc.area
}

func (cc *chatChannel) getKey() string {
	if cc.key != "" {
		return cc.key
	}

	cc.key = chatMgr.genChannelKey(cc.channel, cc.area, cc.campaignCryID)
	return cc.key
}

func (cc *chatChannel) subscribe(uid common.UUid) *pb.ChatItemList {
	cc.subscribes.Add(uint64(uid))

	agent := logic.GetPlayerAgent(uid)
	if agent != nil {
		key := cc.getKey()
		agent.SetClientFilter(key, "1")
		if blackUidList.Contains(uint64(uid)) {
			agent.SetClientFilter(chatMgr.genBlackChannelKey(key), "1")
		}
	}

	l := &pb.ChatItemList{}
	for chatElem := cc.chatLets.Front(); chatElem != nil; chatElem = chatElem.Next() {
		l.ChatItems = append(l.ChatItems, chatElem.Value.(*pb.ChatItem))
	}
	return l
}

func (cc *chatChannel) unsubscribe(uid common.UUid, isLogout bool) {
	cc.subscribes.Remove(uint64(uid))
	agent := logic.GetPlayerAgent(uid)
	if agent != nil && !isLogout {
		key := cc.getKey()
		agent.SetClientFilter(key, "0")
		if blackUidList.Contains(uint64(uid)) {
			agent.SetClientFilter(chatMgr.genBlackChannelKey(key), "0")
		}
	}
}

func (cc *chatChannel) broadcast(senderUid common.UUid, sender *pb.CSendChatArg, msg string) {
	chat := &pb.ChatItem{
		Uid:        uint64(senderUid),
		Name:       sender.Name,
		HeadImgUrl: sender.HeadImgUrl,
		Time:       int32(time.Now().Unix()),
		Msg:        msg,
		PvpLevel: sender.PvpLevel,
		Country: sender.Country,
		HeadFrame: sender.HeadFrame,
		CityID: sender.CityID,
		CityJob: sender.CityJob,
		CountryJob: sender.CountryJob,
		ChatPop:sender.ChatPop,
		CountryFlag: sender.CountryFlag,
	}

	if !blackUidList.Contains(uint64(senderUid)) {
		cc.chatLets.PushBack(chat)
		if cc.chatLets.Len() > 40 {
			cc.chatLets.Remove(cc.chatLets.Front())
		}
	}

	clet := &pb.Chatlet{Type: pb.Chatlet_Normal}
	clet.Data, _ = chat.Marshal()
	chatMsg := &pb.ChatNotify{
		Channel: cc.channel,
		Chat:    clet,
	}

	key := cc.getKey()
	if !blackUidList.Contains(uint64(senderUid)) {
		api.BroadcastClient(pb.MessageID_S2C_CHAT_NOTIFY, chatMsg, &gpb.BroadcastClientFilter{
			OP:  gpb.BroadcastClientFilter_EQ,
			Key: key,
			Val: "1",
		})
	} else {

		api.BroadcastClient(pb.MessageID_S2C_CHAT_NOTIFY, chatMsg, &gpb.BroadcastClientFilter{
			OP:  gpb.BroadcastClientFilter_EQ,
			Key: chatMgr.genBlackChannelKey(key),
			Val: "1",
		})
	}

}
