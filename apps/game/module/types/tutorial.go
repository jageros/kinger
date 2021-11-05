package types

import "kinger/proto/pb"

type ITutorialComponent interface {
	IPlayerComponent
	GetCampID() int32
	PackBeginBattleArg() *pb.BeginBattleArg
	//BeginGuideBattle() error
	OnBattleEnd(fighterData *pb.EndFighterData, isWin bool)
}
