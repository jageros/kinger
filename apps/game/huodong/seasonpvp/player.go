package seasonpvp

import (
	htypes "kinger/apps/game/huodong/types"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/proto/pb"
	"math/rand"
)

type seasonPvpHdPlayerData struct {
	htypes.BaseHdPlayerData
}

func (hpd *seasonPvpHdPlayerData) GetAttr() *attribute.MapAttr {
	return hpd.Attr
}

func (hpd *seasonPvpHdPlayerData) isReward() bool {
	return hpd.Attr.GetBool("reward")
}

func (hpd *seasonPvpHdPlayerData) markReward() {
	hpd.Attr.SetBool("reward", true)
}

func (hpd *seasonPvpHdPlayerData) getSession() int {
	return hpd.Attr.GetInt("session")
}

func (hpd *seasonPvpHdPlayerData) isResetLevel() bool {
	return hpd.Attr.GetBool("resetLevel")
}

func (hpd *seasonPvpHdPlayerData) markResetLevel() {
	hpd.Attr.SetBool("resetLevel", true)
}

func (hpd *seasonPvpHdPlayerData) beginSession(session int) {
	if session == hpd.getSession() {
		return
	}
	hpd.Attr.SetInt("session", session)
	hpd.Attr.SetBool("resetLevel", false)
	hpd.Attr.SetBool("reward", false)
	hpd.Attr.SetInt("firstHandAmount", 0)
	hpd.Attr.SetInt("backHandAmount", 0)
	hpd.Attr.SetInt("firstHandWinAmount", 0)
	hpd.Attr.SetInt("backHandWinAmount", 0)
	hpd.Attr.SetInt("handPro", 0)
	hpd.Attr.SetInt("handWinCnt", 0)
	hpd.Attr.Del("handCards")
	hpd.Attr.SetInt("camp", 0)
	hpd.Attr.Del("chooseCards2")
	hpd.Attr.SetBool("quit", false)
	hpd.Attr.SetInt("freeRefreshChooseCardCnt", 0)
	hpd.Attr.SetInt("jadeRefreshChooseCardCnt", 0)
	if !config.GetConfig().IsMultiLan {
		hpd.Player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).SetResource(consts.WinDiff, 0)
	}
}

func (hpd *seasonPvpHdPlayerData) isQuit() bool {
	return hpd.Attr.GetBool("quit")
}

func (hpd *seasonPvpHdPlayerData) quit(seasonData gamedata.ISeasonPvp) {
	hpd.Attr.SetBool("quit", true)
	hpd.delChooseCards(seasonData, []uint32{})
	hpd.delHandCards()
	agent := hpd.Player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SEASON_PVP_STOP, nil)
	}
}

func (hpd *seasonPvpHdPlayerData) join() {
	hpd.Attr.SetBool("quit", false)
}

func (hpd *seasonPvpHdPlayerData) getFirstHandAmount() int {
	return hpd.Attr.GetInt("firstHandAmount")
}

func (hpd *seasonPvpHdPlayerData) getBackHandAmount() int {
	return hpd.Attr.GetInt("backHandAmount")
}

func (hpd *seasonPvpHdPlayerData) getFirstHandWinAmount() int {
	return hpd.Attr.GetInt("firstHandWinAmount")
}

func (hpd *seasonPvpHdPlayerData) getBackHandWinAmount() int {
	return hpd.Attr.GetInt("backHandWinAmount")
}

func (hpd *seasonPvpHdPlayerData) getWinAmount() int {
	return hpd.getFirstHandWinAmount() + hpd.getBackHandWinAmount()
}

func (hpd *seasonPvpHdPlayerData) onPvpEnd(isWin, isFirstHand bool, seasonData gamedata.ISeasonPvp) {
	if isFirstHand {
		hpd.Attr.SetInt("firstHandAmount", hpd.Attr.GetInt("firstHandAmount")+1)
		if isWin {
			hpd.Attr.SetInt("firstHandWinAmount", hpd.Attr.GetInt("firstHandWinAmount")+1)
		}
	} else {
		hpd.Attr.SetInt("backHandAmount", hpd.Attr.GetInt("backHandAmount")+1)
		if isWin {
			hpd.Attr.SetInt("backHandWinAmount", hpd.Attr.GetInt("backHandWinAmount")+1)
		}
	}

	if isWin {
		hpd.Attr.SetInt("handWinCnt", hpd.getHandCardWinCnt()+1)
	}

	if !config.GetConfig().IsMultiLan {
		if isWin {
			module.Player.ModifyResource(hpd.Player, consts.WinDiff, 1)
		} else {
			module.Player.ModifyResource(hpd.Player, consts.WinDiff, -1)
		}
	}

	handPro := hpd.getHandCardCurPro()
	changeHandType := seasonData.GetChangeHandType()
	if len(changeHandType) < 2 {
		return
	}

	maxPro := changeHandType[1]
	switch pb.FetchSeasonHandCardReply_ChangeTypeEnum(changeHandType[0]) {
	case pb.FetchSeasonHandCardReply_Fight:
		handPro++
	case pb.FetchSeasonHandCardReply_Win:
		if isWin {
			handPro++
		}
	case pb.FetchSeasonHandCardReply_Lose:
		if !isWin {
			handPro++
		}
	}

	if handPro >= maxPro {
		hpd.timeToRefreshHandCard()
	} else {
		hpd.Attr.SetInt("handPro", handPro)
	}
}

func (hpd *seasonPvpHdPlayerData) timeToRefreshHandCard() {
	hpd.Attr.SetInt("handPro", 0)
	hpd.delHandCards()
	agent := hpd.Player.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_SEASON_PVP_CHANGE_HAND_CARD, nil)
	}
}

func (hpd *seasonPvpHdPlayerData) getHandCardCurPro() int {
	return hpd.Attr.GetInt("handPro")
}

func (hpd *seasonPvpHdPlayerData) getHandCardWinCnt() int {
	return hpd.Attr.GetInt("handWinCnt")
}

func (hpd *seasonPvpHdPlayerData) getHandCards() []uint32 {
	var handCards []uint32
	handCardsAttr := hpd.Attr.GetListAttr("handCards")
	if handCardsAttr != nil {
		cardCpt := hpd.Player.GetComponent(consts.CardCpt).(types.ICardComponent)
		handCardsAttr.ForEachIndex(func(index int) bool {
			cardID := handCardsAttr.GetUInt32(index)
			card := cardCpt.GetCollectCard(cardID)
			if card != nil && card.GetState() != pb.CardState_InCampaignMs {
				handCards = append(handCards, cardID)
			} else {
				handCards = []uint32{}
				hpd.timeToRefreshHandCard()
				return false
			}
			return true
		})
	}
	return handCards
}

func (hpd *seasonPvpHdPlayerData) delHandCards() {
	handCardsAttr := hpd.Attr.GetListAttr("handCards")
	var cardIDs []uint32
	if handCardsAttr != nil {
		handCardsAttr.ForEachIndex(func(index int) bool {
			cardIDs = append(cardIDs, handCardsAttr.GetUInt32(index))
			return true
		})
		module.Card.SetCardsState(hpd.Player, cardIDs, pb.CardState_NormalCState)
	}
	hpd.Attr.Del("handCards")
}

func (hpd *seasonPvpHdPlayerData) getCamp() int {
	camp := hpd.Attr.GetInt("camp")
	if camp == 0 {
		return consts.Wei
	}
	return camp
}

func (hpd *seasonPvpHdPlayerData) chooseCamp(camp int) {
	hpd.Attr.SetInt("camp", camp)
}

func (hpd *seasonPvpHdPlayerData) getFreeRefreshChooseCardCnt() int {
	return hpd.Attr.GetInt("freeRefreshChooseCardCnt")
}

func (hpd *seasonPvpHdPlayerData) setFreeRefreshChooseCardCnt(cnt int) {
	hpd.Attr.SetInt("freeRefreshChooseCardCnt", cnt)
}

func (hpd *seasonPvpHdPlayerData) getJadeRefreshChooseCardCnt() int {
	return hpd.Attr.GetInt("jadeRefreshChooseCardCnt")
}

func (hpd *seasonPvpHdPlayerData) setJadeRefreshChooseCardCnt(cnt int) {
	hpd.Attr.SetInt("jadeRefreshChooseCardCnt", cnt)
}

func (hpd *seasonPvpHdPlayerData) getShuffledColllectCards(camp int) []types.ICollectCard {
	allCollectCards := module.Card.GetAllCollectCardsByCamp(hpd.Player, []int{camp, consts.Heroes})
	for i := range allCollectCards {
		j := rand.Intn(i + 1)
		allCollectCards[i], allCollectCards[j] = allCollectCards[j], allCollectCards[i]
	}
	return allCollectCards
}

func (hpd *seasonPvpHdPlayerData) getChooseCardData(seasonData gamedata.ISeasonPvp) *pb.SeasonPvpChooseCardData {
	if len(seasonData.GetHandCardType()) < 3 {
		return nil
	}
	var cards []uint32
	chooseCardsAttr := hpd.Attr.GetListAttr("chooseCards2")
	if chooseCardsAttr != nil {
		chooseCardsAttr.ForEachIndex(func(index int) bool {
			cards = append(cards, chooseCardsAttr.GetUInt32(index))
			return true
		})
	}
	if len(cards) <= 0 {
		return nil
	}

	return &pb.SeasonPvpChooseCardData{
		CardIDs:          cards,
		NeedChooseAmount: int32(seasonData.GetHandCardType()[2]),
		FreeRefreshCnt:   int32(hpd.getFreeRefreshChooseCardCnt()),
		JadeRefreshCnt:   int32(hpd.getJadeRefreshChooseCardCnt()),
	}
}

func (hpd *seasonPvpHdPlayerData) randomChooseCards(seasonData gamedata.ISeasonPvp) *pb.SeasonPvpChooseCardData {
	chooseCardsAttr := hpd.Attr.GetListAttr("chooseCards2")
	if chooseCardsAttr == nil {
		hpd.refreshChooseCards(seasonData)
		hpd.setFreeRefreshChooseCardCnt(1)
		if config.GetConfig().IsMultiLan {
			hpd.setJadeRefreshChooseCardCnt(1)
		} else {
			hpd.setJadeRefreshChooseCardCnt(2)
		}
	}
	return hpd.getChooseCardData(seasonData)
}

func (hpd *seasonPvpHdPlayerData) delChooseCards(seasonData gamedata.ISeasonPvp, handCards []uint32) {
	chooseCardData := hpd.getChooseCardData(seasonData)
	if chooseCardData == nil {
		return
	}

	var cardIDs []uint32
L:
	for _, cardID := range chooseCardData.CardIDs {
		for _, cardID2 := range handCards {
			if cardID == cardID2 {
				continue L
			}
		}
		cardIDs = append(cardIDs, cardID)
	}

	module.Card.SetCardsState(hpd.Player, cardIDs, pb.CardState_NormalCState)
	hpd.Attr.Del("chooseCards2")
}

func (hpd *seasonPvpHdPlayerData) refreshChooseCards(seasonData gamedata.ISeasonPvp) {
	hpd.delChooseCards(seasonData, []uint32{})
	chooseCardsAttr := attribute.NewListAttr()
	hpd.Attr.SetListAttr("chooseCards2", chooseCardsAttr)
	camp := hpd.getCamp()
	allCollectCards := hpd.getShuffledColllectCards(camp)

	amount := seasonData.GetHandCardType()[1]
	var cardIDs []uint32
	for _, card := range allCollectCards {
		if amount <= 0 {
			break
		}
		if card.GetState() == pb.CardState_InCampaignMs {
			continue
		}

		cardID := card.GetCardID()
		chooseCardsAttr.AppendUInt32(cardID)
		cardIDs = append(cardIDs, cardID)
		amount--
	}
	module.Card.SetCardsState(hpd.Player, cardIDs, pb.CardState_InSeasonPvp)
}

func (hpd *seasonPvpHdPlayerData) chooseHandCards(seasonData gamedata.ISeasonPvp, cardIDs []uint32) ([]uint32, error) {
	amount := len(cardIDs)
	if amount < seasonData.GetHandCardType()[2] {
		return []uint32{}, gamedata.GameError(1)
	}

	chooseCardData := hpd.randomChooseCards(seasonData)
	if chooseCardData == nil {
		return []uint32{}, gamedata.GameError(3)
	}

	handCardsAttr := attribute.NewListAttr()
	for _, cardID := range cardIDs {
		ok := false
		for _, cid := range chooseCardData.CardIDs {
			if cardID == cid {
				ok = true
				break
			}
		}

		if !ok {
			return []uint32{}, gamedata.GameError(2)
		}

		handCardsAttr.AppendUInt32(cardID)
	}

	var randCards []uint32
	if amount < 5 {
		needAmount := 5 - amount
		camp := hpd.getCamp()
		allCollectCards := hpd.getShuffledColllectCards(camp)
	L:
		for _, card := range allCollectCards {
			if needAmount <= 0 {
				break
			}
			if card.GetState() == pb.CardState_InCampaignMs {
				continue
			}

			cardID := card.GetCardID()
			for _, cid := range cardIDs {
				if cardID == cid {
					continue L
				}
			}
			randCards = append(randCards, card.GetCardID())

			needAmount--
		}
	}

	for _, cardID := range randCards {
		handCardsAttr.AppendUInt32(cardID)
	}

	var allCardIDs []uint32
	handCardsAttr.ForEachIndex(func(index int) bool {
		allCardIDs = append(allCardIDs, handCardsAttr.GetUInt32(index))
		return true
	})

	hpd.delHandCards()
	hpd.Attr.SetListAttr("handCards", handCardsAttr)
	hpd.Attr.SetInt("handPro", 0)
	hpd.Attr.SetInt("handWinCnt", 0)
	hpd.delChooseCards(seasonData, []uint32{})
	module.Card.SetCardsState(hpd.Player, allCardIDs, pb.CardState_InSeasonPvp)

	return randCards, nil
}
