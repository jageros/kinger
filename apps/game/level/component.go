package level

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"kinger/apps/game/module"
	"strconv"
	"kinger/common/utils"
	"fmt"
	"kinger/gopuppy/common/glog"
)

var _ types.ILevelComponent = &levelComponent{}

type levelComponent struct {
	player types.IPlayer
	gdata  *gamedata.LevelGameData
	attr   *attribute.MapAttr
	log *levelLog
	help *helpRecord
}

func (lc *levelComponent) ComponentID() string {
	return consts.LevelCpt
}

func (lc *levelComponent) GetPlayer() types.IPlayer {
	return lc.player
}

func (lc *levelComponent) OnInit(player types.IPlayer) {
	lc.player = player
	lc.gdata = gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData)
	lc.GetCurLevel()
	lc.log = newLevelLog(lc.attr)
}

func (lc *levelComponent) OnLogin(isRelogin, isRestore bool) {
}

func (lc *levelComponent) OnLogout() {
}

// 当前已解锁的最大关卡
func (lc *levelComponent) GetCurLevel() int {
	levelID := lc.attr.GetInt("curLevel")
	if levelID <= 0 {
		levelID = 1
		lc.attr.SetInt("curLevel", levelID)
	}
	return levelID
}

func (lc *levelComponent) unlockLevel(levelID int) []*pb.ChangeCardInfo {
	lc.attr.SetInt("curLevel", levelID)
	//levelData := lc.gdata.GetLevelData(levelID - 1)
	var changeCard []*pb.ChangeCardInfo
	/*
		if levelData != nil {
			if levelData.GeneralUnlock > 0 {
				cardCpt := lc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
				changeCard = cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
					levelData.GeneralUnlock: &pb.CardInfo{
						CardId: levelData.GeneralUnlock,
						Amount: 1,
					},
				})
			}
		}
	*/

	return changeCard
}

func (lc *levelComponent) IsClearLevel(levelID int) bool {
	return levelID < lc.GetCurLevel()
}

func (lc *levelComponent) getHelpRecord() *helpRecord {
	if lc.help != nil {
		return lc.help
	}
	lc.help = newHelpRecord(lc.attr)
	return lc.help
}

func (lc *levelComponent) beginHelpBattle(levelID int, beHelpUid common.UUid) error {
	levelData := lc.gdata.GetLevelData(levelID)
	if levelData == nil {
		return gamedata.InternalErr
	}

	lc.getHelpRecord().setHelpInfo(beHelpUid, levelID)

	agent := lc.player.GetAgent()
	fighterData := &pb.FighterData{
		Uid:      uint64(lc.player.GetUid()),
		ClientID: uint64(agent.GetClientID()),
		GateID:   agent.GetGateID(),
		IsRobot:  false,
		Region: agent.GetRegion(),
	}
	allCollectCards := lc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetAllCollectCards()
	for _, card := range allCollectCards {
		fighterData.DrawCardPool = append(fighterData.DrawCardPool, &pb.SkinGCard{
			GCardID: card.GetCardGameData().GCardID,
			Skin: card.GetSkin(),
			Equip: card.GetEquip(),
		})
	}
	_, err := logic.CallBackend("", 0, pb.MessageID_L2B_BEGIN_LEVEL_HELP_BATTLE, &pb.BeginLevelBattleArg{
		LevelID:  int32(levelID),
		Fighter1: fighterData,
	})
	return err
}

func (lc *levelComponent) isRechargeLock() bool {
	//if !config.GetConfig().IsXfServer() {
		return false
	//}
	//return !lc.attr.GetBool("rechargeUnLock")
}

func (lc *levelComponent) onRecharge() {
	if !lc.isRechargeLock() {
		return
	}
	lc.attr.SetBool("rechargeUnLock", true)
	lc.player.GetAgent().PushClient(pb.MessageID_S2C_LEVEL_ON_RECHARGE, nil)
}

func (lc *levelComponent) beginBattle(levelID int) (interface{}, error) {
	curLevel := lc.GetCurLevel()
	if levelID > curLevel {
		return nil, gamedata.LevelLockErr
	}

	levelData := lc.gdata.GetLevelData(levelID)
	if levelData == nil {
		return nil, gamedata.InternalErr
	}

	if lc.player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel() < levelData.RankCondition {
		return nil, gamedata.GameError(1)
	}

	if levelID == curLevel && levelData.IsRechargeUnlock && lc.isRechargeLock() {
		return nil, gamedata.GameError(2)
	}

	agent := lc.player.GetAgent()
	fighterData := &pb.FighterData{
		Uid:      uint64(lc.player.GetUid()),
		ClientID: uint64(agent.GetClientID()),
		GateID:   agent.GetGateID(),
		IsRobot:  false,
		Region: agent.GetRegion(),
	}
	allCollectCards := lc.player.GetComponent(consts.CardCpt).(types.ICardComponent).GetAllCollectCards()
	for _, card := range allCollectCards {
		fighterData.DrawCardPool = append(fighterData.DrawCardPool, &pb.SkinGCard{
			GCardID: card.GetCardGameData().GCardID,
			Skin: card.GetSkin(),
			Equip: card.GetEquip(),
		})
	}

	lc.log.onLevelBegin(levelID)
	module.Player.LogMission(lc.player, fmt.Sprintf("level_%d", levelID), 1)

	return logic.CallBackend("", 0, pb.MessageID_L2B_BEGIN_LEVEL_BATTLE, &pb.BeginLevelBattleArg{
		LevelID:  int32(levelID),
		Fighter1: fighterData,
	})
	//return module.Battle.BeginBattle(consts.BtLevel, newPlayerFighterData(lc.player, levelData), newRobotFighterData(levelData),
	//	levelData.Offensive, newBattleEndHandler(lc.player.GetUid(), levelID), nil, levelData)
	//return module.Battle.BeginLevelFight(lc.player, levelData, newBattleEndHandler(lc.player.GetUid(), levelID)), nil
}

func (lc *levelComponent) GetUnlockCards() []uint32 {
	return lc.gdata.GetUnlockCards(lc.GetCurLevel())
}

func (lc *levelComponent) getOpenedTreasureChapters() *attribute.ListAttr {
	openedTreasureChaptersAttr := lc.attr.GetListAttr("openedTreasureChapters")
	if openedTreasureChaptersAttr == nil {
		openedTreasureChaptersAttr = attribute.NewListAttr()
		lc.attr.SetListAttr("openedTreasureChapters", openedTreasureChaptersAttr)
	}
	return openedTreasureChaptersAttr
}

func (lc *levelComponent) packMsg() *pb.LevelInfo {
	openedTreasureChaptersAttr := lc.getOpenedTreasureChapters()
	msg := &pb.LevelInfo{
		CurLevel: int32(lc.GetCurLevel()),
		AskHelpLevels: lc.getHelpRecord().packNeedAskHelpLevels(),
		IsRechargeLock: lc.isRechargeLock(),
	}

	openedTreasureChaptersAttr.ForEachIndex(func(index int) bool {
		msg.OpenedTreasureChapters = append(msg.OpenedTreasureChapters, int32(openedTreasureChaptersAttr.GetInt(index)))
		return true
	})
	return msg
}

func (lc *levelComponent) openTreasure(chapterID int) (*pb.OpenTreasureReply, error) {
	treasureModelID, ok := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData).ChapterTreasure[chapterID]
	if !ok {
		return nil, gamedata.InternalErr
	}

	openedTreasureChaptersAttr := lc.getOpenedTreasureChapters()
	canOpen := true
	openedTreasureChaptersAttr.ForEachIndex(func(index int) bool {
		if chapterID == openedTreasureChaptersAttr.GetInt(index) {
			canOpen = false
			return false
		}
		return true
	})

	if !canOpen {
		return nil, gamedata.InternalErr
	}

	treasureReward := lc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
		treasureModelID, false)
	if treasureReward.OK {
		openedTreasureChaptersAttr.AppendInt(chapterID)
	}
	return treasureReward, nil
}

func (lc *levelComponent) doLevelReward(levelData *gamedata.Level) ([]*pb.ChangeCardInfo, []*pb.ChangeResInfo) {
	resComponent := lc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	cardComponent := lc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	rewardCards := make(map[uint32]*pb.CardInfo)
	for _, cardId := range levelData.RewardCer {
		rewardCards[cardId] = &pb.CardInfo{
			Amount: 1,
		}
	}

	changeCards := cardComponent.ModifyCollectCards(rewardCards)
	rewardRes := resComponent.BatchModifyResource(map[int]int{
		//consts.Weap:   levelData.RewardWeap,
		//consts.Horse:  levelData.RewardHor,
		//consts.Mat:    levelData.RewardMat,
		//consts.Forage: levelData.RewardFor,
		consts.Gold:   levelData.RewardGold,
		//consts.Med:    levelData.RewardMed,
		//consts.Ban:    levelData.RewardBan,
	}, consts.RmrClearLevel)

	maxLevelID := lc.gdata.GetMaxLevelID()
	if levelData.ID == maxLevelID {
		module.Televise.SendNotice(pb.TeleviseEnum_ClearanceLevel, lc.player.GetName())
	}

	return changeCards, rewardRes
}

func (lc *levelComponent) getVideo(levelID int) common.UUid {
	videosAttr := lc.attr.GetMapAttr("videos")
	if videosAttr == nil {
		return 0
	}
	return common.UUid(videosAttr.GetUInt64(strconv.Itoa(levelID)))
}

func (lc *levelComponent) addVideo(levelID int, battleID common.UUid) {
	videosAttr := lc.attr.GetMapAttr("videos")
	if videosAttr == nil {
		videosAttr = attribute.NewMapAttr()
		lc.attr.SetMapAttr("videos", videosAttr)
	}

	levelKey := strconv.Itoa(levelID)
	if videosAttr.GetUInt64(levelKey) <= 0 {
		videosAttr.SetUInt64(levelKey, uint64(battleID))
	}
}

func (lc *levelComponent) OnBattleEnd(fighterData *pb.EndFighterData, isWin bool, levelID int, battleID common.UUid) {
	msgid := pb.MessageID_S2C_BATTLE_END
	_msg := &pb.BattleResult{}

	if !isWin {
		_msg.WinUid = 1
		lc.player.GetAgent().PushClient(msgid, _msg)

		if levelID == lc.GetCurLevel() {
			lc.getHelpRecord().addNeedAskHelpLevel(levelID)
		}

		return
	}

	_msg.WinUid = uint64(lc.player.GetUid())
	gdata := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData)
	levelData := gdata.GetLevelData(levelID)
	if levelData == nil {
		lc.player.GetAgent().PushClient(msgid, _msg)
		return
	}

	lc.addVideo(levelID, battleID)

	isFirstWin := levelID >= lc.GetCurLevel()
	if isFirstWin {
		_msg.ChangeCards = lc.unlockLevel(levelID + 1)
		changeCards, changeRes := lc.doLevelReward(levelData)
		_msg.ChangeCards = append(_msg.ChangeCards, changeCards...)
		_msg.ChangeRes = changeRes
		lc.getHelpRecord().delNeedAskHelpLevel(levelID)
	}

	lc.player.GetAgent().PushClient(msgid, _msg)

	lc.log.onLevelWin(levelID)
	module.Player.LogMission(lc.player, fmt.Sprintf("level_%d", levelID), 2)
}

func (lc *levelComponent) ClearLevel() {
	gdata := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData)
	lc.unlockLevel(gdata.GetMaxLevelID() + 1)
}

func (lc *levelComponent) OnHelpBattleEnd(isWin bool, battleID common.UUid) {
	msg := &pb.BattleResult{}
	if !isWin {
		battleID = 0
		msg.WinUid = 1
	} else {
		msg.WinUid = uint64(lc.player.GetUid())
	}
	lc.player.GetAgent().PushClient(pb.MessageID_S2C_BATTLE_END, msg)

	help := lc.getHelpRecord()
	helpUid := help.getBeHelpUid()
	levelID := help.getLevelID()
	help.delHelp()

	helpPlayer := module.Player.GetPlayer(helpUid)
	if helpPlayer != nil {
		helpPlayer.GetComponent(consts.LevelCpt).(*levelComponent).OnBeHelpBattle(lc.player.GetUid(), lc.player.GetName(),
			levelID, battleID)
	} else {
		utils.PlayerMqPublish(helpUid, pb.RmqType_HelpLevel, &pb.RmqHelpLevel{
			HelperUid: uint64(lc.player.GetUid()),
			HelperName: lc.player.GetName(),
			LevelID: int32(levelID),
			BattleID: uint64(battleID),
		})
	}
}

func (lc *levelComponent) OnBeHelpBattle(helperUid common.UUid, helperName string, levelID int, battleID common.UUid) {

	if lc.GetCurLevel() > levelID {
		return
	}

	levelData := lc.gdata.GetLevelData(levelID)
	if levelData == nil {
		return
	}

	lc.getHelpRecord().recordBeHelp(helperUid, helperName, levelID, battleID)
	if battleID > 0 && lc.GetCurLevel() <= levelID {
		lc.unlockLevel(levelID + 1)
		lc.player.GetAgent().PushClient(pb.MessageID_S2C_LEVEL_BE_HELP, &pb.LevelBeHelpArg{
			HelperName: helperName,
			LevelID: int32(levelID),
		})
	}
}

func (lc *levelComponent) getLevelBeHelpRecord(levelID int) *pb.LevelHelpRecord {
	return lc.getHelpRecord().packBeHelpMsg(levelID)
}

func (lc *levelComponent) clearChapter() error {
	curLevel := lc.GetCurLevel()
	maxLevelID := lc.gdata.GetMaxLevelID()
	if curLevel > maxLevelID {
		return gamedata.GameError(1)
	}

	levelAmount := maxLevelID - curLevel + 1
	needJade := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).ClearLevel * levelAmount
	resCpt := lc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Jade, needJade) {
		return gamedata.GameError(2)
	}
	resCpt.ModifyResource(consts.Jade, - needJade, consts.RmrClearChapter)

	glog.Infof("clearChapter uid=%d, curLevel=%d, jade=%d", lc.player.GetUid(), curLevel, needJade)

	lc.unlockLevel(maxLevelID + 1)
	for levelID := curLevel; levelID <= maxLevelID; levelID++ {
		lvData := lc.gdata.GetLevelData(levelID)
		if lvData != nil {
			lc.doLevelReward(lvData)
		}
	}
	return nil
}
