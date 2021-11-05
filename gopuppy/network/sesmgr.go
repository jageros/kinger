package network

import "sync"

type sessionMgr struct {
	guard    sync.RWMutex
	sesIDAcc int64
	sesMap   map[int64]*Session
}

func newSessionMgr() *sessionMgr {
	return &sessionMgr{
		sesMap: make(map[int64]*Session),
	}
}

func (sm *sessionMgr) addSession(ses *Session) {
	sm.guard.Lock()
	sm.sesIDAcc++
	sm.sesMap[sm.sesIDAcc] = ses
	ses.setSesID(sm.sesIDAcc)
	sm.guard.Unlock()
}

func (sm *sessionMgr) removeSession(ses *Session) {
	sm.guard.Lock()
	delete(sm.sesMap, ses.GetSesID())
	sm.guard.Unlock()
}

func (sm *sessionMgr) getSession(sesID int64) *Session {
	sm.guard.RLock()
	ses := sm.sesMap[sesID]
	sm.guard.RUnlock()
	return ses
}

func (sm *sessionMgr) closeAllSession() {
	sm.guard.Lock()
	var cs []chan struct{}
	for _, ses := range sm.sesMap {
		cs = append(cs, ses.Close())
	}
	sm.guard.Unlock()

	for _, c := range cs {
		<-c
	}
}
