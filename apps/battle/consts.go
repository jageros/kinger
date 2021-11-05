package main

const (
	battleAttrVersion = 44
)

const (
	attrGrid = iota
	attrCard
	attrFort
	attrFighter
)

const (
	campaignBonus = 1
	pvpBonus = 2
)

const (
	// skill target type
	stOwner       = 1  // 技能属主
	stEnterDesk   = 2  // 进场的牌
	stInDesk      = 3  // 在场的牌
	stHand        = 4  // 手牌
	stPlayer      = 5  // 玩家
	stEmptyGrid   = 6  // 空格
	stGrid        = 7  // 格子
	stDrawCard    = 8  //  补的牌
	stDisCard     = 9  // 弃的牌
	stMoveCard    = 10 // 移动的牌
	stDestroyCard = 11 // 消灭的牌
	stPreMoveCard = 12 // 准备移动的牌
	stFort = 13   // 城防
	stReturnCard    = 14  //  回手的牌
)

const (
	// 技能触发时机
	awaysTrigger            = 0   // 总会触发
	boutBeginTrigger        = 1   // 回合开始
	enterBattleTrigger      = 2   // 进场
	turnTrigger              = 3  // 翻面
	beTurnTrigger            = 4  // 被翻面
	afterAttackTrigger       = 5  // 进攻比点后
	attackTrigger            = 6  // 进攻比点前
	beAttackTrigger          = 7  // 被进攻比点
	findAttackTargetTrigger  = 8  // 寻找进攻目标
	tryPlayCardTrigger       = 9  // 尝试放置卡牌
	preBeTurnTrigger         = 10 // 被翻面前
	preBeAttackTrigger       = 11 // 被进攻比点前
	boutEndTrigger           = 12 // 回合结束
	preGameEndTrigger        = 13 // 游戏结束前
	afterMoveTrigger         = 14 // 移动后
	beforeDrawCardTrigger    = 15 // 补牌前
	afterBoutEndTrigger      = 16 // 回合结束后
	afterBoutBeginTrigger    = 17 // 回合开始后
	afterDestroyTrigger      = 18 // 消灭卡牌后
	loseSkillTrigger         = 19 // 失去技能后
	afterDrawCardTrigger     = 20 // 补牌后
	beforeMoveTrigger        = 21 // 移动前
	afterReturnTrigger       = 22 // 回手后
	afterCleanDestroyTrigger = 23 //  根除卡牌后
	preAddSkillTrigger       = 24 // 获得技能前
	preReturnTrigger         = 25 // 回手前
	preEnterBattleTrigger    = 26 // 进场前
	surrenderTrigger         = 27 // 投降时
	preAddValueTrigger       = 28 // 加点前
	preSubValueTrigger       = 29 // 减点前
	afterAddValueTrigger    = 30  // 加点后
	afterSubValueTrigger    = 31  // 减点后
)

const (
	// battle state
	bsCreate = 1 + iota
	bsReady
	bsWaitSeason
	bsInBout
	bsWaitClient
	bsBoutTimeout
	bsEnd
)

const (
	atNormal  = 1 // 只能攻击敌人
	atScuffle = 2 // 可以攻击队友
)

const (
	// 比点结果
	bWin  = 1 // 赢
	bLose = 2 // 输
	bEq   = 3 // 等于
	bGt   = 4 // 大于
	bLt   = 5 // 小于
)

const (
	sOwn               = 1 // 我方
	sEnemy             = 2 // 敌方

	sInitOwn1               = 3 // 我方
	sInitEnemy1             = 4 // 敌方
	sInitOwn           = 5 // 初始我方
	sInitEnemy         = 6 // 初始敌方
	sRelativeInitOwn   = 7 // 相对者的初始我方
	sRelativeInitEnemy = 8 // 相对者的初始敌方
)

const (
	tTurner   = 1 // 翻面者
	tBeTurner = 2 // 被翻面者
	tPreTurner = 3 // 翻面前者
	tPreBeTurner = 4 // 被翻面前者
	tCantBeTurner = 5  // 不能被翻者
)

const (
	atAttacker   = 1 // 攻击者
	atBeAttacker = 2 // 被攻击者
	atForceAttacker = 3  // 出手的攻击者
	atForceBeAttacker = 4  // 被出手者
)

const (
	// pos type  位置
	pAdjoin      = 1 // 相邻
	pAll         = 2 // 所有
	pApartEmpty  = 3 // 隔着空格
	pApart       = 4 // 隔着格子
	pOwn         = 5
	pApart2Empty = 3 // 隔着2空格
	pApart2      = 4 // 隔着2格子
	pApart3Empty = 3 // 隔着2空格
	pApart3      = 4 // 隔着2格子
	pNotAdjoin = 10  // 非相邻
)

const (
	mustTurn = 1 // 强制翻面
	cantTurn = 2 // 阻止翻面
)

const (
	boutTimeOut  = 35 + 1
	boutTimeOut2 = 21 + 1
)

const (
	surrenderor   = 1 // 投降者
	beSurrenderor = 2 // 被投降者
)

const (
	// 改变牌点数
	mvtAll = 0        // 所有pos都改
	mvtMinPos = 1     // 只改点数最小的pos
	mvtAllBecome = 2  // 所有pos成为最大值
)
