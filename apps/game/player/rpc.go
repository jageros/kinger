package player

import (
	"crypto/md5"
	"fmt"
	"io"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gamedata"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	gconfig "kinger/gopuppy/common/config"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/wordfilter"
	"kinger/gopuppy/network"
	gpb "kinger/gopuppy/proto/pb"
	"kinger/proto/pb"
	"kinger/sdk"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func genNewbieName(uid common.UUid) string {
	//idx, _ := common.Gen32UUid("playerNameIdx")
	return fmt.Sprintf("%s%d", mod.getNewbieNamePrefix(), uid)
}

func doLogout(uid common.UUid, isKickOut bool) (accountAttr *attribute.AttrMgr) {
	_player := mod.GetPlayer(uid)
	mod.onPlayerLogout(uid)
	if _player == nil {
		return nil
	}

	mod.delPlayer(uid)
	if !isKickOut {
		mod.addCachePlayer(_player.(*Player))
	}

	_player.(*Player).onLogout()
	playerAttr := _player.(*Player).getAttr()
	playerAttr.Save(true)

	lastLoginTime := _player.GetLastLoginTime()
	onlineTime := int64(_player.GetLastOnlineTime()) - lastLoginTime
	if onlineTime <= 0 {
		onlineTime = time.Now().Unix() - lastLoginTime - 90
		if onlineTime < 0 {
			onlineTime = 0
		}
	}

	glog.JsonInfo("logout", glog.Uint64("uid", uint64(_player.GetUid())), glog.String("accountType",
		_player.GetAccountType().String()), glog.String("channel", _player.GetChannel()), glog.String("channelID",
		_player.GetChannelUid()), glog.String("loginChannel", _player.GetLoginChannel()), glog.Int("area",
		_player.GetArea()), glog.Int("onlineTime", int(onlineTime)), glog.String("subChannel", _player.GetSubChannel()))

	return nil
}

func onLogout(args ...interface{}) {
	agent := args[0].(*logic.PlayerAgent)
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid)
	if _player == nil {
		return
	}

	//if _player.GetAgent() != agent || _player.GetAgent().GetClientID() != agent.GetClientID() {
	//	return
	//}

	doLogout(uid, false)
}

func onPlayerKickOut(args ...interface{}) {
	//glog.Infof("onPlayerKickOut 11111111111")
	agent := args[0].(*logic.PlayerAgent)
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid)

	if _player != nil {
		p := _player.(*Player)
		p.kickOut(Relogin)
		p.delAgent()
		doLogout(uid, true)
	}

	mod.delCachePlayer(uid)
	//if _player.GetAgent() != agent || _player.GetAgent().GetClientID() != agent.GetClientID() {
	//	mod.delPlayer(uid)
	//	return
	//}
}

func onRestoreAgent(args ...interface{}) {
	clients := args[0].([]*gpb.PlayerClient)
	for _, cli := range clients {
		if cli.GateID <= 0 || cli.ClientID <= 0 || cli.Uid <= 0 {
			continue
		}

		uid := common.UUid(cli.Uid)
		if mod.GetPlayer(uid) != nil {
			continue
		}

		agent := logic.NewPlayerAgent(cli)
		playerAttr := attribute.NewAttrMgr("player", uid)
		err := playerAttr.Load()
		if err != nil {
			glog.Errorf("onRestoreAgent uid=%d, err=%s", uid, err)
			continue
		}

		if mod.GetPlayer(uid) != nil {
			continue
		}

		mod.playerCache.RemoveWithoutCallback(uid)
		agent.SetUid(uid)
		player := newPlayer(uid, agent, playerAttr)
		mod.addPlayer(player)
		if ok := mod.onPlayerLogin(player, false, true); !ok {
			mod.delPlayer(uid)
			return
		}
		glog.Infof("onRestoreAgent ok uid=%d", uid)
	}
}

func onReloadConfig(ev evq.IEvent) {
	if module.Service.GetAppID() == 1 {
		api.BroadcastClient(pb.MessageID_S2C_UPDATE_WX_EXAMINE_STATE, &pb.WxExamineState{
			IsExamined: config.GetConfig().Wxgame.IsExamined,
		}, nil)
	}

	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		agent := player.GetAgent()
		if agent == nil {
			return
		}

		agent.PushClient(pb.MessageID_S2C_UPDATE_VERSION, config.GetConfig().GetVersion(player.GetAccountType()))
	})
}

func accountLoginCheckRegion(channel, loginChannel, channelID string, isTourist bool) *pb.AccountArchives {
	bindRegion := mod.getBindRegion(genAccountID(channel, loginChannel, channelID, isTourist))
	if bindRegion > 0 && bindRegion != module.Service.GetRegion() {
		gate := gconfig.RandomGateByRegion(bindRegion)
		if gate != nil {
			redirectHost := &pb.GateHost{
				Host: gate.Host,
			}

			for _, lInfo := range gate.Listens {
				if lInfo.Network != "wss" {
					redirectHost.Port = int32(lInfo.Port)
				} else {
					redirectHost.WssPort = int32(lInfo.Port)
				}
			}

			return &pb.AccountArchives{
				Ok:           false,
				RedirectHost: redirectHost,
			}
		}
	}
	return nil
}

func loginCheckServerStatus(channel string, uid common.UUid) *pb.AccountArchives {
	serverStatus := module.GM.GetServerStatus()
	if serverStatus == nil || serverStatus.Status != pb.ServerStatus_Maintain {
		return nil
	}

	cfg := config.GetConfig()
	for _, gmChannel := range cfg.GmChannels {
		if gmChannel == channel {
			return nil
		}
	}

	if uid > 0 {
		for _, gmUid := range cfg.GmUids {
			if gmUid == uid {
				return nil
			}
		}
	}

	return &pb.AccountArchives{ServerSt: serverStatus}
}

func rpc_C2S_AccountLogin(_ *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.AccountLoginArg)
	if reply := accountLoginCheckRegion(_arg.Channel, _arg.LoginChannel, _arg.ChannelID, _arg.IsTourist); reply != nil {
		return reply, nil
	}

	a, err := loadAccount(_arg.Channel, _arg.LoginChannel, _arg.ChannelID, _arg.IsTourist)
	var uid common.UUid
	if a != nil {
		uid = a.getUid()
	}

	if reply := loginCheckServerStatus(_arg.Channel, uid); reply != nil {
		return reply, nil
	}

	if err != nil {
		if err == attribute.NotExistsErr {
			if _arg.IsTourist {
				// new account
				a.setPwd(md5HashPassword(_arg.Password))
				if err = a.save(true); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}

		} else {
			return nil, err
		}
	}

	accountPwd := a.getPwd()
	if accountPwd != "" {
		if md5HashPassword(_arg.Password) != accountPwd {
			return nil, gamedata.InternalErr
		}
	}

	return a.packMsg(), nil
}

func rpc_C2S_Login(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {

	_arg := arg.(*pb.LoginArg)
	a, err := loadAccount(_arg.Channel, _arg.LoginChannel, _arg.ChannelID, _arg.IsTourist)
	if err != nil {
		return nil, err
	}

	if loginCheckServerStatus(_arg.Channel, a.getUid()) != nil {
		return nil, gamedata.GameError(1000)
	}

	if reply := accountLoginCheckRegion(_arg.Channel, _arg.LoginChannel, _arg.ChannelID, _arg.IsTourist); reply != nil {
		return nil, gamedata.GameError(1001)
	}

	arc := a.getArchive(int(_arg.ArchiveID))
	var player *Player
	isRelogin := false
	var isNew bool

	if arc == nil {
		// new archive, new player
		isNew = true
		uid, err := a.genArchive(int(_arg.ArchiveID))
		if err != nil {
			return nil, err
		}
		player = mod.createNewbiePlayer(uid, agent, _arg.Channel, _arg.ChannelID, a.getArea())
	} else {

		uid := common.UUid(arc.GetUInt64("uid"))
		var ok bool
		player, ok = mod.GetPlayer(uid).(*Player)

		api.NotifyPlayerBeginLogin(uid, agent.GetClientID(), agent.GetGateID(), agent.GetRegion())
		mod.playerCache.RemoveWithoutCallback(uid)

		if ok {
			//player.relogin(ses)
			player.agent = agent
			isRelogin = true
			//doLogout(player.GetUid())
		} else {
			playerAttr := attribute.NewAttrMgr("player", uid)
			err := playerAttr.Load()
			if err != nil {
				if err == attribute.NotExistsErr {
					player = mod.createNewbiePlayer(uid, agent, _arg.Channel, _arg.ChannelID, a.getArea())
				} else {
					return nil, err
				}
			} else {
				player = newPlayer(common.UUid(uid), agent, playerAttr)
			}
		}
	}

	glog.JsonInfo("login", glog.Uint64("uid", uint64(player.GetUid())), glog.String("accountType",
		_arg.AccountType.String()), glog.String("channel", _arg.Channel), glog.String("channelID", _arg.ChannelID),
		glog.String("loginChannel", _arg.LoginChannel), glog.Bool("isNew", isNew), glog.Int("area",
			player.GetArea()), glog.String("ip", agent.GetIP()), glog.String("subChannel", _arg.SubChannel))

	mod.playerCache.RemoveWithoutCallback(player.GetUid())
	mod.addPlayer(player)
	agent.SetUid(player.GetUid())

	if ok := mod.onPlayerLogin(player, isRelogin, false); !ok {
		mod.delPlayer(player.GetUid())
		return nil, gamedata.InternalErr
	}

	if player.IsForbidLogin() {
		glog.Infof("player.IsForbidLogin uid=%d", player.GetUid())
		evq.Await(func() {
			time.Sleep(2 * time.Second)
		})
		if player.IsForbidLogin() {
			glog.Infof("player.IsForbidLogin uid=%d", player.GetUid())
			return nil, gamedata.GameError(1)
		}
	}

	ipAddr := agent.GetIP()
	player.SetIP(ipAddr)
	if utils.IsForbidIPAddr(ipAddr) {
		glog.Infof("player'IP is forbid uid=%d, ip=%s", player.GetUid(), ipAddr)
		return nil, gamedata.GameError(1)
	}

	reply := &pb.LoginReply{
		Uid:     uint64(player.GetUid()),
		Name:    player.GetName(),
		HeadImg: player.GetHeadImgUrl(),
		// for test
		ServerID:           player.GetServerID(),
		SeasonPvpLimitTime: module.Huodong.GetSeasonPvpLimitTime(player),
		HeadFrame:          player.GetHeadFrame(),
		CardSkins:          module.Bag.GetAllItemIDsByType(player, consts.ItCardSkin),
		Area:               int32(player.GetArea()),
		ChatPop:            player.GetChatPop(),
		Notice:             player.getCanShowLoginNotice(),
		CountryFlag:        player.GetCountryFlag(),
	}

	cardComponent := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	for _, card := range cardComponent.GetAllCollectCards() {
		reply.Cards = append(reply.Cards, card.PackMsg())
	}

	onceCards := module.Card.GetOnceCards(player)
	for _, card := range onceCards {
		if card.GetMaxUnlockLevel() > 0 {
			reply.Cards = append(reply.Cards, card.PackMsg())
		}
	}

	module.OutStatus.ForEachClientStatus(player, func(st types.IOutStatus) {
		reply.OutStatuses = append(reply.OutStatuses, st.PackMsg())
	})

	//reply.DiyCards = cardComponent.PackDiyCardMsg()
	player.setAccountType(_arg.AccountType)
	player.setCountry(_arg.Country)
	resComponent := player.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	reply.Res = resComponent.packMsg()
	reply.FightID = uint64(player.GetBattleID())
	reply.GuideCamp = player.GetComponent(consts.TutorialCpt).(types.ITutorialComponent).GetCampID()
	reply.Ver = config.GetConfig().GetVersion(_arg.AccountType)
	reply.IsExamined = config.GetConfig().Wxgame.IsExamined
	reply.CumulativePay = int32(player.GetComponent(consts.ShopCpt).(types.IShopComponent).GetCumulativePay())
	reply.SharedState = int32(module.WxGame.GetDailyShareState(player))
	_, _, _, _, seasonPvpHandCard := module.Huodong.GetSeasonPvpHandCardInfo(player)
	reply.IsSeasonPvpChooseCard = seasonPvpHandCard != nil
	reply.IsInCampaignMatch = module.Campaign.IsInCampaignMatch(player)
	reply.Huodongs = module.Huodong.PackEventHuodongs(player)
	reply.RebornCnt = int32(module.Reborn.GetRebornCnt(player))

	player.forEachHint(func(type_ pb.HintType, count int) {
		reply.Hints = append(reply.Hints, &pb.Hint{
			Type:  type_,
			Count: int32(count),
		})
	})

	api.NotifyPlayerLoginDone(agent.GetUid(), agent.GetClientID(), agent.GetGateID(), agent.GetRegion())
	player.onNetConnect()
	now := time.Now().Unix()
	if _arg.Channel == "lzd_pkgsdk" && a.isCpAccount() {
		player.setFire233BindAccount(_arg.ChannelID)
	}
	player.setLastOnlineTime(int(now))
	player.setLoginTime(now)
	player.setAccountID(a.getAccountID())
	player.setChannel(_arg.Channel)
	player.setSubChannel(_arg.SubChannel)
	player.setLoginChannel(_arg.LoginChannel)
	player.setChannelUid(_arg.ChannelID)

	evq.PostEvent(evq.NewCommonEvent(consts.EvLogin, player.GetUid(), isRelogin))
	if isNew {
		doReward190430(a, player)
	}
	doRewardFengce(a, player)

	return reply, nil
}

func rpc_C2S_FinishGuide(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.FinishGuide)
	_player.finishGuide(int(_arg.GuideID))
	return nil, nil
}

func rpc_C2S_FetchGuide(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.AllFinishGuide{}
	guideAttr := _player.getGuideAttr()
	guideAttr.ForEachIndex(func(index int) bool {
		reply.GuideIDs = append(reply.GuideIDs, int32(guideAttr.GetInt(index)))
		return true
	})
	return reply, nil
}

func rpc_C2S_GmCommand(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid)
	if _player == nil {
		return nil, gamedata.InternalErr
	}

	if !config.GetConfig().Debug {
		return nil, gamedata.InternalErr
	}

	_arg := arg.(*pb.GmCommand)
	commandInfo := strings.Split(_arg.Command, " ")
	if len(commandInfo) <= 0 {
		return nil, gamedata.InternalErr
	}

	if commandInfo[0] == "act" {
		if len(commandInfo) < 2 {
			return nil, gamedata.GameError(100000)
		}
		_, err := module.Activitys.TestARpc(agent, commandInfo)
		return nil, err
	}

	if commandInfo[0] == "addcard" {
		if len(commandInfo) != 3 {
			return nil, gamedata.InternalErr
		}

		cardId, _ := strconv.Atoi(commandInfo[1])
		amount, _ := strconv.Atoi(commandInfo[2])

		cardComponent := _player.GetComponent(consts.CardCpt).(types.ICardComponent)
		cardComponent.ModifyCollectCards(map[uint32]*pb.CardInfo{
			uint32(cardId): &pb.CardInfo{
				Amount: int32(amount),
			},
		})
		return nil, nil

	} else if commandInfo[0] == "addres" {
		if len(commandInfo) != 3 {
			return nil, gamedata.InternalErr
		}

		resType, _ := strconv.Atoi(commandInfo[1])
		amount, _ := strconv.Atoi(commandInfo[2])

		resComponent := _player.GetComponent(consts.ResourceCpt).(*ResourceComponent)
		resComponent.ModifyResource(resType, amount)
		return nil, nil
	} else if commandInfo[0] == "addtreasure" {
		if len(commandInfo) < 2 {
			return nil, gamedata.InternalErr
		}

		treasureType := commandInfo[1]
		var treasureID string

		if len(commandInfo) >= 3 {
			treasureID = commandInfo[2]
		} else {
			treasureID = ""
		}
		treasureComponent := _player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent)

		if treasureType == "daily" {
			if treasureID == "" {
				treasureComponent.AddDailyTreasure(false)
			} else {
				treasureComponent.AddDailyTreasureByID(treasureID, 0)
			}
		} else if treasureType == "reward" {
			if treasureID == "" {
				treasureComponent.AddRewardTreasure(true, false)
			} else {
				treasureComponent.AddRewardTreasureByID(treasureID, false)
			}
		} else {
			return nil, gamedata.InternalErr
		}

		return nil, nil
	} else if commandInfo[0] == "refreshrank" {
		// TODO
		//module.Pvp.RefreshRank()
		return nil, nil
	} else if commandInfo[0] == "clearlevel" {
		_player.GetComponent(consts.LevelCpt).(types.ILevelComponent).ClearLevel()
		return nil, nil
	} else if commandInfo[0] == "opentreasure" {
		if len(commandInfo) < 2 {
			return nil, gamedata.InternalErr
		}
		_player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).OpenTreasureByModelID(commandInfo[1], false)
		return nil, nil
	} else if commandInfo[0] == "addms" {
		return nil, module.Mission.GmAddMission(_player, commandInfo)
	} else if commandInfo[0] == "fuckms" {
		module.Mission.GmCompleteMission(_player)
		return nil, nil
	} else if commandInfo[0] == "addhf" {
		module.Bag.AddHeadFrame(_player, commandInfo[1])
		return nil, nil
	} else if commandInfo[0] == "addeq" {
		module.Bag.AddEquip(_player, commandInfo[1])
		return nil, nil
	} else if commandInfo[0] == "addcp" {
		module.Bag.AddChatPop(_player, commandInfo[1])
		return nil, nil
	} else if commandInfo[0] == "ca" {
		return agent.CallBackend(pb.MessageID_G2CA_GM_COMMAND, arg)
	} else if commandInfo[0] == "resetdt" {
		_player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).AddDailyTreasure(true)
		return nil, nil
	} else if commandInfo[0] == "fuckdt" {
		_player.GetComponent(consts.TreasureCpt).(types.ITreasureComponent).AddDailyTreasureStar(10)
		return nil, nil
	} else if commandInfo[0] == "resetst" {
		_player.GetComponent(consts.ShopCpt).(types.ICrossDayComponent).OnCrossDay(-1)
		_player.GetComponent(consts.ActivityCpt).(types.ICrossDayComponent).OnCrossDay(-1)
		return nil, nil
	} else if commandInfo[0] == "allcard" && commandInfo[1] == "uplevel" {
		module.Card.GmAllCardUpLevel(_player)
		return nil, nil
	} else if commandInfo[0] == "addsk" {
		module.Bag.AddCardSkin(_player, commandInfo[1])
		return nil, nil
	} else if commandInfo[0] == "setrt" {
		module.Shop.GM_setRecruitVer(commandInfo[1], _player)
		return nil, nil
	} else if commandInfo[0] == "league" {
		if len(commandInfo) >= 2 {
			module.Pvp.GM_CrossSeason(agent, commandInfo[1])
		}
		return nil, nil
	} else {
		return nil, gamedata.InternalErr
	}
}

func rpc_C2S_LoadFight(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		glog.Errorf("rpc_C2S_LoadFight no player uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	fightID := _player.GetBattleID()
	if fightID <= 0 {
		glog.Errorf("rpc_C2S_LoadFight no battleID uid=%d", uid)
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.C2SLoadFightArg)
	battleType := _player.GetBattleType()
	if battleType == consts.BtLevelHelp || (arg2.IsIgnorePve && (battleType == consts.BtLevel || battleType == consts.BtLevelHelp)) {

		_player.clearBattleID()
		return nil, gamedata.InternalErr
	}

	glog.Infof("begin LoadFight uid=%d, battleID=%d, battleAppID=%d", uid, fightID, _player.GetBattleAppID())

	reply, err := logic.CallBackend(consts.AppBattle, _player.GetBattleAppID(), pb.MessageID_G2B_LOAD_BATTLE, &pb.LoadBattleArg{
		ClientID: uint64(agent.GetClientID()),
		Uid:      uint64(uid),
		GateID:   agent.GetGateID(),
		BattleID: uint64(fightID),
		Region:   agent.GetRegion(),
	})

	glog.Infof("rpc_C2S_LoadFight uid=%d, battleID=%d, battleAppID=%d", uid, fightID, _player.GetBattleAppID())
	if reply == nil || err != nil {
		_player.clearBattleID()
		glog.Infof("rpc_C2S_LoadFight err uid=%d, battleID=%d, battleAppID=%d", uid, fightID, _player.GetBattleAppID())
		return nil, gamedata.InternalErr
	}
	return reply, nil
}

func rpc_C2S_RegisterAccount(_ *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.RegisterAccount)
	a, err := doRegistAccount(arg2.Channel, arg2.Account, arg2.Password)
	if err != nil {
		return nil, err
	}
	return a.packMsg(), nil
}

func rpc_C2S_FetchSurveyInfo(agent *logic.PlayerAgent, _ interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		return nil, gamedata.InternalErr
	}

	return _player.GetComponent(consts.SurveyCpt).(*surveyComponent).packMsg(), nil
}

func rpc_C2S_CompleteSurvey(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		return nil, nil
	}

	_player.GetComponent(consts.SurveyCpt).(*surveyComponent).answer(arg.(*pb.SurveyAnswer))
	return nil, nil
}

func rpc_C2S_GetSurveyReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	_player := mod.GetPlayer(uid).(*Player)
	if _player == nil {
		return nil, nil
	}

	reply := _player.GetComponent(consts.SurveyCpt).(*surveyComponent).getReward()
	if reply == nil {
		return nil, gamedata.InternalErr
	}
	return reply, nil
}

func rpc_C2S_ModifyName(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid).(*Player)
	if player == nil {
		return nil, nil
	}

	arg2 := arg.(*pb.ModifyNameArg)
	if _, hasDirty, _, wTy := wordfilter.ContainsDirtyWords(arg2.Name, false); hasDirty && wTy == gconsts.GeneralWords {
		return nil, gamedata.GameError(101)
	}
	if mod.isNameExist(arg2.Name) {
		return nil, gamedata.InternalErr
	}
	player.setName(arg2.Name)
	mod.addName(arg2.Name, uid)
	//evq.PostEvent(evq.NewCommonEvent(consts.EvModifyName, _player.GetUid(), arg2.Name))

	return nil, nil
}

func rpc_C2S_UpdateName(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid).(*Player)
	if player == nil {
		return nil, nil
	}

	arg2 := arg.(*pb.UpdateNameArg)
	if _, hasDirty, _, wTy := wordfilter.ContainsDirtyWords(arg2.Name, false); hasDirty && wTy == gconsts.GeneralWords {
		return nil, gamedata.GameError(101)
	}

	if mod.isNameExist(arg2.Name) {
		return nil, gamedata.InternalErr
	}

	priceGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	if !player.HasBowlder(priceGameData.ModifyName) {
		return nil, gamedata.GameError(1)
	}

	nameTime := player.GetModifytime()
	if nameTime != "" {
		int64Name, _ := strconv.ParseInt(nameTime, 10, 64)
		isModifyDay := timer.GetDayNo(int64Name)
		now := time.Now().Unix()
		curDay := timer.GetDayNo(now)
		if surplus := curDay - isModifyDay; surplus <= 0 {
			//resCpt.ModifyResource(consts.Jade, 50,true)
			return nil, gamedata.GameError(4)
		}
	}

	module.Shop.LogShopBuyItem(player, "modifyName", "改名", 1, "gameplay",
		strconv.Itoa(consts.Jade), module.Player.GetResourceName(consts.Jade), priceGameData.ModifyName, "")

	player.SubBowlder(priceGameData.ModifyName, consts.RmrModifyName)
	player.setModifytime()
	mod.delName(player.GetName())
	player.setName(arg2.Name)
	mod.addName(arg2.Name, uid)
	return nil, nil
}

func sdkAccountAuth(arg *pb.AccountLoginArg) (*accountSt, error) {
	s := sdk.GetSdk(arg.Channel, arg.LoginChannel)
	if s != nil {
		if err := s.LoginAuth(arg.ChannelID, arg.SdkToken); err != nil {
			glog.Errorf("rpc_C2S_SdkAccountLogin channel=%s, loginChannel=%s, channelID=%s, account=%s, err=%s",
				arg.Channel, arg.LoginChannel, arg.ChannelID, arg.Account, err)
			return nil, gamedata.GameError(1)
		} else {
			glog.Infof("rpc_C2S_SdkAccountLogin channel=%s, loginChannel=%s, channelID=%s, account=%s",
				arg.Channel, arg.LoginChannel, arg.ChannelID, arg.Account)
		}
	}

	channelUid := arg.ChannelID
	if channelUid == "" {
		channelUid = arg.Account
	}
	a, err := loadAccount(arg.Channel, arg.LoginChannel, channelUid, arg.IsTourist)
	if err != nil {
		if err == attribute.NotExistsErr {

			// for old account
			//if exists, _ := db.Exists("account_old",
			//	genAccountID(arg.Channel, arg.LoginChannel, channelUid, arg.IsTourist)); exists {

			//	return nil, gamedata.GameError(100)
			//}

			// new account
			if err = a.save(true); err != nil {
				return nil, err
			}

			glog.JsonInfo("account", glog.String("channel", arg.Channel), glog.String("loginChannel",
				arg.LoginChannel), glog.String("channelID", channelUid))
		} else {
			return nil, err
		}
	}

	a.setWxOpenID(arg.WxOpenID)
	return a, nil
}

func rpc_C2S_SdkAccountLogin(_ *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	_arg := arg.(*pb.AccountLoginArg)
	if reply := accountLoginCheckRegion(_arg.Channel, _arg.LoginChannel, _arg.ChannelID, _arg.IsTourist); reply != nil {
		return reply, nil
	}

	a, err := sdkAccountAuth(_arg)
	if err != nil {
		return nil, err
	}

	if reply := loginCheckServerStatus(_arg.Channel, a.getUid()); reply != nil {
		return reply, nil
	}

	return a.packMsg(), nil
}

func rpc_C2S_FetchVersion(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	player := mod.GetPlayer(agent.GetUid())
	var accountType pb.AccountTypeEnum
	if player != nil {
		accountType = player.GetAccountType()
	}

	return config.GetConfig().GetVersion(accountType), nil
}

func rpc_LoadPlayer(uid common.UUid) ([]byte, error) {
	return mod.loadPlayer(uid, true)
}

func rpc_C2S_UpdateSdkUserInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.SdkUserInfo)
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, nil
	}
	player2 := player.(*Player)
	oldName := player2.GetName()
	if oldName == "" || strings.Index(oldName, mod.getNewbieNamePrefix()) == 0 || strings.Index(oldName, "微信用户") == 0 ||
		strings.Index(oldName, "无名军师") == 0 {
		player2.setName(arg2.NickName)
	}
	player2.setHeadImgUrl(arg2.HeadImgUrl)

	if arg2.InviterUid > 0 && player2.GetPvpScore() <= 0 {
		evq.CallLater(func() {
			module.Social.WxInviteFriend(common.UUid(arg2.InviterUid), player)
			//module.Social.AddFriendApply(player, common.UUid(arg2.InviterUid), true)
		})
	}

	return nil, nil
}

func rpc_C2S_RecordCurGuideGroup(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player.(*Player).setCurGuideGroup(int(arg.(*pb.GuideGroup).GroupID))
	return nil, nil
}

func rpc_C2S_FetchCurGuideGroup(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return &pb.GuideGroup{
		GroupID: int32(player.(*Player).getCurGuideGroup()),
	}, nil
}

func rpc_C2S_UpdateHeadImg(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player.(*Player).setHeadImgUrl(arg.(*pb.UpdateHeadImgArg).HeadImg)
	return nil, nil
}

func rpc_C2S_ShareVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.ShareVideoArg)
	if _, hasDirty, _, wTy := wordfilter.ContainsDirtyWords(arg2.Name, false); hasDirty && wTy == gconsts.GeneralWords {
		return nil, gamedata.GameError(101)
	}

	reply, err := agent.CallBackend(pb.MessageID_G2V_SHARE_VIDEO, &pb.GShareVideoArg{
		VideoID: arg2.VideoID,
		Name:    arg2.Name,
		Area:    int32(player.GetArea()),
	})
	if err == nil {
		module.Mission.OnShareVideo(player)
		eventhub.Publish(consts.EvShareBattleReport, player)
	}
	return reply, err
}

func rpc_C2S_CommentsVideo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.CommentsVideoArg)
	if _, hasDirty, _, wTy := wordfilter.ContainsDirtyWords(arg2.Content, false); hasDirty && wTy == gconsts.GeneralWords {
		return nil, gamedata.GameError(101)
	}

	return agent.CallBackend(pb.MessageID_G2V_COMMENTS_VIDEO, &pb.GCommentsVideoArg{
		VideoID:     arg2.VideoID,
		Content:     arg2.Content,
		Name:        player.GetName(),
		HeadImgUrl:  player.GetHeadImgUrl(),
		Country:     player.GetCountry(),
		HeadFrame:   player.GetHeadFrame(),
		CountryFlag: player.GetCountryFlag(),
	})
}

func rpc_C2S_TouristRegisterAccount(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TouristRegisterAccountArg)
	accountSeq := common.GenUUid("tourist")
	account := fmt.Sprintf("tourist%d", accountSeq)
	a, err := loadAccount(arg2.Channel, "", account, true)
	if err == nil {
		return nil, gamedata.GameError(1)
	}

	if err != attribute.NotExistsErr {
		return nil, err
	}

	pwdBuilder := strings.Builder{}
	for i := 0; i < 10; i++ {
		pwdBuilder.WriteString(strconv.Itoa(rand.Intn(10)))
	}
	pwd := pwdBuilder.String()
	glog.Infof("rpc_C2S_TouristRegisterAccount account=%s, pwd=%d, channel=%s", account, pwd, arg2.Channel)

	md5Writer := md5.New()
	io.WriteString(md5Writer, pwdPrefix+pwd)
	md5pwd := fmt.Sprintf("%x", md5Writer.Sum(nil))
	a.setPwd(md5HashPassword(md5pwd))
	if err = a.save(true); err != nil {
		return nil, err
	}

	return &pb.TouristRegisterAccountRelpy{
		Account:  account,
		Password: pwd,
	}, nil
}

func rpc_C2S_TouristBindAccount(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.TouristBindAccountArg)
	a, err := sdkAccountAuth(arg2.BindAccount)
	if err != nil {
		return nil, err
	}

	arcID := 1
	bindArc := a.getArchive(arcID)
	if bindArc != nil {
		return nil, gamedata.GameError(2)
	}

	touristAccount, err := loadAccount(arg2.Channel, "", arg2.TouristAccount, true)
	if err != nil {
		return nil, err
	}

	accountPwd := touristAccount.getPwd()
	md5Writer := md5.New()
	io.WriteString(md5Writer, pwdPrefix+arg2.TouristPassword)
	md5pwd := fmt.Sprintf("%x", md5Writer.Sum(nil))
	if md5HashPassword(md5pwd) != accountPwd {
		return nil, gamedata.GameError(3)
	}

	arc := touristAccount.getArchive(arcID)
	if arc == nil {
		return nil, gamedata.GameError(4)
	}

	arcData := arc.ToMap()
	bindArc = attribute.NewMapAttr()
	bindArc.AssignMap(arcData)
	a.setArchive(arcID, bindArc)
	glog.Infof("rpc_C2S_TouristBindAccount, channel=%s, TouristAccount=%s, TouristPwd=%s, BindLoginChannel=%s, "+
		"BindChannelID=%s, uid=%d", arg2.Channel, arg2.TouristAccount, arg2.TouristPassword, arg2.BindAccount.LoginChannel,
		arg2.BindAccount.ChannelID, arcData["uid"])
	return a.packMsg(), nil
}

func rpc_GT2C_OnSnetDisconnect(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*gpb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	clientID := common.UUid(arg2.ClientID)
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, nil
	}

	agent := player.GetAgent()
	if agent == nil || agent.GetClientID() != clientID {
		return nil, nil
	}

	player.(*Player).onNetDisconnect()
	return nil, nil
}

func rpc_GT2C_OnSnetReconnect(ses *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*gpb.PlayerClient)
	uid := common.UUid(arg2.Uid)
	clientID := common.UUid(arg2.ClientID)
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, nil
	}

	agent := player.GetAgent()
	if agent == nil || agent.GetClientID() != clientID {
		return nil, nil
	}

	player.(*Player).onNetConnect()
	return nil, nil
}

func rpc_C2S_FbAdvertReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player2 := player.(*Player)
	if player2.isFbAdvertReward() {
		return nil, gamedata.GameError(1)
	}

	player2.setFbAdvertReward()
	glog.Infof("rpc_C2S_FbAdvertReward uid=%d", player2.GetUid())
	sender := module.Mail.NewMailSender(uid)
	sender.SetTypeAndArgs(pb.MailTypeEnum_FbAdvert)
	reward := sender.GetRewardObj()
	reward.AddItem(pb.MailRewardType_MrtTreasure, "BX0204", 1)
	sender.Send()
	return nil, nil
}

func rpc_C2S_FetchHead(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := &pb.HeadData{}
	cards := module.Card.GetOnceCards(player)
	for _, card := range cards {
		reply.OnceCards = append(reply.OnceCards, card.GetCardID())
	}
	return reply, nil
}

func rpc_C2S_FetchVipRemainTime(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	return &pb.VipRemainTime{
		RemainTime: int32(module.Shop.GetVipRemainTime(player)),
	}, nil
}

func rpc_C2S_Fire233BindAccount(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.RegisterAccount)
	md5Writer := md5.New()
	io.WriteString(md5Writer, pwdPrefix+arg2.Password)
	password := fmt.Sprintf("%x", md5Writer.Sum(nil))
	a, err := doRegistAccount(arg2.Channel, arg2.Account, password)
	if err != nil {
		return nil, err
	}

	a.setRawPwd(arg2.Password)
	a.newArchiveByUid(1, uid)
	player.(*Player).setFire233BindAccount(arg2.Account)

	return nil, nil
}

func rpc_C2S_FetchFire233BindAccount(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	player2 := player.(*Player)
	account := player2.getFire233BindAccount()
	if account == "" {
		return nil, gamedata.GameError(1)
	}

	a, err := loadAccount(player2.GetChannel(), "", account, false)
	if err != nil {
		return nil, gamedata.GameError(2)
	}

	return &pb.RegisterAccount{
		Channel:  player2.GetChannel(),
		Account:  account,
		Password: a.getRawPwd(),
	}, nil
}

func rpc_C2S_FetchAccountCode(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	acCode := player.(*Player).getAccountCode()
	reply := &pb.AccountCode{}
	if acCode != nil {
		reply.Code = acCode.getCode()
	}
	return reply, nil
}

func rpc_C2S_GenAccountCode(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if config.GetConfig().HostID < 1000 {
		return nil, gamedata.GameError(1)
	}

	accountType := player.GetAccountType()
	if accountType == pb.AccountTypeEnum_Wxgame {
		return nil, gamedata.GameError(2)
	}

	if player.GetChannel() == "lzd_xianfeng_recharge" {
		return nil, gamedata.GameError(5)
	}

	if player.GetPvpTeam() < 4 {
		return nil, gamedata.GameError(3)
	}

	acCode := genAccountCode(player.(*Player))
	if acCode == nil {
		return nil, gamedata.GameError(4)
	}

	code := acCode.getCode()
	glog.Infof("rpc_C2S_GenAccountCode, uid=%d, code=%s", uid, code)

	return &pb.AccountCode{
		Code: code,
	}, nil
}

func rpc_C2S_FetchAccountCodePlayerInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	arg2 := arg.(*pb.AccountCode)
	acCode := loadAccountCode(arg2.Code)
	if acCode == nil {
		return nil, gamedata.GameError(1)
	}

	p := acCode.loadPlayer()
	if p == nil {
		return nil, gamedata.GameError(2)
	}

	return &pb.AccountCodePlayerInfo{
		Uid:       p.Uid,
		Name:      p.Name,
		PvpScore:  p.PvpScore,
		HeadImg:   p.HeadImgUrl,
		HeadFrame: p.HeadFrame,
	}, nil
}

func rpc_C2S_BindAccountCode(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if config.GetConfig().HostID < 1000 {
		return nil, gamedata.GameError(1)
	}

	myPvpLevel := player.GetPvpLevel()
	if myPvpLevel < 2 || myPvpLevel > 10 {
		return nil, gamedata.GameError(2)
	}

	//resCpt := player.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	//funcPrice := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	//if !resCpt.HasResource(consts.Jade, funcPrice.AccountTransfer) {
	//	return nil, gamedata.GameError(3)
	//}

	p := player.(*Player)
	arg2 := arg.(*pb.AccountCode)
	acCode := loadAccountCode(arg2.Code)
	if acCode == nil {
		return nil, gamedata.GameError(4)
	}

	if acCode.getPvpLevel() < myPvpLevel {
		return nil, gamedata.GameError(5)
	}

	toBindAccount, err := loadAccountByAccountID(acCode.getAccountID(), acCode.getRegion())
	if err != nil || toBindAccount == nil {
		return nil, gamedata.GameError(6)
	}

	myAccount, err := loadAccountByAccountID(p.getAccountID())
	if err != nil || myAccount == nil {
		return nil, gamedata.GameError(7)
	}

	//resCpt.ModifyResource(consts.Jade, - funcPrice.AccountTransfer, consts.RmrAccountTransfer)
	glog.Infof("rpc_C2S_BindAccountCode, uid=%d, code=%s, targetUid=%d", uid, arg2.Code, acCode.getUid())

	err = acCode.del(true)
	if err != nil {
		return nil, gamedata.GameError(8)
	}

	err = myAccount.bindOthAccount(toBindAccount)
	if err != nil {
		return nil, gamedata.GameError(9)
	}

	return nil, nil
}

func rpc_C2S_FetchVideoList(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	return agent.CallBackend(pb.MessageID_G2V_FETCH_VIDEO_LIST, &pb.TargetArea{
		Area: int32(mod.GetPlayer(agent.GetUid()).GetArea()),
	})
}

func rpc_C2S_AdultCertification(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	module.Player.GetPlayer(agent.GetUid()).(*Player).adultCertification(arg.(*pb.AdultCertificationArg).IsAdult)
	return nil, nil
}

func rpc_C2S_OnLoginNoticeShow(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	module.Player.GetPlayer(agent.GetUid()).(*Player).onLoginNoticeShow(int(arg.(*pb.OnLoginNoticeShowArg).Version))
	return nil, nil
}

func rpc_C2S_DelHint(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	hintArg := arg.(*pb.Hint)
	if hintArg.Count == 0 {
		player.DelHint(hintArg.Type)
		return nil, nil
	}
	player.SubHint(hintArg.Type, int(hintArg.Count))
	return nil, nil
}

func rpc_C2S_FetchUpdateCountryFlagCD(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	return &pb.UpdateCountryFlagRemainTime{
		RemainTime: int32(player.(*Player).getUpdateCountryFlagCD()),
	}, nil
}

func rpc_C2S_UpdateCountryFlag(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := mod.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	p := player.(*Player)
	if p.getUpdateCountryFlagCD() > 0 {
		return nil, gamedata.GameError(1)
	}
	p.setCountryFlag(arg.(*pb.UpdateNationalFlagArg).CountryFlag)
	return nil, nil
}

func rpc_C2S_Ping(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	return mod.pong, nil
}

func rpc_G2G_GetOnlineInfo(_ *network.Session, arg interface{}) (interface{}, error) {
	infos := mod.getOnlineInfo()
	reply := &pb.OnlineInfo{}
	for area, accountType2info := range infos {
		for accountType, info := range accountType2info {
			reply.Infos = append(reply.Infos, &pb.AreaAccountTypeOnlineInfo{
				AccountType:     accountType,
				PlayerAmount:    int32(info.onlineAmount),
				TotalOnlineTime: int32(info.onlineTime),
				Area:            int32(area),
			})
		}
	}
	return reply, nil
}

func rpc_G2G_OnBindAccount(_ *network.Session, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.OnBindAccountArg)
	mod.onBindAccount(arg2.AccountID, arg2.BindRegion)
	return nil, nil
}

func registerRpc() {
	//evq.HandleEvent(pconsts.SESSION_ON_CLOSE_EVENT, onLogout)
	eventhub.Subscribe(logic.CLIENT_CLOSE_EV, onLogout)
	eventhub.Subscribe(logic.PLAYER_KICK_OUT_EV, onPlayerKickOut)
	eventhub.Subscribe(logic.RESTORE_AGENT_EV, onRestoreAgent)
	evq.HandleEvent(consts.EvReloadConfig, onReloadConfig)

	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ACCOUNT_LOGIN, rpc_C2S_AccountLogin)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LOGIN, rpc_C2S_Login)
	//peer.RegisterRpcHandler(pb.MessageID_C2S_EXCHANGE_RESOURCE, rpc_C2S_ExchangeResource)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GM_COMMAND, rpc_C2S_GmCommand)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LOAD_FIGHT, rpc_C2S_LoadFight)
	//peer.RegisterRpcHandler(pb.MessageID_C2S_DEL_ARCHIVES, rpc_C2S_DelArchive)
	//peer.RegisterRpcHandler(pb.MessageID_C2S_PLAYER_LOGOUT, rpc_C2S_PlayerLogout)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FINISH_GUIDE, rpc_C2S_FinishGuide)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_GUIDE, rpc_C2S_FetchGuide)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REGISTER_ACCOUNT, rpc_C2S_RegisterAccount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SURVEY_INFO, rpc_C2S_FetchSurveyInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_COMPLETE_SURVEY, rpc_C2S_CompleteSurvey)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_SURVEY_REWARD, rpc_C2S_GetSurveyReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_MODIFY_NAME, rpc_C2S_ModifyName)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_NAME, rpc_C2S_UpdateName)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SDK_ACCOUNT_LOGIN, rpc_C2S_SdkAccountLogin)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_VERSION, rpc_C2S_FetchVersion)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_SDK_USER_INFO, rpc_C2S_UpdateSdkUserInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_RECORD_CUR_GUIDE_GROUP, rpc_C2S_RecordCurGuideGroup)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_CUR_GUIDE_GROUP, rpc_C2S_FetchCurGuideGroup)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_HEADIMG, rpc_C2S_UpdateHeadImg)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SHARE_VIDEO, rpc_C2S_ShareVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_COMMENTS_VIDEO, rpc_C2S_CommentsVideo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_TOURIST_REGISTER_ACCOUNT, rpc_C2S_TouristRegisterAccount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_TOURIST_BIND_ACCOUNT, rpc_C2S_TouristBindAccount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GET_FBADVERT_REWARD, rpc_C2S_FbAdvertReward)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_HEAD, rpc_C2S_FetchHead)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_VIP_REMAIN_TIME, rpc_C2S_FetchVipRemainTime)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FIRE233_BIND_ACCOUNT, rpc_C2S_Fire233BindAccount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_FIRE233_BIND_ACCOUNT, rpc_C2S_FetchFire233BindAccount)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_ACCOUNT_CODE, rpc_C2S_FetchAccountCode)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_GEN_ACCOUNT_CODE, rpc_C2S_GenAccountCode)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_ACCOUNT_CODE_PLAYER_INFO, rpc_C2S_FetchAccountCodePlayerInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BIND_ACCOUNT_CODE, rpc_C2S_BindAccountCode)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_VIDEO_LIST, rpc_C2S_FetchVideoList)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ADULT_CERTIFICATION, rpc_C2S_AdultCertification)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_ON_LOGIN_NOTICE_SHOW, rpc_C2S_OnLoginNoticeShow)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_DEL_HINT, rpc_C2S_DelHint)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_UPDATE_COUNTRY_FLAG_CD, rpc_C2S_FetchUpdateCountryFlagCD)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_UPDATE_COUNTRY_FLAG, rpc_C2S_UpdateCountryFlag)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_PING, rpc_C2S_Ping)

	logic.RegisterRpcHandler(pb.MessageID_G2G_GET_ONLINE_INFO, rpc_G2G_GetOnlineInfo)
	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_BIND_ACCOUNT, rpc_G2G_OnBindAccount)

	api.RegisterCenterRpcHandler(gpb.MessageID_GT2C_ON_SNET_DISCONNECT, rpc_GT2C_OnSnetDisconnect)
	api.RegisterCenterRpcHandler(gpb.MessageID_GT2C_ON_SNET_RECONNECT, rpc_GT2C_OnSnetReconnect)

	logic.RegisterLoadPlayerHandler(rpc_LoadPlayer)
	//logic.RegisterAgentRpcHandler(pb.MessageID_G2G_LOAD_PLAYER, rpc_G2G_LoadPlayer)
}
