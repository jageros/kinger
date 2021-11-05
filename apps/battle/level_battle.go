package main

import (
	"kinger/gopuppy/common"
	"kinger/proto/pb"
	"kinger/gamedata"
	"kinger/common/consts"
	"kinger/gopuppy/common/glog"
	"kinger/common/utils"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/apps/logic"
)

type levelBattle struct {
	battle
	levelData *gamedata.Level
	fighterData *pb.FighterData
	needChooseCardAmount int
	canChooseCard map[uint32]*fightCard  // map[gcardID]*fightCard
}

func newLevelBattle(battleID common.UUid, fighterData *pb.FighterData, levelData *gamedata.Level, isHelp bool) *levelBattle {
	battleObj := &levelBattle{
		levelData: levelData,
		fighterData: fighterData,
	}
	battleObj.battleID = battleID
	battleObj.levelID = levelData.ID
	battleObj.needVideo = true
	if isHelp {
		battleObj.battleType = consts.BtLevelHelp
	} else {
		battleObj.battleType = consts.BtLevel
	}
	battleObj.i = battleObj

	poolGameData := bService.getPoolGameData()
	for i, gcardID := range levelData.OwnHand {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData == nil {
			glog.Errorf("newLevelBattle OwnHand no card %d", gcardID)
			continue
		}
		if i == 0 && fighterData.Camp == 0 {
			fighterData.Camp = int32(cardData.Camp)
		}
		fighterData.HandCards = append(fighterData.HandCards, &pb.SkinGCard{
			GCardID: gcardID,
		})
	}

	ownGridCards := levelData.GetOwnGridCards()
	for gridID, gcardID := range ownGridCards {
		fighterData.GridCards = append(fighterData.GridCards, &pb.InGridCard{
			GridID: int32(gridID),
			GCardID: gcardID,
		})
	}

	fighterData2 := &pb.FighterData{
		Uid: 1,
		NameText: int32(levelData.Name),
		IsRobot: true,
	}

	var camp int
	for i, gcardID := range levelData.EnemyHand {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData == nil {
			glog.Errorf("newLevelBattle enemyHandCards no card %d", gcardID)
			continue
		}
		if i == 0 {
			camp = cardData.Camp
		}
		fighterData2.HandCards = append(fighterData2.HandCards, &pb.SkinGCard{
			GCardID: gcardID,
		})
	}
	fighterData2.Camp = int32(camp)

	enemyGridCards := levelData.GetEnemyGridCards()
	for gridID, gcardID := range enemyGridCards {
		fighterData2.GridCards = append(fighterData2.GridCards, &pb.InGridCard{
			GridID: int32(gridID),
			GCardID: gcardID,
		})
	}

	battleObj.situation = newBattleSituation(fighterData, fighterData2, battleID, levelData.Offensive, 0,
		consts.BtScale33, 0, false, battleObj.battleType, nil)
	battleObj.initCanChooseCard(levelData, fighterData)
	battleObj.situation.fighter2.initDrawCard(battleObj.battleType, fighterData2)
	battleObj.situation.fighter1.agent.SetDispatchApp(consts.AppBattle, bService.AppID)
	return battleObj
}

func (b *levelBattle) initCanChooseCard(levelData *gamedata.Level, fighterData *pb.FighterData) {
	myHandCards := levelData.OwnHand
	enemyHandCards := levelData.EnemyHand
	poolGameData := bService.getPoolGameData()
	forbidCards := common.UInt32Set{}
	myGridCards := levelData.GetOwnGridCards()
	enemyGridCards := levelData.GetEnemyGridCards()
	for _, gcardID := range myHandCards {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData != nil {
			forbidCards.Add(cardData.CardID)
		}
	}
	for _, gcardID := range enemyHandCards {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData != nil {
			forbidCards.Add(cardData.CardID)
		}
	}
	for _, gcardID := range myGridCards {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData != nil {
			forbidCards.Add(cardData.CardID)
		}
	}
	for _, gcardID := range enemyGridCards {
		cardData := poolGameData.GetCardByGid(gcardID)
		if cardData != nil {
			forbidCards.Add(cardData.CardID)
		}
	}

	b.canChooseCard = map[uint32]*fightCard{}
	b.needChooseCardAmount = b.situation.getMaxHandCardAmount() - len(myHandCards)
	if b.needChooseCardAmount < 0 {
		b.needChooseCardAmount = 0
	}

	for _, c := range fighterData.DrawCardPool {
		cardData := poolGameData.GetCardByGid(c.GCardID)
		if cardData == nil {
			continue
		}
		if forbidCards.Contains(cardData.CardID) {
			continue
		}

		if len(levelData.CampsLimit) > 0 {
			ok := false
			for _, camp := range levelData.CampsLimit {
				if cardData.Camp == camp {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}

		c := newCardByData(b.situation.genObjID(), cardData, c.Skin, c.Equip, b.situation)
		b.canChooseCard[cardData.GCardID] = c
	}
}

func (b *levelBattle) syncReadyFight() {
	msg := b.packMsg()
	fighter1 := b.situation.getFighter1()
	fighter1.agent.PushClient(pb.MessageID_S2C_READY_LEVEL_FIGHT, msg)
}

func (b *levelBattle) packMsg() interface{} {
	battleMsg := b.battle.packMsg().(*pb.FightDesk)
	msg := &pb.LevelBattle{
		Desk:          battleMsg,
		NeedChooseNum: int32(b.needChooseCardAmount),
		LevelID: int32(b.levelData.ID),
	}
	for _, c := range b.canChooseCard {
		msg.ChoiceCards = append(msg.ChoiceCards, c.packMsg())
	}
	return msg
}

func (b *levelBattle) readyDone(cards ...uint32) error {
	if len(cards) != b.needChooseCardAmount {
		return gamedata.GameError(2)
	}

	f := b.getSituation().getFighter1()
	for _, gcardID := range cards {
		card, ok := b.canChooseCard[gcardID]
		if !ok {
			return gamedata.GameError(3)
		}
		f.addHandCard(card)
	}

	f.initDrawCard(b.battleType, b.fighterData)
	b.fighterData = nil

	b.situation.state = bsWaitClient
	f.setInitialHand()
	b.situation.fighter2.setInitialHand()

	if b.needVideo {
		b.videoBegin()
	}
	b.boutReadyDone(b.situation.fighter2)
	b.boutReadyDone(f)

	utils.PlayerMqPublish(f.getUid(), pb.RmqType_BattleBegin, &pb.RmqBattleBegin{
		BattleID:   uint64(b.battleID),
		AppID:      bService.AppID,
		BattleType: int32(b.battleType),
	})

	return nil
}

func (b *levelBattle) packAttr() *attribute.AttrMgr {
	attr := b.battle.packAttr()
	attr.SetInt("levelID", b.levelData.ID)
	return attr
}

func (b *levelBattle) packRestoredMsg() *pb.RestoredFightDesk {
	msg := b.battle.packRestoredMsg()
	msg.LevelID = int32(b.levelData.ID)
	return msg
}

func (b *levelBattle) restoredFromAttr(attr *attribute.AttrMgr, agent *logic.PlayerAgent) {
	(&b.battle).restoredFromAttr(attr, agent)
	b.i = b
	levelID := attr.GetInt("levelID")
	b.levelData = gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData).GetLevelData(levelID)
	b.levelID = levelID
}
