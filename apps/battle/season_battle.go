package main

import (
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"kinger/gamedata"
	"math/rand"
	"kinger/common/consts"
)

type seasonBattle struct {
	battle
	handType pb.BattleHandType
	fighter1HandCards []*pb.SkinGCard
	fighter2HandCards []*pb.SkinGCard
	randomArg *pb.SeasonRandomHand
	banArg *seasonBanHand
	switchArg *pb.SeasonSwitchHand
}

func newSeasonBattle(battleID common.UUid, battleType int, fighterData1, fighterData2 *pb.FighterData, upperType, bonusType,
	scale, battleRes int, needVideo bool, needFortifications bool, seasonPvpData *gamedata.SeasonPvp) *seasonBattle {

	battleObj := &seasonBattle {
		handType: pb.BattleHandType_Default,
	}
	battleObj.battleID = battleID
	battleObj.battleType = battleType
	battleObj.needVideo = needVideo
	battleObj.i = battleObj
	if len(seasonPvpData.HandCardType) > 0 && pb.BattleHandType(seasonPvpData.HandCardType[0]) != pb.BattleHandType_UnknowType {
		battleObj.handType = pb.BattleHandType(seasonPvpData.HandCardType[0])
	}
	battleObj.situation = newBattleSituation(fighterData1, fighterData2, battleID, upperType, bonusType, scale, battleRes,
		needFortifications, battleType, seasonPvpData)
	return battleObj
}

func (b *seasonBattle) createHandArg(fighterData1, fighterData2 *pb.FighterData, seasonPvpData *gamedata.SeasonPvp) {
	switch b.handType {
	case pb.BattleHandType_Random:
	case pb.BattleHandType_Ban:
		b.fighter1HandCards = fighterData1.HandCards
		b.fighter2HandCards = fighterData2.HandCards
		b.banArg = &seasonBanHand{}
		b.banArg.ban(fighterData1, fighterData2, b.fighter1HandCards, b.fighter2HandCards, seasonPvpData)
	case pb.BattleHandType_Switch:
		b.fighter1HandCards = fighterData1.HandCards
		b.fighter2HandCards = fighterData2.HandCards
	default:
		return
	}
}

func (b *seasonBattle) readyDone(cards ...uint32) error {
	if b.handType == pb.BattleHandType_Default {
		return b.battle.readyDone()
	}

	b.situation.setState(bsWaitSeason)
	f1 := b.situation.getFighter1()
	f2 := b.situation.getFighter2()
	b.syncReadyFight()

	if f1.isRobot {
		b.boutReadyDone(f1)
	}
	if f2.isRobot {
		b.boutReadyDone(f2)
	}
	return nil
}

func (b *seasonBattle) packSeasonMsg(isFighter1 bool) *pb.SeasonBattle {
	msg := &pb.SeasonBattle{
		Battle: b.packMsg().(*pb.FightDesk),
		HandType: b.handType,
	}

	if isFighter1 {
		msg.MyHandCards = b.fighter1HandCards
	} else {
		msg.MyHandCards = b.fighter2HandCards
	}

	switch b.handType {
	case pb.BattleHandType_Random:
		msg.Arg, _ = b.randomArg.Marshal()

	case pb.BattleHandType_Ban:
		arg := &pb.SeasonBanHand{}
		if isFighter1 {
			arg.BanEnemyCardIdxs = b.banArg.f2BanCardIdxs
			arg.BanMyCardIdxs = b.banArg.f1BanCardIdxs
			arg.EnemyBanCards = b.banArg.f2RandomCards
			arg.MyRandomCards = b.banArg.f1RandomCards
		} else {
			arg.BanEnemyCardIdxs = b.banArg.f1BanCardIdxs
			arg.BanMyCardIdxs = b.banArg.f2BanCardIdxs
			arg.EnemyBanCards = b.banArg.f1RandomCards
			arg.MyRandomCards = b.banArg.f2RandomCards
		}
		msg.Arg, _ = arg.Marshal()

	case pb.BattleHandType_Switch:
		msg.Arg, _ = b.switchArg.Marshal()

	}
	return msg
}

func (b *seasonBattle) syncReadyFight() {
	if b.handType == pb.BattleHandType_Default {
		b.battle.syncReadyFight()
		return
	}

	fighter1 := b.situation.getFighter1()
	fighter2 := b.situation.getFighter2()
	if !fighter1.isRobot && fighter1.agent != nil {
		fighter1.agent.PushClient(pb.MessageID_S2C_BEGIN_SEASON_BATTLE, b.packSeasonMsg(true))
	}
	if !fighter2.isRobot && fighter2.agent != nil {
		fighter2.agent.PushClient(pb.MessageID_S2C_BEGIN_SEASON_BATTLE, b.packSeasonMsg(false))
	}


}

type seasonBanHand struct {
	f1BanCardIdxs []int32
	f2BanCardIdxs []int32
	f1RandomCards []*pb.SkinGCard
	f2RandomCards []*pb.SkinGCard
}

func (sbh *seasonBanHand) ban(fighterData1, fighterData2 *pb.FighterData, f1HandCards, f2HandCards []*pb.SkinGCard,
	seasonPvpData *gamedata.SeasonPvp) {

	arg := seasonPvpData.HandCardType[1]
	fighterDatas := []*pb.FighterData{fighterData1, fighterData2}
	handCards := [][]*pb.SkinGCard{f1HandCards, f2HandCards}
	banCardIdxs := [][]int32{[]int32{}, []int32{}}
	randomCards := [][]*pb.SkinGCard{[]*pb.SkinGCard{}, []*pb.SkinGCard{}}
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)

	for i, fighterData := range fighterDatas {
		hand := handCards[i]
		banIdxs := banCardIdxs[i]
		randomCard := randomCards[i]
		amount := arg
		handAmount := len(hand)
		if amount > handAmount {
			amount = handAmount
		}
		handMinLevel := 10000
		for idx, card := range hand {
			if rand.Intn(handAmount) < amount {
				banIdxs = append(banIdxs, int32(idx))
				amount--
			}
			handAmount--
			cardData := poolGameData.GetCardByGid(card.GCardID)
			if cardData != nil && cardData.Level > 0 && cardData.Level < handMinLevel {
				handMinLevel = cardData.Level
			}
		}

		amount = len(banIdxs)
		for a := range fighterData.DrawCardPool {
			b := rand.Intn(a + 1)
			fighterData.DrawCardPool[a], fighterData.DrawCardPool[b] = fighterData.DrawCardPool[b], fighterData.DrawCardPool[a]
		}

	L:	for _, card := range fighterData.DrawCardPool {
			if amount <= 0 {
				break
			}

			for _, c := range hand {
				if card.GCardID == c.GCardID {
					continue L
				}
			}

			cardData := poolGameData.GetCardByGid(card.GCardID)
			if cardData != nil {
				if fighterData.IsRobot && cardData.Level > handMinLevel {
					continue
				}
				randomCard = append(randomCard, card)
				amount--
			}
		}

		banCardIdxs[i] = banIdxs
		randomCards[i] = randomCard
	}

	sbh.f1BanCardIdxs = banCardIdxs[0]
	sbh.f2BanCardIdxs = banCardIdxs[1]
	sbh.f1RandomCards = randomCards[0]
	sbh.f2RandomCards = randomCards[1]
}
