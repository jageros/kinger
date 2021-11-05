package consts

const (
	AppBattle = "battle"
	AppMatch  = "match"
	AppChat   = "chat"
	AppRank   = "rank"
	AppVideo  = "video"
	AppCampaign = "campaign"
	AppSdk = "sdk"
)

const (
	MaxHandCardAmount = 5
)

const (
	// gamedata
	Level              = "level"
	Exchange           = "exchange"
	Pool               = "pool"
	Bonus              = "bonus"
	Diy                = "diy"
	Duel               = "duel"
	Campaign           = "campaign"
	Skill              = "skill"
	Target             = "target"
	Treasure           = "treasure_config"
	TreasureRewardFake = "reward_fake"
	TreasureDailyFake  = "daily_fake"
	TreasureShare      = "treasure_share"
	Rank               = "rank"
	Tutorial           = "tech_battle"
	Text               = "text"
	City               = "city"
	Road               = "road"
	Robber             = "robber"
	Name1              = "name1"
	Name2              = "name2"
	Name3              = "name3"
	Name4              = "name4"
	Name5              = "name5"
	NameEn1 = "name_en_1"
	NameEn2 = "name_en_2"
	GiftCode            = "exchange_code"
	IosRecharge         = "ios_recharge"
	IosLimitGift        = "ios_limit_gift"
	SoldTreasure        = "sold_treasure"
	SoldGold            = "sold_gold"
	FreeGoldAds         = "free_gold"
	FreeTreasureAds     = "free_treasure"
	FreeGoodTreasureAds = "free_good_treasure"
	AndroidRecharge = "android_recharge"
	AndroidLimitGift = "android_limit_gift"
	FreeJddeAds         = "free_jade"
	Mission             = "quest_config"
	MissionTreasure     = "quest_treasure"
	NewbiePvp           = "newbie_pvp"
	WxInviteReward      = "invite_reward"
	WxRecharge          = "wxgame_recharge"
	WxLimitGift         = "wxgame_limit_gift"
	SeasonPvp           = "season_config"
	HeadFrame           = "headFrame_config"
	CardSkin = "skin_config"
	SeasonReward           = "season_reward"
	RebornSoldCard = "sold_general"
	RebornSoldPriv = "sold_priv"
	RebornSoldSkin = "sold_skin"
	RebornCardCacul = "card_caculation"
	RebornGoldCacul = "gold_caculation"
	RebornTreausre = "reborn_treausre"
	RebornCnt = "reborn_cnt"
	RebornDayCacul = "day_caculation"
	Equip = "item"
	Emoji = "emoji_config"
	RebornSoldEquip = "sold_equip"
	CampaignParam = "parameter"
	AndroidHandjoyLimitGift = "android_limit_gift_handjoy"
	IosHandjoyLimitGift = "ios_limit_gift_lzd_handjoy"
	SoldGoldHandjoy = "sold_gold_handjoy"
	SoldTreasureHandjoy = "sold_treasure_handjoy"
	WarShopCard = "war_shop_card"
	WarShopEquip = "war_shop_equip"
	WarShopSkin = "war_shop_skin"
	WarShopRes = "war_shop_res"
	SeasonPvpHandjoy           = "season_config_handjoy"
	FunctionPrice = "function_price"
	HuodongConfig               = "event_config"
	HuodongReward               = "event_reward"
	LuckyBagReward = "luckyBagReward"
	PieceCard = "card_exchange"
	PieceSkin = "skin_exchange"
	AiMatch = "aiMatch"
	RecruitTreausre = "gotcha_treasure"
	ActivityConfig = "event_config_opt"
	ActivityTime = "event_time"
	ActivityOpenCondition = "event_condition"
	RandomShop = "random_shop_config"
	AreaConfig = "area_config"
	MatchParam = "match_parameter"
	TreasureEvent      = "treasure_event"
	PrivConfig             = "priv_config"
	RecruitTreausreCardRewardTbl = "info_reward_gotcha"
	RecruitTreausreSkinRewardTbl = "info_reward_gotcha2"
	ChatPopConfig = "chatpop_config"
	RankHonorReward = "ranking_honor_reward"
	SoldGoldGift            = "sold_gold_gift"
	WinningRate = "winning_rate"
	RecruitRefreshConfig = "recruit_refresh_config"
	EventFirstRechargeReward = "event_reward_first_recharge"
	League = "league"
	LeagueReward = "league_end_reward"
	LeagueRankReward = "league_rank_reward"
)

const (
	// IPlayerComponent
	CardCpt     = "card"
	LevelCpt    = "level"
	ResourceCpt = "resource"
	CampaignCpt = "campaign"
	PvpCpt      = "pvp"
	TreasureCpt = "treasure"
	TutorialCpt = "tutorial"
	SurveyCpt   = "survey"
	GiftCodeCpt = "giftCode"
	SocialCpt   = "social"
	WxgameCpt   = "wxgame"
	ShopCpt = "shop"
	MissionCpt = "mission"
	MailCpt    = "mail"
	HuodongCpt = "huodong"
	BagCpt = "bag"
	RebornCpt = "reborn"
	OutStatusCpt = "outstatus"
	BuffCpt = "buff"
	ActivityCpt = "activity"
	FatigueCpt = "fatigue"
)

const (
	// res type
	Weap             = 1
	Horse            = 2
	Mat              = 3
	Gold             = 4
	Forage           = 5
	Med              = 6
	Ban              = 7
	Wine             = 8
	Book             = 9
	Mmr              = 10
	Jade             = 11
	Feats            = 12  // 功勋，新版本不再投
	Prestige         = 13  // 名望，新版本不再投
	//ExchangeEquipCnt = 14  // 可以换多少把装备
	Reputation     = 14 // 声望
	Bowlder        = 15 // 玉石
	SkyBook        = 16 // 遁甲天书
	EventItem1     = 17 // 活动物品（鞭炮）
	CardPiece = 18  // 卡碎片
	SkinPiece = 19  // 皮肤碎片
	CrossAreaHonor = 20  // 跨区荣誉勋章
	//CrossAreaBlotHonor = 21  // 跨区耻辱勋章
	Score          = 100
	GuidePro       = 101
	MaxScore       = 102
	AccTreasureCnt = 103 // 宝箱加速还剩多少次
	NotSubStarCnt  = 104 // 不掉星还剩多少次
	PvpTreasureCnt = 105  // 今天战斗宝箱已获得多少个
	PvpGoldCnt     = 106  // 今天战斗金币奖励已获得多少次
	WinDiff        = 107  // 锦标赛净胜场
	MatchScore = 108  // 匹配用的积分
	MaxMatchScore = 109 //本赛季最高分
	KingFlag       = 110 //王者标记
)

const (
	// battle type
	BtPvp                = 1
	BtLevel              = 2
	BtCampaignAtk        = 3
	BtCampaignDef        = 4
	BtTraining           = 5
	BtGuide              = 6
	BtCampaignCity       = 7
	BtCampaignPincerCity = 8
	BtCampaignThief      = 9
	BtCampaignEncounter  = 10
	BtCampaignField      = 11
	BtCampaignDefend     = 12
	BtFriend             = 13
	BtLevelHelp              = 14
	BtCampaign        = 15
)

const (
	SitOne = 1 // 下方座位
	SitTwo = 2 // 上方座位
	SitPubEnemy = 3
)

const (
	BtScale33 = 1
	BtScale43 = 2
	BtScale53 = 3
)

const (
	// camp
	Wei    = 1
	Shu    = 2
	Wu     = 3
	Heroes = 4
)

const (
	// card num pos
	UP    = 1
	DOWN  = 2
	LEFT  = 3
	RIGHT = 4
)

const (
	// 战役类型
	CaWeiShu = 1
	CaWeiWu  = 2
	CaShuWei = 3
	CaShuWu  = 4
	CaWuWei  = 5
	CaWuShu  = 6
)

const (
	TreasureMaxReward = 4
	TreasureMaxDaily  = 1
	TreasureDayAmountLimit = 24
)

const (
	CtGeneral = 1
	CtSoldier = 2
)

const (
	MaxGuidePro = 5
)

const (
	// 物品类型
	ItHeadFrame = iota + 1  // 头像框
	ItCardSkin
	ItEquip
	ItEmoji
	ItChatPop
)

const (
	// 外部状态
	OtVipCard    = "vc"     // vip
	OtMinVipCard = "mvc"
	OtBuffPrefix = "buff_"  // 特权前缀
	OtFatigue = "fatigue"   // 防沉迷
	OtAdvertProtecter = "aniAd"  // 反广告
	OtForbid = "forbid_"    //账号禁用状态前缀
)

const (
	// 物品来自哪里，用于回退
	FromReborn = iota
	FromCampaign
	FromPieceShop
	FromSpringHd
)

const (
	// 重生商店特权
	PrivTreasureCard = 1    // 开启对战宝箱可额外获得2张卡
	PrivTreasureGold = 2    // 开启对战宝箱可额外获得5%金币
	PrivTreasureCnt = 3     // 获得对战宝箱上限额外提高1个
	PrivAccTreasureCnt = 4  // 优惠加速次数额外提高1次
	PrivDayTreasureCard = 5 // 开启每日宝箱可额外获得10张卡
	PrivTreasureTime = 6    // 解锁对战宝箱的所需时间减少5%
	PrivAddPvpGold = 7      // 对战金币，额外获得100%
	PrivPvpAddStar = 8      // 战斗胜利时，有10%几率额外获得一颗星
	PrivPvpNoSubStar = 9    // 战斗失败时，有10%几率不掉星
	PrivPvpAddTreasure  = 10        //对战胜利且有空余宝箱位时有几率额外获得1个宝箱
	PrivDayTreasureGold = 11        //每日宝箱金币提高
	PrivDoubleRewardOfVip = 12      //vip每日宝箱自动双倍
	PrivAddCardOfVip = 13           //vip对战宝箱卡牌+2
	PrivAutoOpenTreasureOfVip = 14  //vip自动开宝箱
)

const (
	//账号禁止状态
	ForbidAccount = 1 //封号
	ForbidChat = 2 //禁言
	ForbidMonitor = 3 //监控
)

const (
	//活动类型
	ActivityUnknow     = 0  //未知活动类型
	ActivityOfLogin    = 1 //登录类活动
	ActivityOfRecharge = 2 //充值类活动
	ActivityOfOnline   = 3 //在线类活动
	ActivityOfFight    = 4 //对战次数类活动
	ActivityOfVictory  = 5 //胜利次数类活动
	ActivityOfRank     = 6 //段位类活动
	ActivityOfConsume  = 7 //消费类活动
	ActivityOfLoginRecharge  = 8  //登录类活动
	ActivityOfFirstRecharge  = 9  //首充活动
	ActivityOfGrowPlan       = 10 //成长基金活动
	ActivityOfDailyRecharge  = 11 //每日累计充值
	ActivityOfDailyShare     = 12 //每日分享
	ActivityOfSpring = 13  // 烟花兑换
)

const (
	//时间限制类型
	TimeToTime  = 1 //时间段
	CreateDurationDay = 2 //建号持续多少天
	DayOfWeek   = 3 //周几
)

const (
	// 资源从哪里产出或消耗

	// 产出
	RmrRecharge             = "recharge"  // 充值
	RmrActivityRewardPrefix = "activityReward_" // 运营活动前缀，完整格式是 activityReward_活动类型_活动id_rewardID
	RmrResetEquip           = "resetEquip" // 回退装备
	RmrCampaignMission      = "caMission"  // 国战政令
	RmrSpringHuodong        = "spring"     // 春节活动
	RmrForeverVip           = "foreverVip" // 永久vip（老服），每天10元宝
	RmrRewardTbl            = "rewardTbl_" // 奖励表获得奖励前缀，完整格式是 rewardTbl_奖励表名
	RmrDailyShareReturn     = "dailyShareReturn"  // 每日分享回礼
	RmrResetCard           = "resetCard" // 回退卡
	RmrBackCardUnlock	= "backCardUnlock" // 回退卡突破
	RmrUpLevelCard      = "upLevelCard"  // 升级卡
	RmrClearLevel     = "clearLevel" // 关卡通关
	RmrMailReward = "mail"  // 邮件奖励
	RmrMission = "mission"  // 日常任务
	RmrBonus = "bonus"  // 战斗红利
	RmrBattleWin = "battleWin" // 战斗胜利
	RmrReborn = "reborn"  // 下野
	RmrWxInviteReward = "wxInvite"  // 微信邀请奖励
	RmrTreasure = "treasure_"  // 开箱子前缀，完整格式是 treasure_箱子id
	RmrDailyShare = "dailyShare"  // 每日分享
	RmrGiftCodeExchange = "giftCodeExchange"  //兑换码兑换
	RmrUnknownOutput = "unknownOutput"  // 未知产出
	RmrLeagueReward = "leagueReward_" //王者联赛奖励前缀 leagueReward_league表ID

	// 消耗
	RmrDeEquip              = "deEquip" // 脱装备
	RmrCityCapitalInjection = "cityCapitalInjection" // 国战城市注资
	RmrCountryModifyName    = "countryModifyName" // 国战国家改名
	RmrCreateCountry        = "createCountry" // 国战国家竞选
	RmrEscapedFromJail      = "escapedFromJail" // 国战越狱
	RmrMoveCity             = "moveCity" // 国战迁移
	RmrUnlockCardLevel      = "unlockCardLevel" // 解锁5级卡
	RmrLuckBag              = "luckBag" // 春节抽福袋
	RmrClearChapter         = "clearChapter" // 元宝通关关卡章节
	RmrAccountTransfer      = "accountTransfer" // 账号转移
	RmrRebornBuyEquip       = "rebornBuyEquip" // 重生商店换装备
	RmrShopAds              = "shopAds" // 商城赞助
	RmrShopBuyGold          = "shopBuyGold" // 商城买金币（消耗宝玉，产出金币）
	RmrRandomShop           = "randomShop" // 买随机商店商品（消耗元宝，可能会产出金币）
	RmrRefreshRandomShop    = "refreshRandomShop" // 刷新随机商店
	RmrRecruitTreasure      = "recruitTreasure" // 买招募宝箱
	RmrLimitGift            = "limitGift" // 商城礼包（可能消耗宝玉，可能产出金币等）
	RmrBuyVip               = "buyVip" // 买vip
	RmrSoldTreasure         = "soldTreasure" // 买军备宝箱
	RmrRefreshSeasonPvp     = "refreshSeasonPvp" // 锦标赛刷卡
	RmrModifyName           = "modifyName" // 改名
	RmrAccTreasure          = "accTreasure" // 加速宝箱
	RmrDailyTreasureDouble  = "dailyTreasureDouble" // 每日宝箱双倍
	RmrTreasureAddCard      = "treasureAddCard" // 宝箱卡牌+2
	RmrTreasureUpRare       = "treasureUpRare" // 宝箱升级
	RmrBuyGoldGift             = "buyGoldGift" // 买金币宝箱
	RmrBuyRecommendGift     = "buyRecommendGift" //买推荐宝箱
	RmrUnknownConsume       = "unknownConsume" //未知消耗
)

const (
	// 防止客户端连点
	FmcMatch = iota
	FmcGuideBattle
	FmcRecruit
)