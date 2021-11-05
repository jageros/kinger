package main

import (
	"fmt"
	//"kinger/gopuppy/common/glog"
	"math/rand"
)

const (
	infinityValue float32 = 1000000000.0
)

type battleAction struct {
	cardObjID int
	gridID    int
	value float32
}

func (ba *battleAction) String() string {
	return fmt.Sprintf("[CardObjID=%d, GridID=%d]", ba.cardObjID, ba.gridID)
}

func (ba *battleAction) getCardObjID() int {
	return ba.cardObjID
}

func (ba *battleAction) getGridID() int {
	return ba.gridID
}

func newBattleAction(cardObjID, gridID int) *battleAction {
	return &battleAction{
		cardObjID: cardObjID,
		gridID:    gridID,
	}
}

type actionQueue struct {
	actions []*battleAction
	idx int
	maxIdx int
	isPositive bool
}

func newActionQueue(positive, negative int) *actionQueue {
	if positive <= 0 && negative <= 0 {
		positive = 1
	}

	q := &actionQueue{idx: -1}
	if positive > 0 {
		q.maxIdx = positive - 1
		q.actions = make([]*battleAction, positive)
		q.isPositive = true
	} else {
		q.maxIdx = negative - 1
		q.actions = make([]*battleAction, negative)
	}
	return q
}

func (q *actionQueue) push(action *battleAction) {
	index := q.idx + 1
	for i := 0; i <= q.idx; i++ {
		if q.isPositive {
			if action.value > q.actions[i].value {
				index = i
				break
			}
		} else {
			if action.value < q.actions[i].value {
				index = i
				break
			}
		}
	}

	if index <= q.maxIdx {
		for j := q.idx; j >= index; j-- {
			newIndex := j + 1
			if newIndex > q.maxIdx {
				continue
			}
			q.actions[newIndex] = q.actions[j]
		}

		q.actions[index] = action

		if q.idx < q.maxIdx {
			q.idx ++
		}
	}
}

func (q *actionQueue) random() *battleAction {
	if q.idx >= 0 {
		return q.actions[ rand.Intn(q.idx + 1) ]
	} else {
		return nil
	}
}

func alphaBetaSearch(situation *battleSituation, depth int, maxDepth int, alpha, beta float32) (float32, *battleAction) {

	if depth <= 0 {
		return situation.evaluateAiValue(), nil
	}
	winUid, _ := situation.checkResult()
	if winUid > 0 {
		return situation.evaluateAiValue(), nil
	}

	var bestAction *battleAction
	allActions := situation.genAllActions()
	var bakSituation *battleSituation
	for _, action := range allActions {
		if bestAction == nil {
			bestAction = action
		}

		bakSituation = situation.copy()
		result, _, errcode := bakSituation.doAction(bakSituation.curBoutFighter, action.cardObjID, action.gridID)
		if errcode != 0 {
			continue
		}

		bakSituation.boutEnd()
		if result.WinUid == 0 {
			bakSituation.boutBegin()
		}

		//glog.Infof("act=%s value1=%f", action, bakSituation.evaluateAiValue())

		_value, _ := alphaBetaSearch(bakSituation, depth-1, maxDepth, -beta, -alpha)
		_value = -_value

		if _value >= beta {
			return _value, action
		}

		if _value > alpha {
			alpha = _value
			if depth == maxDepth {
				bestAction = action
			}
		} else if _value == alpha && rand.Int()%2 == 0 {
			alpha = _value
			if depth == maxDepth {
				bestAction = action
			}
		}

	}

	if bestAction == nil {
		bestAction = &battleAction{}
	}

	return alpha, bestAction
}


func searchBattleAction(situation *battleSituation, positive, negative int) *battleAction {

	var bestAction *battleAction
	allActions := situation.genAllActions()
	var bakSituation *battleSituation
	myUid := situation.getCurBoutFighter().getUid()
	actQueue := newActionQueue(positive, negative)

	for _, action := range allActions {
		if bestAction == nil {
			bestAction = action
		}

		bakSituation = situation.copy()
		result, _, errcode := bakSituation.doAction(bakSituation.curBoutFighter, action.cardObjID, action.gridID)
		if errcode != 0 {
			continue
		}

		bakSituation.boutEnd()
		if result.WinUid == 0 {
			bakSituation.boutBegin()
		}

		value := bakSituation.evaluateAiValue()
		if myUid == bakSituation.getCurBoutFighter().getUid() {
			action.value = value
		} else {
			action.value = - value
		}
		actQueue.push(action)
	}

	bestAction = actQueue.random()
	if bestAction == nil {
		bestAction = &battleAction{}
	}

	return bestAction
}
