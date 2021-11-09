package reward

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/common/glog"
	"math/rand"
	"strconv"
)

type rewardModule struct {
	rewarders map[string]map[string]iRewarder
}

func (m *rewardModule) getRewarder(type_, itemID string) iRewarder {
	if m.rewarders == nil {
		m.rewarders = map[string]map[string]iRewarder{}
	}

	item2Rewarder, ok := m.rewarders[type_]
	if !ok {
		item2Rewarder = map[string]iRewarder{}
		m.rewarders[type_] = item2Rewarder
	}

	r, ok := item2Rewarder[itemID]
	if ok {
		return r
	}

	switch type_ {
	case "card":
		id, _ := strconv.Atoi(itemID)
		cardID := uint32(id)
		cardData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData).GetCard(cardID, 1)
		if cardData == nil {
			return nil
		}

		if cardData.Rare >= 99 {
			r = &spCardRewarder{cardID: cardID}
		} else {
			r = &cardRewarder{cardID: cardID}
		}

	case "skin":
		r = &cardSkinRewarder{skinID: itemID}

	case "equip":
		r = &equipRewarder{equipID: itemID}

	case "headFrame":
		r = &headFrameRewarder{headFrameID: itemID}

	case "emoji":
		emojiTeam, _ := strconv.Atoi(itemID)
		r = &emojiRewarder{emojiTeam: emojiTeam}

	case "resource":
		resType, _ := strconv.Atoi(itemID)
		r = &resourceRewarder{resType: resType}

	case "priv":
		privID, _ := strconv.Atoi(itemID)
		r = &privilegeRewarder{privID: privID}

	case "tryVip":
		r = &tryVipRewarder{}

	case "tryPriv":
		privID, _ := strconv.Atoi(itemID)
		r = &tryPrivilegeRewarder{privID: privID}

	default:
		return nil
	}

	item2Rewarder[itemID] = r
	return r
}

func (m *rewardModule) GiveReward(player types.IPlayer, rewardTbl string) types.IRewardResult {
	data := gamedata.GetGameData(rewardTbl)
	if data == nil {
		glog.Infof("GiveReward no rewardTbl %s", rewardTbl)
		return nil
	}

	rewardGameData, ok := data.(*gamedata.RewardTblGameData)
	if !ok {
		glog.Infof("GiveReward no rewardTbl2 %s", rewardTbl)
		return nil
	}

	var rr *rewardResult = nil
	for _, rewardTeamData := range rewardGameData.Team2Rewards {
		var pro int
		rate := rand.Intn(10000) + 1

		for _, rewardItem := range rewardTeamData.Rewards {

			if !rewardItem.AreaLimit.IsEffective(player.GetArea()) {
				continue
			}

			if rewardTbl == consts.RecruitTreausreCardRewardTbl || rewardTbl == consts.RecruitTreausreSkinRewardTbl {
				var flag int
				curids := module.Shop.GetRecruitCurIDs(player)
				for _, cid := range curids {
					if int(cid) == rewardItem.ID {
						flag = 1
						break
					}
				}
				if flag == 0 {
					continue
				}
			}

			pro += rewardItem.Pro
			if rate <= pro {

				r := m.getRewarder(rewardItem.Type, rewardItem.ItemID)
				if r != nil {

					if rr == nil {
						rr = &rewardResult{rewardTblName: rewardTbl}
					}
					r.doReward(player, rewardItem.Amount, rr)
					rr.addRewardIdx(rewardItem.ID)
				}
				break
			}
		}
	}

	if rr == nil {
		return nil
	}
	return rr
}

func Initialize() {
	module.Reward = &rewardModule{}
}
