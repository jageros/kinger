package player

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/common/utils"
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/eventhub"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"strconv"
	"time"
)

var _ types.IPlayer = &Player{}

type Player struct {
	isLogout        bool
	uid             common.UUid
	agent           *logic.PlayerAgent
	attr            *attribute.AttrMgr
	components      map[string]types.IPlayerComponent
	componentList   []types.IPlayerComponent
	hints           map[pb.HintType]int
	multiClickGuard *forbidMultiClick
}

func newPlayer(uid common.UUid, agent *logic.PlayerAgent, attr *attribute.AttrMgr) *Player {
	p := &Player{
		uid:             uid,
		agent:           agent,
		attr:            attr,
		components:      make(map[string]types.IPlayerComponent),
		hints:           map[pb.HintType]int{},
		multiClickGuard: newForbidMultiClick(),
	}

	if attr != nil {
		p.createComponent()
	}

	return p
}

func (p *Player) addComponent(cpt types.IPlayerComponent) {
	cptID := cpt.ComponentID()
	if _, ok := p.components[cptID]; ok {
		return
	}
	p.components[cptID] = cpt
	p.componentList = append(p.componentList, cpt)
	cpt.OnInit(p)
}

func (p *Player) createComponent() {
	fatigueCpt := newFatigueComponent(p.attr, false)
	if fatigueCpt != nil {
		p.addComponent(fatigueCpt)
	}

	p.addComponent(module.Level.NewLevelComponent(p.attr))
	p.addComponent(module.Bag.NewComponent(p.attr))
	p.addComponent(module.Card.NewCardComponent(p.attr))
	p.addComponent(&ResourceComponent{})
	p.addComponent(module.Pvp.NewPvpComponent(p.attr))
	p.addComponent(module.Reborn.NewComponent(p.attr))
	p.addComponent(module.OutStatus.NewComponent(p.attr))
	p.addComponent(module.Treasure.NewTreasureComponent(p.attr))
	p.addComponent(module.Tutorial.NewTutorialComponent(p.attr))
	p.addComponent(newSurveyComponent(p))
	p.addComponent(module.GiftCode.NewComponent(p.attr))
	p.addComponent(module.Social.NewComponent(p.attr))
	p.addComponent(module.WxGame.NewComponent(p.attr))
	p.addComponent(module.Shop.NewComponent(p.attr))
	p.addComponent(module.Mission.NewComponent(p.attr))
	p.addComponent(module.Mail.NewComponent(p.attr))
	p.addComponent(module.Huodong.NewComponent(p.attr))
	p.addComponent(module.Campaign.NewComponent(p.attr))
	p.addComponent(module.Activitys.NewComponent(p.attr))
}

func (p *Player) getAttr() *attribute.AttrMgr {
	return p.attr
}

func (p *Player) Save(needReply bool) {
	if p.attr != nil {
		p.attr.Save(needReply)
	}
}

func (p *Player) kickOut(reason int32) {
	if p.agent != nil {
		p.agent.PushClient(pb.MessageID_S2C_KICK_OUT, &pb.KickOut{Reason: reason})
	}
}

func (p *Player) delAgent() {
	if p.agent != nil {
		logic.DelPlayerAgent(p.agent.GetUid(), p.agent.GetClientID())
		p.agent = nil
	}
}

func (p *Player) relogin(agent *logic.PlayerAgent) {
	p.kickOut(Relogin)
	p.delAgent()
	p.agent = agent
}

func (p *Player) finishGuide(guideID int) {
	guideAttr := p.getGuideAttr()

	isFinish := false
	guideAttr.ForEachIndex(func(index int) bool {
		guideID2 := guideAttr.GetInt(index)
		if guideID == guideID2 {
			isFinish = true
			return false
		}
		return true
	})

	if isFinish {
		return
	}

	module.Player.LogMission(p, fmt.Sprintf("guideStep_%d", guideID), 3)
	guideAttr.AppendInt(guideID)
}

func (p *Player) getGuideAttr() *attribute.ListAttr {
	guideAttr := p.attr.GetListAttr("guide")
	if guideAttr == nil {
		guideAttr = attribute.NewListAttr()
		p.attr.SetListAttr("guide", guideAttr)
	}
	return guideAttr
}

func (p *Player) GetModifytime() string {
	return p.attr.GetStr("modifyTime")
}

func (p *Player) setModifytime() {
	now := time.Now().Unix()
	strNow := strconv.FormatInt(now, 10)
	p.attr.SetStr("modifyTime", strNow)
}

func (p *Player) GetUid() common.UUid {
	return p.uid
}

func (p *Player) GetName() string {
	//return p.attr.GetStr("name")
	name := p.attr.GetStr("name")
	rebornCnt := module.Reborn.GetRebornCnt(p)
	switch rebornCnt {
	case 0:
	case 1:
		name = "#c5cfc60" + name + "#n"
	case 2:
		name = "#c5ca4fc" + name + "#n"
	case 3:
		name = "#cc95cfc" + name + "#n"
	case 4:
		name = "#cfc855c" + name + "#n"
	case 5:
		name = "#cfce75c" + name + "#n"
	default:
		name = "#cfc5c5c" + name + "#n"
	}
	return name
}

func (p *Player) GetChannel() string {
	return p.attr.GetStr("channel")
}
func (p *Player) setChannel(channel string) {
	p.attr.SetStr("channel", channel)
}

func (p *Player) GetSubChannel() string {
	return p.attr.GetStr("subChannel")
}

func (p *Player) setSubChannel(subChannel string) {
	p.attr.SetStr("subChannel", subChannel)
}

func (p *Player) GetLoginChannel() string {
	return p.attr.GetStr("loginChannel")
}
func (p *Player) setLoginChannel(channel string) {
	p.attr.SetStr("loginChannel", channel)
}

func (p *Player) GetServerID() string {
	return p.attr.GetStr("serverID")
}

func (p *Player) GetIP() string {
	return p.attr.GetStr("ipAddr")
}

func (p *Player) SetIP(ipAddr string) {
	p.attr.SetStr("ipAddr", ipAddr)
}

func (p *Player) setDefaultName() {
	p.attr.SetStr("name", p.attr.GetStr("channelID"))
}

func (p *Player) setName(name string) {
	oldName := p.GetName()
	p.attr.SetStr("name", name)
	if oldName != name {
		p.OnSimpleInfoUpdate()
	}
}

func (p *Player) OnSimpleInfoUpdate() {
	logic.BroadcastBackend(pb.MessageID_L2L_UPDATE_SIMPLE_PLAYER, &pb.UpdateSimplePlayerArg{
		Uid:        uint64(p.GetUid()),
		Name:       p.GetName(),
		HeadImgUrl: p.GetHeadImgUrl(),
		HeadFrame:  p.GetHeadFrame(),
		PvpScore:   int32(p.GetPvpScore()),
	})
}

func (p *Player) GetAgent() *logic.PlayerAgent {
	return p.agent
}

func (p *Player) GetComponent(componentID string) types.IPlayerComponent {
	if cpt, ok := p.components[componentID]; ok {
		return cpt
	} else {
		return nil
	}
}

func (p *Player) IsRobot() bool {
	return false
}

func (p *Player) IsOnline() bool {
	if p.agent == nil {
		return false
	}
	pa := logic.GetPlayerAgent(p.agent.GetUid())
	return pa != nil && pa.GetClientID() == p.agent.GetClientID()
}

func (p *Player) OnBeginBattle(battleID common.UUid, battleType int, battleAppID uint32) {
	glog.Infof("OnBeginBattle, uid=%d, battleID=%d, battletype=%d, battleAppID=%d, oldBattleID=%d, oldBattleAppID=%d, oldBattleType=%d",
		p.GetUid(), battleID, battleType, battleAppID, p.attr.GetUInt64("battleID"), p.attr.GetUInt32("battleAppID"), p.attr.GetInt("battleType"))
	if p.attr.GetUInt64("battleID") > 0 {
		glog.Infof("fuck OnBeginBattle, uid=%d, battleID=%d, battletype=%d, battleAppID=%d, oldBattleID=%d, oldBattleAppID=%d, oldBattleType=%d",
			p.GetUid(), battleID, battleType, battleAppID, p.attr.GetUInt64("battleID"), p.attr.GetUInt32("battleAppID"), p.attr.GetInt("battleType"))
	}

	module.Pvp.CancelPvpBattle(p)
	p.attr.SetUInt64("battleID", uint64(battleID))
	p.attr.SetUInt32("battleAppID", battleAppID)
	p.attr.SetInt("battleType", battleType)
	module.Social.OnBattleBegin(p)
	eventhub.Publish(consts.EvBeginBattle, p.GetUid(), battleID, battleType)
}

func (p *Player) OnBattleEnd(battleID common.UUid, battleType int, winner, loser *pb.EndFighterData, levelID int,
	isWonderful bool) {

	if p.attr.GetUInt64("battleID") != uint64(battleID) {
		return
	}

	p.attr.SetUInt64("battleID", 0)

	fighter := winner
	isWin := true
	opponent := loser
	oppMMr := int(loser.Mmr)
	oppCamp := int(loser.Camp)
	oppArea := int(loser.Area)
	if common.UUid(fighter.Uid) != p.GetUid() {
		fighter = loser
		opponent = winner
		isWin = false
		oppMMr = int(winner.Mmr)
		oppCamp = int(winner.Camp)
		oppArea = int(winner.Area)
	}

	var initHandCards []uint32
	for _, c := range fighter.InitHandCards {
		initHandCards = append(initHandCards, c.GCardID)
	}

	if battleType == consts.BtPvp || battleType == consts.BtGuide {
		isFirstHand := fighter.IsFirstHand
		if isFirstHand {
			p.attr.SetInt("firstHandAmount", p.attr.GetInt("firstHandAmount")+1)
		} else {
			p.attr.SetInt("backHandAmount", p.attr.GetInt("backHandAmount")+1)
		}

		if isWin {
			if isFirstHand {
				p.attr.SetInt("firstHandWinAmount", p.attr.GetInt("firstHandWinAmount")+1)
			} else {
				p.attr.SetInt("backHandWinAmount", p.attr.GetInt("backHandWinAmount")+1)
			}
		}
	}

	if battleType == consts.BtPvp {
		//p.attr.SetUInt64("lastBattleID", uint64(battleID))
		if !opponent.IsRobot {
			p.attr.SetUInt64("lastOpponent", opponent.Uid)
		}

		p.GetComponent(consts.PvpCpt).(types.IPvpComponent).OnBattleEnd(fighter, isWin, isWonderful, oppMMr, oppCamp, oppArea)
		p.GetComponent(consts.CardCpt).(types.ICardComponent).OnPvpBattleEnd(initHandCards)
	} else if battleType == consts.BtLevel {
		p.GetComponent(consts.LevelCpt).(types.ILevelComponent).OnBattleEnd(fighter, isWin, levelID, battleID)
	} else if battleType == consts.BtGuide {
		p.GetComponent(consts.TutorialCpt).(types.ITutorialComponent).OnBattleEnd(fighter, isWin)
		p.GetComponent(consts.CardCpt).(types.ICardComponent).OnPvpBattleEnd(initHandCards)
	} else if battleType == consts.BtFriend {
		p.GetComponent(consts.SocialCpt).(types.ISocialComponent).OnBattleEnd(fighter, isWin)
	} else if battleType == consts.BtLevelHelp {
		p.GetComponent(consts.LevelCpt).(types.ILevelComponent).OnHelpBattleEnd(isWin, battleID)
	} else if battleType == consts.BtCampaign {
		module.Campaign.OnBattleEnd(p, isWin)
	} else if battleType == consts.BtTraining {
		p.GetComponent(consts.PvpCpt).(types.IPvpComponent).OnTrainingBattleEnd(fighter, isWin)
	} else {
		glog.Errorf("onBattleEnd err battletype %d, uid=%d", battleType, p.GetUid())
	}
}

func (p *Player) GetBattleID() common.UUid {
	return common.UUid(p.attr.GetUInt64("battleID"))
}

func (p *Player) GetBattleType() int {
	return p.attr.GetInt("battleType")
}

func (p *Player) clearBattleID() {
	p.attr.SetUInt64("battleID", 0)
}

func (p *Player) GetBattleAppID() uint32 {
	return p.attr.GetUInt32("battleAppID")
}

func (p *Player) GetLastBattleID() common.UUid {
	return common.UUid(p.attr.GetUInt64("lastBattleID"))
}

func (p *Player) GetPvpScore() int {
	resCpt := p.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	return resCpt.GetResource(consts.Score)
}

func (p *Player) GetHeadImgUrl() string {
	return p.attr.GetStr("headImgUrl")
}

func (p *Player) setHeadImgUrl(url string) {
	p.attr.SetStr("headImgUrl", url)
}

func (p *Player) GetCountryFlag() string {
	return p.attr.GetStr("cryFlag")
}
func (p *Player) setCountryFlag(flag string) {
	if flag == p.GetCountryFlag() {
		return
	}
	p.attr.SetStr("cryFlag", flag)
	p.attr.SetInt64("updateCryFlagTime", time.Now().Unix())
}

func (p *Player) getUpdateCountryFlagCD() int {
	lastUpdateTime := p.attr.GetInt64("updateCryFlagTime")
	if lastUpdateTime <= 0 {
		return 0
	}

	t := lastUpdateTime + 7*24*60*60
	now := time.Now().Unix()
	cd := now - t
	if cd < 0 {
		cd = 0
	}
	return int(cd)
}

func (p *Player) IsInBattle() bool {
	return p.GetBattleID() > 0
}

func (p *Player) GetLastOnlineTime() int {
	return p.attr.GetInt("lastOnlineTime")
}

func (p *Player) setLastOnlineTime(t int) {
	p.attr.SetInt("lastOnlineTime", t)
}

func (p *Player) setLoginTime(t int64) {
	p.attr.SetInt64("loginTime", t)
}

func (p *Player) getLoginTime() int64 {
	return p.attr.GetInt64("loginTime")
}

func (p *Player) GetLastLoginTime() int64 {
	return p.getLoginTime()
}

func (p *Player) isNetAlive() bool {
	return p.attr.GetBool("isNetAlive")
}

func (p *Player) onNetDisconnect() {
	p.setLastOnlineTime(int(time.Now().Unix()))
	p.attr.SetBool("isNetAlive", false)
}

func (p *Player) onNetConnect() {
	p.attr.SetBool("isNetAlive", true)
}

func (p *Player) GetLastOpponent() common.UUid {
	return common.UUid(p.attr.GetUInt64("lastOpponent"))
}

func (p *Player) GetRankScore() int {
	return p.GetComponent(consts.ResourceCpt).(*ResourceComponent).GetResource(consts.MatchScore)
}

func (p *Player) GetMaxRankScore() int {
	return p.GetComponent(consts.ResourceCpt).(*ResourceComponent).GetResource(consts.MaxMatchScore)
}

func (p *Player) GetFirstHandAmount() int {
	return p.attr.GetInt("firstHandAmount")
}

func (p *Player) GetBackHandAmount() int {
	return p.attr.GetInt("backHandAmount")
}

func (p *Player) GetFirstHandWinAmount() int {
	return p.attr.GetInt("firstHandWinAmount")
}

func (p *Player) GetBackHandWinAmount() int {
	return p.attr.GetInt("backHandWinAmount")
}

// []gcardID
func (p *Player) GetLastFightCards() []*pb.SkinGCard {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetLastFightCards()
}

// []gcardID
func (p *Player) GetFavoriteCards() []*pb.SkinGCard {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetFavoriteCards()
}

func (p *Player) GetCardFromCollectCardById(cid uint32) types.ICollectCard {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetCollectCard(cid)
}

func (p *Player) GetPvpCardPoolsByCamp(camp int) []types.ICollectCard {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetPvpCardPoolByCamp(camp)
}

//统计指定等级卡的数量
func (p *Player) CalcCollectCardNumByLevel(lvl int) int {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetCollectCardNumByLevel(lvl)
}

//统计指定星的卡
func (p *Player) CalcCollectCardNumByStar(star int) int {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetCollectCardNumByStar(star)
}

func (p *Player) setAccountType(accountType pb.AccountTypeEnum) {
	if p.attr.GetInt32("accountType") != int32(accountType) {
		p.attr.SetInt32("accountType", int32(accountType))
	}
}

func (p *Player) GetAccountType() pb.AccountTypeEnum {
	return pb.AccountTypeEnum(p.attr.GetInt32("accountType"))
}

func (p *Player) GetLogAccountType() pb.AccountTypeEnum {
	accountType := pb.AccountTypeEnum(p.attr.GetInt32("accountType"))
	if accountType == pb.AccountTypeEnum_WxgameIos {
		return pb.AccountTypeEnum_Wxgame
	}
	return accountType
}

func (p *Player) setCurGuideGroup(groupID int) {
	p.attr.SetInt("groupID", groupID)
}

func (p *Player) getCurGuideGroup() int {
	return p.attr.GetInt("groupID")
}

func (p *Player) IsWxgameAccount() bool {
	return mod.IsWxgameAccount(p.GetAccountType())
}

func (p *Player) packSimpleMsg() *pb.SimplePlayerInfo {
	var statusIDs []string
	module.OutStatus.ForEachClientStatus(p, func(st types.IOutStatus) {
		if _, ok := st.(types.IBuff); ok || st.GetID() == consts.OtVipCard {
			statusIDs = append(statusIDs, st.GetID())
		}
	})

	return &pb.SimplePlayerInfo{
		Uid:                uint64(p.GetUid()),
		Name:               p.GetName(),
		PvpScore:           int32(p.GetPvpScore()),
		HeadImgUrl:         p.GetHeadImgUrl(),
		FirstHandAmount:    int32(p.GetFirstHandAmount()),
		BackHandAmount:     int32(p.GetBackHandAmount()),
		FirstHandWinAmount: int32(p.GetFirstHandWinAmount()),
		BackHandWinAmount:  int32(p.GetBackHandWinAmount()),
		RankScore:          int32(p.GetRankScore()),
		FavoriteCards:      p.GetFavoriteCards(),
		FightCards:         p.GetLastFightCards(),
		IsWechatFriend:     false,
		IsOnline:           p.IsOnline(),
		IsInBattle:         p.IsInBattle(),
		LastOnlineTime:     int32(p.GetLastOnlineTime()),
		PvpCamp:            int32(p.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()),
		Country:            p.GetCountry(),
		HeadFrame:          p.GetHeadFrame(),
		RebornCnt:          int32(module.Reborn.GetRebornCnt(p)),
		StatusIDs:          statusIDs,
		CrossAreaHonor:     int32(module.Player.GetResource(p, consts.CrossAreaHonor)),
		CountryFlag:        p.GetCountryFlag(),
		Area:               int32(p.GetArea()),
		MaxRankScore:       int32(p.GetMaxRankScore()),
	}
}

func (p *Player) onLogin() bool {
	p.isLogout = false
	mq.AddConsumer(fmt.Sprintf("player:%d", p.GetUid()), module.Service.GetRegion(), &mqConsumer{
		player: p,
	})

	// 兼容老数据
	if p.attr.GetBool("isVip") {
		p.attr.Del("isVip")
		module.OutStatus.AddStatus(p, consts.OtVipCard, -1, false, false)
	}

	return true
}

func (p *Player) onLogout() {
	//glog.Infof("onPlayerKickOut onLogout 11111111111")
	if p.isLogout {
		return
	}
	p.isLogout = true
	mq.RemoveConsumer(fmt.Sprintf("player:%d", p.GetUid()), module.Service.GetRegion())

	for _, cpt := range p.components {
		cpt.OnLogout()
	}
}

func (p *Player) GetMaxPvpLevel() int {
	return p.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpLevel()
}

func (p *Player) GetMaxPvpTeam() int {
	return p.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetMaxPvpTeam()
}

func (p *Player) GetPvpLevel() int {
	return p.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpLevel()
}

func (p *Player) GetPvpTeam() int {
	return p.GetComponent(consts.PvpCpt).(types.IPvpComponent).GetPvpTeam()
}

func (p *Player) IsForbidLogin() bool {
	if p.isForbidAccount() {
		return true
	}
	return p.attr.GetBool("isForbidLogin")
}

//func (p *Player) ForbidLogin(isForbid bool) {
//	uid := p.GetUid()
//	area := p.GetArea()
//	err := utils.ForbidAccount(area, uid, consts.ForbidAccount, isForbid, -1, "")
//	if err == nil {
//		glog.Infof("ForbidLogin uid=%d, area=%d, isForbid=%v", uid, area, isForbid)
//	}
//
//	//p.attr.SetBool("isForbidLogin", isForbid)
//	//if isForbid {
//	//	p.kickOut(Relogin)
//	//}
//	//p.attr.Save(false)
//}

func (p *Player) Forbid(forbidType int, isForbid bool, overTimes int64, msg string, isAuto bool) {
	uid := p.GetUid()
	area := p.GetArea()
	err := utils.ForbidAccount(uid, forbidType, isForbid, overTimes, msg, isAuto)
	if err == nil {
		glog.Infof("Forbid uid=%d, area=%d, forbidType=%d, isForbid=%v", uid, area, forbidType, isForbid)
	}
}

func (p *Player) OnForbidLogin() {
	p.kickOut(Relogin)
	evq.CallLater(func() {
		p.Save(false)
	})
}

func (p *Player) isForbidChat() bool {
	return module.OutStatus.GetForbidStatus(p, consts.ForbidChat) != nil
}

func (p *Player) IsForbidChat() bool {
	if p.isForbidChat() {
		return true
	}
	return p.attr.GetBool("isForbidChat")
}

//func (p *Player) ForbidChat(isForbid bool) {
//	uid := p.GetUid()
//	area := p.GetArea()
//	err := utils.ForbidAccount(uid, consts.ForbidChat, isForbid, -1, "")
//	if err == nil {
//		glog.Infof("forbid chat uid=%d, area=%d, isForbid=%v", uid, area, isForbid)
//	}
//	//p.attr.SetBool("isForbidChat", isForbid)
//	//logic.BroadcastBackend(pb.MessageID_L2CA_FORBID_CHAT, &pb.ForbidChatArg{
//	//	Uid: uint64(p.GetUid()),
//	//	IsForbid: isForbid,
//	//})
//	//p.attr.Save(false)
//}

func (p *Player) GetCountry() string {
	return p.attr.GetStr("country")
}

func (p *Player) setCountry(country string) {
	p.attr.SetStr("country", country)
}

func (p *Player) GetChannelUid() string {
	return p.attr.GetStr("channelID")
}

func (p *Player) setChannelUid(channelUid string) {
	p.attr.SetStr("channelID", channelUid)
}

func (p *Player) IsVip() bool {
	if p.IsForeverVip() {
		return true
	}
	return module.OutStatus.GetStatus(p, consts.OtVipCard) != nil
}

func (p *Player) isForbidAccount() bool {
	return module.OutStatus.GetForbidStatus(p, consts.ForbidAccount) != nil
}

func (p *Player) IsForbidMonitor() bool {
	return module.OutStatus.GetForbidStatus(p, consts.ForbidMonitor) != nil
}

func (p *Player) IsForeverVip() bool {
	st := module.OutStatus.GetStatus(p, consts.OtVipCard)
	return st != nil && st.GetRemainTime() < 0
}

func (p *Player) BuyVip() {
	//p.attr.SetBool("isVip", true)
	module.OutStatus.AddStatus(p, consts.OtVipCard, -1, false, false)
}

func (p *Player) GetSkipAdsNeedJade() int {
	if p.IsVip() && !config.GetConfig().IsMultiLan {
		return 0
	} else {
		return 3
	}
}

func (p *Player) GetHeadFrame() string {
	headFrame := p.attr.GetStr("headFrame")
	if headFrame == "" {
		headFrame = module.Bag.GetDefHeadFrame()
	}
	return headFrame
}

func (p *Player) SetHeadFrame(headFrame string) {
	p.attr.SetStr("headFrame", headFrame)
}

func (p *Player) GetChatPop() string {
	chatPop := p.attr.GetStr("chatPop")
	if chatPop == "" {
		chatPop = module.Bag.GetDefChatPop()
	}
	return chatPop
}

func (p *Player) SetChatPop(chatPop string) {
	p.attr.SetStr("chatPop", chatPop)
}

func (p *Player) GetFriendsNum() int {
	pcm := p.GetComponent(consts.SocialCpt).(types.ISocialComponent)
	return pcm.GetFriendsNum()
}

func (p *Player) isFbAdvertReward() bool {
	return p.attr.GetBool("isFbAdvertReward")
}

func (p *Player) setFbAdvertReward() {
	p.attr.SetBool("isFbAdvertReward", true)
}

func (p *Player) Tellme(msg string, text int) {
	agent := p.GetAgent()
	if agent != nil {
		agent.PushClient(pb.MessageID_S2C_TELL_ME, &pb.TellMe{
			Msg:  msg,
			Text: int32(text),
		})
	}
}

func (p *Player) getCompensateVersion() int {
	return p.attr.GetInt("cpsVer")
}

func (p *Player) setCompensateVersion(version int) {
	p.attr.SetInt("cpsVer", version)
}

func (p *Player) HasBowlder(amount int) bool {
	if amount < 0 {
		return false
	}
	resCpt := p.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	bowlder := resCpt.GetResource(consts.Bowlder)
	if bowlder >= amount {
		return true
	}

	diff := amount - bowlder
	return resCpt.HasResource(consts.Jade, diff)
}

func (p *Player) SubBowlder(amount int, reason string) {
	if amount <= 0 {
		return
	}
	resCpt := p.GetComponent(consts.ResourceCpt).(*ResourceComponent)
	bowlder := resCpt.GetResource(consts.Bowlder)
	if bowlder >= amount {
		resCpt.ModifyResource(consts.Bowlder, -amount, reason)
		return
	}

	resCpt.BatchModifyResource(map[int]int{
		consts.Bowlder: -bowlder,
		consts.Jade:    bowlder - amount,
	}, reason)
}

func (p *Player) setFire233BindAccount(account string) {
	p.attr.SetStr("fire233Account", account)
}

func (p *Player) getFire233BindAccount() string {
	return p.attr.GetStr("fire233Account")
}

func (p *Player) getAccountCode() *accountCodeSt {
	code := p.attr.GetStr("accountCode")
	if code == "" {
		return nil
	}

	acCode := loadAccountCode(code)
	if acCode == nil {
		p.attr.Del("accountCode")
	}

	return acCode
}

func (p *Player) setAccountCode(code string) {
	p.attr.SetStr("accountCode", code)
}

// accountID = channel_channelID_loginChannel
func (p *Player) getAccountID() string {
	return p.attr.GetStr("accountID")
}

func (p *Player) setAccountID(accountID string) {
	p.attr.SetStr("accountID", accountID)
}

func (p *Player) GetDataDayNo() int {
	return p.attr.GetInt("dayno")
}

func (p *Player) setDataDayno(dayno int) {
	p.attr.SetInt("dayno", dayno)
}

func (p *Player) onCrossDay(dayno int) {
	if dayno == p.GetDataDayNo() {
		return
	}

	p.setDataDayno(dayno)
	if p.IsForeverVip() {
		mod.ModifyResource(p, consts.Jade, 10, consts.RmrForeverVip)
	}
}

func (p *Player) GetCreateTime() int64 {
	return int64(p.attr.GetUInt32("createTime"))
}

func (p *Player) GetInitCamp() int {
	return int(p.GetComponent(consts.TutorialCpt).(types.ITutorialComponent).GetCampID())
}

func (p *Player) GetCurCamp() int {
	return p.GetComponent(consts.CardCpt).(types.ICardComponent).GetFightCamp()
}

func (p *Player) AddHint(type_ pb.HintType, count int) {
	if count <= 0 {
		return
	}
	p.hints[type_] = count
}

func (p *Player) UpdateHint(type_ pb.HintType, count int) {
	oldCount := p.hints[type_]
	if oldCount == count {
		return
	}

	if count <= 0 {
		p.DelHint(type_)
		return
	}

	p.hints[type_] = count
	if p.agent != nil {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_HINT, &pb.Hint{
			Type:  type_,
			Count: int32(count),
		})
	}
}

func (p *Player) DelHint(type_ pb.HintType) {
	count, ok := p.hints[type_]
	if !ok {
		return
	}

	delete(p.hints, type_)
	if count > 0 && p.agent != nil {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_HINT, &pb.Hint{
			Type:  type_,
			Count: 0,
		})
	}
}

func (p *Player) SubHint(type_ pb.HintType, num int) {
	count, ok := p.hints[type_]
	if !ok {
		return
	}
	count = count - num
	if count <= 0 {
		p.DelHint(type_)
		return
	}

	p.hints[type_] = count
	if p.agent != nil {
		p.agent.PushClient(pb.MessageID_S2C_UPDATE_HINT, &pb.Hint{
			Type:  type_,
			Count: int32(count),
		})
	}
}

func (p *Player) GetHintCount(type_ pb.HintType) int {
	count, ok := p.hints[type_]
	if !ok {
		return 0
	}
	return int(count)
}

func (p *Player) forEachHint(callback func(type_ pb.HintType, count int)) {
	for t, c := range p.hints {
		callback(t, c)
	}
}

func (p *Player) GetArea() int {
	area := p.attr.GetInt("area")
	if area <= 0 {
		p.attr.SetInt("area", 1)
		area = 1
	}
	return area
}

func (p *Player) adultCertification(isAdult bool) {
	if !isAdult {
		return
	}

	module.OutStatus.DelStatus(p, consts.OtFatigue)
	cpt, ok := p.components[consts.FatigueCpt]
	if ok {
		delete(p.components, consts.FatigueCpt)
		for i, cpt2 := range p.componentList {
			if cpt2 == cpt {
				p.componentList = append(p.componentList[:i], p.componentList[i+1:]...)
			}
		}
	}
}

func (p *Player) getLoginNoticeVersion() int {
	return p.attr.GetInt("lnVer")
}
func (p *Player) setLoginNoticeVersion(version int) {
	p.attr.SetInt("lnVer", version)
}

func (p *Player) getCanShowLoginNotice() *pb.LoginNotice {
	return module.GM.GetCanShowLoginNotice(p.getLoginNoticeVersion(), p.GetChannel())
}

func (p *Player) onLoginNoticeShow(version int) {
	if version > p.getLoginNoticeVersion() {
		p.setLoginNoticeVersion(version)
	}
}

func (p *Player) IsMultiRpcForbid(type_ int) bool {
	return p.multiClickGuard.isForbid(type_)
}
