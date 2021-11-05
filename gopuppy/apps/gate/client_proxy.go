package main

import (
	"fmt"
	"kinger/gopuppy/common"
	"kinger/gopuppy/network"
)

type clientProxy struct {
	clientID    common.UUid
	uid         common.UUid
	ses         *network.Session
	filterProps map[string]string
	isKickout   bool
	beMonitor bool
	monitorPending []func()
}

func newClientProxy(ses *network.Session, clientID common.UUid) *clientProxy {
	ses.SetProp("clientID", clientID)
	return &clientProxy{
		clientID:    clientID,
		ses:         ses,
		filterProps: make(map[string]string),
	}
}

func (cp *clientProxy) String() string {
	return fmt.Sprintf("<clientProxy %d %d>", cp.clientID, cp.uid)
}

func (cp *clientProxy) addMonitorTask(f func()) {
	cp.monitorPending = append(cp.monitorPending, f)
}

func (cp *clientProxy) executeMonitorTask() {
	if cp.beMonitor {
		for _, f := range cp.monitorPending {
			f()
		}
	}
	cp.monitorPending = []func(){}
}


