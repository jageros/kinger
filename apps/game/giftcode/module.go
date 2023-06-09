package giftcode

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
)

var mod *giftCodeModule

type giftCodeModule struct {
}

func (m *giftCodeModule) NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	giftCodeAttr := playerAttr.GetMapAttr("giftCode")
	if giftCodeAttr == nil {
		giftCodeAttr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("giftCode", giftCodeAttr)
	}
	return &giftCodeComponent{attr: giftCodeAttr}
}

func Initialize() {
	mod = &giftCodeModule{}
	module.GiftCode = mod
	registerRpc()
}
