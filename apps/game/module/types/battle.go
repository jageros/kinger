package types

import (
	"kinger/gopuppy/common"
)

type IFighterData interface {
	GetPlayer() IPlayer
	GetSit() int
	IsRobot() bool
	GetCamp() int
	GetHandCards() []IFightCardData
	GetGridCards(isFirstHand bool, gridIDs []int) map[int]IFightCardData // map[gridID]card
	//GetKingForbidCards() []uint32         // []cardID
	GetCasterSkills() []int32 // []skillID
	//GetMaxHandCardAmount() int
}

type IEndFighterData interface {
	GetPlayer() IPlayer
	IsRobot() bool
	IsSurrender() bool
	GetCamp() int
	GetFightCards() []IFightCardData
	GetHandCards() []IFightCardData
	GetInitHandCards() []IFightCardData
	GetLoseCards() []IFightCardData
	IsFighter1() bool
	IsFirstHand() bool
}

type IBattleEndHandler interface {
	HandleBattleEnd(winer IEndFighterData, loser IEndFighterData, videoID common.UUid)
}

type IBattle interface {
	PackMsg() interface{}
	ReadyDone(choiceCardIDs ...uint32) error
}
