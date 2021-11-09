package types

import (
	"kinger/gamedata"
	"kinger/proto/pb"
)

type ICardComponent interface {
	IPlayerComponent
	GetCollectCard(cardID uint32) ICollectCard
	GetPvpCardPoolByCamp(camp int) []ICollectCard
	GetAllCollectCards() []ICollectCard
	GetAllCollectCardDatas() []*gamedata.Card
	PackDiyCardMsg() []*pb.DiyCardData
	//统计指定等级卡的数量
	GetCollectCardNumByLevel(lvl int) int
	//统计指定星的卡
	GetCollectCardNumByStar(star int) int

	// 批量修改卡
	//@param  cardsChange map[cardID]*pb.CardInfo  要修改的卡
	//@return  卡的改变
	ModifyCollectCards(cardsChange map[uint32]*pb.CardInfo) []*pb.ChangeCardInfo
	NewbieInitPvpCardPool(camp int, cardIDs []uint32)
	CreatePvpHandCards(camp int) []*pb.SkinGCard // []gcardID
	GetCollectCardLevelInfo() (maxLevel, minLevel, avlLevel int)
	OnPvpBattleEnd(fightCards []uint32)
	GetLastFightCards() []*pb.SkinGCard
	GetFavoriteCards() []*pb.SkinGCard
	GetFightCamp() int
}

type ICollectCard interface {
	IRestAble
	GetCardGameData() *gamedata.Card
	IsDead() bool
	GetCardID() uint32
	GetAmount() int
	GetEnergy() float32
	GetLevel() int
	GetSkin() string
	SetAmount(amount int)
	IsMaxLevel() bool
	GetEquip() string
	WearEquip(equipID string)
	DeEquip()
	GetState() pb.CardState
	PackMsg() *pb.CardInfo
	IsMaxCanUpLevel() bool
	GetMaxUnlockLevel() int
	IsSpCard() bool
}

type IFightCardData interface {
	GetName() string
	GetGCardID() uint32
	GetCardID() uint32
	RandomUp() int
	RandomDown() int
	RandomLeft() int
	RandomRight() int
	GetUpValueRate() float32
	GetDownValueRate() float32
	GetLeftValueRate() float32
	GetRightValueRate() float32
	GetAdjFValue() float32
	GetCardValue() float32
	GetCamp() int
	GetSkillIds() []int32
	PackDiyFightCardInfo() *pb.DiyFightCardInfo
	GetLevel() int
	GetCardType() int
}
