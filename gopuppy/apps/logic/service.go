package logic

import (
	"fmt"
	"kinger/gopuppy/apps/center/api"
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/common/async"
	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/opmon"
	"kinger/gopuppy/db"
	_ "kinger/gopuppy/meta"
	"kinger/gopuppy/proto/pb"
	"net/http"
)

var lService *LogicService

type LogicService struct {
	AppID   uint32
	Region  uint32
	AppName string
}

func (ls *LogicService) OnStart(appID uint16, appName string) {
	lService = ls
	ls.AppID = uint32(appID)
	ls.AppName = appName
	opcfg := config.GetConfig().Opmon
	opmon.Initialize(appName, ls.AppID, opcfg.DumpInterval, opcfg.FalconAgentPort)
	cfg := config.GetConfig().GetLogicConfig(appName, appID)
	ls.Region = cfg.Region
	api.Initialize(appID, appName, ls.Region, nil)
	registerRpc()
	mq.InitClient()

	if cfg.HttpPort > 0 {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.HttpPort), nil)
			if err != nil {
				panic(err)
			}
		}()
	}
}

func (ls *LogicService) OnStop() {
	api.OnAppClose()
	async.WaitClear()
	db.Shutdown()
	evq.Stop()
	glog.Close()
	async.WaitClear()
}

func (ls *LogicService) ReportRpcHandlers() {
	arg := &pb.RpcHandlers{
		AppID:   ls.AppID,
		AppName: ls.AppName,
	}

	for msgID, _ := range noAgentRpcHandlers {
		arg.Handlers = append(arg.Handlers, &pb.RpcHandler{
			MsgID: msgID,
		})
	}

	for msgID, _ := range agentRpcHandlers {
		arg.Handlers = append(arg.Handlers, &pb.RpcHandler{
			MsgID:    msgID,
			IsPlayer: true,
		})
	}

	api.BroadcastCenter(pb.MessageID_L2C_REPORT_RPC, arg)
}

func (ls *LogicService) ReportOnStop() {
	api.CallAllCenter(pb.MessageID_L2C_BEGIN_HOT_FIX, nil)
}
