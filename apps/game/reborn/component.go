package reborn

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var _ types.IPlayerComponent = &rebornComponent{}

type rebornComponent struct {
	attr   *attribute.MapAttr
	player types.IPlayer
}

func (rc *rebornComponent) ComponentID() string {
	return consts.RebornCpt
}

func (rc *rebornComponent) GetPlayer() types.IPlayer {
	return rc.player
}

func (rc *rebornComponent) OnInit(player types.IPlayer) {
	rc.player = player
}

func (rc *rebornComponent) OnLogin(isRelogin, isRestore bool) {
	if rc.attr.GetInt("version") < version {
		rc.attr.SetInt("version", version)
	}
}

func (rc *rebornComponent) OnLogout() {

}

// 重生后的最高段位
func (rc *rebornComponent) getMaxPvpLevel() int {
	return rc.attr.GetInt("maxPvpLevel")
}

func (rc *rebornComponent) buyCard(goodsID int) (string, string, int, int, error) {
	rebornCardGoods := gamedata.GetGameData(consts.RebornSoldCard).(*gamedata.RebornSoldCardGameData)
	goodsData, ok := rebornCardGoods.ID2Card[goodsID]
	if !ok {
		return "", "", 0, 0, gamedata.GameError(2)
	}

	resCpt := rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Feats, goodsData.HonorPrice) {
		return "", "", 0, 0, gamedata.GameError(3)
	}

	cardCpt := rc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	if cardCpt.GetCollectCard(goodsData.CardID) != nil {
		return "", "", 0, 0, gamedata.GameError(4)
	}

	resCpt.ModifyResource(consts.Feats, -goodsData.HonorPrice, consts.RmrUnknownConsume)
	cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
		goodsData.CardID: &pb.CardInfo{CardId: goodsData.CardID, Amount: 1},
	})

	card := cardCpt.GetCollectCard(goodsData.CardID)
	if card != nil {
		card.SetFrom(consts.FromPieceShop)
	}

	return fmt.Sprintf("card%d", goodsData.CardID), card.GetCardGameData().GetName(),
		consts.Feats, goodsData.HonorPrice, nil
}

func (rc *rebornComponent) buyPrivilege(goodsID int) (string, string, int, int, error) {
	rebornPrivGoods := gamedata.GetGameData(consts.RebornSoldPriv).(*gamedata.RebornSoldPrivGameData)
	goodsData, ok := rebornPrivGoods.ID2Priv[goodsID]
	if !ok {
		return "", "", 0, 0, gamedata.GameError(2)
	}

	resCpt := rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Prestige, goodsData.PrestigePrice) {
		return "", "", 0, 0, gamedata.GameError(3)
	}

	statusID := fmt.Sprintf("%s%d", consts.OtBuffPrefix, goodsData.PrivID)
	if module.OutStatus.GetStatus(rc.player, statusID) != nil {
		return "", "", 0, 0, gamedata.GameError(4)
	}

	resCpt.ModifyResource(consts.Prestige, -goodsData.PrestigePrice, consts.RmrUnknownConsume)
	module.OutStatus.AddStatus(rc.player, statusID, -1)

	itemID := fmt.Sprintf("priv%d", goodsData.PrivID)
	return itemID, itemID, consts.Prestige, goodsData.PrestigePrice, nil
}

func (rc *rebornComponent) buyCardSkin(goodsID int) (string, string, int, int, error) {
	rebornSkinGoods := gamedata.GetGameData(consts.RebornSoldSkin).(*gamedata.RebornSoldSkinGameData)
	goodsData, ok := rebornSkinGoods.ID2Skin[goodsID]
	if !ok {
		return "", "", 0, 0, gamedata.GameError(2)
	}

	resCpt := rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Prestige, goodsData.HonorPrice) {
		return "", "", 0, 0, gamedata.GameError(3)
	}

	if module.Bag.HasItem(rc.player, consts.ItCardSkin, goodsData.SkinID) {
		return "", "", 0, 0, gamedata.GameError(4)
	}

	resCpt.ModifyResource(consts.Prestige, -goodsData.HonorPrice, consts.RmrUnknownConsume)
	it := module.Bag.AddCardSkin(rc.player, goodsData.SkinID)
	return it.GetGmID(), it.GetName(), consts.Prestige, goodsData.HonorPrice, nil
}

func (rc *rebornComponent) buyEquip(goodsID int) (string, string, int, int, error) {
	rebornEquipGoods := gamedata.GetGameData(consts.RebornSoldEquip).(*gamedata.RebornSoldEquipGameData)
	goodsData, ok := rebornEquipGoods.ID2Equip[goodsID]
	if !ok {
		return "", "", 0, 0, gamedata.GameError(2)
	}

	resCpt := rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	if !resCpt.HasResource(consts.Reputation, goodsData.Price) {
		return "", "", 0, 0, gamedata.GameError(3)
	}

	if module.Bag.HasItem(rc.player, consts.ItEquip, goodsData.EquipID) {
		return "", "", 0, 0, gamedata.GameError(4)
	}

	resCpt.ModifyResource(consts.Reputation, -goodsData.Price, consts.RmrRebornBuyEquip)
	glog.Infof("reborn buyEquip, uid=%d, equipID=%s", rc.player.GetUid(), goodsData.EquipID)
	module.Bag.AddEquip(rc.player, goodsData.EquipID)

	it := module.Bag.GetItem(rc.player, consts.ItEquip, goodsData.EquipID)
	if it != nil {
		eit, ok := it.(types.IEquipItem)
		if ok {
			eit.SetFrom(consts.FromReborn)
			eit.SetObtainTime(time.Now().Unix())
		}
		return it.GetGmID(), it.GetName(), consts.Reputation, goodsData.Price, nil
	}

	return "", "", 0, 0, nil
}

func (rc *rebornComponent) canReborn() (treausreModelID string, gold int, allCollectCards []types.ICollectCard,
	resetUnlockCards common.UInt32Set, skyBook int, err error) {

	err = gamedata.GameError(10)
	return

	// FIXME 首冲貂蝉没算解锁

	if module.Campaign.IsInWar() {
		err = gamedata.GameError(4)
		return
	}

	pvpTeam := rc.player.GetPvpTeam()
	rebornTreausreGameData := gamedata.GetGameData(consts.RebornTreausre).(*gamedata.RebornTreausreGameData)
	treausreModelID = rebornTreausreGameData.Team2Treausre[pvpTeam]
	gold = rebornTreausreGameData.Team2Gold[pvpTeam]
	if treausreModelID == "" && gold <= 0 {
		err = gamedata.GameError(1)
		return
	}

	resCpt := rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
	pvpScore := resCpt.GetResource(consts.Score)
	if module.Pvp.GetMinStarByPvpLevel(rebornPvpLevel) >= pvpScore {
		err = gamedata.GameError(3)
		return
	}

	curCnt := rc.attr.GetInt("rebornCnt") + 1
	rebornCntGameData := gamedata.GetGameData(consts.RebornCnt).(*gamedata.RebornCntGameData)
	skyBook = rebornCntGameData.Cnt2BookAmount[curCnt]
	if skyBook <= 0 {
		err = gamedata.GameError(7)
		return
	}

	levelGameData := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData)
	levelCpt := rc.player.GetComponent(consts.LevelCpt).(types.ILevelComponent)
	if levelCpt.GetCurLevel() <= levelGameData.GetMaxLevelID() {
		err = gamedata.GameError(8)
		return
	}

	// 这次重生后，理应共获得过多少天数
	skyBook = 0
	for cnt := 1; cnt <= curCnt; cnt++ {
		skyBook += rebornCntGameData.Cnt2BookAmount[cnt]
	}

	// 找出所有已解锁卡牌
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	rankData := rankGameData.Ranks[rebornPvpLevel]
	// 重生后的段位解锁哪些牌
	resetUnlockCards = common.UInt32Set{}
	resetUnlockCards.AddList(rankData.Unlock)

	rankData = rankGameData.Ranks[rc.player.GetPvpLevel()]
	// 当前段位解锁哪些牌
	unlockCards := common.UInt32Set{}
	unlockCards.AddList(rankData.Unlock)

	for _, levelData := range levelGameData.LevelList {
		if rebornPvpLevel >= levelData.RankCondition {
			if !levelCpt.IsClearLevel(levelData.ID) {
				break
			}
			if levelData.GeneralUnlock > 0 {
				resetUnlockCards.Add(levelData.GeneralUnlock)
				unlockCards.Add(levelData.GeneralUnlock)
			}
		} else {
			break
		}
	}

	cardCpt := rc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	allCollectCards = cardCpt.GetAllCollectCards()

	// 所有已解锁卡是否达到最高级
	isAllUnlockMaxLevel := true
	var usedSkyBook int
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	for _, card := range allCollectCards {
		if card.GetCardGameData().IsSpCard() {
			continue
		}
		if !card.IsMaxCanUpLevel() {
			err = gamedata.GameError(5)
			return
		}

		if card.GetMaxUnlockLevel() <= 0 {
			isAllUnlockMaxLevel = false
		} else {
			data := poolGameData.GetCard(card.GetCardID(), card.GetMaxUnlockLevel()-1)
			if data != nil {
				usedSkyBook += data.ConsumeBook
			}
		}

		unlockCards.Remove(card.GetCardID())
	}

	if unlockCards.Size() > 0 {
		err = gamedata.GameError(5)
		return
	}

	if !isAllUnlockMaxLevel && resCpt.HasResource(consts.SkyBook, 1) {
		err = gamedata.GameError(6)
		return
	}

	skyBook -= usedSkyBook
	if skyBook < 0 {
		skyBook = 0
	}
	rc.attr.SetInt("rebornCnt", curCnt)
	return
}

func (rc *rebornComponent) rebornOneCard(card types.ICollectCard, resetUnlockCards common.UInt32Set,
	modifyCards map[uint32]*pb.CardInfo) int {

	cardID := card.GetCardID()
	if config.GetConfig().IsXfServer() {
		// 保留溢出数量，变成1级
		modifyCards[cardID] = &pb.CardInfo{
			Level: 1 - int32(card.GetLevel()),
		}
		return 0
	} else {

		amount := card.GetAmount()
		if resetUnlockCards.Contains(cardID) {
			// 重生后已解锁，变成3级0张
			modifyCards[cardID] = &pb.CardInfo{
				Level:  3 - int32(card.GetLevel()),
				Amount: -int32(amount),
			}

		} else {
			// 重生后未解锁，删除
			modifyCards[cardID] = &pb.CardInfo{
				Level: -int32(card.GetLevel()),
			}
		}
		return amount
	}
}

func (rc *rebornComponent) reborn() (*pb.RebornReply, error) {
	treausreModelID, rewardGold, allCollectCards, resetUnlockCards, skyBook, err := rc.canReborn()
	if err != nil {
		return nil, err
	}

	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	modifyCards := map[uint32]*pb.CardInfo{}
	cardCaculGameData := gamedata.GetGameData(consts.RebornCardCacul).(*gamedata.RebornCardCaculGameData)
	goldCaculGameData := gamedata.GetGameData(consts.RebornGoldCacul).(*gamedata.RebornGoldCaculGameData)
	var feats float32
	var gold int
	for _, card := range allCollectCards {
		cardID := card.GetCardID()
		if _, ok := modifyCards[cardID]; ok {
			continue
		}

		cardData := card.GetCardGameData()
		f, ok := cardCaculGameData.Start2Feats[cardData.Rare]
		if !ok {
			continue
		}

		amount := rc.rebornOneCard(card, resetUnlockCards, modifyCards)
		level := card.GetLevel() - 1
		cardData = poolGameData.GetCard(cardID, level)
		for cardData != nil {
			amount += cardData.LevelupNum
			gold += cardData.LevelupGold
			level--
			cardData = poolGameData.GetCard(cardID, level)
		}
		feats += f * float32(amount)
	}

	cardCpt := rc.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	prestige := int(float32(gold) / float32(goldCaculGameData.Cacul.Gold) * goldCaculGameData.Cacul.Honor)

	/*
		resetUnlockCards.ForEach(func(cardID uint32) bool {
			if _, ok := modifyCards[cardID]; ok {
				return true
			}
			if cardCpt.GetCollectCard(cardID) != nil {
				return true
			}
			modifyCards[cardID] = &pb.CardInfo{
				Level: 3,
			}
			return true
		})
	*/

	glog.Infof("reborn uid=%d, pvpTeam=%d, gold=%d, feats=%d, prestige=%d, skyBook=%d, cards=%v, rewardGold=%d",
		rc.player.GetUid(), rc.player.GetPvpTeam(), gold, int(feats), prestige, skyBook, allCollectCards, rewardGold)
	reputation := prestige + int(feats)
	rc.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).BatchModifyResource(map[int]int{
		consts.Score: module.Pvp.GetMinStarByPvpLevel(rebornPvpLevel) - rc.player.GetPvpScore(),
		//consts.Feats: int(feats),
		//consts.Prestige: prestige + int(feats),
		consts.Reputation: reputation,
		consts.SkyBook:    skyBook,
		consts.Gold:       rewardGold,
	}, consts.RmrReborn)
	cardCpt.ModifyCollectCards(modifyCards)
	rc.attr.SetInt("maxPvpLevel", rebornPvpLevel)
	rc.attr.SetInt("rebornDay3", timer.GetDayNo())
	eventhub.Publish(consts.EvReborn, rc.player)

	logic.PushBackend("", 0, pb.MessageID_G2R_UPDATE_PVP_SCORE, module.Player.PackUpdateRankMsg(rc.player,
		[]*pb.SkinGCard{}, 0))

	return &pb.RebornReply{
		TreasureReward: rc.player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(
			treausreModelID, false),
		Reputation: int32(reputation),
		NewName:    rc.player.GetName(),
		Gold:       int32(rewardGold),
	}, nil
}

func (rc *rebornComponent) onPvpLevelUpdate(pvpLevel int) {
	maxLevel := rc.getMaxPvpLevel()
	if maxLevel <= 0 || pvpLevel <= maxLevel {
		return
	}
	rc.attr.SetInt("maxPvpLevel", pvpLevel)
	var cards []uint32
	rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
	levelGameData := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData)
	levelCpt := rc.player.GetComponent(consts.LevelCpt).(types.ILevelComponent)
	var curLevelID int

	for pvpLevel2 := maxLevel + 1; pvpLevel2 <= pvpLevel; pvpLevel2++ {
		rankData := rankGameData.Ranks[pvpLevel2]
		if rankData != nil && len(rankData.OriginUnlockCard) > 0 {
			cards = append(cards, rankData.OriginUnlockCard...)
		}

		for _, levelData := range levelGameData.LevelList {
			if levelData.ID <= curLevelID {
				continue
			}
			if levelData.RankCondition <= maxLevel {
				curLevelID = levelData.ID
				continue
			}
			if levelData.RankCondition > pvpLevel2 || !levelCpt.IsClearLevel(levelData.ID) {
				break
			}
			curLevelID = levelData.ID
			if levelData.GeneralUnlock > 0 {
				cards = append(cards, levelData.GeneralUnlock)
			}
		}
	}

	glog.Infof("reborn onPvpLevelUpdate uid=%d, pvpLevel=%d, cards=%v", rc.player.GetUid(), pvpLevel, cards)
	if len(cards) > 0 {
		modifyCards := map[uint32]*pb.CardInfo{}
		for _, cardID := range cards {
			info, ok := modifyCards[cardID]
			if !ok {
				info = &pb.CardInfo{}
				modifyCards[cardID] = info
			}
			info.Amount = 1
		}
		rc.player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(modifyCards)
	}
	return
}

func (rc *rebornComponent) getRebornCnt() int {
	return rc.attr.GetInt("rebornCnt")
}

func (rc *rebornComponent) getRebornRemainDay() int {
	return 0
	rebornDay := rc.attr.GetInt("rebornDay3")
	if rebornDay <= 0 {
		return 0
	} else {
		day := timer.GetDayNo() - rebornDay
		remainDay := rebornCD - day
		if remainDay < 0 {
			remainDay = 0
		}
		return remainDay
	}
}

// 老版本回收特权
func (rc *rebornComponent) resetPrivileges() []int {
	var privileges []int
	privilegesAttr := rc.attr.GetMapAttr("privileges")
	if privilegesAttr == nil {
		return privileges
	}

	rc.attr.Del("privileges")
	privilegesAttr.ForEachKey(func(key string) {
		privID, _ := strconv.Atoi(key)
		privileges = append(privileges, privID)
	})
	return privileges
}
