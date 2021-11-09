package springskin

import (
	"kinger/apps/game/huodong/event"
	"kinger/gopuppy/attribute"
)

type springSkinHdPlayerData struct {
	event.EventPlayerData
	springSkinsAttr *attribute.ListAttr
}

func (hpd *springSkinHdPlayerData) Reset(version int) {
	if hpd.GetVersion() == version {
		return
	}
	hpd.EventPlayerData.Reset(version)

	hpd.springSkinsAttr = nil
	hpd.Attr.Del("springSkins")
}

func (hpd *springSkinHdPlayerData) getSpringSkinsAttr() *attribute.ListAttr {
	if hpd.springSkinsAttr == nil {
		hpd.springSkinsAttr = hpd.Attr.GetListAttr("springSkins")
		if hpd.springSkinsAttr == nil {
			hpd.springSkinsAttr = attribute.NewListAttr()
			hpd.Attr.SetListAttr("springSkins", hpd.springSkinsAttr)
		}
	}
	return hpd.springSkinsAttr
}

func (hpd *springSkinHdPlayerData) onGetSkin(skin string) {
	hpd.getSpringSkinsAttr().AppendStr(skin)
}

func (hpd *springSkinHdPlayerData) forEachSkin(callback func(skin string) bool) {
	springSkinsAttr := hpd.getSpringSkinsAttr()
	if springSkinsAttr == nil {
		return
	}
	springSkinsAttr.ForEachIndex(func(index int) bool {
		return callback(springSkinsAttr.GetStr(index))
	})
}
