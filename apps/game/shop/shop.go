package shop

import (
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/proto/pb"
	"fmt"
	"strconv"
)

// 未来可能基于此，重构游戏里的所有类似商店的东西
type shopSt struct {
	type_ string
	id2Goods map[int]iGoods
}

func newCardPieceShop() *shopSt {
	s := &shopSt{type_: "cardPiece"}
	gdata := gamedata.GetGameData(consts.PieceCard)
	gdata.AddReloadCallback(s.initCardPieceShop)
	s.initCardPieceShop(gdata)
	return s
}

func newSkinPieceShop() *shopSt {
	s := &shopSt{type_: "skinPiece"}
	gdata := gamedata.GetGameData(consts.PieceSkin)
	gdata.AddReloadCallback(s.initSkinPieceShop)
	s.initSkinPieceShop(gdata)
	return s
}

func (s *shopSt) initCardPieceShop(gdata gamedata.IGameData) {
	gameData := gdata.(*gamedata.PieceCardGameData)
	s.id2Goods = map[int]iGoods{}
	for _, data := range gameData.Goods {
		g := &pieceCardGoods{data: data}
		s.id2Goods[g.getID()] = g
	}
}

func (s *shopSt) initSkinPieceShop(gdata gamedata.IGameData) {
	gameData := gdata.(*gamedata.PieceSkinGameData)
	s.id2Goods = map[int]iGoods{}
	for _, data := range gameData.Goods {
		g := &pieceSkinGoods{data: data}
		s.id2Goods[g.getID()] = g
	}
}

func (s *shopSt) buyGoods(player types.IPlayer, goodsID int) error {
	g, ok := s.id2Goods[goodsID]
	if !ok {
		return gamedata.GameError(10)
	}

	if !g.canBuy(player) {
		return gamedata.GameError(11)
	}

	if !g.hasEnoughMoney(player) {
		return gamedata.GameError(12)
	}

	g.subMoney(player)
	err := g.buy(player)
	if err != nil {
		return err
	}

	resType, resAmount := g.getConsumeMoneyType(), g.getPrice()
	mod.LogShopBuyItem(player, g.getItemID(), g.getItemName(), 1, s.type_, strconv.Itoa(resType),
		module.Player.GetResourceName(resType), resAmount, "")
	return nil
}

type iGoods interface {
	getID() int
	canBuy(player types.IPlayer) bool
	hasEnoughMoney(player types.IPlayer) bool
	subMoney(player types.IPlayer)
	buy(player types.IPlayer) error
	getItemID() string
	getItemName() string
	getConsumeMoneyType() int
	getPrice() int
}

type pieceCardGoods struct {
	data *gamedata.PieceCardGoods
}

func (g *pieceCardGoods) getConsumeMoneyType() int {
	return consts.CardPiece
}

func (g *pieceCardGoods) getPrice() int {
	return g.data.Price
}

func (g *pieceCardGoods) getID() int {
	return g.data.ID
}

func (g *pieceCardGoods) canBuy(player types.IPlayer) bool {
	if !g.data.AreaLimit.IsEffective(player.GetArea()) {
		return false
	}
	return module.Card.GetCollectCard(player, g.data.CardID) == nil
}

func (g *pieceCardGoods) hasEnoughMoney(player types.IPlayer) bool {
	return module.Player.HasResource(player, consts.CardPiece, g.data.Price)
}

func (g *pieceCardGoods) subMoney(player types.IPlayer) {
	module.Player.ModifyResource(player, consts.CardPiece, - g.data.Price)
}

func (g *pieceCardGoods) buy(player types.IPlayer) error {
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
		g.data.CardID: &pb.CardInfo{Amount: 1},
	})
	card := cardCpt.GetCollectCard(g.data.CardID)
	if card != nil {
		card.SetFrom(consts.FromPieceShop)
	}
	return nil
}

func (g *pieceCardGoods) getItemID() string {
	return fmt.Sprintf("card%d", g.data.CardID)
}

func (g *pieceCardGoods) getItemName() string {
	cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(g.data.CardID, 1)
	if cardData != nil {
		return cardData.GetName()
	} else {
		return g.getItemID()
	}
}

type pieceSkinGoods struct {
	data *gamedata.PieceSkinGoods
}

func (g *pieceSkinGoods) getID() int {
	return g.data.ID
}

func (g *pieceSkinGoods) getConsumeMoneyType() int {
	return consts.SkinPiece
}

func (g *pieceSkinGoods) getPrice() int {
	return g.data.Price
}

func (g *pieceSkinGoods) canBuy(player types.IPlayer) bool {
	if !g.data.AreaLimit.IsEffective(player.GetArea()) {
		return false
	}
	return module.Bag.GetItem(player, consts.ItCardSkin, g.data.SkinID) == nil
}

func (g *pieceSkinGoods) hasEnoughMoney(player types.IPlayer) bool {
	return module.Player.HasResource(player, consts.SkinPiece, g.data.Price)
}

func (g *pieceSkinGoods) subMoney(player types.IPlayer) {
	module.Player.ModifyResource(player, consts.SkinPiece, - g.data.Price)
}

func (g *pieceSkinGoods) buy(player types.IPlayer) error {
	module.Bag.AddCardSkin(player, g.data.SkinID)
	return nil
}

func (g *pieceSkinGoods) getItemID() string {
	return g.data.SkinID
}

func (g *pieceSkinGoods) getItemName() string {
	skinData := gamedata.GetGameData(consts.CardSkin).(*gamedata.CardSkinGameData).ID2CardSkin[g.data.SkinID]
	if skinData != nil {
		return skinData.GetName()
	} else {
		return g.getItemID()
	}
}
