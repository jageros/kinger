package mail

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strings"
)

type mailReward struct {
	mailID int
	attrs  *attribute.ListAttr
}

func newMailReward() *mailReward {
	return &mailReward{
		attrs: attribute.NewListAttr(),
	}
}

func newMailRewardByAttr(attr *attribute.ListAttr) *mailReward {
	return &mailReward{
		attrs: attr,
	}
}

func copyMailRewardFromAttr(attr *attribute.ListAttr) *mailReward {
	attrs := attribute.NewListAttr()
	attr.ForEachIndex(func(index int) bool {
		a := attr.GetMapAttr(index)
		a2 := attribute.NewMapAttr()
		a2.SetInt("type", a.GetInt("type"))
		a2.SetInt("amount", a.GetInt("amount"))
		a2.SetUInt32("cardID", a.GetUInt32("cardID"))
		a2.SetStr("itemID", a.GetStr("itemID"))
		a2.SetInt32("emoji", a.GetInt32("emoji"))
		attrs.AppendMapAttr(a2)
		return true
	})
	return &mailReward{
		attrs: attrs,
	}
}

func newMailRewardByMsg(msg []*pb.MailReward) *mailReward {
	if len(msg) <= 0 {
		return nil
	}
	attrs := attribute.NewListAttr()
	for _, m := range msg {
		a := attribute.NewMapAttr()
		a.SetInt("type", int(m.Type))
		a.SetInt("amount", int(m.Amount))
		if m.CardID > 0 {
			a.SetUInt32("cardID", m.CardID)
		}
		if m.EmojiTeam > 0 {
			a.SetInt32("emoji", m.EmojiTeam)
		}
		a.SetStr("itemID", m.ItemID)
		attrs.AppendMapAttr(a)
	}
	return &mailReward{
		attrs: attrs,
	}
}

func (mr *mailReward) PackMsg() []*pb.MailReward {
	var rewards []*pb.MailReward
	mr.attrs.ForEachIndex(func(index int) bool {
		attr := mr.attrs.GetMapAttr(index)
		r := &pb.MailReward{
			Type:   pb.MailRewardType(attr.GetInt("type")),
			Amount: int32(attr.GetInt("amount")),
			ItemID: attr.GetStr("itemID"),
		}
		if r.Type == pb.MailRewardType_MrtCard {
			r.CardID = attr.GetUInt32("cardID")
		}
		if r.Type == pb.MailRewardType_MrtEmoji {
			r.EmojiTeam = attr.GetInt32("emoji")
		}
		rewards = append(rewards, r)
		return true
	})
	return rewards
}

func (mr *mailReward) String() string {
	s := strings.Builder{}
	s.WriteString("[reward ")
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		s.WriteString(fmt.Sprintf("{type=%s, amount=%d, cardID=%d, itemID=%s, emojiTeam=%d}",
			pb.MailRewardType(a.GetInt("type")), a.GetInt("amount"), a.GetUInt32("cardID"),
			a.GetStr("itemID"), a.GetInt32("emoji")))
		return true
	})
	s.WriteString("]")
	return s.String()
}

func (mr *mailReward) AddGold(amount int) {
	var attr *attribute.MapAttr
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == pb.MailRewardType_MrtGold {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(pb.MailRewardType_MrtGold))
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+amount)
}

func (mr *mailReward) AddJade(amount int) {
	var attr *attribute.MapAttr
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == pb.MailRewardType_MrtJade {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(pb.MailRewardType_MrtJade))
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+amount)
}

func (mr *mailReward) AddAmountByType(rewardType pb.MailRewardType, amount int) {
	var attr *attribute.MapAttr
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == rewardType {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(rewardType))
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+amount)
}

func (mr *mailReward) AddCard(cardID uint32, amount int) {
	var attr *attribute.MapAttr
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == pb.MailRewardType_MrtCard &&
			a.GetUInt32("cardID") == cardID {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(pb.MailRewardType_MrtCard))
		attr.SetUInt32("cardID", cardID)
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+amount)
}

func (mr *mailReward) AddEmoji(emojiTeam int) {
	var attr *attribute.MapAttr
	emoji := int32(emojiTeam)
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == pb.MailRewardType_MrtEmoji &&
			a.GetInt32("emoji") == emoji {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(pb.MailRewardType_MrtEmoji))
		attr.SetInt32("emoji", emoji)
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+1)
}

func (mr *mailReward) AddItem(rewardType pb.MailRewardType, itemID string, amount int) {
	var attr *attribute.MapAttr
	mr.attrs.ForEachIndex(func(index int) bool {
		a := mr.attrs.GetMapAttr(index)
		if pb.MailRewardType(a.GetInt("type")) == rewardType &&
			a.GetStr("itemID") == itemID {
			attr = a
			return false
		}
		return true
	})

	if attr == nil {
		attr = attribute.NewMapAttr()
		attr.SetInt("type", int(rewardType))
		attr.SetStr("itemID", itemID)
		mr.attrs.AppendMapAttr(attr)
	}

	attr.SetInt("amount", attr.GetInt("amount")+amount)
}

func (mr *mailReward) getReward(player types.IPlayer) (amountRewards []*pb.MailRewardAmountArg, treasureRewards []*pb.OpenTreasureReply,
	cards []*pb.MailRewardCard, itemRewards []*pb.MailRewardItemArg, emojis []int32) {

	resReward := map[int]int{}
	cardReward := map[uint32]*pb.CardInfo{}
	var treasureModelIDs []string
	var contribution int
	mr.attrs.ForEachIndex(func(index int) bool {
		attr := mr.attrs.GetMapAttr(index)
		rewardType := pb.MailRewardType(attr.GetInt("type"))
		amount := attr.GetInt("amount")
		switch rewardType {
		case pb.MailRewardType_MrtGold:
			resReward[consts.Gold] += amount
		case pb.MailRewardType_MrtJade:
			resReward[consts.Jade] += amount
		case pb.MailRewardType_MrtFeats:
			resReward[consts.Feats] += amount
		case pb.MailRewardType_MrtPrestige:
			resReward[consts.Prestige] += amount
		case pb.MailRewardType_MrtBowlder:
			resReward[consts.Bowlder] += amount
		case pb.MailRewardType_MrtReputation:
			resReward[consts.Reputation] += amount
		case pb.MailRewardType_MrtCardPiece:
			resReward[consts.CardPiece] += amount
		case pb.MailRewardType_MrtSkinPiece:
			resReward[consts.SkinPiece] += amount
		case pb.MailRewardType_MrtAccTicket:
			resReward[consts.AccTreasureCnt] += amount
		case pb.MailRewardType_MrtContribution:
			contribution += amount
		case pb.MailRewardType_MrtCard:
			cardID := attr.GetUInt32("cardID")
			info := cardReward[cardID]
			if info == nil {
				info = &pb.CardInfo{}
				cardReward[cardID] = info
			}
			info.Amount += int32(amount)
		case pb.MailRewardType_MrtCardSkin:
			skinID := attr.GetStr("itemID")
			module.Bag.AddCardSkin(player, skinID)
			itemRewards = append(itemRewards, &pb.MailRewardItemArg{
				Type:   pb.MailRewardType_MrtCardSkin,
				ItemID: skinID,
			})
		case pb.MailRewardType_MrtHeadFrame:
			headFrameID := attr.GetStr("itemID")
			module.Bag.AddHeadFrame(player, headFrameID)
			itemRewards = append(itemRewards, &pb.MailRewardItemArg{
				Type:   pb.MailRewardType_MrtHeadFrame,
				ItemID: headFrameID,
			})
		case pb.MailRewardType_MrtEquip:
			equipID := attr.GetStr("itemID")
			module.Bag.AddEquip(player, equipID)
			itemRewards = append(itemRewards, &pb.MailRewardItemArg{
				Type:   pb.MailRewardType_MrtEquip,
				ItemID: equipID,
			})
		case pb.MailRewardType_MrtTreasure:
			treasureModelID := attr.GetStr("itemID")
			treasureModelIDs = append(treasureModelIDs, treasureModelID)
			treasureRewards = append(treasureRewards, module.Treasure.OpenTreasureByModelID(player, treasureModelID, false))
		case pb.MailRewardType_MrtEmoji:
			emojiTeam := attr.GetInt32("emoji")
			module.Bag.AddEmoji(player, int(emojiTeam))
			emojis = append(emojis, emojiTeam)
		}
		return true
	})

	glog.Infof("get mail reward, uid=%d, mailID=%d, res=%v, cards=%v, items=%v, treasures=%v, emojis=%v, contribution=%d",
		player.GetUid(), mr.mailID, resReward, cards, itemRewards, treasureModelIDs, emojis, contribution)

	if len(resReward) > 0 {
		player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).BatchModifyResource(resReward, consts.RmrMailReward)
		for resType, amount := range resReward {
			if amount == 0 {
				continue
			}

			var type_ pb.MailRewardType
			switch resType {
			default:
				continue
			case consts.Gold:
				type_ = pb.MailRewardType_MrtGold
			case consts.Jade:
				type_ = pb.MailRewardType_MrtJade
			case consts.Feats:
				type_ = pb.MailRewardType_MrtFeats
			case consts.Prestige:
				type_ = pb.MailRewardType_MrtPrestige
			case consts.Bowlder:
				type_ = pb.MailRewardType_MrtBowlder
			case consts.Reputation:
				type_ = pb.MailRewardType_MrtReputation
			case consts.CardPiece:
				type_ = pb.MailRewardType_MrtCardPiece
			case consts.SkinPiece:
				type_ = pb.MailRewardType_MrtSkinPiece
			case consts.AccTreasureCnt:
				type_ = pb.MailRewardType_MrtAccTicket
			}

			amountRewards = append(amountRewards, &pb.MailRewardAmountArg{
				Type:   type_,
				Amount: int32(amount),
			})
		}
	}
	if len(cardReward) > 0 {
		for cardID, info := range cardReward {
			cards = append(cards, &pb.MailRewardCard{
				CardID: cardID,
				Amount: info.Amount,
			})
		}
		player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(cardReward)
	}
	if contribution > 0 {
		module.Campaign.ModifyContribution(player, contribution)
		amountRewards = append(amountRewards, &pb.MailRewardAmountArg{
			Type:   pb.MailRewardType_MrtContribution,
			Amount: int32(contribution),
		})
	}

	return
}
