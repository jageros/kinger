package reward

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"math/rand"
	"time"
)

type rewardResult struct {
	rewardTblName      string
	rewardIdxs         []int // 获得了哪些奖励？奖励表里的id
	resources          map[int]int
	cards              map[uint32]int
	cardSkins          []string
	equips             []string
	headFrames         []string
	emojis             []int
	privileges         []*pb.GiftPrivilegeReward
	convertResources   map[int]int
	upLevelRewardCards map[uint32]int
}

func (rr *rewardResult) GetRewardIdxs() []int {
	return rr.rewardIdxs
}

func (rr *rewardResult) addRewardIdx(idx int) {
	rr.rewardIdxs = append(rr.rewardIdxs, idx)
}

func (rr *rewardResult) GetResources() map[int]int {
	if rr.resources == nil {
		rr.resources = map[int]int{}
	}
	return rr.resources
}

func (rr *rewardResult) GetConvertResources() map[int]int {
	if rr.convertResources == nil {
		rr.convertResources = map[int]int{}
	}
	return rr.convertResources
}

func (rr *rewardResult) GetCards() map[uint32]int {
	if rr.cards == nil {
		rr.cards = map[uint32]int{}
	}
	return rr.cards
}

func (rr *rewardResult) GetUpLevelRewardCards() map[uint32]int {
	if rr.upLevelRewardCards == nil {
		rr.upLevelRewardCards = map[uint32]int{}
	}
	return rr.upLevelRewardCards
}

func (rr *rewardResult) GetCardSkins() []string {
	return rr.cardSkins
}

func (rr *rewardResult) GetEquips() []string {
	return rr.equips
}

func (rr *rewardResult) GetHeadFrames() []string {
	return rr.headFrames
}

func (rr *rewardResult) GetEmojis() []int {
	return rr.emojis
}

func (rr *rewardResult) PackOpenTreasureMsg() *pb.OpenTreasureReply {
	msg := &pb.OpenTreasureReply{
		OK:         true,
		CardSkins:  rr.cardSkins,
		Headframes: rr.headFrames,
	}

	if rr.cards != nil {
		for cardID, amount := range rr.cards {
			for i := 0; i < amount; i++ {
				msg.CardIDs = append(msg.CardIDs, cardID)
			}
		}
	}

	if rr.upLevelRewardCards != nil {
		for cardID, amount := range rr.upLevelRewardCards {
			for i := 0; i < amount; i++ {
				msg.UpLevelRewardCards = append(msg.UpLevelRewardCards, cardID)
			}
		}
	}

	for _, emoji := range rr.emojis {
		msg.EmojiTeams = append(msg.EmojiTeams, int32(emoji))
	}

	if rr.resources != nil {
		for resType, amount := range rr.resources {
			msg.Resources = append(msg.Resources, &pb.Resource{
				Type:   int32(resType),
				Amount: int32(amount),
			})
		}
	}

	if rr.convertResources != nil {
		for resType, amount := range rr.convertResources {
			msg.ConvertResources = append(msg.ConvertResources, &pb.Resource{
				Type:   int32(resType),
				Amount: int32(amount),
			})
		}
	}

	return msg
}

func (rr *rewardResult) addResource(resType, amount int) {
	resources := rr.GetResources()
	resources[resType] += amount
}

func (rr *rewardResult) addCard(cardID uint32, amount int) {
	cards := rr.GetCards()
	cards[cardID] += amount
}

func (rr *rewardResult) addCardSkin(skin string) {
	rr.cardSkins = append(rr.cardSkins, skin)
}

func (rr *rewardResult) addEquip(equip string) {
	rr.equips = append(rr.equips, equip)
}

func (rr *rewardResult) addHeadFrame(headFrame string) {
	rr.headFrames = append(rr.headFrames, headFrame)
}

func (rr *rewardResult) addEmoji(emoji int) {
	rr.emojis = append(rr.emojis, emoji)
}

func (rr *rewardResult) addConvertResources(resType, amount int) {
	resources := rr.GetConvertResources()
	resources[resType] += amount
}

func (rr *rewardResult) addUpLevelRewardCards(cardID uint32, amount int) {
	cards := rr.GetUpLevelRewardCards()
	cards[cardID] += amount
}

func (rr *rewardResult) addPrivilege(privID int, remainTime int) {
	rr.privileges = append(rr.privileges, &pb.GiftPrivilegeReward{
		PrivID:     int32(privID),
		RemainTime: int32(remainTime),
	})
}

func (rr *rewardResult) GetPrivileges() []*pb.GiftPrivilegeReward {
	return rr.privileges
}

type iRewarder interface {
	doReward(player types.IPlayer, amount int, rr *rewardResult)
}

type resourceRewarder struct {
	resType int
}

func (r *resourceRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	var oldPvpLevel int
	if r.resType == consts.Score {
		oldPvpLevel = player.GetPvpLevel()
	}

	module.Player.ModifyResource(player, r.resType, amount, consts.RmrRewardTbl+rr.rewardTblName)
	rr.addResource(r.resType, amount)

	if r.resType == consts.Score && oldPvpLevel < player.GetPvpLevel() {
		cards := player.GetComponent(consts.PvpCpt).(types.IPvpComponent).UplevelReward()
		for _, cardID := range cards {
			rr.addUpLevelRewardCards(cardID, 1)
		}
	}
}

type cardRewarder struct {
	cardID uint32
}

func (r *cardRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(map[uint32]*pb.CardInfo{
		r.cardID: &pb.CardInfo{Amount: int32(amount)},
	})
	rr.addCard(r.cardID, amount)
}

type spCardRewarder struct {
	cardID uint32
}

func (r *spCardRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	var pieceAmount int
	for i := 0; i < amount; i++ {
		if cardCpt.GetCollectCard(r.cardID) != nil {
			funcPriceGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
			pieceAmount += funcPriceGameData.SpCardToPiece
		} else {
			cardCpt.ModifyCollectCards(map[uint32]*pb.CardInfo{
				r.cardID: &pb.CardInfo{Amount: 1},
			})

			card := cardCpt.GetCollectCard(r.cardID)
			if card != nil {
				card.SetFrom(consts.FromPieceShop)
			}
		}
	}

	rr.addCard(r.cardID, amount)
	if pieceAmount > 0 {
		module.Player.ModifyResource(player, consts.CardPiece, pieceAmount, consts.RmrRewardTbl+rr.rewardTblName)
		rr.addConvertResources(consts.CardPiece, pieceAmount)
	}
}

type cardSkinRewarder struct {
	skinID string
}

func (r *cardSkinRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	var pieceAmount int
	for i := 0; i < amount; i++ {
		if module.Bag.GetItem(player, consts.ItCardSkin, r.skinID) != nil {
			funcPriceGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
			pieceAmount += funcPriceGameData.SkinToPiece
		} else {
			module.Bag.AddCardSkin(player, r.skinID)
		}

		rr.addCardSkin(r.skinID)
	}

	if pieceAmount > 0 {
		module.Player.ModifyResource(player, consts.SkinPiece, pieceAmount, consts.RmrRewardTbl+rr.rewardTblName)
		rr.addConvertResources(consts.SkinPiece, pieceAmount)
	}
}

type equipRewarder struct {
	equipID string
}

func (r *equipRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	module.Bag.AddEquip(player, r.equipID)
	equip := module.Bag.GetItem(player, consts.ItEquip, r.equipID)
	if equip != nil {
		equip.(types.IEquipItem).SetFrom(consts.FromPieceShop)
		equip.(types.IEquipItem).SetObtainTime(time.Now().Unix())
	}
	rr.addEquip(r.equipID)
	glog.Infof("equipRewarder uid=%d, equipID=%s, amount=%d", player.GetUid(), r.equipID, amount)
}

type headFrameRewarder struct {
	headFrameID string
}

func (r *headFrameRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	module.Bag.AddHeadFrame(player, r.headFrameID)
	rr.addHeadFrame(r.headFrameID)
	glog.Infof("headFrameRewarder uid=%d, headFrameID=%s, amount=%d", player.GetUid(), r.headFrameID, amount)
}

type emojiRewarder struct {
	emojiTeam int
}

func (r *emojiRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	module.Bag.AddEmoji(player, r.emojiTeam)
	rr.addEmoji(r.emojiTeam)
	glog.Infof("emojiRewarder uid=%d, emojiTeam=%d, amount=%d", player.GetUid(), r.emojiTeam, amount)
}

type privilegeRewarder struct {
	privID int
}

func (r *privilegeRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	st := module.OutStatus.GetBuff(player, r.privID)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if st == nil {
		st = module.OutStatus.AddBuff(player, r.privID, funcPrice.PrivContinuedTime)
	} else if st.GetRemainTime() >= 0 {
		st.Over(funcPrice.PrivContinuedTime)
	}

	rr.addPrivilege(r.privID, st.GetRemainTime())
}

type tryVipRewarder struct {
}

func (r *tryVipRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	st := module.OutStatus.GetStatus(player, consts.OtVipCard)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if st == nil {
		st = module.OutStatus.AddStatus(player, consts.OtVipCard, funcPrice.TryPrivContinuedTime)
	} else if st.GetRemainTime() >= 0 {
		st.Over(funcPrice.TryPrivContinuedTime)
	}
}

type tryPrivilegeRewarder struct {
	privID int
}

func (r *tryPrivilegeRewarder) doReward(player types.IPlayer, amount int, rr *rewardResult) {
	privID := r.privID
	if r.privID <= 0 {
		privGameData := gamedata.GetGameData(consts.PrivConfig).(*gamedata.PrivilegeGameData)
		privAmount := len(privGameData.Privileges)
		if privAmount > 0 {
			privID = privGameData.Privileges[rand.Intn(privAmount)].ID
		}
	}

	st := module.OutStatus.GetBuff(player, privID)
	funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if st == nil {
		st = module.OutStatus.AddBuff(player, privID, funcPrice.TryPrivContinuedTime)
	} else if st.GetRemainTime() >= 0 {
		st.Over(funcPrice.TryPrivContinuedTime)
	}

	rr.addPrivilege(privID, st.GetRemainTime())
}
