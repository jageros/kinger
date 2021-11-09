package social

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
)

type friend struct {
	uid  common.UUid
	attr *attribute.MapAttr
}

func newFriendByAttr(uid common.UUid, attr *attribute.MapAttr) *friend {
	return &friend{
		uid:  uid,
		attr: attr,
	}
}

func newFriend(uid common.UUid, isWechatFriend bool) *friend {
	attr := attribute.NewMapAttr()
	if isWechatFriend {
		attr.SetBool("isWechatFriend", isWechatFriend)
	}
	return &friend{
		uid:  uid,
		attr: attr,
	}
}

func (f *friend) packMsgAsync() chan *pb.FriendItem {
	c := make(chan *pb.FriendItem, 1)
	player := module.Player.GetPlayer(f.uid)
	if player == nil {
		player = module.Player.GetCachePlayer(f.uid)
	}

	if player == nil {
		c2 := module.Player.LoadSimplePlayerInfoAsync(f.uid)
		go func() {
			var msg *pb.FriendItem
			utils.CatchPanic(func() {
				playerInfo := <-c2
				if playerInfo == nil {
					return
				}

				msg = &pb.FriendItem{
					Uid:            uint64(f.uid),
					Name:           playerInfo.Name,
					PvpScore:       playerInfo.PvpScore,
					IsOnline:       playerInfo.IsOnline,
					HeadImgUrl:     playerInfo.HeadImgUrl,
					IsWechatFriend: f.isWechatFriend(),
					PvpCamp:        playerInfo.PvpCamp,
					Country:        playerInfo.Country,
					HeadFrame:      playerInfo.HeadFrame,
					RebornCnt:      playerInfo.RebornCnt,
					CountryFlag:    playerInfo.CountryFlag,
					RankScore:      playerInfo.RankScore,
				}

				if !msg.IsOnline {
					msg.LastOnlineTime = int32(playerInfo.LastOnlineTime)
				} else {
					msg.IsInBattle = playerInfo.IsInBattle
				}
			})

			c <- msg
		}()
		return c
	}

	msg := &pb.FriendItem{
		Uid:            uint64(f.uid),
		Name:           player.GetName(),
		PvpScore:       int32(player.GetPvpScore()),
		IsOnline:       player.IsOnline(),
		HeadImgUrl:     player.GetHeadImgUrl(),
		IsWechatFriend: f.isWechatFriend(),
		PvpCamp:        int32(player.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()),
		Country:        player.GetCountry(),
		HeadFrame:      player.GetHeadFrame(),
		RebornCnt:      int32(module.Reborn.GetRebornCnt(player)),
		CountryFlag:    player.GetCountryFlag(),
		RankScore:      int32(player.GetRankScore()),
	}

	if !msg.IsOnline {
		msg.LastOnlineTime = int32(player.GetLastOnlineTime())
	} else {
		msg.IsInBattle = player.IsInBattle()
	}

	c <- msg
	return c
}

func (f *friend) isWechatFriend() bool {
	return false
}
