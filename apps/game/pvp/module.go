package pvp

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/eventhub"
	"kinger/proto/pb"
	"math"
)

var (
	mod        *pvpModule
	leagueAttr *leagueSeasonAttr
)

const (
	minMmr = 1000
)

type pvpModule struct {
	gdata *gamedata.RankGameData
}

func newPvpModule() *pvpModule {
	return &pvpModule{
		gdata: gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData),
	}
}

func (p *pvpModule) NewPvpComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent {
	pvpAttr := playerAttr.GetMapAttr("pvp")
	if pvpAttr == nil {
		pvpAttr = attribute.NewMapAttr()
		playerAttr.SetMapAttr("pvp", pvpAttr)
	}
	return &pvpComponent{attr: pvpAttr}
}

func (p *pvpModule) GetPvpLevelByStar(star int) int {
	if star <= 0 {
		return 1
	}

	for i := 0; i < len(p.gdata.RankList); i++ {
		rank := p.gdata.RankList[i]
		if star <= rank.LevelUpStar {
			return rank.ID
		}
	}
	return p.gdata.RankList[len(p.gdata.RankList)-1].ID
}

func (p *pvpModule) getMaxStarByPvpLevel(level int) int {
	rankData, ok := p.gdata.Ranks[level]
	if !ok {
		return 0
	}
	return rankData.LevelUpStar
}

func (p *pvpModule) GetMinStarByPvpLevel(level int) int {
	return p.getMaxStarByPvpLevel(level-1) + 1
}

func (p *pvpModule) CanPvpMatch(player types.IPlayer) bool {
	if player.IsMultiRpcForbid(consts.FmcMatch) {
		return false
	}
	if module.Campaign.IsInCampaignMatch(player) {
		return false
	}
	if module.Social.IsInInviteBattle(player.GetUid()) {
		return false
	}
	if module.WxGame.IsInInviteBattle(player.GetUid()) {
		return false
	}

	battleID := player.GetBattleID()
	if battleID > 0 && player.GetBattleType() == consts.BtPvp {
		_, err := logic.CallBackend(consts.AppBattle, player.GetBattleAppID(), pb.MessageID_G2B_IS_BATTLE_ALIVE, &pb.CancelBattleArg{
			Uid:      uint64(player.GetUid()),
			BattleID: uint64(battleID),
		})
		return err != nil
	}

	return true
}

func (p *pvpModule) CancelPvpBattle(player types.IPlayer) {
	battleID := player.GetBattleID()
	if battleID <= 0 {
		return
	}

	if player.GetBattleType() != consts.BtPvp {
		return
	}

	logic.PushBackend(consts.AppBattle, player.GetBattleAppID(), pb.MessageID_G2B_CANCEL_BATTLE, &pb.CancelBattleArg{
		Uid:      uint64(player.GetUid()),
		BattleID: uint64(battleID),
	})
}

func onResUpdate(args ...interface{}) {
	resType := args[1].(int)
	if resType != consts.Score && resType != consts.MaxScore {
		return
	}

	player := args[0].(types.IPlayer)
	player.GetComponent(consts.PvpCpt).(*pvpComponent).onResUpdate(resType, args[2].(int))
}

func onRecharge(args ...interface{}) {
	player := args[0].(types.IPlayer)
	money := args[1].(int)
	matchParam := gamedata.GetGameData(consts.MatchParam).(*gamedata.MatchParamGameData)
	pvpCpt := player.GetComponent(consts.PvpCpt).(*pvpComponent)
	modifyIndex := -int(math.Ceil(math.Sqrt(float64(money)) * matchParam.RechargeRevise))
	index := pvpCpt.getRechargeMatchIndex() + modifyIndex
	if float64(index) < matchParam.RechargeReviseLimit {
		index = int(matchParam.RechargeReviseLimit)
	}
	pvpCpt.setRechargeMatchIndex(index)
}

func Initialize() {
	leagueAttr = &leagueSeasonAttr{}
	leagueAttr.init()
	mod = newPvpModule()
	module.Pvp = mod
	registerRpc()
	//loadShareVideoList()
	eventhub.Subscribe(consts.EvResUpdate, onResUpdate)
	eventhub.Subscribe(consts.EvRecharge, onRecharge)
	leagueAttr.initCrossSeasonFunction()
}

func (m *pvpModule) GM_CrossSeason(agent *logic.PlayerAgent, com string) {
	uid := agent.GetUid()
	player := module.Player.GetPlayer(uid)
	if player == nil {
		return
	}
	if com == "reset" {
		for area, _ := range leagueAttr.attr {
			leagueAttr.setSeasonSerial(area, 1, false)
		}

		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			player.GetComponent(consts.PvpCpt).(*pvpComponent).setSeassonSerial()
		})
		return
	}
	if com == "cross" {
		leagueAttr.crossSeasonUpdateAttr()
		//module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		//	player.GetComponent(consts.PvpCpt).(*pvpComponent).updateLeagueSeason()
		//})
		return
	}
}
