package spring

import (
	aTypes "kinger/apps/game/activitys/types"
	"kinger/common/consts"
)

var mod *activity

type activity struct {
	aTypes.BaseActivity
}

func (a *activity) initActivityData() {
	a.IAMod = aTypes.IMod.InitDataByType(consts.ActivityOfSpring)
	if htype2Goods == nil {
		initAllGoods()
	}
}
