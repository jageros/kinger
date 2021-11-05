package types

import (
	"kinger/proto/pb"
	"kinger/gopuppy/common"
)

type ILevelComponent interface {
	IPlayerComponent
	GetUnlockCards() []uint32
	OnBattleEnd(fighterData *pb.EndFighterData, isWin bool, levelID int, battleID common.UUid)
	ClearLevel()
	OnBeHelpBattle(helperUid common.UUid, helperName string, levelID int, battleID common.UUid)
	OnHelpBattleEnd(isWin bool, battleID common.UUid)
	GetCurLevel() int
	IsClearLevel(levelID int) bool
}
