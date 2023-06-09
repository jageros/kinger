package tutorial

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/attribute"
)

type tutorialModule struct {
}

func (t *tutorialModule) NewTutorialComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	tutorialAttr := playerAttr.GetMapAttr("tutorial")
	if tutorialAttr == nil {
		tutorialAttr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("tutorial", tutorialAttr)
	}
	return &tutorialComponent{attr: tutorialAttr}
}

func Initialize() {
	module.Tutorial = &tutorialModule{}
	registerRpc()
}
