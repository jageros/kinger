package level

import (
	"kinger/gopuppy/apps/logic"
	"kinger/apps/game/module"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"kinger/gopuppy/common"
	"kinger/gopuppy/attribute"
)

func rpc_C2S_FetchLevelInfo(agent *logic.PlayerAgent, _ interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return player.GetComponent(consts.LevelCpt).(*levelComponent).packMsg(), nil
}

func rpc_C2S_BeginLevelBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.BeginLevelBattle)
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	levelComponent := player.GetComponent(consts.LevelCpt).(*levelComponent)
	reply, err := levelComponent.beginBattle(int(_arg.LevelId))
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func rpc_C2S_OpenLevelTreasure(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.OpenLevelTreasureArg)
	levelComponent := player.GetComponent(consts.LevelCpt).(*levelComponent)
	return levelComponent.openTreasure(int(arg2.ChapterID))
}

func rpc_C2S_LevelHelpOther(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.LevelHelpArg)
	levelComponent := player.GetComponent(consts.LevelCpt).(*levelComponent)
	return nil, levelComponent.beginHelpBattle(int(arg2.LevelID), common.UUid(arg2.HelpUid))
}

func rpc_C2S_WatchHelpVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.WatchHelpVideoArg)
	attr := attribute.NewAttrMgr("battleVideo", arg2.VideoID, true)
	err := attr.Load()
	if err != nil {
		return nil, err
	}

	data := []byte(attr.GetStr("data"))
	videData := &pb.VideoBattleData{}
	videData.Unmarshal(data)
	return videData, nil
}

func rpc_C2S_FetchLevelHelpRecord(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.TargetLevel)
	return player.GetComponent(consts.LevelCpt).(*levelComponent).getLevelBeHelpRecord(int(arg2.LevelID)), nil
}

func rpc_C2S_FetchLevelVideoID(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.FetchLevelVideoIDArg)
	videoID := player.GetComponent(consts.LevelCpt).(*levelComponent).getVideo(int(arg2.LevelID))
	return &pb.FetchLevelVideoIDRely{
		VideoID: uint64(videoID),
	}, nil
}

func rpc_C2S_ClearLevelChapter(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	levelCpt := player.GetComponent(consts.LevelCpt).(*levelComponent)
	err := levelCpt.clearChapter()
	if err != nil {
		return nil, err
	}

	return levelCpt.packMsg(), nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_LEVEL_INFO, rpc_C2S_FetchLevelInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BEGIN_LEVEL_BATTLE, rpc_C2S_BeginLevelBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_OPEN_LEVEL_TREASURE, rpc_C2S_OpenLevelTreasure)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LEVEL_HELP_OTHER, rpc_C2S_LevelHelpOther)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_WATCH_HELP_VIDEO, rpc_C2S_WatchHelpVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_LEVEL_HELP_RECORD, rpc_C2S_FetchLevelHelpRecord)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_LEVEL_VIDEO_ID, rpc_C2S_FetchLevelVideoID)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_CLEAR_LEVEL_CHAPTER, rpc_C2S_ClearLevelChapter)
}
