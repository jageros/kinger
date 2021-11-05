package types

import (
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
)

type ISimplePlayer interface {
	GetUid() common.UUid
	GetName() string
	GetPvpScore() int
	IsOnline() bool
	IsInBattle() bool
	GetHeadImgUrl() string
	GetLastOnlineTime() int
	GetRankScore() int
	GetMaxRankScore() int
	GetFirstHandAmount() int
	GetBackHandAmount() int
	GetFirstHandWinAmount() int
	GetBackHandWinAmount() int
	GetCountry() string
	GetLastFightCards() []*pb.SkinGCard
	GetFavoriteCards() []*pb.SkinGCard
	GetPvpCardPoolsByCamp(camp int) []ICollectCard
	//统计指定等级卡的数量
	CalcCollectCardNumByLevel(lvl int) int
	//统计指定星的卡
	CalcCollectCardNumByStar(star int) int
	//获取卡池中指定id的卡
	GetCardFromCollectCardById(cid uint32) ICollectCard

	IsVip() bool
	GetHeadFrame() string
	GetLastLoginTime() int64
	GetFriendsNum() int
	GetCountryFlag() string
}

type IPlayer interface {
	ISimplePlayer
	Save(needReply bool)
	GetAgent() *logic.PlayerAgent
	GetComponent(componentID string) IPlayerComponent
	IsRobot() bool
	//OnBeginBattle(battleID common.UUid)
	//OnBattleEnd(battleID common.UUid, battleType int, winner, loser IEndFighterData)
	GetLastBattleID() common.UUid
	GetBattleID() common.UUid
	GetBattleType() int
	GetBattleAppID() uint32
	GetLastOpponent() common.UUid
	GetAccountType() pb.AccountTypeEnum
	GetLogAccountType() pb.AccountTypeEnum
	IsWxgameAccount() bool
	GetChannel() string
	GetLoginChannel() string
	GetServerID() string
	GetPvpLevel() int
	GetMaxPvpLevel() int
	GetMaxPvpTeam() int
	IsForbidLogin() bool
	IsForbidChat() bool
	Forbid(forbidType int, isForbid bool, overTimes int64, msg string, isAuto bool)
	//ForbidChat(isForbid bool)
	GetChannelUid() string
	GetPvpTeam() int
	GetSkipAdsNeedJade() int
	Tellme(msg string, text int)
	IsForeverVip() bool
	BuyVip()
	SetHeadFrame(headFrame string)
	SetChatPop(chatPop string)
	GetChatPop() string
	HasBowlder(amount int) bool
	SubBowlder(amount int, reason string)
	GetInitCamp() int
	GetCurCamp() int
	GetDataDayNo() int
	GetCreateTime() int64
	// 添加客户端红点，不会同步给客户端
	AddHint(type_ pb.HintType, count int)
	// 改变客户端红点数量，会同步给客户端
	UpdateHint(type_ pb.HintType, count int)
	// 删除客户端红点，会同步给客户端
	DelHint(type_ pb.HintType)
	// 红点数量减1，同步给客户端
	SubHint(type_ pb.HintType, num int)
	//返回当前红点数量
	GetHintCount(type_ pb.HintType) int
	// 玩家所在区
	GetArea() int
	OnSimpleInfoUpdate()
	OnForbidLogin()
	GetIP() string
	SetIP(ipAddr string)
	GetSubChannel() string
	// 防止客户端连点
	IsMultiRpcForbid(type_ int) bool
}

type IPlayerComponent interface {
	ComponentID() string
	GetPlayer() IPlayer
	OnInit(IPlayer)
	OnLogin(isRelogin, isRestore bool)
	OnLogout()
}

type ICrossDayComponent interface {
	OnCrossDay(dayno int)
}

type IHeartBeatComponent interface {
	OnHeartBeat()
}

type IPlayerReloginComponent interface {
	OnRelogin()
}

type IResourceComponent interface {
	IPlayerComponent
	HasResource(resType int, amount int) bool

	// 修改资源数量
	//@param args[0]：string，改变资源的原因，看consts/资源从哪里产出或消耗
	//@param args[1]：bool，是否需要同步给客户端，默认true
	ModifyResource(resType int, amount int, args ...interface{})

	//@param modify map[resType]modifyAmount
	//@param args[0]：改变资源的原因，看consts/资源从哪里产出或消耗
	BatchModifyResource(modify map[int]int, args ...string) []*pb.ChangeResInfo
	GetResource(resType int) int
	SetResource(resType int, amount int)
}
