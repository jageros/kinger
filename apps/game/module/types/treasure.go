package types

import "kinger/proto/pb"

type ITreasureComponent interface {
	IPlayerComponent
	AddRewardTreasure(ignoreLimit bool, canTriggerUpRare bool) (modelID string, reason pb.NoTreasureReasonEnum,
		upRareTreasureModelID string)
	AddRewardTreasureByID(modelID string, canTriggerUpRare bool) (ok bool, upRareTreasureModelID string)
	AddDailyTreasure(isReset bool) bool
	AddDailyTreasureByID(modelID string, dayIdx int) bool
	AddDailyTreasureStar(star int)
	OpenTreasureByModelID(modelID string, isDouble bool, ignoreRewardTbl ...bool) *pb.OpenTreasureReply
	IsDailyTreasure(treasureID uint32) bool
	HelpOpenTreasure(treasureID uint32)
	DailyTreasureBeDobule(isConsumeJade bool) bool
	AccTreasure(treasureID uint32, isConsumeJade bool) (int32, error)
	UpTreasureRare(isConsumeJade bool) (*pb.Treasure, error)
	WatchTreasureAddCardAds(treasureID uint32, isConsumeJade bool) (*pb.WatchTreasureAddCardAdsReply, error)
	CancelTreasureAddCardAds()
}
