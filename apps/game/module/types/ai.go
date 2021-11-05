package types

import (
	"fmt"
)

type BattleAction struct {
	cardObjID int
	gridID    int
}

func (ba *BattleAction) String() string {
	return fmt.Sprintf("[CardObjID=%d, GridID=%d]", ba.cardObjID, ba.gridID)
}

func (ba *BattleAction) GetCardObjID() int {
	return ba.cardObjID
}

func (ba *BattleAction) GetGridID() int {
	return ba.gridID
}

func NewBattleAction(cardObjID, gridID int) *BattleAction {
	return &BattleAction{
		cardObjID: cardObjID,
		gridID:    gridID,
	}
}

type IBattleSituation interface {
	Copy() IBattleSituation
	GenAllAction() []*BattleAction
	BoutBegin()
	BoutEnd()
	IsEnd() bool
	PlayAction(*BattleAction) (bool, bool)
	Evaluate() float32
}

type IRobotPlayer interface {
	IPlayer
	SetName(name string)
	SetPvpScore(val int)
}
