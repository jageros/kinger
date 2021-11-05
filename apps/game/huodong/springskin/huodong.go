package springskin

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/huodong/event"
	"github.com/gogo/protobuf/proto"
	"kinger/proto/pb"
)

var _ htypes.IEventHuodong = &springSkinHd{}

type springSkinHd struct {
	event.EventHd
}

func (hd *springSkinHd) NewPlayerData(player types.IPlayer) htypes.IHdPlayerData {
	attr := attribute.NewMapAttr()
	return hd.NewPlayerDataByAttr(player, attr)
}

func (hd *springSkinHd) NewPlayerDataByAttr(player types.IPlayer, attr *attribute.MapAttr) htypes.IHdPlayerData {
	hpd := &springSkinHdPlayerData{}
	hpd.Player = player
	hpd.Attr = attr
	return hpd
}

func (hd *springSkinHd) PackEventDetailMsg(data event.IEventHdPlayerData) proto.Marshaler {
	data2, ok := data.(*springSkinHdPlayerData)
	if !ok {
		return nil
	}

	msg := &pb.SpringSkinHuodong{}
	data2.forEachSkin(func(skin string) bool {
		msg.SkinIDs = append(msg.SkinIDs, skin)
		return true
	})
	return msg
}
