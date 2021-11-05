package main

import (
	"kinger/common/consts"
	"math/rand"
)

var (
	allAttackPos = []int{consts.LEFT, consts.RIGHT, consts.UP, consts.DOWN}

	normalAttackTargetFinder iAttackTargetFinder = &normalAttackTargetFinderSt{}
	_ iAttackTargetFinder = &arrowAttackTargetFinderSt{}
	_ iAttackTargetFinder = &peerlessAttackTargetFinderSt{}
	_ iAttackTargetFinder = &lightningAttackTargetFinderSt{}
	_ iAttackTargetFinder = &riprapAttackTargetFinderSt{}
	_ iAttackTargetFinder = &breakthroughAttackTargetFinderSt{}
	_ iAttackTargetFinder = &appointAttackTargetFinderSt{}
	_ iAttackTargetFinder = &buffAppointAttackTargetFinderSt{}
	_ iAttackTargetFinder = &aoeAttackTargetFinderSt{}

	normalAttacker iAttacker = &normalAttackerSt{}
	pierceAttacker iAttacker = &pierceAttackerSt{}
	normalDefenser iDefenser = &normalDefenserSt{}
	shieldDefenser iDefenser = &shieldDefenserSt{}
)

type (
	iAttackTargetFinder interface {
		findTarget(situation *battleSituation, attackCxt *attackContext, attackCard *fightCard, attackType int) map[int]*fightCard
		setNext(iAttackTargetFinder) iAttackTargetFinder
		getNext() iAttackTargetFinder
	}

	iAttacker interface {
		attack(atkCard *fightCard, defCard *fightCard, attackCxt *attackContext) (atkPos, defPos, bat int)
	}

	iDefenser interface {
		defense(atkPos int, defCard *fightCard) (defNum int, defPos int)
	}

	attackContext struct {
		attackTargetFinders map[int]iAttackTargetFinder // 怎样找攻击目标
		attackTypes         map[int]int                 // 能否打队友
		attackers           map[int]iAttacker           // 怎样比点
		defensers           map[int]iDefenser           // 怎样被比点
		attackPoses         map[int]int                 // 攻击方向
	}
)

func (cxt *attackContext) getAttackFinder(objID int) iAttackTargetFinder {
	var attackTargetFinder iAttackTargetFinder = nil
	if cxt.attackTargetFinders != nil {
		var ok bool
		attackTargetFinder, ok = cxt.attackTargetFinders[objID]
		if !ok {
			attackTargetFinder = normalAttackTargetFinder
		}
	} else {
		attackTargetFinder = normalAttackTargetFinder
	}
	return attackTargetFinder
}

func (cxt *attackContext) findAttackTarget(situation *battleSituation, attackCard *fightCard) []*fightCard {
	objID := attackCard.getObjID()
	var attackType int
	var ok bool

	if cxt.attackTypes != nil {
		attackType, ok = cxt.attackTypes[objID]
		if !ok {
			attackType = atNormal
		}
	} else {
		attackType = atNormal
	}

	var cards []*fightCard
	beAttackCards := cxt.getAttackFinder(objID).findTarget(situation, cxt, attackCard, attackType)
	for _, c := range beAttackCards {
		cards = append(cards, c)
	}
	return cards
}

func (cxt *attackContext) setAttackTargetFinder(objID int, finder iAttackTargetFinder) {
	if cxt.attackTargetFinders == nil {
		cxt.attackTargetFinders = map[int]iAttackTargetFinder{}
	}
	finder2, ok := cxt.attackTargetFinders[objID]
	if !ok {
		cxt.attackTargetFinders[objID] = finder
	} else {
		finder2.setNext(finder)
	}
}

func (cxt *attackContext) getAttacker(cardObjID int) iAttacker {
	if cxt.attackers != nil {
		attacker, ok := cxt.attackers[cardObjID]
		if !ok {
			return normalAttacker
		} else {
			return attacker
		}
	} else {
		return normalAttacker
	}
}

func (cxt *attackContext) setAttacker(cardObjID int, attacker iAttacker) {
	if cxt.attackers == nil {
		cxt.attackers = map[int]iAttacker{}
	}
	cxt.attackers[cardObjID] = attacker
}

func (cxt *attackContext) getDefenser(cardObjID int) iDefenser {
	if cxt.defensers != nil {
		defenser, ok := cxt.defensers[cardObjID]
		if !ok {
			return normalDefenser
		} else {
			return defenser
		}
	} else {
		return normalDefenser
	}
}

func (cxt *attackContext) setDefenser(cardObjID int, defenser iDefenser) {
	if cxt.defensers == nil {
		cxt.defensers = map[int]iDefenser{}
	}
	cxt.defensers[cardObjID] = defenser
}

func (cxt *attackContext) setAttackType(cardObjID, attackType int) {
	if cxt.attackTypes == nil {
		cxt.attackTypes = map[int]int{}
	}
	cxt.attackTypes[cardObjID] = attackType
}

func (cxt *attackContext) getAttackPos(beAttackObjID int) int {
	if cxt.attackPoses != nil {
		return cxt.attackPoses[beAttackObjID]
	} else {
		return 0
	}
}

func (cxt *attackContext) setAttackPos(beAttackObjID, pos int) {
	if cxt.attackPoses == nil {
		cxt.attackPoses = map[int]int{}
	}
	cxt.attackPoses[beAttackObjID] = pos
}

func (cxt *attackContext) attack(attackCard *fightCard, beAttackCard *fightCard) (atkPos, defPos, bat int) {
	beAttackObjID := beAttackCard.getObjID()
	var pos int
	if cxt.attackPoses != nil {
		pos = cxt.attackPoses[beAttackCard.getObjID()]
	}
	var atkNum int
	var defNum int
	switch pos {
	case consts.UP:
		atkNum = attackCard.getUp()
	case consts.DOWN:
		atkNum = attackCard.getDown()
	case consts.LEFT:
		atkNum = attackCard.getLeft()
	default:
		atkNum = attackCard.getRight()
	}

	var defenser iDefenser
	var ok bool
	if cxt.defensers != nil {
		defenser, ok = cxt.defensers[beAttackObjID]
		if !ok {
			defenser = normalDefenser
		}
	} else {
		defenser = normalDefenser
	}

	defNum, defPos = defenser.defense(pos, beAttackCard)

	if atkNum > defNum {
		bat = bGt
	} else if atkNum < defNum {
		bat = bLt
	} else {
		bat = bEq
	}
	return
}

type normalAttackTargetFinderSt struct {
	next iAttackTargetFinder
}

func (atf *normalAttackTargetFinderSt) setNext(targetFinder iAttackTargetFinder) iAttackTargetFinder {
	if apTargetFinder, ok := targetFinder.(*appointAttackTargetFinderSt); ok {
		apTargetFinder.next = atf
		return apTargetFinder
	} else {
		if atf.next == nil {
			atf.next = targetFinder
		} else {
			atf.next.setNext(targetFinder)
		}
	}
	return atf
}

func (atf *normalAttackTargetFinderSt) getNext() iAttackTargetFinder {
	return atf.next
}

func (atf *normalAttackTargetFinderSt) nextFindTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	var targets map[int]*fightCard
	if atf.next != nil {
		targets = atf.next.findTarget(situation, attackCxt, attackCard, attackType)
		atf.next = nil
	} else {
		targets = map[int]*fightCard{}
	}
	return targets
}

func (atf *normalAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	targetCards := map[int]*fightCard{}
	grid := attackCard.getGrid()

	for _, pos := range allAttackPos {
		t := situation.getPosTargetByGrid(grid, pos, 1)
		if card, ok := t.(*fightCard); ok && atf.canBeTarget(attackCard, card, attackType) {
			attackCxt.setAttackPos(card.getObjID(), pos)
			targetCards[card.getObjID()] = card
		}
	}

	return targetCards
}

func (atf *normalAttackTargetFinderSt) canBeTarget(attackCard *fightCard, targetCard *fightCard, attackType int) bool {
	attackUid := attackCard.getControllerUid()
	targetUid := targetCard.getControllerUid()
	isEnemy := attackUid != targetUid || targetCard.isPublicEnemy()
	if isEnemy {
		if attackType == atScuffle {
			return false
		}
		if targetCard.isInFog() {
			return false
		}
	} else {
		if attackType != atScuffle {
			return false
		}
	}
	return true
}

type arrowAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
	n int
}

func newArrowAttackTargetFinder(n int) *arrowAttackTargetFinderSt {
	if n <= 0 {
		n = 1
	}
	return &arrowAttackTargetFinderSt{n: n}
}

func (atf *arrowAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 可以隔着0到n个空格攻击
	targetCards := atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	grid := attackCard.getGrid()

	for _, pos := range allAttackPos {
		for i := 0; i <= atf.n; i++ {
			t := situation.getPosTargetByGrid(grid, pos, 1 + i)
			if t == nil {
				break
			}

			if card, ok := t.(*fightCard); ok {
				if atf.canBeTarget(attackCard, card, attackType) {
					attackCxt.setAttackPos(card.getObjID(), pos)
					targetCards[card.getObjID()] = card
				}
				break
			}
		}
	}

	return targetCards
}

type peerlessAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
}

func (atf *peerlessAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 对相邻敌军连同其身后的敌军一起攻击
	targetCards := atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	grid := attackCard.getGrid()
	for _, pos := range allAttackPos {
		t := situation.getPosTargetByGrid(grid, pos, 1)
		if t != nil {
			if card, ok := t.(*fightCard); ok {

				if atf.canBeTarget(attackCard, card, attackType) {
					attackCxt.setAttackPos(card.getObjID(), pos)
					targetCards[card.getObjID()] = card

					t = situation.getPosTargetByGrid(grid, pos, 2)
					if card, ok := t.(*fightCard); ok && atf.canBeTarget(attackCard, card, attackType) {
						attackCxt.setAttackPos(card.getObjID(), pos)
						targetCards[card.getObjID()] = card
					}
				}
			} else {
				// 空格
			}
		}
	}

	return targetCards
}

type riprapAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
	n int
}

func newRiprapAttackTargetFinder(n int) *riprapAttackTargetFinderSt {
	if n <= 0 {
		n = 1
	}
	return &riprapAttackTargetFinderSt{n: n}
}

func (atf *riprapAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 只能攻击1到n个格子后面的敌军
	targetCards := atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	grid := attackCard.getGrid()
	for _, pos := range allAttackPos {
		for i := 1; i <= atf.n; i++ {
			t := situation.getPosTargetByGrid(grid, pos, 1 + i)
			if t == nil {
				break
			}

			if card, ok := t.(*fightCard); ok {
				if atf.canBeTarget(attackCard, card, attackType) {
					attackCxt.setAttackPos(card.getObjID(), pos)
					targetCards[card.getObjID()] = card
				}
				break
			}
		}
	}

	return targetCards
}

type lightningAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
	targets []iTarget
}

func (atf *lightningAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 无视位置，随机对场上某个敌军进行攻击，但数字方位必须符合正常攻击
	var targetCards []*fightCard
	grid := attackCard.getGrid()
	attacker := attackCxt.getAttacker(attackCard.getObjID())
	column := situation.getGridColumn()

	if len(atf.targets) <= 0 {
		atf.targets = situation.getGridsTarget()
	}

	for _, t := range atf.targets {
		card, ok := t.(*fightCard)
		if !ok || card.getObjID() == attackCard.getObjID() {
			continue
		}
		cardObjID := card.getObjID()

		if !atf.canBeTarget(attackCard, card, attackType) {
			continue
		}

		var attackPos []int
		if card.getGrid()/column < grid/column {
			attackPos = append(attackPos, consts.UP)
		}
		if card.getGrid()/column > grid/column {
			attackPos = append(attackPos, consts.DOWN)
		}
		if card.getGrid()%column < grid%column {
			attackPos = append(attackPos, consts.LEFT)
		}
		if card.getGrid()%column > grid%column {
			attackPos = append(attackPos, consts.RIGHT)
		}

		for _, pos := range attackPos {
			attackCxt.setAttackPos(cardObjID, pos)
			_, _, bat := attacker.attack(attackCard, card, attackCxt)
			if bat == bGt {
				targetCards = append(targetCards, card)
				break
			}
		}
	}

	if len(targetCards) > 0 {
		card := targetCards[rand.Intn(len(targetCards))]
		return map[int]*fightCard{card.getObjID(): card}
	}
	return map[int]*fightCard{}
}

type breakthroughAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
}

func (atf *breakthroughAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 贯矢：可以无视任意个队友或空格的阻挡对相隔的敌军发起进攻
	targetCards := atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	grid := attackCard.getGrid()
	for _, pos := range allAttackPos {
		for i := 0; true; i++ {
			t := situation.getPosTargetByGrid(grid, pos, 1 + i)
			if t == nil {
				break
			}

			if card, ok := t.(*fightCard); ok {
				if atf.canBeTarget(attackCard, card, attackType) {
					attackCxt.setAttackPos(card.getObjID(), pos)
					targetCards[card.getObjID()] = card
					break
				}
			}
		}
	}

	return targetCards
}

type appointAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
	targets    []iTarget
	attackType int
}

func newAppointAttackTargetFinderSt(targets []iTarget, attackType int) *appointAttackTargetFinderSt {
	targetFinder := &appointAttackTargetFinderSt{
		targets:    targets,
		attackType: attackType,
	}
	return targetFinder
}

func (atf *appointAttackTargetFinderSt) setNext(targetFinder iAttackTargetFinder) iAttackTargetFinder {
	if atf.attackType != 0 {
		if _, ok := targetFinder.(*buffAppointAttackTargetFinderSt); !ok {
			return atf
		}
	}

	if lightningTargetFinder, ok := targetFinder.(*lightningAttackTargetFinderSt); ok {
		lightningTargetFinder.targets = atf.targets
	}
	if atf.next == nil {
		atf.next = targetFinder
	} else {
		atf.next.setNext(targetFinder)
	}
	return atf
}

func (atf *appointAttackTargetFinderSt) findTarget2(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	targetCards := map[int]*fightCard{}
	grid := attackCard.getGrid()
	column := situation.getGridColumn()

	attacker := attackCxt.getAttacker(attackCard.getObjID())

	for _, t := range atf.targets {
		card, ok := t.(*fightCard)
		if !ok || card.getObjID() == attackCard.getObjID() {
			continue
		}

		if !atf.canBeTarget(attackCard, card, attackType) {
			continue
		}

		targetCards[card.getObjID()] = card

		var attackPos []int
		if card.getGrid()/column < grid/column {
			attackPos = append(attackPos, consts.UP)
		}
		if card.getGrid()/column > grid/column {
			attackPos = append(attackPos, consts.DOWN)
		}
		if card.getGrid()%column < grid%column {
			attackPos = append(attackPos, consts.LEFT)
		}
		if card.getGrid()%column > grid%column {
			attackPos = append(attackPos, consts.RIGHT)
		}

		for _, pos := range attackPos {
			attackCxt.setAttackPos(card.getObjID(), pos)
			_, _, bat := attacker.attack(attackCard, card, attackCxt)
			if bat == bGt {
				break
			}
		}
	}

	return targetCards
}

func (atf *appointAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	targets := map[int]*fightCard{}
	var targetCards map[int]*fightCard
	if atf.next == nil {
		if atf.attackType != 0 {
			targetCards = atf.findTarget2(situation, attackCxt, attackCard, attackType)
		} else {
			targetCards = normalAttackTargetFinder.findTarget(situation, attackCxt, attackCard, attackType)
		}
	} else {
		targetCards = atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	}

	for _, t := range atf.targets {
		if card, ok := targetCards[t.getObjID()]; ok {
			targets[card.getObjID()] = card
		}
	}
	return targets
}

type buffAppointAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
	targets []iTarget
}

func newBuffAppointAttackTargetFinderSt(targets []iTarget) *buffAppointAttackTargetFinderSt {
	return &buffAppointAttackTargetFinderSt{
		targets: targets,
	}
}

func (atf *buffAppointAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	targetCards := map[int]*fightCard{}
	for _, t := range atf.targets {
		if card, ok := t.(*fightCard); ok {
			targetCards[card.getObjID()] = card
		}
	}
	return targetCards

}

type aoeAttackTargetFinderSt struct {
	normalAttackTargetFinderSt
}

func (atf *aoeAttackTargetFinderSt) findTarget(situation *battleSituation, attackCxt *attackContext,
	attackCard *fightCard, attackType int) map[int]*fightCard {

	// 阉党?：可以攻击所有敌军
	targetCards := atf.nextFindTarget(situation, attackCxt, attackCard, attackType)
	grid := attackCard.getGrid()
	attacker := attackCxt.getAttacker(attackCard.getObjID())
	column := situation.getGridColumn()
	targets := situation.getGridsTarget()

	for _, t := range targets {
		card, ok := t.(*fightCard)
		if !ok || card.getObjID() == attackCard.getObjID() {
			continue
		}
		cardObjID := card.getObjID()

		if !atf.canBeTarget(attackCard, card, attackType) {
			continue
		}

		var attackPos []int
		if card.getGrid()/column < grid/column {
			attackPos = append(attackPos, consts.UP)
		}
		if card.getGrid()/column > grid/column {
			attackPos = append(attackPos, consts.DOWN)
		}
		if card.getGrid()%column < grid%column {
			attackPos = append(attackPos, consts.LEFT)
		}
		if card.getGrid()%column > grid%column {
			attackPos = append(attackPos, consts.RIGHT)
		}

		for _, pos := range attackPos {
			attackCxt.setAttackPos(cardObjID, pos)
			_, _, bat := attacker.attack(attackCard, card, attackCxt)
			if bat == bGt {
				targetCards[cardObjID] = card
				break
			}
		}
	}

	return targetCards
}

type normalAttackerSt struct {
}

func (at *normalAttackerSt) attack(atkCard *fightCard, defCard *fightCard, attackCxt *attackContext) (atkPos, defPos, bat int) {
	atkPos = attackCxt.getAttackPos(defCard.getObjID())
	var atkNum int
	var defNum int
	switch atkPos {
	case consts.UP:
		atkNum = atkCard.getUp()
	case consts.DOWN:
		atkNum = atkCard.getDown()
	case consts.LEFT:
		atkNum = atkCard.getLeft()
	default:
		atkNum = atkCard.getRight()
	}

	defNum, defPos = attackCxt.getDefenser(defCard.getObjID()).defense(atkPos, defCard)

	if atkNum > defNum {
		bat = bGt
	} else if atkNum < defNum {
		bat = bLt
	} else {
		bat = bEq
	}
	return
}

type pierceAttackerSt struct {
}

func (at *pierceAttackerSt) attack(atkCard *fightCard, defCard *fightCard, attackCxt *attackContext) (atkPos, defPos, bat int) {
	atkPos = attackCxt.getAttackPos(defCard.getObjID())
	defPos = consts.UP
	defNum := defCard.getUp()
	if defCard.getDown() < defNum {
		defNum = defCard.getDown()
		defPos = consts.DOWN
	}
	if defCard.getLeft() < defNum {
		defNum = defCard.getLeft()
		defPos = consts.LEFT
	}
	if defCard.getRight() < defNum {
		defNum = defCard.getRight()
		defPos = consts.RIGHT
	}

	var atkNum int
	switch atkPos {
	case consts.UP:
		atkNum = atkCard.getUp()
	case consts.DOWN:
		atkNum = atkCard.getDown()
	case consts.LEFT:
		atkNum = atkCard.getLeft()
	default:
		atkNum = atkCard.getRight()
	}

	if atkNum > defNum {
		bat = bGt
	} else if atkNum < defNum {
		bat = bLt
	} else {
		bat = bEq
	}
	return
}

type normalDefenserSt struct {
}

func (df *normalDefenserSt) defense(atkPos int, defCard *fightCard) (defNum int, defPos int) {
	switch atkPos {
	case consts.UP:
		defNum = defCard.getDown()
		defPos = consts.DOWN
	case consts.DOWN:
		defNum = defCard.getUp()
		defPos = consts.UP
	case consts.LEFT:
		defNum = defCard.getRight()
		defPos = consts.RIGHT
	default:
		defNum = defCard.getLeft()
		defPos = consts.LEFT
	}
	return
}

type shieldDefenserSt struct {
}

func (at *shieldDefenserSt) defense(atkPos int, defCard *fightCard) (defNum int, defPos int) {
	defPos = consts.UP
	defNum = defCard.getUp() + defCard.getDown() + defCard.getLeft() + defCard.getRight()

	switch atkPos {
	case consts.UP:
		defPos = consts.DOWN
	case consts.DOWN:
		defPos = consts.UP
	case consts.LEFT:
		defPos = consts.RIGHT
	default:
		defPos = consts.LEFT
	}
	return
}
