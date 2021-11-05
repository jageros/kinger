package tutorial

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"fmt"
)

var _ types.ITutorialComponent = &tutorialComponent{}

type tutorialComponent struct {
	player types.IPlayer
	gdata  *gamedata.TutorialGameData
	attr   *attribute.MapAttr
}

func (tc *tutorialComponent) ComponentID() string {
	return consts.TutorialCpt
}

func (tc *tutorialComponent) GetPlayer() types.IPlayer {
	return tc.player
}

func (tc *tutorialComponent) OnInit(player types.IPlayer) {
	tc.player = player
	tc.gdata = gamedata.GetGameData(consts.Tutorial).(*gamedata.TutorialGameData)
}

func (tc *tutorialComponent) OnLogin(isRelogin, isRestore bool) {
}

func (tc *tutorialComponent) OnLogout() {
}

func (tc *tutorialComponent) GetCampID() int32 {
	campID := tc.attr.GetInt32("campID")
	return campID
}

func (tc *tutorialComponent) setCampID(campID_ int32) bool {
	if tc.attr.GetInt32("campID") > 0 {
		return false
	}

	campID := int(campID_)
	var cardIDs []uint32

	if campID == consts.Wei {
		cardIDs = []uint32{7, 9, 10, 62, 64}
	} else if campID == consts.Shu {
		cardIDs = []uint32{24, 25, 26, 62, 64}
	} else if campID == consts.Wu {
		cardIDs = []uint32{45, 49, 50, 62, 64}
	} else {
		return false
	}

	cardComponent := tc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardComponent.NewbieInitPvpCardPool(campID, cardIDs)
	tc.attr.SetInt32("campID", campID_)

	module.Mission.RefreashMission(tc.player)

	module.Player.LogMission(tc.player, fmt.Sprintf("newbieCamp_%d", campID_), 3)
	return true
}

func (tc *tutorialComponent) PackBeginBattleArg() *pb.BeginBattleArg {
	campID := int(tc.attr.GetInt32("campID"))
	if campID <= 0 {
		campID = consts.Wei
		tc.setCampID(int32(campID))
	}
	battles := tc.gdata.BattlesOfCamp[campID]

	if battles == nil {
		return nil
	}

	resCpt := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	battleID := resCpt.GetResource(consts.GuidePro) + 1
	battle := battles[int(battleID)]

	if battle == nil {
		return nil
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	var gridCards []*pb.InGridCard
	for _, gridCardInfo := range battle.OwnSide {
		cardData := poolGameData.GetCardByGid(uint32(gridCardInfo[0]))
		if cardData != nil {
			gridCards = append(gridCards, &pb.InGridCard{
				GCardID: uint32(gridCardInfo[0]),
				GridID:  int32(gridCardInfo[1]),
			})
		}
	}

	var robotGridCards []*pb.InGridCard
	for _, gridCardInfo := range battle.EnemySide {
		cardData := poolGameData.GetCardByGid(uint32(gridCardInfo[0]))
		if cardData != nil {
			robotGridCards = append(robotGridCards, &pb.InGridCard{
				GCardID: uint32(gridCardInfo[0]),
				GridID:  int32(gridCardInfo[1]),
			})
		}
	}

	var enemyHands []*pb.SkinGCard
	for _, gcardID := range battle.EnemyHand {
		enemyHands = append(enemyHands, &pb.SkinGCard{
			GCardID: gcardID,
		})
	}

	var winRate int32 = 100
	if battle.NormalAI <= 0 {
		totalBattleCnt := tc.player.GetFirstHandAmount() + tc.player.GetBackHandAmount()
		if totalBattleCnt > 0 {
			winRate = int32(float64(tc.player.GetFirstHandWinAmount()+tc.player.GetBackHandWinAmount()) /
				float64(totalBattleCnt) * 100)
		}
	}

	agent := tc.player.GetAgent()
	return &pb.BeginBattleArg{
		BattleType: int32(consts.BtGuide),
		UpperType:  int32(battle.Offensive),
		Fighter1: &pb.FighterData{
			Uid:       uint64(tc.player.GetUid()),
			ClientID:  uint64(agent.GetClientID()),
			GateID:    agent.GetGateID(),
			HandCards: tc.player.GetComponent(consts.CardCpt).(types.ICardComponent).CreatePvpHandCards(campID),
			Name:      tc.player.GetName(),
			Camp:      int32(campID),
			IsRobot:   false,
			GridCards: gridCards,
			HeadImgUrl: tc.player.GetHeadImgUrl(),
			HeadFrame: tc.player.GetHeadFrame(),
			WinRate: winRate,
			Region: agent.GetRegion(),
			CountryFlag: tc.player.GetCountryFlag(),
		},

		Fighter2: &pb.FighterData{
			Uid:       1,
			HandCards: enemyHands,
			Camp:      int32(battle.Country),
			IsRobot:   true,
			GridCards: robotGridCards,
			HeadImgUrl: battle.Head,
			HeadFrame: battle.HeadFrame,
			NameText: int32(battle.Name),
		},
	}
}

func (tc *tutorialComponent) startTutorialBattle() (interface{}, error) {
	arg := tc.PackBeginBattleArg()
	if arg == nil {
		return nil, gamedata.InternalErr
	}

	if tc.player.IsMultiRpcForbid(consts.FmcGuideBattle) {
		return nil, gamedata.GameError(1)
	}

	resCpt := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	battleID := resCpt.GetResource(consts.GuidePro) + 1
	module.Player.LogMission(tc.player, fmt.Sprintf("guideBattle_%d", battleID), 1)
	return logic.CallBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, arg)

	//return module.Battle.BeginBattle(consts.BtGuide, pd, rd, battle.Offensive, endHandler, nil, consts.BtScale33,
	//	0, false)
}

/*
func (tc *tutorialComponent) BeginGuideBattle() error {
	battleObj, err := tc.startTutorialBattle()
	if err == nil {
		tc.GetPlayer().GetAgent().PushClient(pb.MessageID_S2C_READY_FIGHT, battleObj.PackMsg())
		return nil
	} else {
		return err
	}
}
*/

func (tc *tutorialComponent) OnBattleEnd(fighterData *pb.EndFighterData, isWin bool) {
	if !isWin {
		tc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
			WinUid: 1,
		})
		return
	}

	guideCamp := tc.GetCampID()
	gdata := gamedata.GetGameData(consts.Tutorial).(*gamedata.TutorialGameData)
	battles := gdata.BattlesOfCamp[int(guideCamp)]
	resCpt := tc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	battleID := resCpt.GetResource(consts.GuidePro) + 1

	module.Player.LogMission(tc.player, fmt.Sprintf("guideBattle_%d", battleID), 2)

	if battles == nil {
		tc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
			WinUid: uint64(tc.player.GetUid()),
		})
		return
	}

	battle := battles[int(battleID)]
	treasureComponent := tc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent)
	msg := &pb.BattleResult{
		WinUid: uint64(tc.player.GetUid()),
	}

	for _, treasureID := range battle.TreasureReward {
		treasureComponent.AddRewardTreasureByID(treasureID, false)
		msg.TreasureID = []string{treasureID}
	}

	resCpt.ModifyResource(consts.GuidePro, 1)

	isGuideEnd := false
	if battles[int(battleID+1)] == nil {
		isGuideEnd = true
		//tc.attr.Del("lastBattleID")
		//tc.attr.SetInt32("campID", -1)
	}

	rewardGold := 0
	rewardStar := 0
	rewardMatchScore := 0
	if isGuideEnd {
		rewardMatchScore = 30
		rewardStar = 1
	}

	pvpComponent := tc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent)
	//rewardGoldCnt := resCpt.GetResource(consts.PvpGoldCnt)
	//if rewardGoldCnt < 24 {
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	pvpLevel := rankGameData.GetPvpLevelByStar(resCpt.GetResource(consts.Score) + rewardStar)
	rankData := rankGameData.Ranks[pvpLevel]
	if rankData != nil {
		rewardGold = rankData.GoldReward
		//resCpt.ModifyResource(consts.PvpGoldCnt, 1, true)
	}
	//}


	msg.ChangeRes = resCpt.BatchModifyResource(map[int]int{
		consts.Score: rewardStar,
		consts.Gold:  rewardGold,
		consts.MatchScore: rewardMatchScore,
	}, consts.RmrBattleWin)
	msg.UpPvpLevelRewardCards = pvpComponent.UplevelReward()

	//treasureComponent.AddDailyTreasureStar(1)
	tc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, msg)

	//if rewardStar > 0 && !tc.player.IsVip() {
	//	module.OutStatus.AddStatus(tc.player, consts.OtVipCard, 24 * 3600, true)
	//}
}

type playerData struct {
	campID    int
	player    types.IPlayer
	handCards []types.IFightCardData
	gridCards map[int]types.IFightCardData
}

func (pd *playerData) Initialize(campID int, player types.IPlayer, battle *gamedata.TutorialBattle) {
	//cardComponent := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	//handCards := cardComponent.CreatePvpHandCards(campID)

	pd.campID = campID
	pd.player = player
	//pd.handCards = handCards

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	pd.gridCards = make(map[int]types.IFightCardData)
	for _, gridCardInfo := range battle.OwnSide {
		cardData := poolGameData.GetCardByGid(uint32(gridCardInfo[0]))
		if cardData != nil {
			pd.gridCards[gridCardInfo[1]] = cardData
		}
	}
}

func (pd *playerData) GetMaxHandCardAmount() int {
	return consts.MaxHandCardAmount
}

func (pd *playerData) GetPlayer() types.IPlayer {
	return pd.player
}

func (pd *playerData) GetSit() int {
	return consts.SitOne
}

func (pd *playerData) IsRobot() bool {
	return false
}

func (pd *playerData) GetCamp() int {
	return pd.campID
}

func (pd *playerData) GetHandCards() []types.IFightCardData {
	return pd.handCards
}

func (pd *playerData) GetGridCards(isFirstHand bool, gridIDs []int) map[int]types.IFightCardData {
	return pd.gridCards
}

func (pd *playerData) GetKingForbidCards() []uint32 {
	return []uint32{}
}

func (pd *playerData) GetCasterSkills() []int32 {
	return []int32{}
}

type robotData struct {
	campID    int
	player    types.IPlayer
	handCards []types.IFightCardData
	gridCards map[int]types.IFightCardData
}

func (rd *robotData) Initialize(campID int, battle *gamedata.TutorialBattle) {
	robot := module.AI.NewRobotPlayer(nil)
	textGameData := gamedata.GetGameData(consts.Text).(*gamedata.TextGameData)
	robot.SetName(textGameData.TEXT(battle.Name))
	var handCards []types.IFightCardData
	pgd := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)

	for _, CardID := range battle.EnemyHand {
		handCards = append(handCards, pgd.GetCardByGid(uint32(CardID)))
	}

	rd.campID = campID
	rd.player = robot
	rd.handCards = handCards

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	rd.gridCards = make(map[int]types.IFightCardData)
	for _, gridCardInfo := range battle.EnemySide {
		cardData := poolGameData.GetCardByGid(uint32(gridCardInfo[0]))
		if cardData != nil {
			rd.gridCards[gridCardInfo[1]] = cardData
		}
	}
}

func (rd *robotData) GetMaxHandCardAmount() int {
	return consts.MaxHandCardAmount
}

func (rd *robotData) GetPlayer() types.IPlayer {
	return rd.player
}

func (rd *robotData) GetSit() int {
	return consts.SitTwo
}

func (rd *robotData) IsRobot() bool {
	return true
}

func (rd *robotData) GetCamp() int {
	return rd.campID
}

func (rd *robotData) GetHandCards() []types.IFightCardData {
	return rd.handCards
}

func (rd *robotData) GetGridCards(isFirstHand bool, gridIDs []int) map[int]types.IFightCardData {
	return rd.gridCards
}

func (rd *robotData) GetKingForbidCards() []uint32 {
	return []uint32{}
}

func (rd *robotData) GetCasterSkills() []int32 {
	return []int32{}
}
