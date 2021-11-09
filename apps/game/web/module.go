package web

import (
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/config"
	"kinger/common/consts"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/network"
	"kinger/proto/pb"
)

var mod *gmModule

type gmModule struct {
	maxNoticeVersion   int
	channel2NoticeAttr map[string]*attribute.AttrMgr
	channel2Notice     map[string]*pb.LoginNotice

	serverStatusAttr *attribute.AttrMgr
	serverStatus     *pb.ServerStatus
}

func newGmModule() *gmModule {
	m := &gmModule{
		channel2Notice:     map[string]*pb.LoginNotice{},
		channel2NoticeAttr: map[string]*attribute.AttrMgr{},
		serverStatusAttr:   attribute.NewAttrMgr("server_status", 1, true),
		serverStatus:       &pb.ServerStatus{},
	}

	err := m.serverStatusAttr.Load()
	if err == nil {
		m.serverStatus = &pb.ServerStatus{
			//Status: pb.ServerStatus_StatusEnum(m.serverStatusAttr.GetInt("status")),
			Message: m.serverStatusAttr.GetStr("msg"),
		}
	}

	if config.GetConfig().IsServerMaintain {
		m.serverStatus.Status = pb.ServerStatus_Maintain
	}

	channelNoticeAttrs, _ := attribute.LoadAll("login_notice", true)
	for _, channelNoticeAttr := range channelNoticeAttrs {
		channel := channelNoticeAttr.GetAttrID().(string)
		m.channel2NoticeAttr[channel] = channelNoticeAttr
		version := channelNoticeAttr.GetInt("version")
		if version > m.maxNoticeVersion {
			m.maxNoticeVersion = version
		}

		notice := &pb.LoginNotice{
			Version: int32(version),
		}
		m.channel2Notice[channel] = notice

		noticesAttr := channelNoticeAttr.GetListAttr("notices")
		noticesAttr.ForEachIndex(func(index int) bool {
			noticeAttr := noticesAttr.GetMapAttr(index)
			notice.Notices = append(notice.Notices, &pb.LoginNoticeData{
				Title:   noticeAttr.GetStr("title"),
				Content: noticeAttr.GetStr("content"),
			})
			return true
		})
	}

	return m
}

func (m *gmModule) GetCanShowLoginNotice(curVersion int, channel string) *pb.LoginNotice {
	othChannelNotice, othChannelNoticeExist := m.channel2Notice["__oth__"]

	if notice, ok := m.channel2Notice[channel]; ok && int(notice.Version) > curVersion {
		if !othChannelNoticeExist || notice.Version >= othChannelNotice.Version {
			return notice
		}
	}

	if othChannelNoticeExist && int(othChannelNotice.Version) > curVersion {
		return othChannelNotice
	}
	return nil
}

func (m *gmModule) GetServerStatus() *pb.ServerStatus {
	return m.serverStatus
}

func (m *gmModule) updateServerMaintainMessage(message string) {
	if m.serverStatus == nil {
		m.serverStatus = &pb.ServerStatus{}
	}
	m.serverStatus.Message = message

	if module.Service.GetAppID() == 1 {
		m.serverStatusAttr.SetStr("msg", message)
		m.serverStatusAttr.Save(false)
		logic.BroadcastBackend(pb.MessageID_G2G_ON_SERVER_STATUS_UPDATE, m.serverStatus)
	}
}

func (m *gmModule) onServerStatusUpdate(status pb.ServerStatus_StatusEnum) {
	needNotifyClient := status == pb.ServerStatus_Maintain && (m.serverStatus == nil ||
		m.serverStatus.Status == pb.ServerStatus_Normal)
	if m.serverStatus == nil {
		m.serverStatus = &pb.ServerStatus{}
	}
	m.serverStatus.Status = status

	if needNotifyClient {
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			agent := player.GetAgent()
			if agent != nil {
				agent.PushClient(pb.MessageID_S2C_ON_SERVER_STATUS_UPDATE, m.serverStatus)
			}
		})
	}
}

func (m *gmModule) genLoginNoticeVersion() int {
	m.maxNoticeVersion += 1
	return m.maxNoticeVersion
}

func (m *gmModule) onLoginNoticeUpdate(channel string, noticeMsg *pb.LoginNotice) {
	if channel == "" {
		channel = "__oth__"
	}

	if int(noticeMsg.Version) > m.maxNoticeVersion {
		m.maxNoticeVersion = int(noticeMsg.Version)
	}
	m.channel2Notice[channel] = noticeMsg

	if module.Service.GetAppID() != 1 {
		return
	}

	noticeAttr, ok := m.channel2NoticeAttr[channel]
	if !ok {
		noticeAttr = attribute.NewAttrMgr("login_notice", channel, true)
		m.channel2NoticeAttr[channel] = noticeAttr
	}

	noticeAttr.SetInt("version", int(noticeMsg.Version))
	noticeDatasAttr := attribute.NewListAttr()
	noticeAttr.SetListAttr("notices", noticeDatasAttr)
	for _, data := range noticeMsg.Notices {
		dataAttr := attribute.NewMapAttr()
		dataAttr.SetStr("title", data.Title)
		dataAttr.SetStr("content", data.Content)
		noticeDatasAttr.AppendMapAttr(dataAttr)
	}
	noticeAttr.Save(false)

	logic.BroadcastBackend(pb.MessageID_G2G_ON_LOGIN_NOTICE_UPDATE, &pb.GmLoginNotice{
		Channel: channel,
		Notice:  noticeMsg,
	})
}

func (m *gmModule) forEachNotice(callback func(channel string, notice *pb.LoginNotice)) {
	othChannelNotice, othChannelNoticeExist := m.channel2Notice["__oth__"]
	for channel, notice := range m.channel2Notice {
		if channel == "__oth__" {
			callback(channel, notice)
		} else if !othChannelNoticeExist || notice.Version > othChannelNotice.Version {
			callback(channel, notice)
		}
	}
}

func rpc_G2G_OnServerStatusUpdate(_ *network.Session, arg interface{}) (interface{}, error) {
	if module.Service.GetAppID() == 1 {
		return nil, nil
	}

	arg2 := arg.(*pb.ServerStatus)
	mod.updateServerMaintainMessage(arg2.Message)
	return nil, nil
}

func rpc_G2G_OnLoginNoticeUpdate(_ *network.Session, arg interface{}) (interface{}, error) {
	if module.Service.GetAppID() == 1 {
		return nil, nil
	}
	arg2 := arg.(*pb.GmLoginNotice)
	mod.onLoginNoticeUpdate(arg2.Channel, arg2.Notice)
	return nil, nil
}

func onReloadConfig(ev evq.IEvent) {
	status := pb.ServerStatus_Normal
	if config.GetConfig().IsServerMaintain {
		status = pb.ServerStatus_Maintain
	}
	mod.onServerStatusUpdate(status)
}

func Initialize() {
	mod = newGmModule()
	module.GM = mod
	initializeGmtool()
	//initializeSdk()

	evq.HandleEvent(consts.EvReloadConfig, onReloadConfig)
	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_SERVER_STATUS_UPDATE, rpc_G2G_OnServerStatusUpdate)
	logic.RegisterRpcHandler(pb.MessageID_G2G_ON_LOGIN_NOTICE_UPDATE, rpc_G2G_OnLoginNoticeUpdate)
}
