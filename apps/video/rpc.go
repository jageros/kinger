package main

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/network"
	"kinger/gamedata"
	"kinger/proto/pb"
)

func onLogout(args ...interface{}) {
	agent := args[0].(*logic.PlayerAgent)
	videoMgr.delPlayer(agent.GetUid())
}

func rpc_B2V_SaveVideo(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.SaveVideoArg)
	videoMgr.saveVideoItem(common.UUid(arg2.VideoID), arg2.Fighter1, arg2.Fighter2, common.UUid(arg2.Winner))
	return nil, nil
}

func rpc_G2V_FetchVideoList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	vis := videoMgr.randomVideos(int(arg.(*pb.TargetArea).Area))
	reply := &pb.VideoList{}
	for _, vi := range vis {
		reply.Videos = append(reply.Videos, vi.packMsg(uid))
	}
	return reply, nil
}

func rpc_C2S_WatchVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.WatchVideoArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	if vi == nil {
		videoData := videoMgr.loadVideo(common.UUid(arg2.VideoID))
		if videoData == nil {
			return nil, gamedata.InternalErr
		} else {
			return &pb.WatchVideoResp{
				VideoData:     videoData,
			}, nil
		}
	}

	videoData, watchTimes, like := vi.watch()
	if videoData == nil {
		return nil, gamedata.InternalErr
	}
	videoData.ShareUid = uint64(vi.getSharePlayer())
	return &pb.WatchVideoResp{
		VideoData:     videoData,
		CurWatchTimes: int32(watchTimes),
		CurLike:       int32(like),
	}, nil
}

func rpc_C2S_FetchSelfVideoList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.FetchSelfVideoListArg)
	vis := videoMgr.getPlayer(uid).getHistoryVideos(int(arg2.Page))
	reply := &pb.VideoList{}
	for _, vi := range vis {
		reply.Videos = append(reply.Videos, vi.packMsg(uid))
	}
	return reply, nil
}

func rpc_C2S_LikeVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.LikeVideoArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	if vi == nil {
		return &pb.LikeVideoResp{
			CurWatchTimes: 0,
			CurLike:       0,
		}, nil
	}
	watchTimes, like := vi.like(uid)
	return &pb.LikeVideoResp{
		CurWatchTimes: int32(watchTimes),
		CurLike:       int32(like),
	}, nil
}

func rpc_C2S_FetchVideoComments(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchVideoCommentsArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	reply := &pb.FetchVideoCommentsReply{}
	if vi == nil {
		return reply, nil
	}

	cs, hasMore := vi.getComments(agent.GetUid(), int(arg2.CurAmount))
	reply.CommentsList = cs
	reply.HasMore = hasMore
	return reply, nil
}

func rpc_C2S_LikeVideoComments(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.LikeVideoCommentsArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	reply := &pb.LikeVideoCommentsReply{}
	if vi == nil {
		return reply, nil
	}

	reply.CurLike = int32(vi.likeComments(agent.GetUid(), int(arg2.CommentsID)))
	return reply, nil
}

func rpc_G2V_ShareVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	arg2 := arg.(*pb.GShareVideoArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	if vi == nil {
		return nil, gamedata.InternalErr
	}

	vi.share(uid, arg2.Name, int(arg2.Area))
	return nil, nil
}

func rpc_G2V_CommentsVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.GCommentsVideoArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false)
	if vi == nil {
		return nil, gamedata.InternalErr
	}

	c := vi.comments(common.UUid(arg2.Uid), arg2.Name, arg2.Content, arg2.HeadImgUrl, arg2.Country, arg2.HeadFrame,
		arg2.CountryFlag)
	return &pb.CommentsVideoReply{
		CommentsID: int32(c.getID()),
		Time: int32(c.getTime()),
	}, nil
}

func rpc_C2S_FetchVideoItem(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.FetchVideoItemArg)
	vi := videoMgr.loadVideoItem(common.UUid(arg2.VideoID), false, true)
	if vi == nil {
		return nil, gamedata.InternalErr
	}
	return vi.packMsg(agent.GetUid()), nil
}

func registerRpc() {
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onLogout)
	logic.RegisterRpcHandler(pb.MessageID_B2V_SAVE_VIDEO, rpc_B2V_SaveVideo)

	logic.RegisterAgentRpcHandler(pb.MessageID_G2V_FETCH_VIDEO_LIST, rpc_G2V_FetchVideoList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_VIDEO, rpc_C2S_WatchVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SELF_VIDEO_LIST, rpc_C2S_FetchSelfVideoList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LIKE_VIDEO, rpc_C2S_LikeVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_VIDEO_COMMENTS, rpc_C2S_FetchVideoComments)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LIKE_VIDEO_COMMENTS, rpc_C2S_LikeVideoComments)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2V_SHARE_VIDEO, rpc_G2V_ShareVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_G2V_COMMENTS_VIDEO, rpc_G2V_CommentsVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_VIDEO_ITEM, rpc_C2S_FetchVideoItem)
}
