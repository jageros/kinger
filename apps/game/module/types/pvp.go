package types

import "kinger/proto/pb"

type IPvpComponent interface {
	IPlayerComponent
	GetPvpLevel() int
	//GetRewardGoldCnt() int
	//SetRewardGoldCnt(cnt int)
	GetMaxPvpLevel() int
	GetMaxPvpTeam() int
	GetPvpTeam() int
	UplevelReward() []uint32
	OnBattleEnd(fighterData *pb.EndFighterData, isWin, isWonderful bool, oppMMr, oppCamp, oppArea int)
	OnTrainingBattleEnd(fighterData *pb.EndFighterData, isWin bool)
	GetPvpFighterData() *pb.FighterData
}
