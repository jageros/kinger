package outstatus

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/proto/pb"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

type iBuff interface {
	iClientOutStatus

	GetBuffID() int
	// 开箱子加卡
	buffTreasureCard(oldAmount float64, buffEffect []int) (newAmount float64)
	// 开箱子加金币
	buffTreasureGold(oldAmount float64, buffEffect []int) (newAmount float64)
	// 加每天箱子个数
	buffTreasureCnt(oldAmount float64, buffEffect []int) (newAmount float64)
	// 加每天加速劵数量(加0)
	buffAccTreasureCnt(oldAmount float64, buffEffect []int) (newAmount float64)
	// 加每天加速劵数量(加2)
	buffAccTreasureCntByActivity(oldAmount float64, buffEffect []int) float64
	// 每日宝箱加卡
	buffDayTreasureCard(oldAmount float64, buffEffect []int) (newAmount float64)
	// 开箱子减时间
	buffTreasureTime(oldAmount float64, buffEffect []int) (newAmount float64)
	// 加pvp金币
	buffAddPvpGold(oldAmount float64, buffEffect []int) (newAmount float64)
	// 加pvp星
	buffPvpAddStar(oldAmount float64, buffEffect []int) (newAmount float64)
	// pvp不减星
	buffPvpNoSubStar(oldAmount float64, buffEffect []int) (newAmount float64)
	// pvp加一个宝箱
	buffPvpAddTreasure(oldAmount float64, buffEffect []int) (newAmount float64)
	//每日宝箱加金币
	buffDayTreasureGlod(oldAmount float64, buffEffect []int) (newAmount float64)
	//vip每日宝箱自动双倍特权
	buffDoubleRewardOfVip(oldAmount float64, buffEffect []int) float64
	//vip对战宝箱卡牌+2特权
	buffAddCardOfVip(oldAmount float64, buffEffect []int) float64
}

type baseBuff struct {
	clientStatus
	buffID int
}

func newBuff(statusID string) iOutStatus {
	if !strings.HasPrefix(statusID, consts.OtBuffPrefix) {
		return nil
	}
	info := strings.Split(statusID, "_")
	if len(info) < 2 {
		return nil
	}

	buffID, _ := strconv.Atoi(info[1])
	switch buffID {
	case consts.PrivTreasureCard:
		return &treasureCardBuff{}
	case consts.PrivTreasureGold:
		return &treasureGoldBuff{}
	case consts.PrivTreasureCnt:
		return &treasureCntBuff{}
	case consts.PrivAccTreasureCnt:
		return &accTreasureCntBuff{}
	case consts.PrivDayTreasureCard:
		return &dayTreasureCardBuff{}
	case consts.PrivTreasureTime:
		return &treasureTimeBuff{}
	case consts.PrivAddPvpGold:
		return &addPvpGoldBuff{}
	case consts.PrivPvpAddStar:
		return &pvpAddStarBuff{}
	case consts.PrivPvpNoSubStar:
		return &pvpNoSubStarBuff{}
	case consts.PrivPvpAddTreasure:
		return &pvpAddTreasureBuff{}
	case consts.PrivDayTreasureGold:
		return &dayTreasureGlodBuff{}
	case consts.PrivDoubleRewardOfVip:
		return &VipDoubleRewardBuff{}
	case consts.PrivAddCardOfVip:
		return &VipAddCardBuff{}
	case consts.PrivAutoOpenTreasureOfVip:
		return &VipAutoOpenTreasure{}
	default:
		return nil
	}
}

func (b *baseBuff) GetBuffID() int {
	if b.buffID <= 0 {
		info := strings.Split(b.GetID(), "_")
		if len(info) < 2 {
			return 0
		}
		b.buffID, _ = strconv.Atoi(info[1])
	}
	return b.buffID
}

func (b *baseBuff) buffTreasureCard(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffTreasureGold(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffTreasureCnt(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffAccTreasureCnt(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffAccTreasureCntByActivity(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffDayTreasureCard(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffTreasureTime(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffAddPvpGold(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffPvpAddStar(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffPvpNoSubStar(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffPvpAddTreasure(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffDayTreasureGlod(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffDoubleRewardOfVip(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

func (b *baseBuff) buffAddCardOfVip(oldAmount float64, buffEffect []int) float64 {
	return oldAmount
}

type treasureCardBuff struct {
	baseBuff
}

func (b *treasureCardBuff) buffTreasureCard(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type treasureGoldBuff struct {
	baseBuff
}

func (b *treasureGoldBuff) buffTreasureGold(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type treasureCntBuff struct {
	baseBuff
}

func (b *treasureCntBuff) onLogin() {
	buffID := b.GetBuffID()
	st := module.OutStatus.GetBuff(b.player, buffID)
	if st == nil {
		return
	}
	remainTime := float64(b.GetRemainTime())
	jadeNum := math.Ceil(remainTime/86400.0) * 10
	buffIDStr := fmt.Sprintf("%s%d", consts.OtBuffPrefix, b.buffID)
	b.player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).delStatus(buffIDStr)
	sender := module.Mail.NewMailSender(common.UUid(b.player.GetUid()))
	sender.SetTypeAndArgs(pb.MailTypeEnum_BackPrivReimburse, buffID)
	mailReward := sender.GetRewardObj()
	mailReward.AddAmountByType(pb.MailRewardType_MrtJade, int(jadeNum))
	sender.Send()
}

func (b *treasureCntBuff) buffTreasureCnt(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type accTreasureCntBuff struct {
	baseBuff
}

func (b *accTreasureCntBuff) onAdd(args ...interface{}) {
	b.baseBuff.onAdd(args...)
	evq.CallLater(func() {
		resCpt := b.player.GetComponent(consts.ResourceCpt).(types.IResourceComponent)
		resCpt.SetResource(consts.AccTreasureCnt, module.OutStatus.BuffAccTreasureCnt(b.player, resCpt.GetResource(
			consts.AccTreasureCnt)))
	})
}

func (b *accTreasureCntBuff) buffAccTreasureCnt(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

func (b *accTreasureCntBuff) buffAccTreasureCntByActivity(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type dayTreasureCardBuff struct {
	baseBuff
}

func (b *dayTreasureCardBuff) buffDayTreasureCard(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type treasureTimeBuff struct {
	baseBuff
}

func (b *treasureTimeBuff) buffTreasureTime(oldAmount float64, buffEffect []int) float64 {
	if rand.Intn(100) < buffEffect[0] {
		if buffEffect[2] == 1 {
			return oldAmount - float64(buffEffect[1])
		} else {
			return math.Floor(oldAmount * (1 - float64(buffEffect[1])/float64(buffEffect[2])))
		}
	}
	return oldAmount
}

type addPvpGoldBuff struct {
	baseBuff
}

func (b *addPvpGoldBuff) onLogin() {
	buffID := b.GetBuffID()
	st := module.OutStatus.GetBuff(b.player, buffID)
	if st == nil {
		return
	}
	remainTime := float64(b.GetRemainTime())
	jadeNum := math.Ceil(remainTime/86400.0) * 10
	buffIDStr := fmt.Sprintf("%s%d", consts.OtBuffPrefix, b.buffID)
	b.player.GetComponent(consts.OutStatusCpt).(*outstatusComponent).delStatus(buffIDStr)

	sender := module.Mail.NewMailSender(common.UUid(b.player.GetUid()))
	sender.SetTypeAndArgs(pb.MailTypeEnum_BackPrivReimburse, buffID)
	mailReward := sender.GetRewardObj()
	mailReward.AddAmountByType(pb.MailRewardType_MrtJade, int(jadeNum))
	sender.Send()
}

func (b *addPvpGoldBuff) buffAddPvpGold(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type pvpAddStarBuff struct {
	baseBuff
}

func (b *pvpAddStarBuff) buffPvpAddStar(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type pvpNoSubStarBuff struct {
	baseBuff
}

func (b *pvpNoSubStarBuff) buffPvpNoSubStar(oldAmount float64, buffEffect []int) float64 {
	if rand.Intn(100) < buffEffect[0] {
		return 0
	}
	return oldAmount
}

type pvpAddTreasureBuff struct {
	baseBuff
}

func (b *pvpAddTreasureBuff) buffPvpAddTreasure(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type dayTreasureGlodBuff struct {
	baseBuff
}

func (b *dayTreasureGlodBuff) buffDayTreasureGlod(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type VipDoubleRewardBuff struct {
	baseBuff
}

func (b *VipDoubleRewardBuff) buffDoubleRewardOfVip(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type VipAddCardBuff struct {
	baseBuff
}

func (b *VipAddCardBuff) buffAddCardOfVip(oldAmount float64, buffEffect []int) float64 {
	return calculateNewAmount(oldAmount, buffEffect)
}

type VipAutoOpenTreasure struct {
	baseBuff
}

func calculateNewAmount(oldAmount float64, buffEffect []int) float64 {
	if rand.Intn(100) < buffEffect[0] {
		if buffEffect[2] == 1 {
			return oldAmount + float64(buffEffect[1])
		} else {
			return math.Ceil(oldAmount * (float64(buffEffect[1])/float64(buffEffect[2]) + 1))
		}
	}
	return oldAmount
}
