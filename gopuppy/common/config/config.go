package config

import (
	"github.com/BurntSushi/toml"
	"kinger/gopuppy/attribute"
	"fmt"
	"math/rand"
)

var (
	configFile = "gopuppy.toml"
	regionConfigFile = "region.toml"
	cfg        *GoPuppyConfig
	regionCfg *RegionConfig
)

type (
	GoPuppyConfig struct {
		ReadTimeout   int
		WriteTimeout  int
		MaxPacketSize uint32
		LogLevel      string
		Region int
		MonitorUrl string
		Opmon         *OpmonConfig
		Gates         []*GateConfig
		Centers       []*CenterConfig
		Logics        []*LogicConfig

		id2Gate    map[uint16]*GateConfig
		id2Center  map[uint16]*CenterConfig
		id2Logic   map[uint16]*LogicConfig
		name2Logic map[string][]*LogicConfig
		AllRegions []uint32
	}

	OpmonConfig struct {
		DumpInterval    int // s
		FalconAgentPort int
	}

	DBConfig struct {
		Type     string
		Region uint32
		IsGlobal bool
		LocalHost string
		PublicHost string
		Port int
		Addr     string
		DB       string
		User     string
		Password string
	}

	RedisConfig struct {
		Region int
		Addr string
	}

	RedisClusterConfig struct {
		Nodes []string
	}

	ListenConfig struct {
		Network  string
		BindIP   string
		Port     int
		Certfile string
		Keyfile  string
	}

	baseConfig struct {
		ID       uint16
		HttpPort int
		Region uint32
	}

	IAppConfig interface {
		GetAppID() uint16
	}

	GateConfig struct {
		baseConfig
		Host string
		Listens []*ListenConfig
	}

	CenterConfig struct {
		baseConfig
		LocalHost string
		PublicHost string
		Listen *ListenConfig
	}

	LogicConfig struct {
		baseConfig
		Host string
		Name string
	}

	EsConfig struct {
		Url string
		Index string
	}

	RegionConfig struct {
		Region uint32
		Databases []*DBConfig `toml:"databases"`
		DB        *DBConfig
		GlobalDB        *DBConfig
		Redis *RedisConfig
		ES            *EsConfig           `toml:"es"`
	}
)

func SetConfigFile(file string) {
	configFile = file
}

func LoadConfig() error {
	cfg = &GoPuppyConfig{}
	_, err := toml.DecodeFile(configFile, cfg)
	if err != nil {
		return err
	}
	return nil
}

func doLoadConfigFromDB(isSync bool) (*GoPuppyConfig, error) {
	attr := attribute.NewAttrMgr("config", "gopuppy", true)
	err := attr.Load(isSync)
	if err != nil {
		return nil, err
	}

	newCfg := &GoPuppyConfig{}
	cfgData := attr.GetStr("data")
	_, err = toml.Decode(cfgData, newCfg)
	newCfg.id2Gate = make(map[uint16]*GateConfig)
	newCfg.id2Center = make(map[uint16]*CenterConfig)
	newCfg.id2Logic = make(map[uint16]*LogicConfig)
	newCfg.name2Logic = make(map[string][]*LogicConfig)
	newCfg.AllRegions = []uint32{}

	for _, gcfg := range newCfg.Gates {
		newCfg.id2Gate[gcfg.ID] = gcfg
		newCfg.addRegion(gcfg.Region)
	}

	for _, ccfg := range newCfg.Centers {
		newCfg.id2Center[ccfg.ID] = ccfg
		newCfg.addRegion(ccfg.Region)
	}

	for _, lcfg := range newCfg.Logics {
		newCfg.id2Logic[lcfg.ID] = lcfg
		ls := newCfg.name2Logic[lcfg.Name]
		newCfg.name2Logic[lcfg.Name] = append(ls, lcfg)
		newCfg.addRegion(lcfg.Region)
	}

	return newCfg, err
}

func LoadConfigFromDB() error {
	newCfg, err := doLoadConfigFromDB(true)
	cfg = newCfg
	return err
}

func ReLoadConfig() error {
	newCfg, err := doLoadConfigFromDB(false)
	if err != nil {
		return err
	}
	cfg = newCfg
	return nil
}

func LoadRegionConfig() error {
	regionCfg = &RegionConfig{}
	_, err := toml.DecodeFile(regionConfigFile, regionCfg)
	if err != nil {
		return err
	}
	regionCfg.init()
	return nil
}

func LoadRegionConfigFromDB() error {
	attr := attribute.NewAttrMgr("config", "region")
	err := attr.Load(true)
	if err != nil {
		return err
	}

	regionCfg = &RegionConfig{}
	cfgData := attr.GetStr("data")
	_, err = toml.Decode(cfgData, regionCfg)
	regionCfg.init()
	return err
}

func GetConfig() *GoPuppyConfig {
	return cfg
}

func (c *baseConfig) GetAppID() uint16 {
	return c.ID
}

func (c *GoPuppyConfig) addRegion(region uint32) {
	for _, r := range c.AllRegions {
		if r == region {
			return
		}
	}
	c.AllRegions = append(c.AllRegions, region)
}

func (c *GoPuppyConfig) GetGateConfig(id uint16) *GateConfig {
	return c.id2Gate[id]
}

func (c *GoPuppyConfig) GetCenterConfig(id uint16) *CenterConfig {
	return c.id2Center[id]
}

func (c *GoPuppyConfig) GetLogicConfig(name string, id uint16) *LogicConfig {
	cfgs := c.GetLogicConfigsByName(name)
	for _, cfg := range cfgs {
		if cfg.ID == id {
			return cfg
		}
	}
	return nil
}

func (c *GoPuppyConfig) GetLogicConfigsByName(name string) []*LogicConfig {
	return c.name2Logic[name]
}

func (c *DBConfig) String() string {
	return fmt.Sprintf("[%d]", c.Region)
}

func (c *DBConfig) GetType() string {
	return c.Type
}

func (c *DBConfig) GetAddr() string {
	return c.Addr
}

func (c *DBConfig) GetDB() string {
	return c.DB
}

func (c *DBConfig) GetUser() string {
	return c.User
}

func (c *DBConfig) GetPassword() string {
	return c.Password
}

func (c *EsConfig) GetUrl() string {
	return c.Url
}

func (c *EsConfig) GetIndex() string {
	return c.Index
}

func (c *RegionConfig) init() {
	for _, db := range c.Databases {
		if db.Region == c.Region {
			db.Addr = fmt.Sprintf("%s:%d", db.LocalHost, db.Port)
			c.DB = db
		} else {
			db.Addr = fmt.Sprintf("%s:%d", db.PublicHost, db.Port)
		}

		if db.IsGlobal {
			c.GlobalDB = db
		}
	}
}

func (c *RegionConfig) GetDbByRegion(region uint32) *DBConfig {
	for _, db := range c.Databases {
		if db.Region == region {
			return db
		}
	}
	return nil
}

func GetRegionConfig() *RegionConfig {
	return regionCfg
}

func GetGlobalDbConfig() *DBConfig {
	return regionCfg.GlobalDB
}

func GetRegionDbConfig() *DBConfig {
	return regionCfg.DB
}

func RandomGateByRegion(region uint32) *GateConfig {
	var regionGates []*GateConfig
	for _, gateCfg := range cfg.Gates {
		if gateCfg.Region == region {
			regionGates = append(regionGates, gateCfg)
		}
	}

	n := len(regionGates)
	if n <= 0 {
		return nil
	}

	return regionGates[rand.Intn(n)]
}
