package spring

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var htype2Goods map[pb.HuodongTypeEnum]map[int]iHuodongGoods

func getGoods(htype pb.HuodongTypeEnum, goodsID int) iHuodongGoods {
	if id2Goods, ok := htype2Goods[htype]; ok {
		if goods, ok := id2Goods[goodsID]; ok {
			return goods
		}
	}
	return nil
}

type iHuodongGoods interface {
	getGameData() *gamedata.HuodongGoods
	canExchange(player types.IPlayer) bool
	exchange(player types.IPlayer) (itemID, itemName string, treasure *pb.OpenTreasureReply)
	getPrice() int
}

type baseHuodongGoods struct {
	data *gamedata.HuodongGoods
}

func (g *baseHuodongGoods) getGameData() *gamedata.HuodongGoods {
	return g.data
}

func (g *baseHuodongGoods) getPrice() int {
	return g.data.Price
}

func (g *baseHuodongGoods) canExchange(player types.IPlayer) bool {
	return true
}

type resourceHuodongGoods struct {
	baseHuodongGoods
	resType int
}

func (g *resourceHuodongGoods) getGoodsName() string {
	return fmt.Sprintf("%s_%d", g.data.Type, g.data.Amount)
}

func (g *resourceHuodongGoods) exchange(player types.IPlayer) (itemID, itemName string, treasure *pb.OpenTreasureReply) {
	module.Player.ModifyResource(player, g.resType, g.data.Amount, consts.RmrSpringHuodong)
	itemID = g.getGoodsName()
	itemName = itemID
	return
}

type itemHuodongGoods struct {
	baseHuodongGoods
	itemID   string
	itemType int
}

func (g *itemHuodongGoods) canExchange(player types.IPlayer) bool {
	if !g.baseHuodongGoods.canExchange(player) {
		return false
	}
	return !module.Bag.HasItem(player, g.itemType, g.itemID)
}

func (g *itemHuodongGoods) exchange(player types.IPlayer) (itemID, itemName string, treasure *pb.OpenTreasureReply) {
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
			it.(types.IEquipItem).SetFrom(consts.FromSpringHd)
			it.(types.IEquipItem).SetObtainTime(time.Now().Unix())
		}
	case consts.ItEmoji:
		emojiTeam, _ := strconv.Atoi(g.itemID)
		module.Bag.AddEmoji(player, emojiTeam)
		it = module.Bag.GetItem(player, consts.ItEmoji, g.itemID)
	default:
		return
	}

	if it != nil {
		itemID = it.GetGmID()
		itemName = it.GetName()
	}
	return
}

type cardHuodongGoods struct {
	baseHuodongGoods
	cardID uint32
}

func (g *cardHuodongGoods) canExchange(player types.IPlayer) bool {
	if !g.baseHuodongGoods.canExchange(player) {
		return false
	}

	cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(g.cardID, 1)
	if cardData == nil {
		return false
	}

	if cardData.Rare >= 99 {
		return module.Card.GetCollectCard(player, g.cardID) == nil
	}
	return true
}

func (g *cardHuodongGoods) exchange(player types.IPlayer) (itemID, itemName string, treasure *pb.OpenTreasureReply) {
	cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(g.cardID, 1)
	if cardData == nil {
		return
	}

	player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(map[uint32]*pb.CardInfo{
		g.cardID: &pb.CardInfo{Amount: int32(g.data.Amount)},
	})
	itemID = fmt.Sprintf("card%d", g.cardID)
	itemName = cardData.GetName()
	return
}

type treasureHuodongGoods struct {
	baseHuodongGoods
	treasureID   string
	treasureData *gamedata.Treasure
}

func (g *treasureHuodongGoods) canExchange(player types.IPlayer) bool {
	if !g.baseHuodongGoods.canExchange(player) {
		return false
	}

	treasureData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[g.treasureID]
	if treasureData == nil {
		return false
	}
	g.treasureData = treasureData
	return true
}

func (g *treasureHuodongGoods) exchange(player types.IPlayer) (itemID, itemName string, treasure *pb.OpenTreasureReply) {
	treasure = module.Treasure.OpenTreasureByModelID(player, g.treasureID, false)
	itemID = g.treasureID
	if g.treasureData != nil {
		itemName = g.treasureData.GetName()
	}
	return
}

func newGoods(data *gamedata.HuodongGoods) iHuodongGoods {
	switch data.Type {
	case "gold":
		goods := &resourceHuodongGoods{resType: consts.Gold}
		goods.data = data
		return goods
	case "jade":
		goods := &resourceHuodongGoods{resType: consts.Jade}
		goods.data = data
		return goods
	case "bowlder":
		goods := &resourceHuodongGoods{resType: consts.Bowlder}
		goods.data = data
		return goods
	case "ticket":
		goods := &resourceHuodongGoods{resType: consts.AccTreasureCnt}
		goods.data = data
		return goods
	case "card":
		cardID, _ := strconv.Atoi(data.RewardID)
		goods := &cardHuodongGoods{cardID: uint32(cardID)}
		goods.data = data
		return goods
	case "skin":
		goods := &itemHuodongGoods{itemType: consts.ItCardSkin, itemID: data.RewardID}
		goods.data = data
		return goods
	case "headFrame":
		goods := &itemHuodongGoods{itemType: consts.ItHeadFrame, itemID: data.RewardID}
		goods.data = data
		return goods
	case "equip":
		goods := &itemHuodongGoods{itemType: consts.ItEquip, itemID: data.RewardID}
		goods.data = data
		return goods
	case "treasure":
		goods := &treasureHuodongGoods{treasureID: data.RewardID}
		goods.data = data
		return goods
	default:
		return nil
	}
}

func genInitGoodsFunc(htype pb.HuodongTypeEnum) func(gamedata.IGameData) {
	return func(data gamedata.IGameData) {
		goodsGameData := data.(gamedata.IHuodongGoodsGameData)
		id2Goods := map[int]iHuodongGoods{}
		goodsDatas := goodsGameData.GetGoods()
		for goodsID, goodsData := range goodsDatas {
			goods := newGoods(goodsData)
			if goods != nil {
				id2Goods[goodsID] = goods
			}
		}

		htype2Goods[htype] = id2Goods
	}
}

func initAllGoods() {
	htype2Goods = map[pb.HuodongTypeEnum]map[int]iHuodongGoods{}
	htype := pb.HuodongTypeEnum_HSpringExchange
	goodsGameData := gamedata.GetGameData(consts.HuodongReward)
	if goodsGameData == nil {
		return
	}

	goodsGameData2, ok := goodsGameData.(gamedata.IHuodongGoodsGameData)
	if !ok {
		return
	}

	id2Goods, ok := htype2Goods[htype]
	if !ok {
		id2Goods = map[int]iHuodongGoods{}
		htype2Goods[htype] = id2Goods
	}

	f := genInitGoodsFunc(htype)
	goodsGameData2.AddReloadCallback(f)
	f(goodsGameData2)
}
