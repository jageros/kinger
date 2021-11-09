package social

import (
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
)

type wxInviteReward struct {
	id   int
	attr *attribute.MapAttr
}

func newWxInviteRewardByAttr(id int, attr *attribute.MapAttr) *wxInviteReward {
	if attr.GetBool("hasReward") {
		attr.SetBool("hasReward", false)
		attr.SetInt("rewardAmount", 1)
	}
	return &wxInviteReward{
		id:   id,
		attr: attr,
	}
}

func (wr *wxInviteReward) getAmount() int {
	return wr.attr.GetInt("amount")
}

func (wr *wxInviteReward) incrAmount() {
	wr.attr.SetInt("amount", wr.attr.GetInt("amount")+1)
}

func (wr *wxInviteReward) getRewardAmount() int {
	return wr.attr.GetInt("rewardAmount")
}

func (wr *wxInviteReward) canReward() bool {
	rewardAmount := wr.getRewardAmount()
	return rewardAmount < wr.getAmount() && rewardAmount < wr.getMaxRewardAmount()
}

func (wr *wxInviteReward) getGameData() *gamedata.WxInviteReward {
	return gamedata.GetGameData(consts.WxInviteReward).(*gamedata.WxInviteRewardGameData).ID2Reward[wr.id]
}

func (wr *wxInviteReward) getMaxRewardAmount() int {
	return wr.getGameData().RewardAmount
}

func (wr *wxInviteReward) packMsg() *pb.WxInviteReward {
	return &pb.WxInviteReward{
		ID:        int32(wr.id),
		CurCnt:    int32(wr.getAmount()),
		RewardCnt: int32(wr.getRewardAmount()),
	}
}

func (wr *wxInviteReward) doReward(player types.IPlayer) (*pb.GetWxInviteRewardReply, error) {
	rewardAmount := wr.attr.GetInt("rewardAmount")
	amount := wr.getAmount()
	maxRewardAmount := wr.getMaxRewardAmount()
	if rewardAmount >= maxRewardAmount || rewardAmount >= amount {
		return nil, gamedata.GameError(2)
	}

	cnt1 := amount - rewardAmount
	cnt2 := maxRewardAmount - rewardAmount
	cnt := cnt1
	if cnt > cnt2 {
		cnt = cnt2
	}

	rewardData := wr.getGameData()
	glog.Infof("getWxInviteFriendReward uid=%d, rewardID=%d, gold=%d, jade=%d, ticket=%d, cards=%v, cnt=%d",
		player.GetUid(), rewardData.GoldReward, rewardData.JadeReward, rewardData.TicketReward,
		rewardData.CardReward, cnt)

	wr.attr.SetInt("rewardAmount", wr.attr.GetInt("rewardAmount")+cnt)

	rewardRes := map[int]int{
		consts.Gold: rewardData.GoldReward * cnt,
		consts.Jade: rewardData.JadeReward * cnt,
		//consts.AccTreasureCnt: rewardData.TicketReward * cnt,
	}
	player.GetComponent(consts.ResourceCpt).(types.IResourceComponent).BatchModifyResource(rewardRes, consts.RmrWxInviteReward)

	var cardReward []uint32
	if len(rewardData.CardReward) > 0 {
		rewardCards := map[uint32]*pb.CardInfo{}
		for _, cardID := range rewardData.CardReward {
			info := rewardCards[cardID]
			if info == nil {
				info = &pb.CardInfo{}
				rewardCards[cardID] = info
			}
			info.Amount += int32(cnt)
			for i := 0; i < cnt; i++ {
				cardReward = append(cardReward, cardID)
			}
		}
		player.GetComponent(consts.CardCpt).(types.ICardComponent).ModifyCollectCards(rewardCards)
	}

	return &pb.GetWxInviteRewardReply{
		Gold:  int32(rewardRes[consts.Gold]),
		Jade:  int32(rewardRes[consts.Jade]),
		Cards: cardReward,
	}, nil
}
