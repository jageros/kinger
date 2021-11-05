package main

import (
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
)

func sessionOnAccept(ev evq.IEvent) {
	ses := ev.(*evq.CommonEvent).GetData()[0].(*network.Session)
	glog.Debugf("sessionOnAccept %s", ses)
	if ses.FromPeer() != peer {
		return
	}

	glog.Debugf("sessionOnAccept newClientProxy %s", ses)
	gService.newClientProxy(ses)
}

func sessionOnClose(ev evq.IEvent) {
	ses := ev.(*evq.CommonEvent).GetData()[0].(*network.Session)
	if ses.FromPeer() != peer {
		return
	}

	gService.onClientProxyClose(ses)
}

func snetOnDisconnect(ev evq.IEvent) {
	ses := ev.(*evq.CommonEvent).GetData()[0].(*network.Session)
	if ses.FromPeer() != peer {
		return
	}

	gService.onSnetEvent(ses, consts.SNET_ON_DISCONNECT)
}

func snetOnReconnect(ev evq.IEvent) {
	ses := ev.(*evq.CommonEvent).GetData()[0].(*network.Session)
	if ses.FromPeer() != peer {
		return
	}

	gService.onSnetEvent(ses, consts.SNET_ON_RECONNECT)
}

func handlerEvent() {
	evq.HandleEvent(consts.SESSION_ON_ACCEPT_EVENT, sessionOnAccept)
	evq.HandleEvent(consts.SESSION_ON_CLOSE_EVENT, sessionOnClose)
	evq.HandleEvent(consts.SNET_ON_DISCONNECT, snetOnDisconnect)
	evq.HandleEvent(consts.SNET_ON_RECONNECT, snetOnReconnect)
}
