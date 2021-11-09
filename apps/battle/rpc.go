package main

import (
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"time"
)

var restoringBattle = map[common.UUid]chan struct{}{}

func onPlayerLogout(args ...interface{}) {
	pa := args[0].(*logic.PlayerAgent)
	mgr.onFighterLogout(pa.GetUid())
}

func onRestoreAgent(args ...interface{}) {
	clients := args[0].([]*gpb.PlayerClient)
	for _, cli := range clients {
		if cli.GateID <= 0 || cli.ClientID <= 0 || cli.Uid <= 0 {
			continue
		}

		uid := common.UUid(cli.Uid)
		playerBattleInfoAttr := attribute.NewAttrMgr("playerBattleInfo", uid)
		err := playerBattleInfoAttr.Load()
		if err != nil {
			continue
		}

		if playerBattleInfoAttr.GetUInt32("appID") != bService.AppID {
			continue
		}

		agent := logic.NewPlayerAgent(cli)
		battleID := common.UUid(playerBattleInfoAttr.GetUInt64("battleID"))
		if c, ok := restoringBattle[battleID]; ok {
			evq.Await(func() {
				<-c
			})
		} else {
			c := make(chan struct{})
			restoringBattle[battleID] = c
			mgr.loadBattle(battleID, agent)
			delete(restoringBattle, battleID)
			close(c)
		}

		battle := mgr.getBattle(battleID)
		if battle != nil {
			agent.SetUid(uid)
			f := battle.getSituation().getFighter(uid)
			f.setBoutTimeout(boutTimeOut)
			battle.boutReadyDone(f)
			glog.Infof("onRestoreAgent ok uid=%d", uid)
		}
	}
}

func rpc_C2S_FightBoutReadyDone(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()

	fighter := mgr.getFighter(uid)
	if fighter == nil {
		glog.Errorf("rpc_C2S_FightBoutReadyDone not in fight uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	battleObj := fighter.getBattle()
	winUid, _ := battleObj.checkResult()

	//glog.Infof("rpc_C2S_FightBoutReadyDone, uid=%d, winUid=%d", uid, winUid)

	if winUid != 0 || battleObj.isEnd() {
		battleObj.battleEnd(winUid, false, false)
		mgr.delBattle(battleObj.getBattleID())
	} else {
		battleObj.boutReadyDone(fighter)
	}

	return nil, nil
}

func rpc_C2S_SendEmoji(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	f := mgr.getFighter(uid)
	if f == nil {
		return nil, nil
	}

	battleObj := f.getBattle()
	if battleObj == nil {
		return nil, nil
	}

	f1 := battleObj.getSituation().getFighter1()
	f2 := battleObj.getSituation().getFighter2()
	for _, f := range []*fighter{f1, f2} {
		if f.agent != nil && !f.isRobot && f.agent.GetUid() != uid {
			f.agent.PushClient(pb.MessageID_S2C_SYNC_EMOJI, arg)
			return nil, nil
		}
	}

	return nil, nil
}

func rpc_C2S_FightBoutCmd(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.FightBoutCmd)
	uid := agent.GetUid()

	f := mgr.getFighter(uid)
	if f == nil {
		glog.Errorf("rpc_C2S_FightBoutCmd not in fight uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	battleObj := f.getBattle()
	if battleObj == nil {
		return nil, gamedata.InternalErr
	}

	reply, errcode := battleObj.doAction(f, int(_arg.UseCardObjID), int(_arg.TargetGridId))
	if reply != nil {
		f.setBoutTimeout(boutTimeOut)
		return reply, nil
	} else {
		return nil, gamedata.GameError(errcode)
	}
}

func rpc_C2S_FightSurrender(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	fighter := mgr.getFighter(uid)
	if fighter == nil {
		glog.Errorf("rpc_C2S_FightSurrender not in fight uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	battleObj := fighter.getBattle()
	if battleObj == nil {
		return nil, gamedata.GameError(1)
	}

	battleType := battleObj.getBattleType()
	if mgr.isPvp(battleType) && battleType != consts.BtFriend {
		doActionFighter := battleObj.getSituation().getDoActionFighter()
		if ((doActionFighter != nil && doActionFighter != fighter) ||
			(battleObj.getSituation().getCurBoutFighter() != fighter)) &&
			time.Now().Sub(battleObj.getBoutBeginTime()) < time.Minute {
			return nil, gamedata.GameError(100)
		}
	}

	return nil, mgr.surrender(fighter)
}

func rpc_C2S_FightGMwin(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	fighter := mgr.getFighter(uid)
	if fighter == nil {
		glog.Errorf("rpc_C2S_FightGMwin not in fight uid=%d", uid)
		return nil, nil
	}

	if !config.GetConfig().Debug {
		return nil, gamedata.InternalErr
	}

	battleObj := fighter.getBattle()
	f1Uid := battleObj.getSituation().getFighter1().getUid()
	f2Uid := battleObj.getSituation().getFighter2().getUid()
	if f1Uid != uid && f2Uid != uid {
		return nil, nil
	}

	battleObj.battleEnd(uid, false, false)
	mgr.delBattle(battleObj.getBattleID())
	return nil, nil
}

func rpc_C2S_LevelReadyDone(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	fighter := mgr.getFighter(uid)
	if fighter == nil {
		glog.Errorf("rpc_C2S_LevelReadyDone not in fight uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	battleObj := fighter.getBattle()
	levelBattleObj, ok := battleObj.(*levelBattle)
	if !ok {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.LevelChooseCard)
	levelBattleObj.readyDone(arg2.Cards...)

	return nil, nil
}

func rpc_M2B_BeginBattle(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.BeginBattleArg)
	battleObj, err := mgr.beginBattle(int(arg2.BattleType), arg2.Fighter1, arg2.Fighter2, int(arg2.UpperType),
		int(arg2.BonusType), consts.BtScale33, 1, arg2.NeedVideo, arg2.NeedFortifications, arg2.IsFirstPvp,
		int(arg2.SeasonPvpSession), int(arg2.IndexDiff))
	if err != nil {
		return nil, err
	} else {
		return battleObj.packMsg(), err
	}
}

func rpc_L2B_BeginLevelBattle(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.BeginLevelBattleArg)
	levelData := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData).GetLevelData(int(arg2.LevelID))
	if levelData == nil {
		return nil, gamedata.InternalErr
	}
	battleObj, _ := mgr.beginLevelBattle(arg2.Fighter1, levelData, false)
	return battleObj.packMsg(), nil
}

func rpc_L2B_BeginLevelHelpBattle(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.BeginLevelBattleArg)
	levelData := gamedata.GetGameData(consts.Level).(*gamedata.LevelGameData).GetLevelData(int(arg2.LevelID))
	if levelData == nil {
		return nil, gamedata.InternalErr
	}
	battleObj, _ := mgr.beginLevelBattle(arg2.Fighter1, levelData, true)
	if battleObj == nil {
		return nil, gamedata.InternalErr
	} else {
		battleObj.syncReadyFight()
		return nil, nil
	}
}

func rpc_G2B_LoadBattle(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.LoadBattleArg)
	agent := logic.NewPlayerAgent(&gpb.PlayerClient{
		Uid:      arg2.Uid,
		ClientID: arg2.ClientID,
		GateID:   arg2.GateID,
		Region:   arg2.Region,
	})
	reply := mgr.loadBattle(common.UUid(arg2.BattleID), agent)

	if reply != nil {
		return reply, nil
	} else {
		return nil, network.InternalErr
	}
}

func rpc_G2B_CancelBattle(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.CancelBattleArg)
	mgr.cancelBattle(common.UUid(arg2.Uid), common.UUid(arg2.BattleID))
	return nil, nil
}

func rpc_G2B_IsBattleAlive(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.CancelBattleArg)
	battleObj := mgr.getBattle(common.UUid(arg2.BattleID))
	if battleObj == nil || battleObj.isEnd() {
		return nil, gamedata.GameError(1)
	}

	situation := battleObj.getSituation()
	uid := common.UUid(arg2.Uid)
	if situation.getFighter1().getUid() == uid || situation.getFighter2().getUid() == uid {
		return nil, nil
	}

	return nil, gamedata.GameError(2)
}

func registerRpc() {
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onPlayerLogout)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, onPlayerLogout)
	eventhub.Subscribe(logic.RESTORE_AGENT_EV, onRestoreAgent)

	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FIGHT_BOUT_READY_DONE, rpc_C2S_FightBoutReadyDone)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SEND_EMOJI, rpc_C2S_SendEmoji)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FIGHT_BOUT_CMD, rpc_C2S_FightBoutCmd)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FIGHT_SURRENDER, rpc_C2S_FightSurrender)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FIGHT_GM_WIN, rpc_C2S_FightGMwin)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LEVEL_READY_DONE, rpc_C2S_LevelReadyDone)

	logic.RegisterRpcHandler(pb.MessageID_M2B_BEGIN_BATTLE, rpc_M2B_BeginBattle)
	logic.RegisterRpcHandler(pb.MessageID_L2B_BEGIN_LEVEL_BATTLE, rpc_L2B_BeginLevelBattle)
	logic.RegisterRpcHandler(pb.MessageID_L2B_BEGIN_LEVEL_HELP_BATTLE, rpc_L2B_BeginLevelHelpBattle)
	logic.RegisterRpcHandler(pb.MessageID_G2B_LOAD_BATTLE, rpc_G2B_LoadBattle)
	logic.RegisterRpcHandler(pb.MessageID_G2B_CANCEL_BATTLE, rpc_G2B_CancelBattle)
	logic.RegisterRpcHandler(pb.MessageID_G2B_IS_BATTLE_ALIVE, rpc_G2B_IsBattleAlive)
}
