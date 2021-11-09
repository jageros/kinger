package app

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kinger/gopuppy/common/config"
	"kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/glog"

	"github.com/sevlyar/go-daemon"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/db"
)

type IService interface {
	Start(appid uint16)
	Stop()
}

type Application struct {
	name          string
	id            uint16
	configFile    string
	daemonContext *daemon.Context
	signalChan    chan os.Signal
	svc           IService
}

func NewApplication(name string, svc IService) *Application {
	return &Application{
		name:       name,
		signalChan: make(chan os.Signal, 1),
		svc:        svc,
	}
}

func (a *Application) parseArgs() {
	var appIdArg int
	flag.IntVar(&appIdArg, "id", 0, "set appid")
	flag.StringVar(&a.configFile, "configfile", "", "set config file path")
	flag.Parse()
	a.id = uint16(appIdArg)
}

func (a *Application) daemonize() {
	context := &daemon.Context{}
	child, err := context.Reborn()

	if err != nil {
		panic(err)
	}

	if child != nil {
		glog.Infof("run in daemon mode")
		os.Exit(0)
	} else {
		glog.Infof("a hahahaha")
		a.daemonContext = context
	}
}

func (a *Application) setupSignals() {
	glog.Infof("Setup signals ...")
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(a.signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-a.signalChan
		if sig == syscall.SIGINT || sig == syscall.SIGTERM {
			glog.Infof("on stop %s %d ...", a.name, a.id)
			if a.svc != nil {
				a.svc.Stop()
			}
			os.Exit(0)
		} else {
			glog.Errorf("unexpected signal: %s", sig)
		}
	}
}

func (a *Application) Run() {
	rand.Seed(time.Now().UnixNano())
	a.parseArgs()

	attribute.DbConfigCreator = func(args ...interface{}) db.IDbConfig {
		var isGlobal bool
		var region uint32
		switch len(args) {
		case 0:
		case 1:
			isGlobal, _ = args[0].(bool)
		default:
			isGlobal, _ = args[0].(bool)
			region, _ = args[1].(uint32)
		}

		if isGlobal {
			return config.GetGlobalDbConfig()
		} else if region > 0 {
			return config.GetRegionConfig().GetDbByRegion(region)
		} else {
			return config.GetRegionDbConfig()
		}
	}

	if a.configFile != "" {
		config.SetConfigFile(a.configFile)
	}
	err := config.LoadRegionConfig()
	if err != nil {
		panic(err)
	}

	err = config.LoadRegionConfigFromDB()
	if err != nil {
		panic(err)
	}
	//db.Shutdown()

	err = config.LoadConfigFromDB()
	if err != nil {
		panic(err)
	}
	db.Shutdown()

	cfg := config.GetConfig()
	var appCfg config.IAppConfig
	if a.name == consts.AppGate {
		appCfg = cfg.GetGateConfig(a.id)
	} else if a.name == consts.AppCenter {
		appCfg = cfg.GetCenterConfig(a.id)
	} else {
		appCfg = cfg.GetLogicConfig(a.name, a.id)
	}

	if appCfg == nil {
		panic(fmt.Sprintf("invalid appid %d", a.id))
	}

	glog.SetupGLog(a.name, a.id, cfg.LogLevel)
	regionConfig := config.GetRegionConfig()
	if regionConfig.ES != nil {
		glog.SetEsConfig(regionConfig.ES)
	}
	//a.daemonize()
	a.svc.Start(a.id)

	a.setupSignals()

}
