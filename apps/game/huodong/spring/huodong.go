package spring

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/huodong/event"
	"github.com/gogo/protobuf/proto"
	"kinger/proto/pb"
)

var _ htypes.IEventHuodong = &springHd{}

type springHd struct {
	event.EventHd
}

func (hd *springHd) NewPlayerData(player types.IPlayer) htypes.IHdPlayerData {
	attr := attribute.NewMapAttr()
	return hd.NewPlayerDataByAttr(player, attr)
}

func (hd *springHd) NewPlayerDataByAttr(player types.IPlayer, attr *attribute.MapAttr) htypes.IHdPlayerData {
	hpd := &springHdPlayerData{}
	hpd.Player = player
	hpd.Attr = attr
	hpd.exchangeCntAttr = attr.GetMapAttr("exchangeCnt")
	return hpd
}

func (hd *springHd) PackEventDetailMsg(data event.IEventHdPlayerData) proto.Marshaler {
	playerHdData, ok := data.(*springHdPlayerData)
	if !ok {
		return nil
	}

	detail := &pb.SpringHuodong{}
	playerHdData.forEachExchangeCnt(func(goodsID, cnt int) {
		detail.ExchangeDatas = append(detail.ExchangeDatas, &pb.SpringExchangeData{
			GoodsID: int32(goodsID),
			ExchangeCnt: int32(cnt),
		})
	})
	return detail
}
