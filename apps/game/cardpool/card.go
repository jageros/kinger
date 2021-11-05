package cardpool

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"time"
)

var _ types.ICollectCard = &collectCard{}

type collectCard struct {
	rare int
	attr *attribute.MapAttr
}

func newCollectCardByAttr(attr *attribute.MapAttr) *collectCard {
	return &collectCard{
		rare: -1,
		attr: attr,
	}
}

func newCollectCard(cardData *gamedata.Card, amount int) *collectCard {
	attr := attribute.NewMapAttr()
	attr.SetUInt32("cardID", cardData.CardID)
	attr.SetInt("level", cardData.Level)
	attr.SetInt("amount", amount)
	attr.SetFloat("energy", cardData.Energy)
	attr.SetInt("version", cardData.Version)
	return newCollectCardByAttr(attr)
}

func newCollectCardByOldData(c *collectCard, cardData *gamedata.Card, amount int) *collectCard {
	c.setLevel(cardData.Level)
	c.SetAmount(amount)
	c.attr.SetInt("version", cardData.Version)
	return c
}

func (c *collectCard) String() string {
	return fmt.Sprintf("[collectCard cardID=%d, level=%d, amount=%d, rare=%d]", c.GetCardID(), c.GetLevel(),
		c.GetAmount(), c.getRare())
}

func (c *collectCard) getAttr() *attribute.MapAttr {
	return c.attr
}

func (c *collectCard) getRare() int {
	if c.rare < 0 {
		c.GetCardGameData()
	}
	return c.rare
}

func (c *collectCard) GetLevel() int {
	return c.attr.GetInt("level")
}

func (c *collectCard) setLevel(lv int) {
	c.attr.SetInt("level", lv)
}

func (c *collectCard) GetCardGameData() *gamedata.Card {
	gdata := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	data := gdata.GetCard(c.GetCardID(), c.GetLevel())
	if data != nil {
		c.rare = data.Rare
	}
	return data
}

func (c *collectCard) IsDead() bool {
	return false
}

func (c *collectCard) GetCardID() uint32 {
	return c.attr.GetUInt32("cardID")
}

func (c *collectCard) GetAmount() int {
	return c.attr.GetInt("amount")
}

func (c *collectCard) SetAmount(amount int) {
	c.attr.SetInt("amount", amount)
}

func (c *collectCard) GetEnergy() float32 {
	return c.attr.GetFloat32("energy")
}

func (c *collectCard) setEnergy(val float32) {
	c.attr.SetFloat("energy", val)
}

func (c *collectCard) getUseCount() int {
	return c.attr.GetInt("useCount")
}

func (c *collectCard) use() {
	c.attr.SetInt("useCount", c.getUseCount()+1)
	c.attr.SetInt("lastUseTime", int(time.Now().Unix()))
}

func (c *collectCard) getLastUseTime() int {
	return c.attr.GetInt("lastUseTime")
}

func (c *collectCard) GetVersion() int {
	return c.attr.GetInt("version")
}

func (c *collectCard) GetFrom() int {
	return c.attr.GetInt("from")
}

func (c *collectCard) SetFrom(from int) {
	c.attr.SetInt("from", from)
}

func (c *collectCard) resetNormalCard(data *gamedata.Card, player types.IPlayer) (returnRes map[int]int, amount, level,
	contribution int) {

	gdata := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	level = c.GetLevel()
	returnRes = map[int]int{}
	cardID := c.GetCardID()
	for level > data.ResetLevel {
		// 回退
		level--
		data2 := gdata.GetCard(cardID, level)
		if data2 == nil {
			level++
			break
		}
		amount += data2.LevelupNum
		returnRes[consts.Gold] += data2.LevelupGold
		returnRes[consts.Heroes] += data2.LevelupHor
		returnRes[consts.Mat] += data2.LevelupMat
		returnRes[consts.Weap] += data2.LevelupWeap
	}
	c.attr.SetInt("level", level)

	if c.GetMaxUnlockLevel() > 0 {
		cardData := gdata.GetCard(cardID, c.GetMaxUnlockLevel() - 1)
		c.setMaxUnlockLevel(0)
		if cardData != nil && cardData.ConsumeBook > 0 {
			returnRes[consts.SkyBook] = cardData.ConsumeBook
		}
	}

	return
}

func (c *collectCard) resetRebornSpCard(data *gamedata.Card, player types.IPlayer) (returnRes map[int]int, amount, level,
	contribution int) {

	cardID := c.GetCardID()
	rebornCardGoods := gamedata.GetGameData(consts.RebornSoldCard).(*gamedata.RebornSoldCardGameData)
	goodsData, ok := rebornCardGoods.CardID2Card[cardID]
	if ok {
		returnRes = map[int]int{}
		returnRes[consts.Feats] = goodsData.HonorPrice
		level = 0
	}
	return
}

func (c *collectCard) resetPieceSpCard(data *gamedata.Card, player types.IPlayer) (returnRes map[int]int, amount, level,
	contribution int) {

	cardID := c.GetCardID()
	pieceCardGoods := gamedata.GetGameData(consts.PieceCard).(*gamedata.PieceCardGameData)
	goodsData, ok := pieceCardGoods.GetCardID2Goods(player.GetArea())[cardID]
	returnRes = map[int]int{}
	level = 0
	if ok {
		returnRes[consts.CardPiece] = goodsData.Price
	} else {
		if len(pieceCardGoods.Goods) > 0 {
			returnRes[consts.CardPiece] = pieceCardGoods.Goods[0].Price
		}
	}
	return
}

func (c *collectCard) resetCampaignSpCard(data *gamedata.Card, player types.IPlayer) (returnRes map[int]int, amount, level,
	contribution int) {

	cardID := c.GetCardID()
	rebornCardGoods := gamedata.GetGameData(consts.WarShopCard).(*gamedata.WarShopCardGameData)
	goodsData, ok := rebornCardGoods.CardID2Goods[cardID]
	if ok {
		contribution = goodsData.FightPrice
		level = 0
	}
	return
}

func (c *collectCard) Reset(player types.IPlayer) {
	data := c.GetCardGameData()
	if data == nil {
		return
	}

	version := c.GetVersion()
	if version >= data.Version {
		return
	}

	c.attr.SetInt("version", data.Version)
	cardID := c.GetCardID()
	var returnRes map[int]int
	level := c.GetLevel()
	var amount, contribution int
	oldLevel := level

	if data.ResetLevel > 0 {
		returnRes, amount, level, contribution = c.resetNormalCard(data, player)
	} else if c.GetFrom() == consts.FromReborn || c.GetFrom() == consts.FromPieceShop {
		returnRes, amount, level, contribution = c.resetPieceSpCard(data, player)
	} else {
		returnRes, amount, level, contribution = c.resetCampaignSpCard(data, player)
	}

	if returnRes != nil && len(returnRes) > 0 {
		player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).BatchModifyResource(returnRes, consts.RmrResetCard)
	}
	if amount > 0 {
		player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(map[uint32]*pb.CardInfo{
			cardID: &pb.CardInfo{Amount:int32(amount)},
		})
	} else if level == 0 {
		cardCpt := player.GetComponent(consts.CardCpt).(*cardComponent)
		cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
			cardID: &pb.CardInfo{Level:int32(- oldLevel)},
		})
		cardCpt.poolDelCard(cardID)
	}

	if contribution > 0 {
		module.Campaign.ModifyContribution(player, contribution)
	}

	glog.Infof("card reset level, uid=%d, cardID=%d, oldLevel=%d, curLevel=%d, returnRes=%v, contribution=%d, amount=%d",
		player.GetUid(), cardID, oldLevel, level, returnRes, contribution, amount)
}

func (c *collectCard) setSkin(skin string) {
	c.attr.SetStr("skin", skin)
}

func (c *collectCard) GetSkin() string {
	return c.attr.GetStr("skin")
}

func (c *collectCard) IsMaxLevel() bool {
	gdata := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	return gdata.GetCard(c.GetCardID(), c.GetLevel() + 1) == nil
}

func (c *collectCard) IsSpCard() bool {
	gdata := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	return gdata.GetCard(c.GetCardID(), c.GetLevel()).IsSpCard()
}

func (c *collectCard) GetEquip() string {
	return c.attr.GetStr("equip")
}

func (c *collectCard) WearEquip(equipID string) {
	c.attr.SetStr("equip", equipID)
}

func (c *collectCard) DeEquip() {
	c.attr.Del("equip")
}

func (c *collectCard) del(player types.IPlayer) {
	c.setLevel(0)
	c.SetAmount(0)
	c.setState(pb.CardState_NormalCState)
	equipID := c.GetEquip()
	c.DeEquip()
	c.setSkin("")
	if equipID != "" {
		it, ok := module.Bag.GetItem(player, consts.ItEquip, equipID).(types.IEquipItem)
		if ok {
			it.SetOwner(0)
		}
	}
}

func (c *collectCard) GetState() pb.CardState {
	return pb.CardState(c.attr.GetInt("state"))
}

func (c *collectCard) setState(state pb.CardState) {
	c.attr.SetInt("state", int(state))
}

func (c *collectCard) GetMaxUnlockLevel() int {
	return c.attr.GetInt("maxUnlockLevel")
}

func (c *collectCard) setMaxUnlockLevel(level int) {
	c.attr.SetInt("maxUnlockLevel", level)
}

func (c *collectCard) PackMsg() *pb.CardInfo {
	return &pb.CardInfo{
		CardId: c.GetCardID(),
		Level: int32(c.GetLevel()),
		Amount: int32(c.GetAmount()),
		Energy: c.GetEnergy(),
		Skin: c.GetSkin(),
		Equip: c.GetEquip(),
		State: c.GetState(),
		MaxUnlockLevel: int32(c.GetMaxUnlockLevel()),
	}
}

func (c *collectCard) IsMaxCanUpLevel() bool {
	if c.IsMaxLevel() {
		return true
	}

	level := c.GetLevel()
	maxUnlockLevel := c.GetMaxUnlockLevel()
	if maxUnlockLevel > 0 {
		return level >= maxUnlockLevel
	} else {
		cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(c.GetCardID(), level)
		if cardData == nil {
			return true
		}
		if cardData.ConsumeBook > 0 {
			return true
		}
		return false
	}
}

func (c *collectCard) pushClient(player types.IPlayer) {
	agent := player.GetAgent()
	agent.PushClient(pb.MessageID_S2C_SYNC_CARD_INFO, &pb.CardDatas{
		Cards: []*pb.CardInfo{c.PackMsg()},
	})
}