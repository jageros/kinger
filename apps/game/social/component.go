package social

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"

	//"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/evq"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"strconv"
	"kinger/common/utils"
	"kinger/gopuppy/common/glog"
	"kinger/gamedata"
	"kinger/gopuppy/common/timer"
	"time"
)

var (
	_ types.ISocialComponent = &socialComponent{}
	_ types.IHeartBeatComponent = &socialComponent{}
)

type socialComponent struct {
	player                types.IPlayer
	attr                  *attribute.MapAttr
	friends               map[common.UUid]*friend
	privateChatAttr       *attribute.MapAttr
	needNotifyPrivateChat bool
	wxInviteFriendAttrMgr *attribute.AttrMgr
	wxInviteFriendAttr    *attribute.ListAttr
	wxInviteRewardAttr    *attribute.MapAttr
	wxInviteFriends       map[common.UUid]*attribute.MapAttr
	wxInviteRewards       map[int]*wxInviteReward
	lastSaveTime          time.Time

	advertProtecter iAdvertProtecter
}

func (sc *socialComponent) OnLogin(isRelogin, isRestore bool) {
	sc.lastSaveTime = time.Now()
	if isRelogin {
		return
	}
	if sc.player.IsWxgameAccount() {
		timer.AfterFunc(5 * time.Second, func() {
			sc.loadWxInviteFriend()

			for _, reward := range sc.wxInviteRewards {
				if reward.canReward() {
					sc.player.GetAgent().PushClient(pb.MessageID_S2C_WX_INVITE_SHOW_RED_DOT, nil)
					return
				}
			}
		})
	}
}

func (sc *socialComponent) OnLogout() {
	sc.needNotifyPrivateChat = false
	if sc.wxInviteFriendAttrMgr != nil {
		sc.wxInviteFriendAttrMgr.Save(false)
	}
}

func (sc *socialComponent) ComponentID() string {
	return consts.SocialCpt
}

func (sc *socialComponent) GetPlayer() types.IPlayer {
	return sc.player
}

func (sc *socialComponent) OnInit(player types.IPlayer) {
	sc.player = player
	sc.friends = make(map[common.UUid]*friend)
	friendsAttr := sc.attr.GetMapAttr("friends")
	if friendsAttr == nil {
		friendsAttr = attribute.NewMapAttr()
		sc.attr.SetMapAttr("friends", friendsAttr)
	}
	friendsAttr.ForEachKey(func(key string) {
		uid, _ := strconv.ParseUint(key, 10, 64)
		f := newFriendByAttr(common.UUid(uid), friendsAttr.GetMapAttr(key))
		sc.friends[common.UUid(uid)] = f
	})

	sc.privateChatAttr = sc.attr.GetMapAttr("privateChat")
	sc.advertProtecter = nil
}

func (sc *socialComponent) getAdvertProtecter() iAdvertProtecter {
	if sc.advertProtecter == nil {
		sc.advertProtecter = newAdvertProtecter(sc.player, sc.attr)
	}
	return sc.advertProtecter
}

func (sc *socialComponent) onMaxPvpLevelUpdate(maxPvpLevel int) {
	if maxPvpLevel == advertProtectMaxPvpLevel + 1 {
		sc.advertProtecter = nil
		module.OutStatus.DelStatus(sc.player, consts.OtAdvertProtecter)
	}
}

func (sc *socialComponent) OnHeartBeat() {
	if sc.wxInviteFriendAttrMgr != nil {
		now := time.Now()
		if now.Sub(sc.lastSaveTime) < 4 * time.Minute {
			return
		}
		sc.lastSaveTime = now
		sc.wxInviteFriendAttrMgr.Save(false)
	}
}

func (sc *socialComponent) OnReceivePrivateChat(fromUid common.UUid, fromName, fromHeadImgUrl string, fromPvpLevel int,
	msg string, chatTime int, fromCountry, fromHeadFrame, fromChatPop, fromCountryFlag string) {

	if sc.privateChatAttr == nil {
		sc.privateChatAttr = attribute.NewMapAttr()
		sc.attr.SetMapAttr("privateChat", sc.privateChatAttr)
	}
	maxID := sc.privateChatAttr.GetInt("maxID") + 1
	sc.privateChatAttr.SetInt("maxID", maxID)

	chatsAttr := sc.privateChatAttr.GetListAttr("chats")
	if chatsAttr == nil {
		chatsAttr = attribute.NewListAttr()
		sc.privateChatAttr.SetListAttr("chats", chatsAttr)
	}

	mattr := attribute.NewMapAttr()
	chatsAttr.AppendMapAttr(mattr)
	mattr.SetInt("id", maxID)
	mattr.SetUInt64("fromUid", uint64(fromUid))
	mattr.SetStr("fromName", fromName)
	mattr.SetStr("fromHeadImgUrl", fromHeadImgUrl)
	mattr.SetInt("time", chatTime)
	mattr.SetStr("msg", msg)
	mattr.SetInt("pvpLevel", fromPvpLevel)
	mattr.SetStr("fromCountry", fromCountry)
	mattr.SetStr("fromHeadFrame", fromHeadFrame)
	mattr.SetStr("fromChatPop", fromChatPop)
	mattr.SetStr("fromCountryFlag", fromCountryFlag)

	if sc.needNotifyPrivateChat {
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_PRIVATE_CHAT_NOTIFY, &pb.PrivateChatItem{
			ID:         int32(maxID),
			Uid:        uint64(fromUid),
			Name:       fromName,
			HeadImgUrl: fromHeadImgUrl,
			Msgs: []*pb.PrivateChatMsg{
				&pb.PrivateChatMsg{
					Time: int32(chatTime),
					Msg:  msg,
				},
			},
			PvpLevel: int32(fromPvpLevel),
			Country: fromCountry,
			HeadFrame: fromHeadFrame,
			ChatPop:fromChatPop,
			CountryFlag: fromCountryFlag,
		})
	}
}

func (sc *socialComponent) getAndDelPrivateChat(maxID int) []*pb.PrivateChatItem {
	sc.needNotifyPrivateChat = true
	var ret []*pb.PrivateChatItem
	if sc.privateChatAttr == nil {
		return ret
	}

	chatsAttr := sc.privateChatAttr.GetListAttr("chats")
	if chatsAttr == nil {
		return ret
	}

	chatsItem := map[common.UUid]*pb.PrivateChatItem{}
	chatsAttr.ForEachIndex(func(index int) bool {
		msgAttr := chatsAttr.GetMapAttr(index)
		uid := common.UUid(msgAttr.GetUInt64("fromUid"))
		id := msgAttr.GetInt("id")

		if id <= maxID {
			return true
		}

		item, ok := chatsItem[uid]
		if !ok {
			item = &pb.PrivateChatItem{
				Uid:        uint64(uid),
				Name:       msgAttr.GetStr("fromName"),
				HeadImgUrl: msgAttr.GetStr("fromHeadImgUrl"),
				PvpLevel: int32(msgAttr.GetInt("pvpLevel")),
				Country: msgAttr.GetStr("fromCountry"),
				HeadFrame: msgAttr.GetStr("fromHeadFrame"),
				ChatPop:msgAttr.GetStr("fromChatPop"),
				CountryFlag: msgAttr.GetStr("fromCountryFlag"),
			}
			chatsItem[uid] = item
		}

		if id > int(item.ID) {
			item.ID = int32(id)
		}

		item.Msgs = append(item.Msgs, &pb.PrivateChatMsg{
			Time: int32(msgAttr.GetInt("time")),
			Msg:  msgAttr.GetStr("msg"),
		})

		return true
	})

	sc.privateChatAttr.Del("chats")

	for _, item := range chatsItem {
		ret = append(ret, item)
	}
	return ret
}

func (sc *socialComponent) getFriendList() ([]*pb.FriendItem, *pb.FriendItem) {
	var fl []*pb.FriendItem
	var cs []chan *pb.FriendItem
	var lastOpp *pb.FriendItem
	lastOpponentUid := sc.player.GetLastOpponent()
	needLoadOpp := true
	var oppc chan *pb.SimplePlayerInfo
	for _, f := range sc.friends {
		if f.uid == lastOpponentUid {
			needLoadOpp = false
		}
		cs = append(cs, f.packMsgAsync())
	}

	if needLoadOpp && lastOpponentUid > 0 {
		oppc = module.Player.LoadSimplePlayerInfoAsync(lastOpponentUid)
	}

	evq.Await(func() {
		if oppc != nil {
			lastOppInfo := <-oppc
			if lastOppInfo != nil {
				lastOpp = &pb.FriendItem{
					Uid:            lastOppInfo.Uid,
					Name:           lastOppInfo.Name,
					PvpScore:       lastOppInfo.PvpScore,
					IsOnline:       lastOppInfo.IsOnline,
					HeadImgUrl:     lastOppInfo.HeadImgUrl,
					IsWechatFriend: lastOppInfo.IsWechatFriend,
					PvpCamp: lastOppInfo.PvpCamp,
					Country: lastOppInfo.Country,
					HeadFrame: lastOppInfo.HeadFrame,
					RebornCnt: lastOppInfo.RebornCnt,
					CountryFlag: lastOppInfo.CountryFlag,
					RankScore: lastOppInfo.RankScore,
				}

				if !lastOpp.IsOnline {
					lastOpp.LastOnlineTime = int32(lastOppInfo.LastOnlineTime)
				} else {
					lastOpp.IsInBattle = lastOppInfo.IsInBattle
				}
			}
		}

		for _, c := range cs {
			fmsg := <-c
			if fmsg != nil {
				fl = append(fl, fmsg)
			}
		}
	})

	return fl, lastOpp
}

func (sc *socialComponent) GetFriendsNum() int {
	fr, _ := sc.getFriendList()
	return len(fr)
}

func (sc *socialComponent) getFriend(uid common.UUid) *friend {
	if f, ok := sc.friends[uid]; ok {
		return f
	} else {
		return nil
	}
}

func (sc *socialComponent) AddFriendApply(fromUid common.UUid, fromName string, isInvite bool) {
	friendApplyAttr := sc.attr.GetListAttr("friendApply")
	if friendApplyAttr == nil {
		friendApplyAttr = attribute.NewListAttr()
		sc.attr.SetListAttr("friendApply", friendApplyAttr)
	}

	canAdd := sc.getFriend(fromUid) == nil
	if !canAdd {
		return
	}

	friendApplyAttr.ForEachIndex(func(index int) bool {
		uid := common.UUid(friendApplyAttr.GetUInt64(index))
		if uid == fromUid {
			canAdd = false
			return false
		}
		return true
	})

	if canAdd {
		friendApplyAttr.AppendUInt64(uint64(fromUid))
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_FRIEND_APPLY_NOTIFY, &pb.FriendApplyNotifyArg{
			FromName: fromName,
			IsInvite: isInvite,
		})
	}
}

func (sc *socialComponent) getFriendApplyList() []*pb.FriendApply {
	var fl []*pb.FriendApply
	friendApplyAttr := sc.attr.GetListAttr("friendApply")
	if friendApplyAttr == nil {
		return fl
	}

	friendApplyAttr.ForEachIndex(func(index int) bool {
		uid := common.UUid(friendApplyAttr.GetUInt64(index))
		player := module.Player.GetSimplePlayerInfo(uid)
		fl = append(fl, &pb.FriendApply{
			Uid:        uint64(player.GetUid()),
			Name:       player.GetName(),
			PvpScore:   int32(player.GetPvpScore()),
			HeadImgUrl: player.GetHeadImgUrl(),
			Country: player.GetCountry(),
			HeadFrame: player.GetHeadFrame(),
			CountryFlag: player.GetCountryFlag(),
		})
		return true
	})

	return fl
}

func (sc *socialComponent) replyFirendApply(uid common.UUid, isAgree bool) {
	friendApplyAttr := sc.attr.GetListAttr("friendApply")
	if friendApplyAttr == nil {
		return
	}

	var targetUid common.UUid
	friendApplyAttr.ForEachIndex(func(index int) bool {
		uid2 := common.UUid(friendApplyAttr.GetUInt64(index))
		if uid2 == uid {
			targetUid = uid2
			friendApplyAttr.DelByIndex(index)
			return false
		}
		return true
	})

	if targetUid > 0 {
		if isAgree {
			sc.AddFriend(targetUid)
		}

		targetPlayer := module.Player.GetPlayer(targetUid)
		if targetPlayer != nil {
			targetPlayer.GetAgent().PushClient(pb.MessageID_S2C_FRIEND_APPLY_RESULT, &pb.FriendApplyResult{
				Name:    sc.player.GetName(),
				IsAgree: isAgree,
			})

			if isAgree {
				socialCpt := targetPlayer.GetComponent(consts.SocialCpt).(*socialComponent)
				socialCpt.AddFriend(sc.player.GetUid())
			}
		} else {
			utils.PlayerMqPublish(targetUid, pb.RmqType_ReplyFriendApply, &pb.RmqReplyFriendApply{
				FromUid:  uint64(sc.player.GetUid()),
				FromName: sc.player.GetName(),
				IsAgree: isAgree,
			})
		}
	}
}

func (sc *socialComponent) AddFriend(uid common.UUid) {
	f := newFriend(uid, false)
	friendsAttr := sc.attr.GetMapAttr("friends")
	if friendsAttr == nil {
		friendsAttr = attribute.NewMapAttr()
		sc.attr.SetMapAttr("friends", friendsAttr)
	}
	friendsAttr.SetMapAttr(strconv.FormatUint(uint64(uid), 10), f.attr)
	sc.friends[uid] = f
	module.Mission.OnAddFriend(sc.player)
	eventhub.Publish(consts.EvAddFriend, sc.player)
}

func (sc *socialComponent) DelFriend(uid common.UUid) {
	delete(sc.friends, uid)
	friendsAttr := sc.attr.GetMapAttr("friends")
	if friendsAttr == nil {
		return
	}
	friendsAttr.Del(strconv.FormatUint(uint64(uid), 10))
}

func (sc *socialComponent) OnBattleEnd(fighterData *pb.EndFighterData, isWin bool) {
	msg := &pb.BattleResult{}
	if isWin {
		msg.WinUid = uint64(sc.player.GetUid())
	} else {
		msg.WinUid = 1
	}
	sc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, msg)
}

func (sc *socialComponent) setWxInviter(uid common.UUid) {
	sc.attr.SetUInt64("wxInviter", uint64(uid))
}

func (sc *socialComponent) getWxInviter() common.UUid {
	return common.UUid(sc.attr.GetUInt64("wxInviter"))
}

func (sc *socialComponent) loadWxInviteFriend() {
	if sc.wxInviteFriendAttrMgr != nil {
		return
	}

	wxInviteFriendAttrMgr := attribute.NewAttrMgr("wxInviteFriend", sc.player.GetUid())
	err := wxInviteFriendAttrMgr.Load()
	if err != nil && err != attribute.NotExistsErr {
		glog.Errorf("loadWxInviteFriend uid=%d, err=%s", sc.player.GetUid(), err)
		return
	}

	if sc.wxInviteFriendAttrMgr != nil {
		return
	}

	sc.wxInviteFriendAttrMgr = wxInviteFriendAttrMgr
	wxInviteFriendAttr := wxInviteFriendAttrMgr.GetListAttr("wxInviteFriend")
	if wxInviteFriendAttr == nil {
		wxInviteFriendAttr = attribute.NewListAttr()
		wxInviteFriendAttrMgr.SetListAttr("wxInviteFriend", wxInviteFriendAttr)
	}
	wxInviteRewardAttr := wxInviteFriendAttrMgr.GetMapAttr("wxInviteReward")
	if wxInviteRewardAttr == nil {
		wxInviteRewardAttr = attribute.NewMapAttr()
		wxInviteFriendAttrMgr.SetMapAttr("wxInviteReward", wxInviteRewardAttr)
	}

	sc.wxInviteFriendAttr = wxInviteFriendAttr
	sc.wxInviteRewardAttr = wxInviteRewardAttr

	sc.wxInviteFriends = map[common.UUid]*attribute.MapAttr{}
	wxInviteFriendAttr.ForEachIndex(func(index int) bool {
		attr := wxInviteFriendAttr.GetMapAttr(index)
		uid := common.UUid(attr.GetUInt64("uid"))
		sc.wxInviteFriends[uid] = attr
		return true
	})

	sc.wxInviteRewards = map[int]*wxInviteReward{}
	sc.wxInviteRewardAttr.ForEachKey(func(key string) {
		id, _ := strconv.Atoi(key)
		sc.wxInviteRewards[id] = newWxInviteRewardByAttr(id, sc.wxInviteRewardAttr.GetMapAttr(key))
	})
}

func (sc *socialComponent) OnWxInviteFriendUpdate(uid common.UUid, headImgUrl string, maxPvpLevel int) {
	sc.loadWxInviteFriend()
	friendAttr, ok := sc.wxInviteFriends[uid]
	isPvpLevelUpdate := false
	var oldMaxPvpLevel int
	if !ok {
		isPvpLevelUpdate = true
		oldMaxPvpLevel = 0
		friendAttr = attribute.NewMapAttr()
		friendAttr.SetUInt64("uid", uint64(uid))
		friendAttr.SetStr("headImgUrl", headImgUrl)
		friendAttr.SetInt("maxPvpLevel", maxPvpLevel)
		sc.wxInviteFriends[uid] = friendAttr
		sc.wxInviteFriendAttr.AppendMapAttr(friendAttr)
	} else {
		friendAttr.SetStr("headImgUrl", headImgUrl)
		oldMaxPvpLevel = friendAttr.GetInt("maxPvpLevel")
		if maxPvpLevel > oldMaxPvpLevel {
			friendAttr.SetInt("maxPvpLevel", maxPvpLevel)
			isPvpLevelUpdate = true
		}
	}

	if !isPvpLevelUpdate {
		return
	}

	canReward := false
	gdata := gamedata.GetGameData(consts.WxInviteReward).(*gamedata.WxInviteRewardGameData)
	for pvpLevel := oldMaxPvpLevel + 1; pvpLevel <= maxPvpLevel; pvpLevel++ {

		rewardDatas, ok := gdata.Level2Rewards[pvpLevel]
		if !ok {
			continue
		}

		for _, rewardData := range rewardDatas {
			reward := sc.wxInviteRewards[rewardData.ID]
			if reward == nil {
				rewardAttr := attribute.NewMapAttr()
				sc.wxInviteRewardAttr.SetMapAttr(strconv.Itoa(rewardData.ID), rewardAttr)
				reward = newWxInviteRewardByAttr(rewardData.ID, rewardAttr)
				sc.wxInviteRewards[rewardData.ID] = reward
			}

			oldCanReward := reward.canReward()
			reward.incrAmount()
			if !oldCanReward && reward.canReward() {
				canReward = true
			}
		}

	}

	if canReward {
		sc.player.GetAgent().PushClient(pb.MessageID_S2C_WX_INVITE_SHOW_RED_DOT, nil)
	}
}

func (sc *socialComponent) packWxInviteFriendMsg() *pb.WxInviteFriendsReply {
	sc.loadWxInviteFriend()
	reply := &pb.WxInviteFriendsReply{}
	sc.wxInviteFriendAttr.ForEachIndex(func(index int) bool {
		friendAttr := sc.wxInviteFriendAttr.GetMapAttr(index)
		reply.Friends = append(reply.Friends, &pb.WxInviteFriend{
			HeadImgUrl: friendAttr.GetStr("headImgUrl"),
			PvpLevel: int32(friendAttr.GetInt("maxPvpLevel")),
		})
		return true
	})

	for _, reward := range sc.wxInviteRewards {
		reply.Rewards = append(reply.Rewards, reward.packMsg())
	}

	return reply
}

func (sc *socialComponent) getWxInviteFriendReward(rewardID int) (*pb.GetWxInviteRewardReply, error) {
	sc.loadWxInviteFriend()
	reward, ok := sc.wxInviteRewards[rewardID]
	if !ok {
		return nil, gamedata.GameError(1)
	}
	return reward.doReward(sc.player)
}
