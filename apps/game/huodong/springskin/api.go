package springskin

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/huodong/event"
	"kinger/proto/pb"
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"math/rand"
)

func Initialize() {
	initSkins()
	registerRpc()
}

func NewSpringSkinHdByAttr(area int, attr *attribute.AttrMgr) htypes.IHuodong {
	return event.NewEventHdByAttr(area, pb.HuodongTypeEnum_HSpringSkin, &springSkinHd{}, attr)
}

func NewSpringSkinHd(area int, attr *attribute.AttrMgr, gdata interface{}) htypes.IHuodong {
	return event.NewEventHd(area, pb.HuodongTypeEnum_HSpringSkin, &springSkinHd{}, attr, gdata)
}

func TreasureRandomSkin(player types.IPlayer, treasureData *gamedata.Treasure) ([]string, int) {
	if treasureData.RandomReward == "" {
		return []string{}, 0
	}

	rw := rand.Intn(10001)
	tw := 0
	var skins []string

	for _, ls := range allLuckBagSkins {
		tw += ls.getProp()
		if rw < tw {
			skinID := ls.getSkin()
			for j := 0; j < ls.amount; j++ {
				skins = append(skins, skinID)
			}
			break
		}
	}

	return skins, addSpringSkin(player, skins)
}
