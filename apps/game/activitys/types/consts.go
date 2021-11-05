package types

const (
	Version     = "version"
	CloseStatus = "closeStatus"

	//登录活动的数据键
	LoginActivity    = "loginActivity"
	ContinueLoginDay = "continueLoginDay"
	TotalLoginDay    = "totalLoginDay"
	LoginRewardBool  = "loginRewardBool"
	SpringActivity    = "springAct"

	//充值活动数据键
	RechargeActivity    = "rechargeActivity"
	TotalRechargeAmount = "totalRechargeAmount"
	RechargeRewardBool  = "rechargeRewardBool"

	//在线活动数据键
	OnlineActivity  = "onlineActivity"
	LastReceiveTime = "lastReceiveTime"

	//对战次数活动数据键
	FightActivity    = "fightActivity"
	TotalFightAmount = "totalFightAmount"
	FightRewardBool  = "fightRewardBool"

	//对战胜利次数活动数据键
	WinActivity    = "winActivity"
	TotalWinAmount = "totalWinAmount"
	WinRewardBool  = "winRewardBool"
	Camp           = "camp_"

	//段位提升活动数据键
	RankActivity   = "rankActivity"
	RankRewardBool = "rankRewardBool"

	//消费活动数据键
	ConsumeActivity    = "consumeActivity"
	TotalConsumeAmount = "totalConsumeAmount"
	ConsumeRewardBool  = "consumeRewardBool"

	//登录充值活动的数据键
	LoginRechargeActivity   = "loginRechargeActivity"
	TotalLoginRechargeDay   = "totalLoginRechargeDay"
	TotalLoginRechargeNum   = "totalLoginRechargeNum"
	LoginRechargeRewardBool = "loginRechargeRewardBool"

	//首充活动数据键
	FirstRechargeActivity     = "firstRechargeActivity"
	FirstRechargeAmount       = "firstRechargeAmount"
	FirstRechargeHasGoodsBool = "firstRechargeHasGoodsBool"
	FirstRechargeRewardBool   = "firstRechargeRewardBool"

	// 成长基金活动
	GrowPlanActivity              = "growPlanActivity"
	GrowPlan_hasBuyGift           = "hasBuyGift"
	GrowPlan_rewardRechiceBool    = "rewardRechiceBool"
	GrowPlan_campCardWin          = "campCardWinNum"
	GrowPlan_TreasureOpenNum      = "treasureOpenNum"
	GrowPlan_jadeConsume          = "jadeConsume"
	GrowPlan_combat               = "combat"
	GrowPlan_watchBattleReportNum = "watchBattleReportNum"
	GrowPlan_sendBattleReportNum  = "sendBattleReportNum"
	GrowPlan_finshMissionNum      = "finshMissionNum"
	GrowPlan_hitOutStarCardNum    = "hitOutStarCardNum"
	GrowPlan_hitOutCampCardNum    = "hitOutCampCardNum"
	GrowPlan_useCampWinNum        = "useCampWinNum"
	GrowPlan_useCampBattleNum     = "useCampBattleNum"
	GrowPlan_continuousWinNum     = "continuousWinNum"
	GrowPlan_totalRecharge        = "totalRecharge"

	//每日充值活动
	DailyRechargeActivity   = "dailyRechargeActivity"
	TodayRechargeAmount     = "todayRechargeAmount"
	DailyRechargeReceiveBool = "dailyRechargeReceiveBool"

	//每日分享活动
	DailyShareActivity    = "dailyShareActivity"
	TodayShareAmount      = "todayShareAmount"
	DailyShareReceiveBool = "DailyShareReceiveBool"
)

const (
	//错误码
	GetPlayerError           = 10000 //获取用户失败
	GetArgError              = 10001 //获取参数失败
	GetPlayComponentError    = 10002 //获取用户数据失败
	GetActivityError         = 10003 //获取活动失败
	GetRewardError           = 10004 //获取奖励失败
	CanNotReceiveRewardError = 10005 //不可获取奖励
	GetTimeConditionError    = 10006 //获取时间限制错误
	RewardArgError           = 10007 //奖励参数出错
	NotConformCondition      = 10008 //不符合开启条件
	GetGameDataError         = 10010 //拉取数据失败
)

const (
	//glog.JsonInfo事件
	ActivityOnStart    = 1 //活动开始
	ActivityOnFinsh    = 2 //档位完成
	ActivityHasReceive = 3 //档位领取
)
