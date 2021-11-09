package api

import (
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/network"
	//"kinger/gopuppy/network/snet"
	"kinger/gopuppy/proto/pb"
	"time"
)

type centerSession struct {
	rawSes               *network.Session
	appID                uint16
	appName              string
	region               uint32
	cfg                  *config.CenterConfig
	centerClientDelegate ICenterClientDelegate
}

func newCenterSession(appID uint16, appName string, cfg *config.CenterConfig, centerClientDelegate ICenterClientDelegate) *centerSession {

	return &centerSession{
		appID:                appID,
		appName:              appName,
		region:               cfg.Region,
		cfg:                  cfg,
		centerClientDelegate: centerClientDelegate,
	}
}

func (cs *centerSession) assureConnected(region uint32, isReconnect bool) {
	for cs.rawSes == nil || !cs.rawSes.IsAlive() {
		var ip string
		if cs.cfg.Region == region && cs.cfg.LocalHost != "" {
			ip = cs.cfg.LocalHost
		} else if cs.cfg.PublicHost != "" {
			ip = cs.cfg.PublicHost
		} else {
			ip = cs.cfg.Listen.BindIP
		}

		ses, err := peer.DialTcp(ip, cs.cfg.Listen.Port, nil, nil)
		if err != nil {
			glog.Errorf("dial center error, id=%d, ip=%s, port=%d, err=%s", cs.cfg.ID, ip,
				cs.cfg.Listen.Port, err)
			time.Sleep(time.Second)
			continue
		}

		glog.Infof("center connected, id=%d, ip=%s, port=%d", cs.cfg.ID, ip, cs.cfg.Listen.Port)
		cs.rawSes = ses
		if !cs.registerApp(region, isReconnect) {
			cs.rawSes = nil
			ses.Close()
		}
	}
}

func (cs *centerSession) registerApp(region uint32, isReconnect bool) bool {
	c := cs.rawSes.CallAsync(pb.MessageID_A2C_REGISTER_APP, &pb.AppInfo{
		AppID:       uint32(cs.appID),
		AppName:     cs.appName,
		Region:      region,
		IsReconnect: isReconnect,
	})
	result := <-c
	return result != nil && result.Err == nil
}
