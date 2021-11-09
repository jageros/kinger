package types

import "kinger/proto/pb"

type IShopComponent interface {
	IPlayerComponent
	WatchShopFreeAds(type_ pb.ShopFreeAdsType, id int, isConsumeJade bool) (*pb.WatchShopFreeAdsReply, error)
	OnShopAdsBeHelp(type_ pb.ShopFreeAdsType, id int) error
	OnSdkRecharge(channelUid, cpOrderID, channelOrderID string, paymentAmount int, needCheckMoney bool)
	GetCumulativePay() int
	CompensateRecharge(cpOrderID, channelOrderID, goodsID string)
}
