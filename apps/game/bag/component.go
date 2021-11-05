package bag

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"strconv"
	"kinger/gamedata"
	"kinger/gopuppy/common/eventhub"
	"time"
)

type bagComponent struct {
	attr *attribute.MapAttr
	player types.IPlayer
	type2Bag map[int]*bagSt
}

func (bc *bagComponent) ComponentID() string {
	return consts.BagCpt
}

func (bc *bagComponent) GetPlayer() types.IPlayer {
	return bc.player
}

func (bc *bagComponent) OnInit(player types.IPlayer) {
	bc.player = player
	bagsAttr := bc.attr.GetMapAttr("type2Bag")
	if bagsAttr == nil {
		bagsAttr = attribute.NewMapAttr()
		bc.attr.SetMapAttr("type2Bag", bagsAttr)
	}
	bc.type2Bag = map[int]*bagSt{}
	for _, itemType := range allItemType {
		key := strconv.Itoa(itemType)
		attr := bagsAttr.GetMapAttr(key)
		if attr == nil {
			attr = attribute.NewMapAttr()
			bagsAttr.SetMapAttr(key, attr)
		}
		b := newBag(itemType, attr)
		bc.type2Bag[itemType] = b
	}

	defHeadFrame := mod.GetDefHeadFrame()
	if bc.getItem(consts.ItHeadFrame, defHeadFrame) == nil {
		bc.addItem(consts.ItHeadFrame, defHeadFrame)
	}
	defEmoji := mod.getDefHeadEmoji()
	if bc.getItem(consts.ItEmoji, defEmoji) == nil {
		bc.addItem(consts.ItEmoji, defEmoji)
	}
	defChatPop := mod.GetDefChatPop()
	if bc.getItem(consts.ItChatPop, defChatPop) == nil {
		bc.addItem(consts.ItChatPop, defChatPop)
	}
}

func (bc *bagComponent) OnLogin(isRelogin, isRestore bool) {
	bc.checkEquipVersion()
	bc.patchObtainTime()
}

func (bc *bagComponent) OnLogout() {
}

func (bc *bagComponent) checkEquipVersion() {
	b, ok := bc.type2Bag[consts.ItEquip]
	if !ok {
		return
	}

	equipGameData := gamedata.GetGameData(consts.Equip).(*gamedata.EquipGameData)
	allItems := b.getAllItems()
	for _, it := range allItems {
		eit, ok := it.(*equipItem)
		if !ok {
			continue
		}

		data := equipGameData.ID2Equip[eit.GetID()]
		if data == nil || data.Version > eit.GetVersion() {
			eit.Reset(bc.player)
		}
	}
}

func (bc *bagComponent) getItem(type_ int, itemID string) types.IItem {
	b, ok := bc.type2Bag[type_]
	if !ok {
		return nil
	}
	return b.getItem(itemID)
}

func (bc *bagComponent) getAllItemsByType(type_ int) []types.IItem {
	b, ok := bc.type2Bag[type_]
	if !ok {
		return []types.IItem{}
	}
	return b.getAllItems()
}

func (bc *bagComponent) addItem(type_ int, itemID string) (types.IItem, bool) {
	b, ok := bc.type2Bag[type_]
	if !ok {
		return nil, false
	}

	it, ok := b.addItemByID(itemID)
	if ok {
		log.modifyItem(bc.player, type_, itemID, 1)
	}
	return it, ok
}

func (bc *bagComponent) delItem(it types.IItem) {
	itemType := it.GetType()
	b, ok := bc.type2Bag[itemType]
	if !ok {
		return
	}

	itemID := it.GetID()
	b.delItem(itemID)
	log.modifyItem(bc.player, it.GetType(), itemID, -1)

	if itemType == consts.ItEquip {
		eventhub.Publish(consts.EvEquipDel, bc.player, itemID)
	}
}

func (bc *bagComponent) patchObtainTime(){
	its := bc.getAllItemsByType(consts.ItEquip)
	for _, it := range its {
		if et, ok := it.(*equipItem); ok {
			tim := et.GetObtainTime()
			if tim == 0 {
				et.SetObtainTime(time.Now().Unix())
			}
		}
	}
}