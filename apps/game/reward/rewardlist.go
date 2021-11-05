package reward

import (
	"fmt"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/utils"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/proto/pb"
	"strconv"
	"strings"
	"unicode"
)

const (
	//奖励类型
	ty_unknow_ = iota
	ty_resource_
	ty_headFrame_
	ty_chatPop_
	ty_card_
	ty_randomCard_
	ty_treasure_
	ty_skin_
)

func (m *rewardModule) GiveRewardList(player types.IPlayer, rwList []string, source string) *pb.RewardList {
	var reply = &pb.RewardList{
		StuffList: map[string]int32{},
	}
	for _, rw := range rwList {
		stuffs := strings.Split(rw, ":")
		if len(stuffs) < 2 {
			continue
		}
		num, err := strconv.Atoi(stuffs[1])
		if err != nil {
			continue
		}
		giveReward(player, stuffs[0], num, source, reply)
	}
	return reply
}

func giveReward(player types.IPlayer, stuff string, num int, source string, rd *pb.RewardList) error {
	rewardTy, resType := getRewardType(stuff)
	switch rewardTy {
	case ty_resource_:
		module.Player.ModifyResource(player, resType, num, source)
		rd.StuffList[stuff] = int32(num)
		glog.Infof("GiveReward give player resource, uid=%d, resType=%d, num=%d, source=%s",player.GetUid(), resType, num, source)

	case ty_headFrame_:
		module.Bag.AddHeadFrame(player, strconv.Itoa(num))
		rd.StuffList[stuff] = int32(num)
		glog.Infof("GiveReward give player headFrame, uid=%d, headFrameID=%d, source=%s", player.GetUid(), num, source)

	case ty_chatPop_:
		module.Bag.AddChatPop(player, strconv.Itoa(num))
		rd.StuffList[stuff] = int32(num)
		glog.Infof("GiveReward give player chatPop, uid=%d, chatpopID=%d, source=%s",player.GetUid(), num, source)

	case ty_randomCard_:
		cardNum := num/8 + 1
		card := utils.RandUInt32Sample(module.Card.GetUnlockCards(player, 0), cardNum, false)
		cardSum := utils.RandFewNumberWithSum(num, cardNum)
		giveCard(player, card, cardSum, rd)

	case ty_card_:
		cardID, err := strconv.Atoi(stuff)
		if err != nil {
			return err
		}
		var card []uint32
		var nums []int
		card = append(card, uint32(cardID))
		nums = append(nums, num)
		giveCard(player, card, nums, rd)

	case ty_treasure_:
		for i := 0; i < num; i++ {
			treasureReward := module.Treasure.OpenTreasureByModelID(player, stuff, false)
			rd.TreasureReward = append(rd.TreasureReward, treasureReward)
		}
		rd.StuffList[stuff] = int32(num)

	case ty_skin_:
		module.Bag.AddCardSkin(player, stuff)
		rd.StuffList[stuff] = int32(num)

	default:

	}
	return nil
}

func giveCard(player types.IPlayer, cardID []uint32, num []int, rd *pb.RewardList) {

	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardMap := map[uint32]*pb.CardInfo{}
	for key, val := range cardID {
		rd.StuffList[strconv.Itoa(int(val))] = int32(num[key])
		cardMap[val] = &pb.CardInfo{
			Amount: int32(num[key]),
		}
	}
	cardCpt.ModifyCollectCards(cardMap)
	var cids string
	for k, cid := range cardID {
		cids = fmt.Sprintf("%s [cardID:%d num:%d]", cids, cid, num[k])
	}
	cids = fmt.Sprintf("{%s }", cids)
	glog.Infof("give player collect cards, uid=%d, cardSet=%s", player.GetUid(), cids)
}

func generateRandomCard(player types.IPlayer, num int)(cardIds []uint32, cardSums []int) {
	cardNum := num/8 + 1
	card := utils.RandUInt32Sample(module.Card.GetUnlockCards(player, 0), cardNum, false)
	cardSum := utils.RandFewNumberWithSum(num, cardNum)
	return card, cardSum
}

func getRewardType(stuff string) (rewardTy, resourceTy int) {
	isNumber := func() bool {
		for _, v := range stuff {
			if !unicode.IsNumber(v) {
				return false
			}
		}
		return true
	}

	switch {
	case stuff == "headFrame":
		return ty_headFrame_, 0
	case stuff == "chatPop":
		return ty_chatPop_, 0
	case stuff == "card":
		return ty_randomCard_, 0
	case strings.HasPrefix(stuff, "BX"):
		return ty_treasure_, 0
	case strings.HasPrefix(stuff, "SK"):
		return ty_skin_, 0
	case isNumber():
		return ty_card_, 0
	default:
		resType := getResourceType(stuff)
		isRes := resType > 0
		if isRes {
			return ty_resource_, resType
		}
		return ty_unknow_, 0
	}
}

func getResourceType(resName string) int {
	switch resName {
	case "weap":
		return consts.Weap
	case "horse":
		return consts.Horse
	case "mat":
		return consts.Mat
	case "gold":
		return consts.Gold
	case "forage":
		return consts.Forage
	case "med":
		return consts.Med
	case "ban":
		return consts.Ban
	case "wine":
		return consts.Wine
	case "book":
		return consts.Book
	case "mmr":
		return consts.Mmr
	case "jade":
		return consts.Jade
	case "feats":
		return consts.Feats
	case "prestige":
		return consts.Prestige
	case "reputation":
		return consts.Reputation
	case "bowlder":
		return consts.Bowlder
	case "skyBook":
		return consts.SkyBook
	case "eventItem1":
		return consts.EventItem1
	case "cardPiece":
		return consts.CardPiece
	case "skinPiece":
		return consts.SkinPiece
	case "crossAreaHonor":
		return consts.CrossAreaHonor
	//case "crossAreaBlotHonor":
	//	return consts.CrossAreaBlotHonor
	case "score":
		return consts.Score
	case "guidePro":
		return consts.GuidePro
	case "maxScore":
		return consts.MaxScore
	case "ticket":
		return consts.AccTreasureCnt
	case "notSubStarCnt":
		return consts.NotSubStarCnt
	case "pvpTreasureCnt":
		return consts.PvpTreasureCnt
	case "pvpGoldCnt":
		return consts.PvpGoldCnt
	case "winDiff":
		return consts.WinDiff
	case "matchScore":
		return consts.MatchScore
	case "maxMatchScore":
		return consts.MaxMatchScore
	case "kingFlag":
		return consts.KingFlag
	default:
		return 0
	}
}

func (m *rewardModule) GetMailItemType(stuff string) (isRes bool, ty pb.MailRewardType) {
	rewardTy, resType := getRewardType(stuff)
	switch rewardTy {
	case ty_resource_:
		switch resType {
		case consts.Gold:
			return true, pb.MailRewardType_MrtGold
		case consts.Jade:
			return true, pb.MailRewardType_MrtJade
		case consts.CardPiece:
			return true, pb.MailRewardType_MrtCardPiece
		case consts.SkinPiece:
			return true, pb.MailRewardType_MrtSkinPiece
		case consts.AccTreasureCnt:
			return true, pb.MailRewardType_MrtAccTicket
		}

	case ty_headFrame_:
		return false, pb.MailRewardType_MrtHeadFrame

	case ty_card_:
		return false, pb.MailRewardType_MrtCard

	case ty_treasure_:
		return false, pb.MailRewardType_MrtTreasure

	case ty_skin_:
		return false, pb.MailRewardType_MrtCardSkin

	default:

	}
	return false, -1
}

