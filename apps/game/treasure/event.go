package treasure

import (
	"time"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"math/rand"
	"kinger/proto/pb"
	"fmt"
	"strconv"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/common/utils"
)

var treasureEventDeadLine int64 = 24 * 60 * 60

type baseTreasureEvent struct {
	attr *attribute.MapAttr
	player types.IPlayer
}

func (te *baseTreasureEvent) init(player types.IPlayer, cptAttr *attribute.MapAttr) {
	attr := cptAttr.GetMapAttr("addCardEvent")
	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt64("lastTime", time.Now().Unix() - treasureEventDeadLine + 7200)
		cptAttr.SetMapAttr("addCardEvent", attr)
	}

	te.player = player
	te.attr = attr
}

func (te *baseTreasureEvent) canTrigger() bool {
	if te.player.GetPvpTeam() < 3 {
		return false
	}

	lastTime := te.attr.GetInt64("lastTime")
	can := false
	now := time.Now().Unix()
	if now - lastTime >= treasureEventDeadLine {
		can = true
	}

	if !can {
		can = rand.Intn(100) < 8
	}

	if can {
		te.attr.SetInt64("lastTime", now)
	}
	return can
}

type addCardEventSt struct {
	baseTreasureEvent
}

func newAddCardEvent(player types.IPlayer, cptAttr *attribute.MapAttr) *addCardEventSt {
	te := &addCardEventSt{}
	te.init(player, cptAttr)
	return te
}

func (te *addCardEventSt) trigger(treasureID uint32, treasureModelID string, cardIDs []uint32) bool {
	if !te.canTrigger() {
		return false
	}

	tdata := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[treasureModelID]
	if tdata == nil {
		return false
	}

	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(tdata.Rare)
	if eventConfig == nil {
		return false
	}

	te.attr.SetUInt32("treasureID", treasureID)
	te.attr.SetStr("modelID", treasureModelID)
	cardsAttr := attribute.NewListAttr()
	te.attr.SetListAttr("cards", cardsAttr)
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	utils.ShuffleUInt32(cardIDs)
	for _, cardID := range cardIDs {
		cardData := poolGameData.GetCard(cardID, 1)
		if cardData == nil || cardData.IsSpCard() {
			continue
		}

		cardsAttr.AppendUInt32(cardID)
		if cardsAttr.Size() >= eventConfig.AddCardCnt {
			break
		}
	}

	return true
}

func (te *addCardEventSt) cancel() {
	te.attr.Del("cards")
	te.attr.Del("treasureID")
	te.attr.Del("modelID")
}

func (te *addCardEventSt) doAction(treasureID uint32) (*pb.WatchTreasureAddCardAdsReply, error) {
	cardsAttr := te.attr.GetListAttr("cards")
	if cardsAttr == nil {
		return nil, gamedata.GameError(1)
	}
	if te.attr.GetUInt32("treasureID") != treasureID {
		return nil, gamedata.GameError(2)
	}

	tdata := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Treasures[te.attr.GetStr("modelID")]
	if tdata == nil {
		return nil, gamedata.GameError(3)
	}

	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(tdata.Rare)
	if eventConfig == nil {
		return nil, gamedata.GameError(4)
	}

	if !te.player.HasBowlder(eventConfig.AddCardPrice) {
		return nil, gamedata.GameError(5)
	}
	te.player.SubBowlder(eventConfig.AddCardPrice, consts.RmrTreasureAddCard)

	te.cancel()

	reply := &pb.WatchTreasureAddCardAdsReply{}
	cardChange := map[uint32]*pb.CardInfo{}
	cardsAttr.ForEachIndex(func(index int) bool {
		cardID := cardsAttr.GetUInt32(index)
		c, ok := cardChange[cardID]
		reply.AddCardIDs = append(reply.AddCardIDs, cardID)
		if !ok {
			cardChange[cardID] = &pb.CardInfo{Amount: 1}
		} else {
			c.Amount += 1
		}
		return true
	})

	module.Shop.LogShopBuyItem(te.player, "tAddCard", "宝箱加卡", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), eventConfig.AddCardPrice,
		fmt.Sprintf("treasure=%s", tdata.ID))

	module.Card.ModifyCollectCards(te.player, cardChange)
	return reply, nil
}


type upRareEventSt struct {
	baseTreasureEvent
	cpt *treasureComponent
}

func newUpRareEventSt(cpt *treasureComponent, cptAttr *attribute.MapAttr) *upRareEventSt {
	te := &upRareEventSt{cpt: cpt}
	te.init(cpt.GetPlayer(), cptAttr)
	return te
}

func (te *upRareEventSt) genUpRareTreasureModelID(rare int) string {
	pvpLevel := te.player.GetPvpLevel()
	rgd := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	r := rgd.Ranks[pvpLevel]
	ts := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).TreasuresOfTeam[r.Team]

	if len(ts) == 0 {
		return ""
	}

	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(rare)
	if eventConfig == nil || eventConfig.UpTreasure <= 0 {
		return ""
	}

	for _, t := range ts {
		if t.Rare == eventConfig.UpTreasure {
			return t.ID
		}
	}
	return ""
}

func (te *upRareEventSt) trigger(t *treasureSt) string {
	tdata := t.getGameData()
	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(tdata.Rare)
	if eventConfig == nil || eventConfig.UpTreasure <= 0 {
		return ""
	}
	if !te.canTrigger() {
		return ""
	}

	upRareTreasureModelID := te.genUpRareTreasureModelID(tdata.Rare)
	if upRareTreasureModelID != "" {
		te.attr.SetUInt32("treasureID", t.getID())
		te.attr.SetStr("modelID", upRareTreasureModelID)
		te.attr.SetInt("oldRare", tdata.Rare)
	}

	return upRareTreasureModelID
}

func (te *upRareEventSt) doAction() (*pb.Treasure, error) {
	upRareTreasureID := te.attr.GetUInt32("treasureID")
	upRareTreasureModelID := te.attr.GetStr("modelID")
	oldRare := te.attr.GetInt("oldRare")
	if upRareTreasureModelID == "" || oldRare <= 0 {
		return nil, gamedata.GameError(1)
	}

	t := te.cpt.getTreasureByID(upRareTreasureID)
	if t == nil {
		return nil, gamedata.GameError(2)
	}

	eventConfig := gamedata.GetGameData(consts.TreasureEvent).(*gamedata.TreasureEventGameData).GetConfigByRare(oldRare)
	if eventConfig == nil {
		return nil, gamedata.GameError(4)
	}

	if !te.player.HasBowlder(eventConfig.UpRarePrice) {
		return nil, gamedata.GameError(3)
	}
	te.player.SubBowlder(eventConfig.UpRarePrice, consts.RmrTreasureUpRare)

	te.attr.Del("treasureID")
	te.attr.Del("modelID")
	te.attr.Del("oldRare")
	t.setModelID(upRareTreasureModelID)

	module.Shop.LogShopBuyItem(te.player, "tUpRare", "宝箱升品质", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), eventConfig.UpRarePrice,
		fmt.Sprintf("treasureID=%d, modelID=%s", t.getID(), upRareTreasureModelID))

	return t.packMsg(te.player).(*pb.Treasure), nil
}
