package types

import "kinger/proto/pb"

type IRewardResult interface {
	GetRewardIdxs() []int
	GetResources() map[int]int
	GetConvertResources() map[int]int
	GetCards() map[uint32]int
	GetUpLevelRewardCards() map[uint32]int
	GetCardSkins() []string
	GetEquips() []string
	GetHeadFrames() []string
	GetEmojis() []int
	PackOpenTreasureMsg() *pb.OpenTreasureReply
	GetPrivileges() []*pb.GiftPrivilegeReward
}
