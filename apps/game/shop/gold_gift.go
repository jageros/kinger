package shop

import (
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/proto/pb"
	"kinger/gamedata"
	"kinger/common/consts"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"strconv"
)

type goldGiftSt struct {
	player types.IPlayer
	attr *attribute.MapAttr
}

func newGoldGiftSt(player types.IPlayer, cptAttr *attribute.MapAttr) *goldGiftSt {
	attr := cptAttr.GetMapAttr("goldGift")
	if attr == nil {
		attr = attribute.NewMapAttr()
		cptAttr.SetMapAttr("goldGift", attr)
	}
	return &goldGiftSt{
		player: player,
		attr: attr,
	}
}

func (gg *goldGiftSt) getCurIdx() int {
	return gg.attr.GetInt("idx")
}

func (gg *goldGiftSt) setCurIdx(idx int) {
	gg.attr.SetInt("idx", idx)
}

func (gg *goldGiftSt) getGameData() *gamedata.SoldGoldGift {
	return gamedata.GetGameData(consts.SoldGoldGift).(*gamedata.SoldGoldGiftGameData).GetGift(gg.player.GetArea(),
		gg.getCurIdx())
}

func (gg *goldGiftSt) getLastGameData() *gamedata.SoldGoldGift {
	return gamedata.GetGameData(consts.SoldGoldGift).(*gamedata.SoldGoldGiftGameData).GetGift(gg.player.GetArea(),
		gg.getCurIdx() - 1)
}

func (gg *goldGiftSt) packMsg() *pb.SoldGoldGift {
	data := gg.getGameData()
	var remainTime int32
	if data == nil {
		data = gg.getLastGameData()
		if data == nil {
			return nil
		}
		remainTime = int32(timer.TimeDelta(0, 0, 0).Seconds())
	}

	return &pb.SoldGoldGift {
		TreasureID: data.TreasureID,
		JadePrice: int32(data.JadePrice),
		CanBuyRemainTime: remainTime,
	}
}

func (gg *goldGiftSt) onCrossDay() {
	oldCurIdx := gg.getCurIdx()
	if oldCurIdx > 0 {
		gg.setCurIdx(0)
		gg.player.GetComponent(consts.ShopCpt).(*shopComponent).onShopDataUpdate(pb.UpdateShopDataArg_GoldGift)
	}
}

func (gg *goldGiftSt) buy() (*pb.BuySoldGoldGiftReply, error) {
	data := gg.getGameData()
	if data == nil {
		return nil, gamedata.GameError(1)
	}

	if !module.Player.HasResource(gg.player, consts.Jade, data.JadePrice) {
		return nil, gamedata.GameError(2)
	}
	module.Player.ModifyResource(gg.player, consts.Jade, - data.JadePrice, consts.RmrBuyGoldGift)

	module.Shop.LogShopBuyItem(gg.player, "gold_gift", "金币礼包", 1, "shop",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), data.JadePrice, "")

	reply := &pb.BuySoldGoldGiftReply{}
	reply.TreasureReward = module.Treasure.OpenTreasureByModelID(gg.player, data.TreasureID, false)
	gg.setCurIdx(gg.getCurIdx() + 1)
	reply.NextGift = gg.packMsg()
	return reply, nil
}
