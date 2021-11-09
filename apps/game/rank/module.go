package rank

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
)

var _ module.IRankModule = &rankModule{}

var mod *rankModule

type rankModule struct {
	area2ranking map[int]map[common.UUid]uint32
	isSuccess    bool
}

func newRankModule() *rankModule {
	r := &rankModule{}
	return r
}

func (rm *rankModule) GetRanking(player types.IPlayer) (uint32, bool) {
	if !rm.isSuccess {
		rm.getRankBoardPre3Ranking()
	}
	area := player.GetArea()
	uid := player.GetUid()
	ranking, ok := rm.area2ranking[area][uid]
	return ranking, ok
}

func (rm *rankModule) getRankBoardPre3Ranking() {
	msg := &pb.MaxRankArg{MaxRank: 3}
	reply, err := logic.CallBackend("", 0, pb.MessageID_G2R_FETCH_PLAYER_RANK, msg)
	rm.isSuccess = false
	if err == nil {
		area2UserRanking := reply.(*pb.Area2UserRanking)
		rm.setArea2Ranking(area2UserRanking)
		rm.isSuccess = true
	}
}

func (rm *rankModule) setArea2Ranking(area2UserRanking *pb.Area2UserRanking) {
	rm.area2ranking = map[int]map[common.UUid]uint32{}
	for i, area := range area2UserRanking.Areas {
		userRankingInfo := area2UserRanking.UserRanking[i]

		userInfo := map[common.UUid]uint32{}
		for j, uid := range userRankingInfo.Uids {
			userInfo[common.UUid(uid)] = uint32(j + 1)
		}
		rm.area2ranking[int(area)] = userInfo
	}
}

func Initialize() {
	mod = newRankModule()
	module.Rank = mod
	registerRpc()
	mod.getRankBoardPre3Ranking()
}
