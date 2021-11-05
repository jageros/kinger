package seasonpvp

import (
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/apps/logic"
	"kinger/gamedata"
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"encoding/json"
	"kinger/proto/pb"
)

type seasonPvpHdSessionData struct {
	attr *attribute.AttrMgr
	area int
	session int
	rewards []*gamedata.SeasonReward
	uid2Rank map[common.UUid]int
}

func newSeasonPvpHdSessionDataByAttr(session, area int, attr *attribute.AttrMgr) *seasonPvpHdSessionData {
	sd := &seasonPvpHdSessionData{
		session: session,
		area: area,
		attr: attr,
	}

	rankAttr := attr.GetListAttr("rank")
	if rankAttr != nil {
		sd.uid2Rank = map[common.UUid]int{}
		rankAttr.ForEachIndex(func(index int) bool {
			uid := common.UUid(rankAttr.GetUInt64(index))
			sd.uid2Rank[uid] = index + 1
			return true
		})
	}

	rewardAttr := attr.GetStr("reward")
	if rewardAttr != "" {
		err := json.Unmarshal([]byte(rewardAttr), &sd.rewards)
		if err != nil {
			glog.Errorf("newSeasonPvpHdSessionDataByAttr Unmarshal reward error, %s, err=%s", rewardAttr, err)
		}
	}
	return sd
}

func (sd *seasonPvpHdSessionData) getRank(uid common.UUid) int {
	if rank, ok := sd.uid2Rank[uid]; ok {
		return rank
	} else {
		return 1000000
	}
}

func (sd *seasonPvpHdSessionData) onSeasonStop() {
	if sd.uid2Rank != nil {
		return
	}

	seasonRewardGameData := gamedata.GetGameData(consts.SeasonReward).(*gamedata.SeasonRewardGameData)
	sd.rewards = seasonRewardGameData.GetRewards(sd.area)
	rewardAttr, err := json.Marshal(sd.rewards)
	if err != nil {
		glog.Infof("onSeasonStop Marshal reward error %s", err)
	} else {
		sd.attr.SetStr("reward", string(rewardAttr))
	}

	sd.uid2Rank = map[common.UUid]int{}
	rankAttr := attribute.NewListAttr()
	sd.attr.SetListAttr("rank", rankAttr)
	reply, err := logic.CallBackend("", 0, pb.MessageID_G2R_SEASON_PVP_END, &pb.TargetArea{Area: int32(sd.area)})

	var rankUids []uint64
	if err == nil {
		reply2 := reply.(*pb.G2RSeasonPvpEndReply)
		rankUids = reply2.RankUids
	}

	glog.Infof("seasonPvpHdSessionData onSeasonStop, session=%d, rank=%v", sd.session, rankUids)
	for i, uid := range rankUids {
		sd.uid2Rank[common.UUid(uid)] = i + 1
		rankAttr.AppendUInt64(uid)
	}
	sd.attr.Save(true)
}
