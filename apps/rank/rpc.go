package main

import (
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
)

func rpc_C2S_FetchRankUser(_ *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchRankUserArg)
	ri := rankMgr.getTodayRankItem(common.UUid(arg2.Uid))
	if ri == nil {
		return &pb.RankUser{}, nil
	} else {
		return ri.packUserMsg(), nil
	}
}

func rpc_C2S_FetchRank(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GFetchRankArg)
	if arg2.Type == pb.RankType_RtUnknow {
		arg2.Type = pb.RankType_RtLadder
	}
	board := rankMgr.getBoard(arg2.Type, int(arg2.Area))
	if board == nil {
		return nil, gamedata.GameError(1)
	}
	return board.packMsg(), nil
}

func rpc_C2S_FetchSeasonRank(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetArea)
	board := rankMgr.getBoard(pb.RankType_RtSeason, int(arg2.Area))
	return board.packMsg(), nil
}

func rpc_G2R_UpdatePvpScore(_ *network.Session, arg interface{}) (interface{}, error) {
	rankMgr.updateRankScore(arg.(*pb.UpdatePvpScoreArg))
	return nil, nil
}

func rpc_G2R_SeasonPvpBegin(_ *network.Session, arg interface{}) (interface{}, error) {
	rankMgr.onSeasonPvpBegin(int(arg.(*pb.TargetArea).Area))
	return nil, nil
}

func rpc_G2R_SeasonPvpEnd(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TargetArea)
	board := rankMgr.getBoard(pb.RankType_RtSeason, int(arg2.Area))
	if board == nil {
		return nil, gamedata.InternalErr
	}

	board.refreshCurRankList(false)
	board.onSeasonPvpEnd()
	return &pb.G2RSeasonPvpEndReply{
		RankUids: board.getCurRankList(10),
	}, nil
}

func rpc_G2R_fetchPlayerRank(_ *network.Session, arg interface{}) (interface{}, error) {
	maxRank := int(arg.(*pb.MaxRankArg).MaxRank)
	return rankMgr.getBoardUsersByArea(pb.RankType_RtLadder, maxRank, 0), nil
}

func rpc_G2R_LeagueSeasonEnd(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GLeagueSeasonEndArg)
	board := rankMgr.getBoard(pb.RankType_RtLadder, int(arg2.Area))
	board.refreshCurRankList(false)
	board.onLeagueSeasonEnd()
	return rankMgr.getBoardUsersByArea(pb.RankType_RtLadder, 200, int(arg2.Area)), nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_RANK_USER, rpc_C2S_FetchRankUser)

	logic.RegisterRpcHandler(pb.MessageID_G2R_FETCH_RANK, rpc_C2S_FetchRank)
	logic.RegisterRpcHandler(pb.MessageID_G2R_FETCH_SEASON_RANK, rpc_C2S_FetchSeasonRank)
	logic.RegisterRpcHandler(pb.MessageID_G2R_UPDATE_PVP_SCORE, rpc_G2R_UpdatePvpScore)
	logic.RegisterRpcHandler(pb.MessageID_G2R_SEASON_PVP_BEGIN, rpc_G2R_SeasonPvpBegin)
	logic.RegisterRpcHandler(pb.MessageID_G2R_SEASON_PVP_END, rpc_G2R_SeasonPvpEnd)
	logic.RegisterRpcHandler(pb.MessageID_G2R_FETCH_PLAYER_RANK, rpc_G2R_fetchPlayerRank)
	logic.RegisterRpcHandler(pb.MessageID_G2R_LEAGUE_SEASON_END, rpc_G2R_LeagueSeasonEnd)
}
