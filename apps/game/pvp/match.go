package pvp

import (
	"github.com/gogo/protobuf/proto"
	"kinger/apps/game/module/types"
	"kinger/proto/pb"
	"kinger/common/consts"
	"fmt"
	"kinger/apps/game/module"
	"kinger/gamedata"
	"kinger/common/aicardpool"
)

var (
	guideMatchStrategy iMatchStrategy = &guideMatchStrategySt{}
	newbiePvpMatchStrategy iMatchStrategy = &newbiePvpMatchStrategySt{}
	pvpMatchStrategy iMatchStrategy = &pvpMatchStrategySt{}
)

type iMatchStrategy interface {
	getMatchMessageID() pb.MessageID
	packMatchArgAndReply(player types.IPlayer, camp int) (proto.Marshaler, *pb.MatchReply, error)
}

// 新手的5场战斗
type guideMatchStrategySt struct {

}

func (ms *guideMatchStrategySt) getMatchMessageID() pb.MessageID {
	return pb.MessageID_G2M_BEGIN_GUIDE_MATCH
}

func (ms *guideMatchStrategySt) packMatchArgAndReply(player types.IPlayer, camp int) (proto.Marshaler, *pb.MatchReply, error) {
	module.Player.LogMission(player, fmt.Sprintf("guideBattle_%d",
		module.Player.GetResource(player, consts.GuidePro)+1), 1)
	return player.GetComponent(consts.TutorialCpt).(types.ITutorialComponent).PackBeginBattleArg(),
		&pb.MatchReply{}, nil
}

type iPvpMatchStrategy interface {
	iMatchStrategy
}

type basePvpMatchStrategySt struct {

}

func (ms *basePvpMatchStrategySt) getDrawCards(player types.IPlayer) []*pb.SkinGCard {
	var drawCards []*pb.SkinGCard
	allCollectCards := module.Card.GetAllCollectCards(player)
	for _, card := range allCollectCards {
		drawCards = append(drawCards, &pb.SkinGCard{
			GCardID: card.GetCardGameData().GCardID,
			Skin:    card.GetSkin(),
			Equip: card.GetEquip(),
		})
	}
	return drawCards
}

func (ms *basePvpMatchStrategySt) getWinRate(player types.IPlayer) int32 {
	var winRate int32 = 100
	totalBattleCnt := player.GetFirstHandAmount() + player.GetBackHandAmount()
	if totalBattleCnt > 0 {
		winRate = int32( float64(player.GetFirstHandWinAmount() + player.GetBackHandWinAmount()) / float64(totalBattleCnt) * 100 )
	}
	return winRate
}

func (ms *basePvpMatchStrategySt) getHandCards(player types.IPlayer, handCardsMsg *pb.FetchSeasonHandCardReply, camp int) (
	[]*pb.SkinGCard, error) {

	var handCards []*pb.SkinGCard
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)

	if handCardsMsg != nil {

		for _, cardID := range handCardsMsg.CardIDs {
			card := cardCpt.GetCollectCard(cardID)
			handCards = append(handCards, &pb.SkinGCard{
				GCardID: card.GetCardGameData().GCardID,
				Skin:    card.GetSkin(),
				Equip:   card.GetEquip(),
			})
		}

	} else {

		handCards = cardCpt.CreatePvpHandCards(camp)
		if len(handCards) < consts.MaxHandCardAmount {
			return handCards, gamedata.GameError(1)
		}
	}

	return handCards, nil
}

type newbiePvpMatchStrategySt struct {
	basePvpMatchStrategySt
}

func (ms *newbiePvpMatchStrategySt) getMatchMessageID() pb.MessageID {
	return pb.MessageID_G2M_BEGIN_NEWBIE_PVP_MATCH
}

func (ms *newbiePvpMatchStrategySt) packMatchArgAndReply(player types.IPlayer, camp int) (proto.Marshaler, *pb.MatchReply, error) {

	handCards, err := ms.getHandCards(player, nil, camp)
	if err != nil {
		return nil, nil, err
	}

	pvpCpt := player.GetComponent(consts.PvpCpt).(*pvpComponent)
	newbiePvpCamp, isFirstPvpBattle := pvpCpt.getNewbiePvpEnemyCamp()

	return &pb.BeginNewbiePvpMatchArg{
		Name:          player.GetName(),
		Camp:          int32(camp),
		PvpScore:      int32(player.GetPvpScore()),
		Mmr:           int32(module.Player.GetResource(player, consts.Mmr)),
		EnemyCamp:     int32(newbiePvpCamp),
		HandCards:     handCards,
		DrawCards:     ms.getDrawCards(player),
		IsFirstBattle: isFirstPvpBattle,
		HeadImgUrl:    player.GetHeadImgUrl(),
		HeadFrame:     player.GetHeadFrame(),
		WinRate: ms.getWinRate(player),
		Area: int32(player.GetArea()),
		CountryFlag: player.GetCountryFlag(),
	}, &pb.MatchReply{}, nil
}

type pvpMatchStrategySt struct {
	basePvpMatchStrategySt
}

func (ms *pvpMatchStrategySt) getMatchMessageID() pb.MessageID {
	return pb.MessageID_G2M_BEGIN_MATCH
}

func (ms *pvpMatchStrategySt) packMatchArgAndReply(player types.IPlayer, camp int) (proto.Marshaler, *pb.MatchReply, error) {
	if !mod.CanPvpMatch(player) {
		return nil, nil, gamedata.GameError(100)
	}

	seasonData, seasonPvpCamp, handCardType, chooseCardData, handCardsMsg := module.Huodong.GetSeasonPvpHandCardInfo(player)
	var seasonDataID int
	if seasonData != nil {
		seasonDataID = seasonData.GetID()
	}

	if handCardType == pb.BattleHandType_Random {
		// 锦标赛

		if handCardsMsg == nil {
			// 锦标赛还没选牌
			return nil, &pb.MatchReply{
				NeedChooseCamp: true,
				LastCamp: int32(seasonPvpCamp),
				ChooseCardData: chooseCardData,
			}, nil
		}
		camp = seasonPvpCamp
	} else {
		// 正常pvp
		handCardsMsg = nil
	}

	handCards, err := ms.getHandCards(player, handCardsMsg, camp)
	if err != nil {
		return nil, nil, err
	}

	var cardStrength int
	handCardIDs := make([]uint32, len(handCards))
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for i, c := range handCards {
		cardData := poolGameData.GetCardByGid(c.GCardID)
		cardStrength += cardData.Strength
		handCardIDs[i] = cardData.CardID
	}

	if seasonDataID <= 0 {
		aicardpool.AddCardPool(player.GetPvpLevel(), camp, handCardIDs)
	}

	pvpCpt := player.GetComponent(consts.PvpCpt).(*pvpComponent)
	return &pb.BeginMatchArg{
		Name:             player.GetName(),
		Camp:             int32(camp),
		PvpScore:         int32(player.GetPvpScore()),
		Mmr:              int32(module.Player.GetResource(player, consts.Mmr)),
		StreakLoseCnt:    int32(pvpCpt.getStreakLoseCnt()),
		HandCards:        handCards,
		DrawCards:        ms.getDrawCards(player),
		CardStrength:     int32(cardStrength),
		HeadImgUrl:       player.GetHeadImgUrl(),
		HeadFrame:        player.GetHeadFrame(),
		SeasonDataID: int32(seasonDataID),
		RebornCnt: int32(module.Reborn.GetRebornCnt(player)),
		WinRate: ms.getWinRate(player),
		Area: int32(player.GetArea()),
		StreakWinCnt: int32(pvpCpt.getStreakWinCnt()),
		LastOppUid: uint64(player.GetLastOpponent()),
		RechargeMatchIndex: int32(pvpCpt.getRechargeMatchIndex()),
		CountryFlag: player.GetCountryFlag(),
		MatchScore: int32(module.Player.GetResource(player, consts.MatchScore)),
	}, &pb.MatchReply{}, nil
}
