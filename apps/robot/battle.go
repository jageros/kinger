package main

import (
	"kinger/gopuppy/common/glog"
	"kinger/proto/pb"
	"math/rand"
	"time"
)

const (
	levelID = 36
)

type battle struct {
	r       *robot
	oppHand []int
	hand    []int
	grid    []int
}

func newBattle(r *robot, msg *pb.LevelBattle) *battle {
	b := &battle{
		r:    r,
		grid: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	var hand []int
	for _, c := range msg.Desk.Fighter1.Hand {
		hand = append(hand, int(c.ObjId))
	}
	if msg.Desk.Fighter1.Uid == r.uid {
		b.hand = hand
	} else {
		b.oppHand = hand
	}

	hand = []int{}
	for _, c := range msg.Desk.Fighter2.Hand {
		hand = append(hand, int(c.ObjId))
	}
	if msg.Desk.Fighter2.Uid == r.uid {
		b.hand = hand
	} else {
		b.oppHand = hand
	}

	return b
}

func (b *battle) onPlayCard(boutResult *pb.FightBoutResult) {
	var hand []int
	if boutResult.BoutUid == b.r.uid {
		hand = b.hand
	} else {
		hand = b.oppHand
	}

	index := -1
	for i, objID := range hand {
		if objID == int(boutResult.UseCardObjID) {
			index = i
			break
		}
	}
	if index >= 0 {
		hand = append(hand[:index], hand[index+1:]...)
		if boutResult.BoutUid == b.r.uid {
			b.hand = hand
		} else {
			b.oppHand = hand
		}
	}

	b.grid[boutResult.TargetGridId] = int(boutResult.UseCardObjID)

	time.Sleep(time.Second)
	b.readDone()
}

func (b *battle) playCard() {
	if len(b.hand) <= 0 {
		glog.Errorf("playCard no hand card")
		return
	}

	var boutResult *pb.FightBoutResult
	for {
		var emptyGrid []int
		for i, objID := range b.grid {
			if objID <= 0 {
				emptyGrid = append(emptyGrid, i)
			}
		}
		var cardObjID int
		var gridID int
		if len(emptyGrid) > 0 && len(b.hand) > 0 {
			cardObjID = b.hand[rand.Intn(len(b.hand))]
			gridID = emptyGrid[rand.Intn(len(emptyGrid))]
		}

		t1 := time.Now()
		c := b.r.ses.CallAsync(pb.MessageID_C2S_FIGHT_BOUT_CMD, &pb.FightBoutCmd{
			UseCardObjID: int32(cardObjID),
			TargetGridId: int32(gridID),
		})
		result := <-c
		t2 := time.Now()
		d := t2.Sub(t1)
		glog.Infof("C2S_FIGHT_BOUT_CMD id=%d, time=%s, err=%s", b.r.id, d, result.Err)
		if result.Err == nil {
			boutResult = result.Reply.(*pb.FightBoutResult)
			break
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	b.onPlayCard(boutResult)
}

func (b *battle) readDone() {
	b.r.ses.Push(pb.MessageID_C2S_FIGHT_BOUT_READY_DONE, nil)
}
