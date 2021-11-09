package main

import (
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/rpubsub"
	"kinger/gopuppy/network"
)

func sessionOnClose(ev evq.IEvent) {
	ses := ev.(*evq.CommonEvent).GetData()[0].(*network.Session)
	_appName := ses.GetProp("appName")
	_appID := ses.GetProp("appID")
	if _appName == nil || _appID == nil {
		return
	}

	appName := _appName.(string)
	appID := _appID.(uint32)
	if appName == consts.AppCenter {
		return
	}

	if appName == consts.AppGate {
		cService.onGateDisconnect(appID)
	} else if appName == consts.AppGame {
		cService.onGameDisconnect(appID)
	} else {
		cService.onLogicDisconnect(appID, appName)
	}
}

func onConfigUpdate(_ map[string]interface{}) {
	err := config.ReLoadConfig()
	glog.Infof("onConfigUpdate err=%s", err)
}

func handlerEvent() {
	evq.HandleEvent(consts.SESSION_ON_CLOSE_EVENT, sessionOnClose)
	regionCfg := config.GetRegionConfig()
	if regionCfg.Redis != nil {
		rpubsub.Initialize(regionCfg.Redis.Addr)
		rpubsub.Subscribe("reload_gconfig", onConfigUpdate)
	}
}
