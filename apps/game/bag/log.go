package bag

import (
	"kinger/gopuppy/attribute"
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"strconv"
)

var log = &logHub{
	itemID2Log: map[string]*itemLog{},
}

type itemLog struct {
	attr *attribute.AttrMgr
	allAreasAttr *attribute.MapAttr
	area2Amounts map[int]*attribute.MapAttr
}

func itemType2LogStr(itemType int) string {
	switch itemType {
	case consts.ItHeadFrame:
		return "headFrame"
	case consts.ItCardSkin:
		return "skin"
	case consts.ItEquip:
		return "equip"
	case consts.ItEmoji:
		return "emoji"
	default:
		return ""
	}
}

func newItemLog(itemType int, itemID string) *itemLog {
	attr := attribute.NewAttrMgr(fmt.Sprintf("itemlog%d", module.Service.GetAppID()), itemID)
	err := attr.Load()
	var amountsAttr *attribute.MapAttr
	if err != nil {
		attr.SetStr("type", itemType2LogStr(itemType))
		amountsAttr = attribute.NewMapAttr()
		attr.SetMapAttr("amounts", amountsAttr)
	}

	return &itemLog{
		attr: attr,
		allAreasAttr: attr.GetMapAttr("amounts"),
	}
}

func (il *itemLog) save() {
	il.attr.Save(false)
}

func (il *itemLog) modifyAmount(player types.IPlayer, amount int) {
	accountType := player.GetLogAccountType().String()
	area := player.GetArea()
	amountsAttr, ok := il.area2Amounts[area]
	if !ok {
		amountsAttr = il.allAreasAttr.GetMapAttr(strconv.Itoa(area))
		if amountsAttr == nil {
			amountsAttr = attribute.NewMapAttr()
			il.allAreasAttr.SetMapAttr(strconv.Itoa(area), amountsAttr)
		}
	}
	amountsAttr.SetInt(accountType, amountsAttr.GetInt(accountType) + amount)
}

type logHub struct {
	itemID2Log map[string]*itemLog
}

func (l *logHub) modifyItem(player types.IPlayer, itemType int, itemID string, amount int) {
	if il, ok := l.itemID2Log[itemID]; ok {
		il.modifyAmount(player, amount)
	} else {
		il = newItemLog(itemType, itemID)
		l.itemID2Log[itemID] = il
		il.modifyAmount(player, amount)
	}
}

func (l *logHub) save() {
	for _, il := range l.itemID2Log {
		il.save()
	}
}
