package level

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
)

var mod = &levelModule{}

type levelModule struct {
}

func (m *levelModule) NewLevelComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	attr := playerAttr.GetMapAttr("level")
	if attr == nil {
		attr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("level", attr)
	}

	return &levelComponent{attr: attr}
}

func (m *levelModule) GetCurLevel(player types.IPlayer) int {
	return player.GetComponent(consts.LevelCpt).(*levelComponent).GetCurLevel()
}

func onRecharge(args ...interface{}) {
	args[0].(types.IPlayer).GetComponent(consts.LevelCpt).(*levelComponent).onRecharge()
}

func Initialize() {
	registerRpc()
	module.Level = mod
	//eventhub.Subscribe(consts.EvRecharge, onRecharge)
	//eventhub.Subscribe(consts.EvFixLevelRechargeUnlock, onRecharge)
}
