package main

import (
	"kinger/gopuppy/network"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/network/protoc"
	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"kinger/gopuppy/common/evq"
)

type logicSession struct {
	*network.Session
	isAlive bool
	msgQueue *xnsyncutil.SyncQueue
}

func newLogicSession(ses *network.Session) *logicSession {
	return &logicSession{
		Session: ses,
		msgQueue: xnsyncutil.NewSyncQueue(),
	}
}

func (ls *logicSession) onReconnect(ses *network.Session) {
	ls.Session = ses
}

func (ls *logicSession) beginHotFix() {
	ls.isAlive = false
	ls.msgQueue = xnsyncutil.NewSyncQueue()
}

func (ls *logicSession) onRestored() {
	appName, _ := ls.GetProp("appName").(string)
	appID, _ := ls.GetProp("appID").(uint32)

	if ls.msgQueue == nil {
		ls.isAlive = true
		glog.Infof("logicSession onRestored ok 1111 %s, %d", appName, appID)
		return
	}

	for {
		f, ok := ls.msgQueue.TryPop()
		if !ok {
			ls.isAlive = true
			ls.msgQueue.Close()
			ls.msgQueue = nil
			glog.Infof("logicSession onRestored ok 22222 %s, %d", appName, appID)
			return
		}

		callback, ok := f.(func())
		if !ok {
			ls.isAlive = true
			ls.msgQueue.Close()
			ls.msgQueue = nil
			glog.Infof("logicSession onRestored ok 33333 %s, %d", appName, appID)
			return
		}

		callback()
	}
}

func (ls *logicSession) CallAsync(msgID protoc.IMessageID, arg interface{}) chan *network.RpcResult {
	if ls.isAlive {
		return ls.Session.CallAsync(msgID, arg)
	}

	c := make(chan *network.RpcResult, 1)
	if ls.msgQueue != nil {
		ls.msgQueue.Push(func() {
			rpcChan := ls.Session.CallAsync(msgID, arg)
			go func() {
				result := <- rpcChan
				c <- result
			}()
		})

		appName, _ := ls.GetProp("appName").(string)
		appID, _ := ls.GetProp("appID").(uint32)
		glog.Infof("logicSession offline CallAsync %s %d %s", appName, appID, msgID)
	} else {
		c <- &network.RpcResult{Err: network.InternalErr}
	}
	return c
}

func (ls *logicSession) Call(msgID protoc.IMessageID, arg interface{}) (reply interface{}, err error) {
	c := ls.CallAsync(msgID, arg)
	var result *network.RpcResult
	evq.Await(func() {
		result = <-c
	})
	reply = result.Reply
	err = result.Err
	return
}

func (ls *logicSession) Push(msgID protoc.IMessageID, arg interface{}) {
	if ls.isAlive {
		ls.Session.Push(msgID, arg)
		return
	}

	if ls.msgQueue != nil {
		ls.msgQueue.Push(func() {
			ls.Session.Write(protoc.GetPushPacket(msgID, arg))
		})

		appName, _ := ls.GetProp("appName").(string)
		appID, _ := ls.GetProp("appID").(uint32)
		glog.Infof("logicSession offline Push %s %d %s", appName, appID, msgID)
	}
}

type logicSessionMgr struct {
	appName string
	gChooseIdx int
	allSessions []*logicSession
	region2ChooseIdx map[uint32]int
	region2Session map[uint32][]*logicSession
}

func newLogicSessionMgr(appName string) *logicSessionMgr {
	return &logicSessionMgr{
		appName: appName,
		region2ChooseIdx: map[uint32]int{},
		region2Session: map[uint32][]*logicSession{},
	}
}

func (lsm *logicSessionMgr) onAppRegister(appID, region uint32, ses *network.Session, isReconnect bool) {
	var logicSes *logicSession
	for i, lses := range lsm.allSessions {
		id := lses.GetProp("appID")
		if id == nil || id.(uint32) == appID {
			logicSes = lses
			glog.Warnf("repeat %s %d", lsm.appName, appID)
			lses.SetProp("appName", nil)
			lses.SetProp("appID", nil)
			lses.SetProp("region", nil)
			lses.Close()
			lsm.allSessions = append(lsm.allSessions[:i], lsm.allSessions[i+1:]...)

			regionSessions := lsm.region2Session[region]
			for i, lses2 := range regionSessions {
				if lses == lses2 {
					lsm.region2Session[region] = append(regionSessions[:i], regionSessions[i+1:]...)
					break
				}
			}

			break
		}
	}

	if logicSes == nil {
		logicSes = newLogicSession(ses)
	} else {
		logicSes.onReconnect(ses)
	}

	ses.SetProp("appName", lsm.appName)
	ses.SetProp("appID", appID)
	ses.SetProp("region", region)
	lsm.allSessions = append(lsm.allSessions, logicSes)
	regionSessions := lsm.region2Session[region]
	lsm.region2Session[region] = append(regionSessions, logicSes)

	if isReconnect {
		logicSes.onRestored()
	}
}

func (lsm *logicSessionMgr) delSessionFromList(appID uint32, sessions []*logicSession) ([]*logicSession, uint32) {
	for i, lses := range sessions {
		id := lses.GetProp("appID").(uint32)
		if id == appID {
			name, _ := lses.GetProp("appName").(string)
			cfg := config.GetConfig().GetLogicConfig(name, uint16(appID))
			if cfg != nil && !lses.isAlive {
				return sessions, 0
			}
			return append(sessions[:i], sessions[i+1:]...), lses.GetProp("region").(uint32)
		}
	}
	return sessions, 0
}

func (lsm *logicSessionMgr) onLogicDisconnect(appID uint32) {
	var region uint32
	lsm.allSessions, region = lsm.delSessionFromList(appID, lsm.allSessions)

	if region > 0 {
		lsm.region2Session[region], _ = lsm.delSessionFromList(appID, lsm.region2Session[region])
	}
}

func (lsm *logicSessionMgr) getAppSession(appID uint32) *logicSession {
	for _, s := range lsm.allSessions {
		appID2 := s.GetProp("appID").(uint32)
		if appID2 == appID {
			return s
		}
	}
	return nil
}

func (lsm *logicSessionMgr) getAppSessions() []*logicSession {
	return lsm.allSessions
}

func (lsm *logicSessionMgr) chooseApp(region uint32) *logicSession {
	if region <= 0 {
		ses := lsm.allSessions[lsm.gChooseIdx % len(lsm.allSessions)]
		lsm.gChooseIdx += 1
		return ses
	}

	ls, ok := lsm.region2Session[region]
	if !ok {
		if lsm.appName != consts.AppGame {
			return lsm.chooseApp(0)
		}
		return nil
	}

	ses := ls[lsm.region2ChooseIdx[region] % len(ls)]
	lsm.region2ChooseIdx[region] += 1
	return ses
}
