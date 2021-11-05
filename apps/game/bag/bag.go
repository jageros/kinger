package bag

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
)

type bagSt struct {
	attr *attribute.MapAttr
	itemsAttr *attribute.MapAttr
	type_ int
	items map[string]types.IItem
}

func newBag(type_ int, attr *attribute.MapAttr) *bagSt {
	b := &bagSt{
		type_: type_,
		attr: attr,
		items: map[string]types.IItem{},
	}

	itemsAttr := attr.GetMapAttr("items")
	if itemsAttr == nil {
		itemsAttr = attribute.NewMapAttr()
		attr.SetMapAttr("items", itemsAttr)
	}
	itemsAttr.ForEachKey(func(itemID string) {
		it := newItemByAttr(type_, itemID, itemsAttr.GetMapAttr(itemID))
		b.items[itemID] = it
	})
	b.itemsAttr = itemsAttr

	return b
}

func (b *bagSt) addItemByID(itemID string) (types.IItem, bool) {
	if it, ok := b.items[itemID]; ok {
		it.ModifyAmount(1)
		return it, false
	} else {
		it := newItem(b.type_, itemID)
		it.ModifyAmount(1)
		b.itemsAttr.SetMapAttr(itemID, it.GetAttr())
		b.items[itemID] = it
		return it, true
	}
}

func (b *bagSt) getItem(itemID string) types.IItem {
	if it, ok := b.items[itemID]; ok {
		return it
	} else {
		return nil
	}
}

func (b *bagSt) getAllItems() []types.IItem {
	var items []types.IItem
	for _, it := range b.items {
		items = append(items, it)
	}
	return items
}

func (b *bagSt) delItem(itemID string) {
	delete(b.items, itemID)
	b.itemsAttr.Del(itemID)
}
