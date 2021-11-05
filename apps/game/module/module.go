package module

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/apps/game/module/types"
	"kinger/gamedata"
	"kinger/proto/pb"
)

var (
	Player IPlayerModule
	Level  ILevelModule
	//Battle   IBattleModule
	AI IAiModule
	//Bonus    IBonusModule
	Pvp      IPvpModule
	Campaign ICampaignModule
	Card     ICardModule
	Treasure ITreasureModule
	Tutorial ITutorialModule
	GiftCode IGiftCodeModule
	Social   ISocialModule
	WxGame   IWxGameModule
	Service  IService
	Shop IShopModule
	Mission IMissionModule
	Mail IMailModule
	Huodong IHuodongModule
	Bag IBagModule
	Reborn IRebornModule
	OutStatus IOutStatusModule
	Reward IRewardModule
	Activitys IActivityModule
	GM IGmModule
	Televise ITeleviseModule
	Rank IRankModule
)

type IPlayerModule interface {
	GetPlayer(uid common.UUid) types.IPlayer
	GetSimplePlayerInfo(uid common.UUid) *pb.SimplePlayerInfo
	LoadSimplePlayerInfoAsync(uid common.UUid) chan *pb.SimplePlayerInfo
	GetCachePlayer(uid common.UUid) types.IPlayer
	NewPlayerByAttr(uid common.UUid, attr *attribute.AttrMgr) types.IPlayer
	IsWxgameAccount(accountType pb.AccountTypeEnum) bool

	// 修改资源数量
	//@param args[0]：string，改变资源的原因，看consts/资源从哪里产出或消耗
	//@param args[1]：bool，是否需要同步给客户端，默认true
	ModifyResource(player types.IPlayer, resType int, amount int, args ...interface{})

	GetResource(player types.IPlayer, resType int) int
	LogMission(player types.IPlayer, missionID string, event int)
	HasResource(player types.IPlayer, resType, amount int) bool
	PackUpdateRankMsg(p types.IPlayer, battleHandCards []*pb.SkinGCard, battleCamp int) *pb.UpdatePvpScoreArg
	SetResource(player types.IPlayer, resType, amount int)
	GetResourceName(resType int) string
	ForEachOnlinePlayer(callback func(player types.IPlayer))
}

type ILevelModule interface {
	NewLevelComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	GetCurLevel(player types.IPlayer) int
}

/*
type IBattleModule interface {
	BeginLevelFight(player types.IPlayer, levelData *gamedata.Level, endHandler types.IBattleEndHandler) types.IBattle
	LoadFight(fightID common.UUid, player types.IPlayer) *pb.RestoredFightDesk
	OnFighterLogout(uid common.UUid)
	OnCardUpLevel(uid common.UUid, cardData *gamedata.Card)
	LoadVideo(battleID common.UUid) *pb.VideoBattleData
	IsPlayerInBattle(uid common.UUid) bool

	//@param battleType  pvp, level, campaign, defCampaign
	//@param upperType 先后手 1.f1先手 2.f2先手 3.随机
	BeginBattle(battleType int, fighterData1, fighterData2 types.IFighterData, upperType int, endHandler types.IBattleEndHandler,
		plugin types.IBattlePlugin, scale, battleRes int, needVideo bool, args ...interface{}) (types.IBattle, error)
}
*/

type IAiModule interface {
	NewRobotPlayer(playerAttr *attribute.AttrMgr) types.IRobotPlayer
	NewRobotByUid(uid common.UUid) types.IRobotPlayer
	RandomName(robot types.IRobotPlayer)
	AlphaBetaSearch(situation types.IBattleSituation, depth int, maxDepth int, alpha, beta float32) (float32, *types.BattleAction)
}

//type IBonusModule interface {
//	NewCampaignPlugin() types.IBattlePlugin
//	NewPvpPlugin() types.IBattlePlugin
//}

type IPvpModule interface {
	NewPvpComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	//StopMatch(uid common.UUid)
	//NewEndHandler() types.IBattleEndHandler
	//RefreshRank()
	//GetPvpLevelByStar(star int) int
	//UpdatePlayerRankScore(fighter types.IEndFighterData)
	GetMinStarByPvpLevel(level int) int
	CanPvpMatch(player types.IPlayer) bool
	CancelPvpBattle(player types.IPlayer)
	GM_CrossSeason(agent *logic.PlayerAgent, com string)
}

/*
type IFightModule interface {
	RandomCamp() int
	BeginLevelFight(player types.IPlayer, levelRes *res.Level, endHandler types.IFightEndHanlder) gtypes.IMsgPacker
	BeginCampaignFight(fighterData1, fighterData2 types.IFighterData, upperType int, endHandler types.IFightEndHanlder,
		fightType int, needVideo bool, plugin types.IFightPlugin, arg ...interface{}) types.IFight
	LoadFight(fightID consts.FightDeskId, player types.IPlayer) *pb.RestoredFightDesk
	OnFighterLogout(uid common.UUid)
	OnCardUpLevel(uid common.UUid, cardRes *res.Card)
}
*/

type ICardModule interface {
	NewCardComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	IsDiyCard(cardId uint32) bool
	UpdateCardSkin(player types.IPlayer, cardID uint32, skin string) error
	GetCollectCard(player types.IPlayer, cardID uint32) types.ICollectCard
	GetAllCollectCards(player types.IPlayer) []types.ICollectCard
	GetAllCollectCardsByCamp(player types.IPlayer, camps []int) []types.ICollectCard
	OnCampaignMissionDone(player types.IPlayer, cardIDs []uint32)
	OnAcceptCampaignMission(player types.IPlayer, cards []uint32)
	SetCardsState(player types.IPlayer, cardIDs []uint32, state pb.CardState)
	GetCardAmountLog(accountType pb.AccountTypeEnum, area int) map[pb.AccountTypeEnum]map[uint32]int
	GetCardLevelLog(accountType pb.AccountTypeEnum, cardID uint32, area int) map[pb.AccountTypeEnum]map[uint32]map[int]int
	//GetCardPoolLog(accountType pb.AccountTypeEnum, pvpLevel, area int) (
	//	cardPoolLogs map[pb.AccountTypeEnum]map[int]map[uint32]int, battleAmountLogs map[pb.AccountTypeEnum]map[int]int)
	// 取玩家已解锁的卡，team <= 0 时，team 取 player.GetPvpTeam()
	GetUnlockCards(player types.IPlayer, team int) []uint32
	GetFirstRechargeUnlockCards(player types.IPlayer) []uint32
	GetOnceCards(player types.IPlayer) []types.ICollectCard
	GmAllCardUpLevel(player types.IPlayer)
	LogBattleCards(player types.IPlayer, cardHand *pb.EndFighterData)

	// 批量修改卡
	//@param  cardsChange map[cardID]*pb.CardInfo  要修改的卡
	//@return  卡的改变
	ModifyCollectCards(player types.IPlayer, cardsChange map[uint32]*pb.CardInfo) []*pb.ChangeCardInfo
}

type ICampaignModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	OnBattleEnd(player types.IPlayer, isWin bool)
	GetPlayerInfo(player types.IPlayer) *pb.GCampaignPlayerInfo
	IsInCampaignMatch(player types.IPlayer) bool
	IsInWar() bool
	GetPlayerOfflineInfo(player types.IPlayer) *pb.GCampaignPlayerInfo
	ModifyContribution(player types.IPlayer, amount int) bool
	OnUnifiedReward(player types.IPlayer, rank, contribution int, yourMajestyName, countryName string)
}

type ITreasureModule interface {
	NewTreasureComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	OpenTreasureByModelID(player types.IPlayer, modelID string, isDobule bool) *pb.OpenTreasureReply
	GetDayAccTicketCanAdd(player types.IPlayer) int
	WxHelpDoubleDailyTreasure(player types.IPlayer, treasureID uint32, helperUid common.UUid, helperHeadImg,
		helperHeadFrame, helperName string) bool
}

type ITutorialModule interface {
	NewTutorialComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
}

type IGiftCodeModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
}

type ISocialModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	OnBattleBegin(player types.IPlayer)
	OnLogout(player types.IPlayer)
	AddFriendApply(player types.IPlayer, targetUid common.UUid, isInvite bool) error
	WxInviteFriend(inviterUid common.UUid, targetPlayer types.IPlayer)
	CancelInviteBattle(uid common.UUid)
	IsInInviteBattle(uid common.UUid) bool
}

type IWxGameModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	CancelInviteBattle(uid common.UUid)
	IsInInviteBattle(uid common.UUid) bool
	GetDailyShareState(player types.IPlayer) int
	ReturnDailyShareReward(player types.IPlayer, playerName string)
}

type IShopModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	GetVipRemainTime(player types.IPlayer) int
	LogShopBuyItem(player types.IPlayer, itemID, itemName string, amount int, shopName, resType,
		resName string, resAmount int, msg string)
	GetLimitGiftPrice(goodsID string, player types.IPlayer) int
	GetRecruitCurIDs(player types.IPlayer) []int32
	GM_setRecruitVer(cmd string, player types.IPlayer)
}

type IService interface {
	GetAppID() uint32
	GetRegion() uint32
}

type IMissionModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	OnOpenTreasure(player types.IPlayer)
	OnInviteBattle(player types.IPlayer)
	OnWxShare(player types.IPlayer, shareTime int64)
	OnWatchVideo(player types.IPlayer)
	OnAddFriend(player types.IPlayer)
	OnShareVideo(player types.IPlayer)
	OnAccTreasure(player types.IPlayer)
	OnPvpBattleEnd(player types.IPlayer, fighterData *pb.EndFighterData, isWin bool)
	GmAddMission(player types.IPlayer, args []string) error
	GmCompleteMission(player types.IPlayer)
	RefreashMission(player types.IPlayer)
}

type IMailModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	NewMailRewardByMsg(msg []*pb.MailReward) types.IMailReward
	NewMailSender(uid common.UUid) types.IMailSender
	UpdateMailDeadLine(mailID int, newbieDeadLine int64) bool
}

type IHuodongModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	GetSeasonPvpLimitTime(player types.IPlayer) int32
	GetSeasonPvpHandCardInfo(player types.IPlayer) (seasonData gamedata.ISeasonPvp, camp int, handCardType pb.BattleHandType,
		chooseCardData *pb.SeasonPvpChooseCardData, handCards *pb.FetchSeasonHandCardReply)
	SeasonPvpChooseCamp(player types.IPlayer, camp int) *pb.SeasonPvpChooseCardData
	SeasonPvpChooseCard(player types.IPlayer, cards []uint32) (randCards []uint32, err error)
	RefreshSeasonPvpChooseCard(player types.IPlayer) (*pb.SeasonPvpChooseCardData, error)
	GetSeasonPvpWinCnt(player types.IPlayer) int
	OnRecharge(player types.IPlayer, oldJade, money int) int
	PackEventHuodongs(player types.IPlayer) []*pb.HuodongData
	GetEventHuodongItemType(player types.IPlayer) int
	GetTreasureHuodongSkin(player types.IPlayer, treasureData *gamedata.Treasure) (skins []string, eventItem int)
}

type IBagModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	HasItem(player types.IPlayer, type_ int, itemID string) bool
	AddCardSkin(player types.IPlayer, skinID string) types.IItem
	AddHeadFrame(player types.IPlayer, headFrame string) types.IItem
	GetDefHeadFrame() string
	AddChatPop(player types.IPlayer, headFrame string) types.IItem
	GetDefChatPop() string
	GetAllItemIDsByType(player types.IPlayer, type_ int) []string
	AddEquip(player types.IPlayer, equipID string)
	GetItem(player types.IPlayer, type_ int, itemID string) types.IItem
	AddEmoji(player types.IPlayer, emojiTeam int)
	GetAllItemsByType(player types.IPlayer, type_ int) []types.IItem
}

type IRebornModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	GetRebornCnt(player types.IPlayer) int
	GetRebornRemainDay(player types.IPlayer) int
}

type IOutStatusModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	GetStatus(player types.IPlayer, statusID string) types.IOutStatus
	GetBuff(player types.IPlayer, buffID int) types.IOutStatus
	AddStatus(player types.IPlayer, statusID string, leftTime int, args ...interface{}) types.IOutStatus
	AddBuff(player types.IPlayer, buffID int, leftTime int, args ...interface{}) types.IOutStatus
	ForEachClientStatus(player types.IPlayer, callback func(st types.IOutStatus))
	DelStatus(player types.IPlayer, statusID string)
	//添加vip指定的两个特权
	AddVipBuff(player types.IPlayer)
	//封禁账号操作
	AddForbidStatus(player types.IPlayer, forbidID int, leftTime int, args ...interface{}) types.IOutStatus
	//解封操作
	DelForbidStatus(player types.IPlayer, forbidID int)
	//获取封禁状态
	GetForbidStatus(player types.IPlayer, forbidID int) types.IOutStatus

	// 开箱子加卡
	BuffTreasureCard(player types.IPlayer, amount int) int
	// 开箱子加金币
	BuffTreasureGold(player types.IPlayer, amount int) int
	// 加每天箱子个数
	BuffTreasureCnt(player types.IPlayer, amount int) int
	// 加每天加速劵数量0
	BuffAccTreasureCnt(player types.IPlayer, amount int) int
	// 加每天加速劵数量2
	BuffAccTreasureCntByActivity(player types.IPlayer, amount int) int
	// 每日宝箱加卡
	BuffDayTreasureCard(player types.IPlayer, amount int) int
	// 开箱子减时间
	BuffTreasureTime(player types.IPlayer, amount int) int
	// 加pvp金币
	BuffAddPvpGold(player types.IPlayer, amount int) int
	// 加pvp星
	BuffPvpAddStar(player types.IPlayer, amount int) int
	// pvp不减星
	BuffPvpNoSubStar(player types.IPlayer, amount int) int
	//pvp加一个宝箱
	BuffPvpAddTreasure(player types.IPlayer, amount int) int
	//每日宝箱加金币
	BuffDayTreasureGlod(player types.IPlayer, amount int) int
	//获取角色特权等级
	GetBuffLevel(player types.IPlayer) int
	//VIP每日宝箱翻倍
	BuffDoubleRewardOfVip(player types.IPlayer, amount int) int
	//VIP对战宝箱卡牌+2
	BuffAddCardOfVip(player types.IPlayer, amount int) int
	// 是否拥有所有勋章
	HasAllPriv(player types.IPlayer) bool
}

type IRewardModule interface {
	GiveReward(player types.IPlayer, rewardTbl string) types.IRewardResult
	GiveRewardList(player types.IPlayer, rwList []string, source string) *pb.RewardList
	GetMailItemType(stuff string) (isRes bool, ty pb.MailRewardType)
}

type IActivityModule interface {
	NewComponent(playerAttr *attribute.AttrMgr) types.IPlayerComponent
	OnGetSpringHuodongItemType(player types.IPlayer, itemAmount int) int
	TestARpc(agent *logic.PlayerAgent, com []string) (rsp interface{}, err error)
}

type IGmModule interface {
	GetCanShowLoginNotice(curVersion int, channel string) *pb.LoginNotice
	GetServerStatus() *pb.ServerStatus
}

type ITeleviseModule interface {
	SendNotice(televiseType pb.TeleviseEnum, args ...interface{})
}

type IRankModule interface {
	GetRanking(player types.IPlayer) (uint32, bool)
}