package types

import (
	"kinger/gopuppy/common"
	"kinger/proto/pb"
)

type ISocialComponent interface {
	IPlayerComponent
	OnBattleEnd(fighterData *pb.EndFighterData, isWin bool)
	AddFriendApply(fromUid common.UUid, fromName string, isInvite bool)
	AddFriend(uid common.UUid)
	DelFriend(uid common.UUid)
	OnReceivePrivateChat(fromUid common.UUid, fromName, fromHeadImgUrl string, fromPvpLevel int, msg string,
		chatTime int, fromCountry, fromHeadFrame, fromChatPop, fromCountryFlag string)
	OnWxInviteFriendUpdate(uid common.UUid, headImgUrl string, maxPvpLevel int)
	GetFriendsNum() int
}
