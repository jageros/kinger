package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
	"sort"
)

type WxInviteReward struct {
	ID           int      `json:"__id__"`
	PvpLevel     int      `json:"pvpLevel"`
	RewardAmount int      `json:"cnt"`
	GoldReward   int      `json:"goldReward"`
	JadeReward   int      `json:"jadeReward"`
	CardReward   []uint32 `json:"cardReward"`
	TicketReward int `json:"ticketReward"`
}

type WxInviteRewardList []*WxInviteReward

func (l WxInviteRewardList) Less(i, j int) bool {
	return l[i].PvpLevel <= l[j].PvpLevel
}

func (l WxInviteRewardList) Len() int {
	return len(l)
}

func (l WxInviteRewardList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type WxInviteRewardGameData struct {
	baseGameData
	Rewards       WxInviteRewardList
	Level2Rewards map[int][]*WxInviteReward
	ID2Reward     map[int]*WxInviteReward
}

func newWxInviteRewardGameData() *WxInviteRewardGameData {
	r := &WxInviteRewardGameData{}
	r.i = r
	return r
}

func (rd *WxInviteRewardGameData) name() string {
	return consts.WxInviteReward
}

func (rd *WxInviteRewardGameData) init(d []byte) error {
	var l WxInviteRewardList

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	sort.Sort(l)
	rd.Rewards = l

	rd.Level2Rewards = map[int][]*WxInviteReward{}
	rd.ID2Reward = map[int]*WxInviteReward{}
	for _, r := range l {
		rs := rd.Level2Rewards[r.PvpLevel]
		rd.Level2Rewards[r.PvpLevel] = append(rs, r)
		rd.ID2Reward[r.ID] = r
	}

	return nil
}
