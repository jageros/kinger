package rank

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
)

func rpc_C2S_FetchRank(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	return logic.CallBackend("", 0, pb.MessageID_G2R_FETCH_RANK, &pb.GFetchRankArg{
		Area: int32(player.GetArea()),
		Type: arg.(*pb.FetchRankArg).Type,
	})
}

func rpc_C2S_FetchSeasonRank(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	return logic.CallBackend("", 0, pb.MessageID_G2R_FETCH_SEASON_RANK, &pb.TargetArea{
		Area: int32(player.GetArea()),
	})
}

func rpc_R2G_SendRankReward(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RankHonorInfo)
	glog.Infof("rpc_R2G_SendRankReward, uids=%v, honors=%v", arg2.Uids, arg2.Honors)
	gdata := gamedata.GetGameData(consts.RankHonorReward).(*gamedata.RankHonorRewardData)

	for _, data := range gdata.RankReward {
		if len(data.Rank) <= 1 {
			continue
		}

		if len(arg2.Uids) == 0 || len(arg2.Uids) < data.Rank[0]{
			return nil, nil
		}

		var sendUIds []uint64
		if len(arg2.Uids) < data.Rank[1]{
			sendUIds = arg2.Uids[data.Rank[0]-1:]
		}else {
			sendUIds = arg2.Uids[data.Rank[0]-1: data.Rank[1]]
		}

		for index, uid := range sendUIds {
			sender := module.Mail.NewMailSender(common.UUid(uid))
			sender.SetTypeAndArgs(pb.MailTypeEnum_RankHonorReward, index+1, arg2.Honors[index])
			mailReward := sender.GetRewardObj()
			if data.Treasure != "" {
				mailReward.AddItem(pb.MailRewardType_MrtTreasure, data.Treasure, 1)
			}

			if data.Gold > 0 {
				mailReward.AddAmountByType(pb.MailRewardType_MrtGold, data.Gold)
			}

			if data.Jade > 0 {
				mailReward.AddAmountByType(pb.MailRewardType_MrtJade, data.Jade)
			}

			sender.Send()
		}
	}

	return nil, nil
}

func rpc_R2G_SendPlayerRank(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.Area2UserRanking)
	mod.setArea2Ranking(arg2)
	return nil, nil
}


func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_RANK, rpc_C2S_FetchRank)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SEASON_RANK, rpc_C2S_FetchSeasonRank)
	//logic.RegisterRpcHandler(pb.MessageID_R2G_SEND_RANK_HONOR, rpc_R2G_SendRankReward)
	logic.RegisterRpcHandler(pb.MessageID_R2G_SEND_PLAYER_RANK, rpc_R2G_SendPlayerRank)
}