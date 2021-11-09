package spring

import (
	"kinger/apps/game/huodong/event"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"strconv"
)

type springHdPlayerData struct {
	event.EventPlayerData
	exchangeCntAttr *attribute.MapAttr
}

func (hpd *springHdPlayerData) Reset(version int) {
	if hpd.GetVersion() == version {
		return
	}
	hpd.EventPlayerData.Reset(version)

	hpd.exchangeCntAttr = nil
	hpd.Attr.Del("exchangeCnt")
	module.Player.SetResource(hpd.Player, consts.EventItem1, 0)
}

func (hpd *springHdPlayerData) getExchangeCnt(goodsID int) int {
	if hpd.exchangeCntAttr == nil {
		return 0
	}
	return hpd.exchangeCntAttr.GetInt(strconv.Itoa(goodsID))
}

func (hpd *springHdPlayerData) onExchangeGoods(goodsData *gamedata.HuodongGoods) {
	if goodsData.ExchangeCnt > 0 {
		if hpd.exchangeCntAttr == nil {
			hpd.exchangeCntAttr = attribute.NewMapAttr()
			hpd.Attr.SetMapAttr("exchangeCnt", hpd.exchangeCntAttr)
		}

		key := strconv.Itoa(goodsData.ID)
		hpd.exchangeCntAttr.SetInt(key, hpd.exchangeCntAttr.GetInt(key)+1)
	}
}

func (hpd *springHdPlayerData) forEachExchangeCnt(callback func(goodsID, cnt int)) {
	if hpd.exchangeCntAttr == nil {
		return
	}
	hpd.exchangeCntAttr.ForEachKey(func(key string) {
		goodsID, _ := strconv.Atoi(key)
		callback(goodsID, hpd.exchangeCntAttr.GetInt(key))
	})
}
