package pvp

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"kinger/common/utils"
)

var _ types.IPvpComponent = &pvpComponent{}

type pvpComponent struct {
	player             types.IPlayer
	gdata              *gamedata.RankGameData
	attr               *attribute.MapAttr
	//historyBattlesAttr *attribute.AttrMgr
	pvpLevel           int
	pvpMaxLevel        int
}

func (pc *pvpComponent) ComponentID() string {
	return consts.PvpCpt
}

func (pc *pvpComponent) GetPlayer() types.IPlayer {
	return pc.player
}

func (pc *pvpComponent) OnInit(player types.IPlayer) {
	pc.player = player
	pc.gdata = gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)

	/*
	historyBattlesAttr := attribute.NewAttrMgr("historyBattles", pc.player.GetUid())
	evq.CallLater(func() {
		err := historyBattlesAttr.Load()
		if err == attribute.NotExistsErr {
			historyBattlesAttr.SetListAttr("battleIDs", attribute.NewListAttr())
			historyBattlesAttr.Save(false)
		} else if err != nil {
			return
		}
		pc.historyBattlesAttr = historyBattlesAttr
	})
	*/

	/*
	cnt := pc.attr.GetInt("rewardGoldCnt")
	if cnt > 0 {
		pc.attr.SetInt("rewardGoldCnt", 0)
		pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).SetResource(consts.PvpGoldCnt, cnt)
	}
	*/
}

func (pc *pvpComponent) fixRankScore() {
	rankScore := pc.player.GetRankScore()
	baseScore := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetScoreById(1)
	if rankScore < baseScore {
		module.Player.ModifyResource(pc.player, consts.MatchScore, baseScore-rankScore)
	}
}

func (pc *pvpComponent) OnLogin(isRelogin, isRestore bool) {
	pc.OnCrossDay(timer.GetDayNo())
	pc.updateLeagueSeason()
	if isRelogin {
		return
	}

	team := pc.player.GetPvpTeam()
	if team >= 9 {
		// 升上王者，积分不够的补够
		pc.fixRankScore()
	}

	resCpt := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	maxScore := resCpt.GetResource(consts.MaxScore)
	score := resCpt.GetResource(consts.Score)
	if maxScore < score {
		resCpt.SetResource(consts.MaxScore, maxScore)
	}

	mmr := resCpt.GetResource(consts.Mmr)
	if mmr < minMmr {
		resCpt.SetResource(consts.Mmr, minMmr)
	}

	pc.pvpLevel = pc.GetPvpLevel()
	pc.pvpMaxLevel = pc.GetMaxPvpLevel()

	if pc.pvpMaxLevel >= 31 && !pc.attr.GetBool("kingRw") {
		pc.attr.SetBool("kingRw", true)
		glog.Infof("pvpComponent OnLogin kingRw, uid=%d", pc.player.GetUid())
		sender := module.Mail.NewMailSender(pc.player.GetUid())
		sender.SetTypeAndArgs(pb.MailTypeEnum_KingReward)
		mailReward := sender.GetRewardObj()
		mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, "5", 1)
		sender.Send()
	}
}

func (pc *pvpComponent) OnLogout() {
}

func (pc *pvpComponent) onResUpdate(resType, amount int) {
	if resType == consts.Score {
		oldLevel := pc.pvpLevel
		pc.pvpLevel = mod.GetPvpLevelByStar(amount)
		if oldLevel != pc.pvpLevel {
			eventhub.Publish(consts.EvPvpLevelUpdate, pc.player, oldLevel, pc.pvpLevel)

			if pc.pvpLevel >= 31 {
				// 升上王者，积分不够的补够
				pc.fixRankScore()
			}
		}

		pc.player.OnSimpleInfoUpdate()

	} else if resType == consts.MaxScore {
		maxPvpLevel := mod.GetPvpLevelByStar(amount)
		if maxPvpLevel > pc.pvpMaxLevel {
			oldLevel := pc.pvpMaxLevel
			pc.pvpMaxLevel = maxPvpLevel
			eventhub.Publish(consts.EvMaxPvpLevelUpdate, pc.player, oldLevel, maxPvpLevel)

			if pc.pvpMaxLevel >= 31 && !pc.attr.GetBool("kingRw") {
				pc.attr.SetBool("kingRw", true)
				glog.Infof("pvpComponent kingRw, uid=%d", pc.player.GetUid())
				sender := module.Mail.NewMailSender(pc.player.GetUid())
				sender.SetTypeAndArgs(pb.MailTypeEnum_KingReward)
				mailReward := sender.GetRewardObj()
				mailReward.AddItem(pb.MailRewardType_MrtHeadFrame, "5", 1)
				sender.Send()
			}
		}
	}
}

func (pc *pvpComponent) GetPvpFighterData() *pb.FighterData {
	cardCpt := pc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	camp := cardCpt.GetFightCamp()
	var drawCards []*pb.SkinGCard
	for _, card := range cardCpt.GetAllCollectCards() {
		drawCards = append(drawCards, &pb.SkinGCard{
			GCardID: card.GetCardGameData().GCardID,
			Skin:    card.GetSkin(),
			Equip:   card.GetEquip(),
		})
	}

	agent := pc.player.GetAgent()
	return &pb.FighterData{
		Uid:          uint64(pc.player.GetUid()),
		ClientID:     uint64(agent.GetClientID()),
		GateID:       agent.GetGateID(),
		HandCards:    cardCpt.CreatePvpHandCards(camp),
		DrawCardPool: drawCards,
		Name:         pc.player.GetName(),
		Camp:         int32(camp),
		PvpScore:     int32(pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).GetResource(consts.Score)),
		HeadImgUrl:   pc.player.GetHeadImgUrl(),
		HeadFrame:    pc.player.GetHeadFrame(),
		Area: int32(pc.player.GetArea()),
		Region: agent.GetRegion(),
		CountryFlag: pc.player.GetCountryFlag(),
	}
}

func (pc *pvpComponent) GetPvpLevel() int {
	if pc.pvpLevel > 0 {
		return pc.pvpLevel
	}
	resComponent := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	return mod.GetPvpLevelByStar(resComponent.GetResource(consts.Score))
}

func (pc *pvpComponent) getTeam() int {
	level := pc.GetPvpLevel()
	return pc.gdata.Ranks[level].Team
}

func (pc *pvpComponent) GetMaxPvpLevel() int {
	if pc.pvpMaxLevel > 0 {
		return pc.pvpMaxLevel
	}
	resComponent := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	return mod.GetPvpLevelByStar(resComponent.GetResource(consts.MaxScore))
}

func (pc *pvpComponent) GetMaxPvpTeam() int {
	level := pc.GetMaxPvpLevel()
	return pc.gdata.Ranks[level].Team
}

func (pc *pvpComponent) GetPvpTeam() int {
	level := pc.GetPvpLevel()
	return pc.gdata.Ranks[level].Team
}

func (pc *pvpComponent) OnCrossDay(dayno int) {
	if dayno == pc.player.GetDataDayNo() {
		return
	}
	//pc.attr.SetInt("dayno", dayno)
	pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).SetResource(consts.PvpGoldCnt, 0)
}

/*
func (pc *pvpComponent) GetRewardGoldCnt() int {
	curDayno := timer.GetDayNo()
	datno := pc.attr.GetInt("dayno")
	cnt := pc.attr.GetInt("rewardGoldCnt")
	if curDayno != datno {
		pc.attr.SetInt("dayno", curDayno)
		cnt = 0
		//pc.attr.SetInt("rewardGoldCnt", cnt)
		pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).SetResource(consts.PvpGoldCnt, 0)
	}
	return cnt
}

func (pc *pvpComponent) SetRewardGoldCnt(cnt int) {
	pc.attr.SetInt("rewardGoldCnt", cnt)
}
*/

func (pc *pvpComponent) incrStreakWinCnt() int {
	n := pc.attr.GetInt("streakWinCnt")
	n++
	pc.attr.SetInt("streakWinCnt", n)
	return n
}

func (pc *pvpComponent) resetStreakWinCnt() {
	pc.attr.SetInt("streakWinCnt", 0)
}

func (pc *pvpComponent) getStreakWinCnt() int {
	return pc.attr.GetInt("streakWinCnt")
}

func (pc *pvpComponent) incrStreakLoseCnt() int {
	n := pc.attr.GetInt("streakLoseCnt")
	n++
	pc.attr.SetInt("streakLoseCnt", n)
	return n
}

func (pc *pvpComponent) resetStreakLoseCnt() {
	pc.attr.SetInt("streakLoseCnt", 0)
}

func (pc *pvpComponent) getStreakLoseCnt() int {
	return pc.attr.GetInt("streakLoseCnt")
}

// 因充值降低匹配指数
func (pc *pvpComponent) getRechargeMatchIndex() int {
	return pc.attr.GetInt("rechargeMatchIndex")
}
func (pc *pvpComponent) setRechargeMatchIndex(index int) {
	if index > 0 {
		index = 0
	}
	pc.attr.SetInt("rechargeMatchIndex", index)
}

func (pc *pvpComponent) UplevelReward() []uint32 {
	curLevel := pc.GetPvpLevel()
	rewardLevelsAttr := pc.attr.GetListAttr("rewardLevels")
	if rewardLevelsAttr == nil {
		rewardLevelsAttr = attribute.NewListAttr()
		pc.attr.SetListAttr("rewardLevels", rewardLevelsAttr)
	}

	var minLevel int
	if rewardLevelsAttr.Size() > 0 {
		minLevel = rewardLevelsAttr.GetInt(rewardLevelsAttr.Size() - 1)
	}

	var rewards []uint32
	modifyCards := map[uint32]*pb.CardInfo{}
	for level := minLevel + 1; level <= curLevel; level++ {
		rankData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).Ranks[level]
		if rankData == nil || len(rankData.RankUpReward) <= 0 {
			continue
		}
		rewardLevelsAttr.AppendInt(level)
		glog.Infof("%d get up pvp level Reward, level=%d reward=%v", pc.player.GetUid(), level, rankData.RankUpReward)

		for _, cardID := range rankData.RankUpReward {
			c, ok := modifyCards[cardID]
			if !ok {
				c = &pb.CardInfo{CardId: cardID}
				modifyCards[cardID] = c
			}
			c.Amount += 1
		}
		rewards = append(rewards, rankData.RankUpReward...)
	}

	if len(modifyCards) > 0 {
		pc.player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(modifyCards)
	}

	return rewards
}

/*  TODO
func (pc *pvpComponent) getHistoryVideos(page int) []*videoItem {
	if pc.historyBattlesAttr == nil {
		return []*videoItem{}
	}
	battleIDsAttr := pc.historyBattlesAttr.GetListAttr("battleIDs")
	amount := battleIDsAttr.Size()
	if amount <= 0 {
		return []*videoItem{}
	}
	var vis []*videoItem
	beginIdx := amount - 1 - ((page - 1) * 10)
	if beginIdx < 0 {
		return vis
	}
	endIdx := amount - 1 - page*10

	for i := beginIdx; i >= 0 && i > endIdx; i-- {
		battleID := common.UUid(battleIDsAttr.GetUInt64(i))
		vi := loadVideoItem(battleID)
		if vi != nil {
			vis = append(vis, vi)
		} else {
			battleIDsAttr.DelBySection(0, beginIdx+1)
			pc.historyBattlesAttr.Save(false)
		}
	}
	return vis
}
*/

func (pc *pvpComponent) getNewbiePvpEnemyCamp() (int, bool) {
	newbiePvpCampAttr := pc.attr.GetListAttr("newbiePvpCamp")
	if newbiePvpCampAttr == nil {
		newbiePvpCampAttr = attribute.NewListAttr()
		pc.attr.SetListAttr("newbiePvpCamp", newbiePvpCampAttr)
	}
	if newbiePvpCampAttr.Size() >= 3 {
		return 0, false
	}

	camps := []int{consts.Wei, consts.Shu, consts.Wu}
	newbiePvpCampAttr.ForEachIndex(func(index int) bool {
		camp := newbiePvpCampAttr.GetInt(index)
		for i, camp2 := range camps {
			if camp2 == camp {
				camps = append(camps[:i], camps[i+1:]...)
			}
		}
		return true
	})

	if len(camps) > 0 {
		return camps[rand.Intn(len(camps))], len(camps) == 3
	}
	return 0, false
}

func (pc *pvpComponent) onBattleWin(fighterData *pb.EndFighterData, oppCamp int, isWonderful bool, oppArea int) (
	changeStar, changeCrossAreaHonor int) {

	pc.resetStreakLoseCnt()
	pc.incrStreakWinCnt()
	changeStar = 1

	matchParam := gamedata.GetGameData(consts.MatchParam).(*gamedata.MatchParamGameData)
	rankData := pc.gdata.Ranks[pc.player.GetPvpLevel()]
	if rankData == nil {
		rankData = pc.gdata.Ranks[2]
	}
	winningRateGameData := gamedata.GetGameData(consts.WinningRate).(*gamedata.WinningRateGameData)

	changeMatchScore := int( math.Ceil( rankData.Kvalue * (1 -
		winningRateGameData.GetExpectedWinningRate(int(fighterData.IndexDiff), pc.player.GetPvpTeam())) * matchParam.WinRevise ) )
	rewardGold := 0
	resComponent := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	rewardGoldCnt := resComponent.GetResource(consts.PvpGoldCnt)
	newPvpLevel := mod.GetPvpLevelByStar(resComponent.GetResource(consts.Score) + changeStar)
	newRankData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData).Ranks[newPvpLevel]

	if rewardGoldCnt < 24 {
		if newRankData != nil {
			rewardGold = rankData.GoldReward
			//pvpComponent.SetRewardGoldCnt(rewardGoldCnt + 1)
		}
	}

	if rewardGold > 0 {
		rewardGold = module.OutStatus.BuffAddPvpGold(pc.player, rewardGold)
	}
	if changeStar > 0 && newRankData.Team < 9 {
		// 王者以下，可能因为特权，多加星、分
		newChangeStar := module.OutStatus.BuffPvpAddStar(pc.player, changeStar)
		changeMatchScore = changeMatchScore * (newChangeStar / changeStar)
		changeStar = newChangeStar
	}

	if oppArea > 0 && pc.player.GetArea() > oppArea {
		changeCrossAreaHonor = 1
	}

	treasureNum := 1
	treasureNum = module.OutStatus.BuffPvpAddTreasure(pc.player, treasureNum)

	treasureComponent := pc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent)
	treasureID1, reason, upRareTreasureModelID := treasureComponent.AddRewardTreasure(false, true)
	hasTreasure := treasureID1 != ""
	treasureIDList := []string{treasureID1}
	if treasureNum > 1 {
		treasureID2, _, _ := treasureComponent.AddRewardTreasure(false, false)
		if treasureID2 != "" {
			treasureIDList = append(treasureIDList, treasureID2)
		} else if !hasTreasure {
			hasTreasure = true
		}
	}

	if rewardGold > 0 {
		resComponent.ModifyResource(consts.PvpGoldCnt, 1)
	}

	resChange := resComponent.BatchModifyResource(map[int]int{
		consts.Score: changeStar,
		consts.Gold:  rewardGold,
		consts.CrossAreaHonor: changeCrossAreaHonor,
		consts.MatchScore: changeMatchScore,
	}, consts.RmrBattleWin)

	pc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
		WinUid:                uint64(pc.player.GetUid()),
		ChangeRes:             resChange,
		TreasureID:            treasureIDList,
		UpPvpLevelRewardCards: pc.UplevelReward(),
		CanShare:              isWonderful,
		NoTreasureReason:      reason,
		UpRareTreasureModelID: upRareTreasureModelID,
	})

	//mod.UpdatePlayerRankScore(winer)
	treasureComponent.AddDailyTreasureStar(changeStar)

	newbiePvpCampAttr := pc.attr.GetListAttr("newbiePvpCamp")
	if newbiePvpCampAttr != nil && newbiePvpCampAttr.Size() < 3 {
		newbiePvpCampAttr.AppendInt(oppCamp)
	}

	return
}

func (pc *pvpComponent) onBattleLose(fighterData *pb.EndFighterData, oppArea int) (changeStar, changeCrossAreaHonor int) {
	pc.incrStreakLoseCnt()
	pc.resetStreakWinCnt()

	matchParam := gamedata.GetGameData(consts.MatchParam).(*gamedata.MatchParamGameData)
	rankData := pc.gdata.Ranks[pc.player.GetPvpLevel()]
	if rankData == nil {
		rankData = pc.gdata.Ranks[2]
	}
	winningRateGameData := gamedata.GetGameData(consts.WinningRate).(*gamedata.WinningRateGameData)

	changeStar = -1
	changeMatchScore := - int( math.Ceil( rankData.Kvalue *
		winningRateGameData.GetExpectedWinningRate(int(fighterData.IndexDiff), pc.player.GetPvpTeam()) * matchParam.WinRevise ))

	resCpt := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)

	curStar := resCpt.GetResource(consts.Score)
	curPvpLevel := pc.GetPvpLevel()

	if curPvpLevel < 5 {
		// 初级不掉星
		changeStar = 0
	} else if rankData, ok := pc.gdata.Ranks[curPvpLevel]; ok && rankData.Protection != 0 {
		// 掉星不掉段位
		levelMaxStar := mod.getMaxStarByPvpLevel(curPvpLevel - 1)
		levelMaxStar++

		if curStar+changeStar < levelMaxStar {
			changeStar = levelMaxStar - curStar
			if changeStar > 0 {
				changeStar = 0
			}
		}
	}

	noSubStarReason := pb.BattleResult_Normal
	if changeStar < 0 && curPvpLevel < 31 {
		// 王者以下，可能因为特权不掉星
		changeStar = module.OutStatus.BuffPvpNoSubStar(pc.player, changeStar)
		if changeStar == 0 {
			noSubStarReason = pb.BattleResult_NoSubStarPriv
		}
	}

	if changeStar == 0 {
		changeMatchScore = 0
	}

	if oppArea > 0 && pc.player.GetArea() < oppArea {
		changeCrossAreaHonor = -1
	}

	var resChange []*pb.ChangeResInfo
	resChange = resCpt.BatchModifyResource(map[int]int{
		consts.Score: changeStar,
		consts.CrossAreaHonor: changeCrossAreaHonor,
		consts.MatchScore: changeMatchScore,
	})

	pc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
		WinUid:          1,
		ChangeRes:       resChange,
		NoSubStarReason: noSubStarReason,
	})

	return
}

func (pc *pvpComponent) OnBattleEnd(fighterData *pb.EndFighterData, isWin, isWonderful bool, oppMMr, oppCamp, oppArea int) {
	pc.setRechargeMatchIndex( pc.getRechargeMatchIndex() + gamedata.GetGameData(consts.MatchParam).(
		*gamedata.MatchParamGameData).RechargeReviseRecovery )
	module.Mission.OnPvpBattleEnd(pc.player, fighterData, isWin)
	resComponent := pc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	//myMmr := resComponent.GetResource(consts.Mmr)
	//ammr := float64(myMmr)
	//bmmr := float64(oppMMr)
	//if !isWin {
	//	ammr, bmmr = bmmr, ammr
	//}

	var changeStar, changeCrossAreaHonor int
	//changeMmr := int(math.Ceil((1.0 - 1.0/(1.0+math.Pow(10.0, (ammr-bmmr)/400.0))) * 50.0))
	if isWin {
		changeStar, changeCrossAreaHonor = pc.onBattleWin(fighterData, oppCamp, isWonderful, oppArea)
	} else {
		changeStar, changeCrossAreaHonor = pc.onBattleLose(fighterData, oppArea)
	}

	oldSeasonWinDiff := resComponent.GetResource(consts.WinDiff)
	oldSeasonWinAmount := module.Huodong.GetSeasonPvpWinCnt(pc.player)

	eventhub.Publish(consts.EvEndPvpBattle, pc.player, isWin, fighterData)

	newSeasonWinDiff := resComponent.GetResource(consts.WinDiff)
	newSeasonWinAmount := module.Huodong.GetSeasonPvpWinCnt(pc.player)

	needUpdateRank := changeStar != 0 || changeCrossAreaHonor != 0 || newSeasonWinDiff != oldSeasonWinDiff ||
		newSeasonWinAmount != oldSeasonWinAmount
	if needUpdateRank {
		logic.PushBackend("", 0, pb.MessageID_G2R_UPDATE_PVP_SCORE, module.Player.PackUpdateRankMsg(
			pc.player, fighterData.InitHandCards, int(fighterData.Camp)))
	}
}

func (pc *pvpComponent) OnTrainingBattleEnd(fighterData *pb.EndFighterData, isWin bool) {
	module.Mission.OnPvpBattleEnd(pc.player, fighterData, isWin)
	winUid := pc.player.GetUid()
	if !isWin {
		winUid = 1
	}
	pc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, &pb.BattleResult{
		WinUid:                uint64(winUid),
	})
}

func (pc *pvpComponent) setRewardReceive(rid int) {
	ridStr := strconv.Itoa(rid)
	attr := pc.attr.GetMapAttr("leagueRewardReceive")
	if attr == nil {
		attr = attribute.NewMapAttr()
		pc.attr.SetMapAttr("leagueRewardReceive", attr)
	}
	attr.SetBool(ridStr, true)
}

func (pc *pvpComponent) resetRewardReceive() {
	attr := pc.attr.GetMapAttr("leagueRewardReceive")
	if attr != nil {
		pc.attr.Del("leagueRewardReceive")
	}
}

func (pc *pvpComponent) getRewardReceiveStatus(rid int) bool {
	ridStr := strconv.Itoa(rid)
	attr := pc.attr.GetMapAttr("leagueRewardReceive")
	if attr != nil {
		return attr.GetBool(ridStr)
	}
	return false
}

func (pc *pvpComponent) getSeassonSerial() int {
	return pc.attr.GetInt("league_serial")
}

func (pc *pvpComponent) setSeassonSerial() {
	s := leagueAttr.getCurSeasonSerial(pc.player.GetArea())
	pc.attr.SetInt("league_serial", s)
}

func (pc *pvpComponent) isInThisSeason() bool {
	s := leagueAttr.getCurSeasonSerial(pc.player.GetArea())
	ps := pc.getSeassonSerial()
	if ps == s {
		return true
	}
	return false
}

func (pc *pvpComponent) crossSeassonResetScore() {
	oldRankScore := pc.player.GetRankScore()
	oldMaxScore := pc.player.GetMaxRankScore()
	modifyMaxScore, modifyScore := utils.CrossLeagueSeasonResetScore(oldMaxScore, oldRankScore)
	if modifyMaxScore != 0 {
		module.Player.ModifyResource(pc.player, consts.MaxMatchScore, modifyMaxScore)
	}
	if modifyScore != 0 {
		module.Player.ModifyResource(pc.player, consts.MatchScore, modifyScore)
	}

	/*
	oldMaxScore := pc.player.GetMaxRankScore()
	baseScore := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetScoreById(1)
	if oldMaxScore < baseScore {
		return
	}
	oldRankScore := pc.player.GetRankScore()
	funGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	pro := float64(funGameData.LeagueResetRewardProp)/100
	compensatoryScore := int(float64(oldRankScore-baseScore)*pro)
	maxCompensatoryScore := funGameData.LeagueResetRewardMax
	if compensatoryScore > maxCompensatoryScore {
		compensatoryScore = maxCompensatoryScore
	}
	newScore := baseScore + compensatoryScore
	modifyScore := newScore-oldRankScore
	modifyMaxScore := newScore-oldMaxScore
	if newScore < baseScore {
		modifyMaxScore = baseScore-oldMaxScore
	}
	if modifyMaxScore < 0 {
		module.Player.ModifyResource(pc.player, consts.MaxMatchScore, modifyMaxScore)
	}
	if modifyScore < 0 && newScore > baseScore{
		module.Player.ModifyResource(pc.player, consts.MatchScore, modifyScore)
	}
	*/
}

func (pc *pvpComponent) updateLeagueSeason() {

	if pc.isInThisSeason() {
		return
	}

	team := pc.player.GetPvpTeam()
	if team < 9 {
		pc.setSeassonSerial()
		return
	}

	pSerial := pc.getSeassonSerial()
	score := pc.player.GetMaxRankScore()
	pc.crossSeassonResetScore()
	newScore := pc.player.GetMaxRankScore()
	leagueLvl := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetLeagueEndRewardLvlByScore(score)

	if leagueLvl > 0 && pSerial > 0 {
		rank, rankRewards, kingFlag, ok := leagueAttr.getPlayerLeagueRankReward(pc.player, pSerial)
		glog.Infof("on cross league season, uid=%d area=%d rank=%d leagueLvl=%d serial=%d", pc.player.GetUid(), 
			pc.player.GetArea(), rank, leagueLvl, pSerial)
		reward := leagueAttr.getLeagueEndReward(pc.player.GetArea(), pSerial, leagueLvl)
		sender := module.Mail.NewMailSender(pc.player.GetUid())
		sender.SetTypeAndArgs(pb.MailTypeEnum_LeagueEnd, score, newScore, rank)
		rewardOdj := sender.GetRewardObj()
		addMailReward := func(rewards []string) {
			for _, rw := range rewards {
				isRes, mty, num, itemId := pc.getMailReward(rw)
				if isRes {
					rewardOdj.AddAmountByType(mty, num)
				}else {
					rewardOdj.AddItem(mty, itemId, num)
				}
			}
		}
		addMailReward(reward)

		if ok {
				addMailReward(rankRewards)
				module.Player.ModifyResource(pc.player, consts.KingFlag, kingFlag, consts.RmrLeagueReward+"kingFlag")
		}

		sender.Send()
	}
	pc.setSeassonSerial()
	pc.resetRewardReceive()
}

func (pc *pvpComponent) getMailReward(rw string) (isRes bool, ty pb.MailRewardType, num int, itemId string){
	rws := strings.Split(rw, ":")
	if len(rws) >= 2 {
		var err error
		itemId = rws[0]
		isRes, ty = module.Reward.GetMailItemType(itemId)
		num, err = strconv.Atoi(rws[1])
		if err != nil {
			glog.Errorf("pvp comment updateLeagueSeason give reward arg error, uid=%d", pc.player.GetUid())
		}
		if ty == pb.MailRewardType_MrtHeadFrame {
			itemId = rws[1]
			num = 1
		}
	}
	return
}

