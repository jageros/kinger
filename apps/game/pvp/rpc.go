package pvp

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/aicardpool"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
	"math/rand"
	"strconv"
)

func rpc_C2S_BeginMatch(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	if module.OutStatus.GetStatus(player, consts.OtFatigue) != nil {
		player.Tellme("由于您的实名认证未达要求，游戏时长已达3小时，现已进入#cb防沉迷#n状态，今天无法继续对战。", 0)
		return nil, gamedata.InternalErr
	}

	var strategy iMatchStrategy
	if module.Player.GetResource(player, consts.GuidePro) < consts.MaxGuidePro {
		// 新手的5场战斗
		strategy = guideMatchStrategy
	} else {
		pvpCpt := player.GetComponent(consts.PvpCpt).(*pvpComponent)
		newbiePvpCamp, _ := pvpCpt.getNewbiePvpEnemyCamp()
		if newbiePvpCamp > 0 {
			strategy = newbiePvpMatchStrategy
		} else {
			strategy = pvpMatchStrategy
		}
	}

	camp := int(arg.(*pb.MatchArg).Camp)
	cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
	if camp <= 0 {
		camp = cardCpt.GetFightCamp()
	}

	matchArg, reply, err := strategy.packMatchArgAndReply(player, camp)
	if err != nil {
		return nil, err
	}

	mod.CancelPvpBattle(player)
	evq.CallLater(func() {
		agent.PushBackend(strategy.getMatchMessageID(), matchArg)
	})
	return reply, nil
}

func rpc_C2S_SeasonPvpChooseCamp(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	chooseCardData := module.Huodong.SeasonPvpChooseCamp(player, int(arg.(*pb.SeasonPvpChooseCampArg).Camp))
	if chooseCardData == nil {
		return nil, gamedata.GameError(1)
	}
	return chooseCardData, nil
}

func rpc_C2S_SeasonPvpChooseCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	randCards, err := module.Huodong.SeasonPvpChooseCard(player, arg.(*pb.SeasonPvpChooseCardArg).CardIDs)
	if err != nil {
		return nil, err
	}

	return &pb.SeasonPvpChooseCardReply{
		CardIDs: randCards,
	}, nil
}

func rpc_C2S_FetchSeasonHandCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	_, _, _, _, handCardsMsg := module.Huodong.GetSeasonPvpHandCardInfo(player)
	if handCardsMsg == nil {
		return nil, gamedata.GameError(1)
	}
	return handCardsMsg, nil
}

func rpc_C2S_RefreshSeasonPvpChooseCard(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	return module.Huodong.RefreshSeasonPvpChooseCard(player)
}

func rpc_L2G_GetPvpFighterData(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}

	reply := player.GetComponent(consts.PvpCpt).(*pvpComponent).GetPvpFighterData()
	arg2 := arg.(*pb.GetFighterDataArg)

	if len(arg2.CardIDs) > 0 {
		reply.Camp = 0
		reply.HandCards = []*pb.SkinGCard{}
		cardCpt := player.GetComponent(consts.CardCpt).(types.ICardComponent)
		for _, cardID := range arg2.CardIDs {
			card := cardCpt.GetCollectCard(cardID)
			if card == nil || card.GetState() == pb.CardState_InCampaignMs {
				return nil, gamedata.GameError(1)
			}

			cardData := card.GetCardGameData()
			if reply.Camp == 0 {
				if cardData.Camp != consts.Heroes {
					reply.Camp = int32(cardData.Camp)
				}
			} else if cardData.Camp != consts.Heroes && cardData.Camp != int(reply.Camp) {
				return nil, gamedata.GameError(2)
			}

			reply.HandCards = append(reply.HandCards, &pb.SkinGCard{
				GCardID: card.GetCardGameData().GetGCardID(),
				Skin:    card.GetSkin(),
				Equip:   card.GetEquip(),
			})
		}
	}

	if len(reply.HandCards) < consts.MaxHandCardAmount {
		return nil, gamedata.GameError(2)
	}
	return reply, nil
}

func rpc_C2S_LogBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	arg2 := arg.(*pb.LogBattleArg)
	glog.Infof("logBattle, uid=%d battleID=%d event=%s error=%s", agent.GetUid(), arg2.BattleID, arg2.Event, arg2.ErrorMsg)
	return nil, nil
}

func rpc_C2S_BeginTrainingBattle(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	fighter1 := player.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpFighterData()
	if len(fighter1.HandCards) != 5 {
		return nil, gamedata.GameError(1)
	}
	fighter1.WinRate = 100

	camp, robotHandCards := aicardpool.RandomCardPool(player.GetPvpLevel(), int(fighter1.Camp),
		fighter1.HandCards, true)
	gcardID := robotHandCards[rand.Intn(len(robotHandCards))].GCardID
	poolGameData := gamedata.GetGameData(consts.Pool).(*gamedata.PoolGameData)
	cardData := poolGameData.GetCardByGid(gcardID)
	var headImgUrl string
	if cardData != nil {
		headImgUrl = fmt.Sprintf("avatar_%s_png", cardData.Head)
	}

	fighter2 := &pb.FighterData{
		Uid:        1,
		HandCards:  robotHandCards,
		Camp:       int32(camp),
		PvpScore:   int32(player.GetPvpScore()),
		HeadImgUrl: headImgUrl,
		HeadFrame:  "1",
		Area:       int32(player.GetArea()),
		Region:     module.Service.GetRegion(),
		NameText:   70426,
		IsRobot:    true,
	}

	logic.PushBackend("", 0, pb.MessageID_M2B_BEGIN_BATTLE, &pb.BeginBattleArg{
		BattleType:         int32(consts.BtTraining),
		Fighter1:           fighter1,
		Fighter2:           fighter2,
		NeedFortifications: true,
		NeedVideo:          true,
	}, module.Service.GetRegion())
	return nil, nil
}

func rpc_C2S_FetchLeagueRewardInfo(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	p := player.GetComponent(consts.PvpCpt).(*pvpComponent)
	reply := &pb.LeagueRewardInfo{
		HasReceive: map[int32]bool{},
	}
	reply.RemainTime = leagueAttr.getRemainTime(player.GetArea())
	ids := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetIdList()
	for _, id := range ids {
		reply.HasReceive[int32(id)] = p.getRewardReceiveStatus(id)
	}
	return reply, nil
}

func rpc_C2S_ReceiveLeagueReward(agent *logic.PlayerAgent, arg interface{}) (interface{}, error) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return nil, gamedata.InternalErr
	}
	p := player.GetComponent(consts.PvpCpt).(*pvpComponent)
	reply := &pb.LeagueRewards{}
	rid := int(arg.(*pb.LeagueRewardId).RewardId)
	if player.GetMaxRankScore() < gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetScoreById(rid) {
		return nil, gamedata.InternalErr
	}
	reward := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetRewardById(rid)
	resource := consts.RmrLeagueReward + strconv.Itoa(rid)
	reply.LeagueReward = module.Reward.GiveRewardList(player, reward, resource)
	p.setRewardReceive(rid)
	return reply, nil
}

func rpc_G2G_SaveLeagueAttrReload(_ *network.Session, arg interface{}) (interface{}, error) {
	attrArg := arg.(*pb.ReloadLeagueAttrArg)
	if attrArg.AppID == module.Service.GetAppID() {
		return nil, nil
	}
	leagueAttr.reloadAttrForArea(int(attrArg.Area), attrArg.IsCrossSeason)
	return nil, nil
}

func registerRpc() {
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BEGIN_MATCH, rpc_C2S_BeginMatch)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SEASON_PVP_CHOOSE_CAMP, rpc_C2S_SeasonPvpChooseCamp)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_SEASON_PVP_CHOOSE_CARD, rpc_C2S_SeasonPvpChooseCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_SEASON_HAND_CARD, rpc_C2S_FetchSeasonHandCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_REFRESH_SEASON_PVP_CHOOSE_CARD, rpc_C2S_RefreshSeasonPvpChooseCard)
	logic.RegisterAgentRpcHandler(pb.MessageID_L2G_GET_PVP_FIGHTER_DATA, rpc_L2G_GetPvpFighterData)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_LOG_BATTLE, rpc_C2S_LogBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_BEGIN_TRAINING_BATTLE, rpc_C2S_BeginTrainingBattle)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_FETCH_LEAGUE_REWARD_INFO, rpc_C2S_FetchLeagueRewardInfo)
	logic.RegisterAgentRpcHandler(pb.MessageID_C2S_RECEIVE_LEAGUE_REWARD, rpc_C2S_ReceiveLeagueReward)
	logic.RegisterRpcHandler(pb.MessageID_G2G_SAVE_LEAGUE_ATTR_RELOAD, rpc_G2G_SaveLeagueAttrReload)
}
