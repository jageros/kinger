package main

import (
	"kinger/gopuppy/common"
	"kinger/common/consts"
	"kinger/gopuppy/common/utils"
	"math/rand"
	"kinger/gamedata"
	"kinger/proto/pb"
	"sort"
	"strings"
	"strconv"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/evq"
	kutils "kinger/common/utils"
)

var bonusPriorityList bonusLists

type bonusContext struct {
	winUid   common.UUid
	ownerUid common.UUid
	// 翻面次数
	turnCnt int
	// 被翻面次数
	beTurnCnt int
	// 我方多少牌
	mySideCnt int
	// 敌方多少牌
	enemyCnt int
}

type bonusCondition struct {
	left string
	op   string
	num  int
}

func newBonusCondition(data string) *bonusCondition {
	if data == "win" {
		return &bonusCondition{
			left: data,
		}
	}
	if data == "lose" {
		return &bonusCondition{
			left: data,
		}
	}

	op := "=="
	i := strings.Index(data, op)
	if i < 0 {
		op = "<="
		i = strings.Index(data, op)
	}
	if i < 0 {
		op = ">="
		i = strings.Index(data, op)
	}
	if i < 0 {
		op = "<"
		i = strings.Index(data, op)
	}
	if i < 0 {
		op = ">"
		i = strings.Index(data, op)
	}
	if i < 0 {
		return nil
	}

	info := strings.Split(data, op)
	num, err := strconv.Atoi(info[1])
	if err != nil {
		glog.Errorf("newCondition %s", err)
		return nil
	}

	return &bonusCondition{
		left: info[0],
		op:   op,
		num:  num,
	}
}

func (bc *bonusCondition) check(uid common.UUid, context *bonusContext) bool {
	if bc.left == "win" {
		return context.winUid == uid
	} else if bc.left == "lose" {
		return context.winUid > 0 && context.winUid != uid
	} else {
		var leftNum int
		switch bc.left {
		case "turn":
			leftNum = context.turnCnt
		case "beturn":
			leftNum = context.beTurnCnt
		case "front":
			leftNum = context.mySideCnt
		case "opp":
			leftNum = context.enemyCnt
		default:
			return false
		}

		switch bc.op {
		case "==":
			return leftNum == bc.num
		case ">":
			return leftNum > bc.num
		case "<":
			return leftNum < bc.num
		case ">=":
			return leftNum >= bc.num
		case "<=":
			return leftNum <= bc.num
		default:
			return false
		}
	}
}

type bonusRewarder struct {
	resType     int
	amountRange []int
}

func (r *bonusRewarder) randAmount(totalMin, totalMax int) int {
	if len(r.amountRange) != 2 {
		return 0
	}

	min := r.amountRange[0]
	max := r.amountRange[1]
	if min > max {
		min, max = max, min
	}

	if min < totalMin {
		min = totalMin
	}
	if max > totalMax {
		max = totalMax
	}
	if min > max {
		return 0
	}

	n := utils.IntAbs(utils.IntAbs(min)-utils.IntAbs(max)) + 1
	return min + rand.Intn(n)
}

type bonusData struct {
	data           *gamedata.Bonus
	conditions     []*bonusCondition
	resRewards     map[int][]interface{} // map[type][]*resRewarder
	rewardTotalMin int
	rewardTotalMax int
}

func (b *bonusData) trigger(uid common.UUid, context *bonusContext, type_ int) *pb.BonusReward {
	//glog.Infof("bonusData trigger uid=%d, type_")
	for _, c := range b.conditions {
		if !c.check(uid, context) {
			return nil
		}
	}
	return b.reward(uid, context, type_)
}

func (b *bonusData) reward(uid common.UUid, context *bonusContext, type_ int) *pb.BonusReward {

	rewarderList, ok := b.resRewards[type_]
	if !ok {
		return nil
	}
	utils.Shuffle(rewarderList)
	resReward := make(map[int]int)
	min := b.rewardTotalMin
	max := b.rewardTotalMax
	var changeRes []*pb.Resource
	for _, rewarder := range rewarderList {
		_rewarder := rewarder.(*bonusRewarder)
		n := _rewarder.randAmount(min, max)
		if n != 0 {
			resReward[_rewarder.resType] = n
			min -= n
			max -= n
		}
	}
	if b.data.Gold != 0 {
		resReward[consts.Gold] = b.data.Gold
		changeRes = append(changeRes, &pb.Resource{
			Type:   int32(consts.Gold),
			Amount: int32(b.data.Gold),
		})
	}

	evq.CallLater(func() {
		kutils.PlayerMqPublish(uid, pb.RmqType_Bonus, &pb.RmqBonus{
			ChangeRes: changeRes,
		})
	})

	return &pb.BonusReward{
		Uid:     uint64(uid),
		BonusID: int32(b.data.ID),
		Res:     changeRes,
	}
}

type bonusLists []*bonusData

func (bl bonusLists) Len() int {
	return len(bl)
}

func (bl bonusLists) Swap(i, j int) {
	bl[i], bl[j] = bl[j], bl[i]
}

func (bl bonusLists) Less(i, j int) bool {
	return bl[i].data.Priority < bl[j].data.Priority
}

type bonus struct {
	type_ int
	situation *battleSituation
	cxt1 *bonusContext
	cxt2 *bonusContext
}

func newBonus(type_ int, situation *battleSituation) *bonus {
	if type_ == pvpBonus || type_ == campaignBonus {
		return &bonus{
			type_: type_,
			situation: situation,
		}
	}
	return nil
}

func (b *bonus) onTurnOver(beTurner *fightCard) {
	if b.cxt1 == nil {
		b.cxt1 = &bonusContext{ownerUid: b.situation.fighter1.getUid()}
	}
	if b.cxt2 == nil {
		b.cxt2 = &bonusContext{ownerUid: b.situation.fighter2.getUid()}
	}

	if beTurner.getControllerUid() == b.cxt1.ownerUid {
		b.cxt1.turnCnt ++
		b.cxt2.beTurnCnt ++
	} else {
		b.cxt2.turnCnt ++
		b.cxt1.beTurnCnt ++
	}
}

func (b *bonus) boutEnd(winUid common.UUid) ([]*clientAction, bool, bool) {
	if b.cxt1 == nil {
		b.cxt1 = &bonusContext{ownerUid: b.situation.fighter1.getUid()}
	}
	if b.cxt2 == nil {
		b.cxt2 = &bonusContext{ownerUid: b.situation.fighter2.getUid()}
	}

	b.cxt1.winUid = winUid
	b.cxt2.winUid = winUid
	f1Uid := b.cxt1.ownerUid
	for _, objID := range b.situation.grids {
		c := b.situation.getTargetMgr().getTargetCard(objID)
		if c == nil {
			continue
		}
		if c.getControllerUid() == f1Uid {
			b.cxt1.mySideCnt += 1
			b.cxt2.enemyCnt += 1
		} else {
			b.cxt2.mySideCnt += 1
			b.cxt1.enemyCnt += 1
		}
	}

	curPriority1 := -1
	curPriority2 := -1
	var act1 *pb.BonusAct
	var act2 *pb.BonusAct
	var actions []*clientAction
	isWonderful1 := false
	isWonderful2 := false
	for _, bonus := range bonusPriorityList {
		if !(curPriority1 >= 0 && bonus.data.Priority != curPriority1 && act1 != nil) &&
			b.cxt1.ownerUid > 1 {
			curPriority1 = bonus.data.Priority
			rw := bonus.trigger(b.cxt1.ownerUid, b.cxt1, b.type_)

			if rw != nil {
				if act1 == nil {
					act1 = &pb.BonusAct{}
					actions = append(actions, &clientAction{
						actID: pb.ClientAction_Bonus,
						actMsg: act1,
					})
				}
				act1.Rewards = append(act1.Rewards, rw)
				if bonus.data.ID >= 2 && bonus.data.ID <= 9 {
					isWonderful1 = true
				}
			}
		}

		if !(curPriority2 >= 0 && bonus.data.Priority != curPriority2 && act2 != nil) &&
			b.cxt2.ownerUid > 1 {
			curPriority2 = bonus.data.Priority
			rw := bonus.trigger(b.cxt2.ownerUid, b.cxt2, b.type_)
			if rw != nil {
				if act2 == nil {
					act2 = &pb.BonusAct{}
					actions = append(actions, &clientAction{
						actID: pb.ClientAction_Bonus,
						actMsg: act2,
					})
				}
				act2.Rewards = append(act2.Rewards, rw)
				if bonus.data.ID >= 2 && bonus.data.ID <= 9 {
					isWonderful2 = true
				}
			}
		}
	}

	b.cxt1 = nil
	b.cxt2 = nil
	return actions, isWonderful1, isWonderful2
}

func doInitBonus(gdata gamedata.IGameData) {
	bonusGameData := gdata.(*gamedata.BonusGameData)
	var _list bonusLists
	allBonus := bonusGameData.GetAllBonusData()
	for _, data := range allBonus {
		b := &bonusData{
			data: data,
			resRewards: map[int][]interface{}{
				campaignBonus: []interface{}{},
				pvpBonus:      []interface{}{},
			},
		}
		for _, cdata := range data.Function {
			c := newBonusCondition(cdata)
			if c != nil {
				b.conditions = append(b.conditions, c)
			}
		}

		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Weap,
			amountRange: data.Weapon,
		})
		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Horse,
			amountRange: data.Horse,
		})
		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Mat,
			amountRange: data.Material,
		})
		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Forage,
			amountRange: data.Forage,
		})
		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Med,
			amountRange: data.Medicine,
		})
		b.resRewards[campaignBonus] = append(b.resRewards[campaignBonus], &bonusRewarder{
			resType:     consts.Ban,
			amountRange: data.Bandage,
		})
		b.resRewards[pvpBonus] = append(b.resRewards[pvpBonus], &bonusRewarder{
			resType:     consts.Wine,
			amountRange: data.WineDuel,
		})
		b.resRewards[pvpBonus] = append(b.resRewards[pvpBonus], &bonusRewarder{
			resType:     consts.Book,
			amountRange: data.BookDuel,
		})

		totalMin := 0
		totalMax := 0
		if len(data.Total) >= 2 {
			totalMin = data.Total[0]
			totalMax = data.Total[1]
		}
		if totalMin > totalMax {
			totalMin, totalMax = totalMax, totalMin
		}
		b.rewardTotalMax = totalMax
		b.rewardTotalMin = totalMin

		_list = append(_list, b)
	}
	sort.Sort(_list)
	bonusPriorityList = _list
}

func initBonus() {
	gdata := gamedata.GetGameData(consts.Bonus)
	gdata.AddReloadCallback(doInitBonus)
	doInitBonus(gdata)
}
