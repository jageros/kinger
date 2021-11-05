package types

type IWxgameComponent interface {
	IPlayerComponent
	TriggerTreasureShareHD(treasureID string) int
	OnShareBeHelp(shareType int, shareTime int64, data []byte)
}
