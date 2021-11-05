package campaign

import (
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"fmt"
	"kinger/apps/game/module"
	"kinger/common/config"
	"kinger/gamedata"
	"time"
)

var type2Goods map[string]map[int]iGoods = nil

func getGoods(type_ string, id int) iGoods {
	if type2Goods == nil {
		return nil
	}

	if id2Goods, ok := type2Goods[type_]; ok {
		if g, ok := id2Goods[id]; ok {
			return g
		} else {
			return nil
		}
	} else {
		return nil
	}
}

type iGoods interface {
	getPrice() int
	canBuy(player types.IPlayer) bool
	buy(player types.IPlayer) (string, string)
}

type baseGoods struct {
	id int
	type_ string
	price int
}

func (g *baseGoods) String() string {
	return fmt.Sprintf("[type_=%s, id=%d, price=%d]", g.type_, g.id, g.price)
}

func (g *baseGoods) getPrice() int {
	return g.price
}

type cardGoods struct {
	baseGoods
	cardID uint32
}

func (g *cardGoods) canBuy(player types.IPlayer) bool {
	return player.GetComponent(consts.CardCpt).(types.ICardComponent).GetCollectCard(g.cardID) == nil
}

func (g *cardGoods) buy(player types.IPlayer) (string, string) {
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
		g.cardID: &pb.CardInfo{
			Amount: 1,
		},
	})

	card := cardCpt.GetCollectCard(g.cardID)
	if card != nil {
		card.SetFrom(consts.FromCampaign)
	}

	return fmt.Sprintf("card%d", g.cardID), card.GetCardGameData().GetName()
}

type itemGoods struct {
	baseGoods
	itemType int
	itemID string
}

func (g *itemGoods) canBuy(player types.IPlayer) bool {
	return !module.Bag.HasItem(player, g.itemType, g.itemID)
}

func (g *itemGoods) buy(player types.IPlayer) (string, string) {
	var it types.IItem = nil
	switch g.itemType {
	case consts.ItHeadFrame:
		it = module.Bag.AddHeadFrame(player, g.itemID)
	case consts.ItCardSkin:
		it = module.Bag.AddCardSkin(player, g.itemID)
	case consts.ItEquip:
		module.Bag.AddEquip(player, g.itemID)
		it = module.Bag.GetItem(player, consts.ItEquip, g.itemID)
		if it != nil {
			eit, ok := it.(types.IEquipItem)
			if ok {
				eit.SetFrom(consts.FromCampaign)
				eit.SetObtainTime(time.Now().Unix())
			}
		}
	}

	if it != nil {
		return it.GetGmID(), it.GetName()
	} else {
		return "", ""
	}
}

type resourceGoods struct {
	baseGoods
	resType int
	amount int
}

func (g *resourceGoods) canBuy(player types.IPlayer) bool {
	return true
}

func (g *resourceGoods) buy(player types.IPlayer) (string, string) {
	module.Player.ModifyResource(player, g.resType, g.amount)
	return "", ""
}

func initCardGoods(gdata gamedata.IGameData) {
	gdata2 := gdata.(*gamedata.WarShopCardGameData)
	goodses := map[string]map[int]iGoods{}
	for _, data := range gdata2.Goods {
		g := &cardGoods{}
		g.type_ = data.Type
		g.id = data.ID
		g.price = data.FightPrice
		g.cardID = data.CardID
		id2Goods, ok := goodses[data.Type]
		if !ok {
			id2Goods = map[int]iGoods{}
		}
		id2Goods[data.ID] = g
		goodses[data.Type] = id2Goods
	}

	for type_, id2Goods := range goodses {
		type2Goods[type_] = id2Goods
	}
}

func initEquipGoods(gdata gamedata.IGameData) {
	gdata2 := gdata.(*gamedata.WarShopEquipGameData)
	goodses := map[string]map[int]iGoods{}
	for _, data := range gdata2.Goods {
		g := &itemGoods{}
		g.type_ = data.Type
		g.id = data.ID
		g.price = data.FightPrice
		g.itemID = data.EquipID
		g.itemType = consts.ItEquip
		id2Goods, ok := goodses[data.Type]
		if !ok {
			id2Goods = map[int]iGoods{}
		}
		id2Goods[data.ID] = g
		goodses[data.Type] = id2Goods
	}

	for type_, id2Goods := range goodses {
		type2Goods[type_] = id2Goods
	}
}

func initSkinGoods(gdata gamedata.IGameData) {
	gdata2 := gdata.(*gamedata.WarShopSkinGameData)
	goodses := map[string]map[int]iGoods{}
	for _, data := range gdata2.Goods {
		g := &itemGoods{}
		g.type_ = data.Type
		g.id = data.ID
		g.price = data.FightPrice
		g.itemID = data.SkinID
		g.itemType = consts.ItCardSkin
		id2Goods, ok := goodses[data.Type]
		if !ok {
			id2Goods = map[int]iGoods{}
		}
		id2Goods[data.ID] = g
		goodses[data.Type] = id2Goods
	}

	for type_, id2Goods := range goodses {
		type2Goods[type_] = id2Goods
	}
}

func initResGoods(gdata gamedata.IGameData) {
	gdata2 := gdata.(*gamedata.WarShopResGameData)
	goodses := map[string]map[int]iGoods{}
	for _, data := range gdata2.Goods {
		g := &resourceGoods{}
		if data.Amount <= 0 {
			continue
		}

		var resType int
		switch data.Type {
		case "gold":
			resType = consts.Gold
		default:
			continue
		}

		g.type_ = data.Type
		g.id = data.ID
		g.price = data.FightPrice
		g.amount = data.Amount
		g.resType = resType
		id2Goods, ok := goodses[data.Type]
		if !ok {
			id2Goods = map[int]iGoods{}
		}
		id2Goods[data.ID] = g
		goodses[data.Type] = id2Goods
	}

	for type_, id2Goods := range goodses {
		type2Goods[type_] = id2Goods
	}
}

func initializeGoods() {
	if config.GetConfig().IsMultiLan {
		return
	}

	type2Goods = map[string]map[int]iGoods{}
	warShopCardGameData := gamedata.GetGameData(consts.WarShopCard)
	warShopCardGameData.AddReloadCallback(initCardGoods)
	initCardGoods(warShopCardGameData)

	warShopEquipGameData := gamedata.GetGameData(consts.WarShopEquip)
	warShopEquipGameData.AddReloadCallback(initEquipGoods)
	initEquipGoods(warShopEquipGameData)

	warShopSkinGameData := gamedata.GetGameData(consts.WarShopSkin)
	warShopSkinGameData.AddReloadCallback(initSkinGoods)
	initSkinGoods(warShopSkinGameData)

	//warShopResGameData := gamedata.GetGameData(consts.WarShopRes)
	//warShopResGameData.AddReloadCallback(initResGoods)
	//initResGoods(warShopResGameData)
}
