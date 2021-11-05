package eventhub

import "kinger/gopuppy/common/utils"

var (
	maxSeq    = 0
	listeners = make(map[int][]*listener)
)

type listener struct {
	seq     int
	handler func(args ...interface{})
}

func Subscribe(eventID int, handler func(args ...interface{})) (seq int) {
	maxSeq++
	seq = maxSeq
	ln := &listener{
		seq:     seq,
		handler: handler,
	}

	ls, ok := listeners[eventID]
	if !ok {
		ls = []*listener{}
	}
	listeners[eventID] = append(ls, ln)

	return seq
}

func Publish(eventID int, args ...interface{}) {
	ls, ok := listeners[eventID]
	if !ok {
		return
	}

	for _, l := range ls {
		utils.CatchPanic(func() {
			l.handler(args...)
		})
	}
}

func Unsubscribe(eventID int, seq int) {
	ls, ok := listeners[eventID]
	if !ok {
		return
	}

	index := -1
	for i, l := range ls {
		if l.seq == seq {
			index = i
			break
		}
	}

	if index >= 0 {
		listeners[eventID] = append(ls[:index], ls[index+1:]...)
	}
}
