package utils

import "kinger/gopuppy/common/glog"

func CatchPanic(f func()) (err interface{}) {
	defer func() {
		err = recover()
		if err != nil {
			glog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
	return
}

func RunPanicless(f func()) (panicless bool) {
	defer func() {
		err := recover()
		panicless = err == nil
		if err != nil {
			glog.TraceError("%s panic: %s", f, err)
		}
	}()

	f()
	return
}

func RepeatUntilPanicless(f func()) {
	for !RunPanicless(f) {
	}
}
