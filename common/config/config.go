package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"kinger/proto/pb"
	"strconv"
	"strings"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
)

var (
	cfg   *KingwarConfig
	wxCfg *WxGameConfig
)

type VersionConfig struct {
	AccountType string
	Version string
}

type KingwarConfig struct {
	Debug    bool
	IsPC     bool
	HttpPort int
	AppStoreCanTest bool
	Wxgame   *WxGameConfig
	Channels []*ChannelConfig
	Ver      *pb.Version
	IsMultiLan bool
	IsXfMultiLan bool
	GmtoolKey string
	HostID int
	IsBanShu bool
	Versions  []*VersionConfig
	GmUids []common.UUid
	GmChannels []string
	IsServerMaintain bool

	accountType2Version map[pb.AccountTypeEnum]*pb.Version
}

type ChannelConfig struct {
	Channel string
	Tdkey string
	LoginChannels []*LoginChannelConfig
}

type LoginChannelConfig struct {
	Channel string
	AppID string
	LoginKey string
	PayKey string
	LoginSecret string
	PaySecret string
	MidasOfferID string
	MidasAppKey string
}

type WxGameConfig struct {
	IsExamined               bool `toml:"isExamined"`
	AccTreasureCnt            int  `toml:"accTreasureCnt"`
	NotSubStarCnt             int  `toml:"notSubStarCnt"`
	AccTreasureTime           int  `toml:"accTreasureTime"`
	DelayRewardTime           int  `toml:"delayRewardTime"`
	TriggerTreasureShareHDCnt int  `toml:"triggerTreasureShareHDCnt"`
	IsExamined2               bool
}

func (kc *KingwarConfig) loadVersion() error {
	kc.accountType2Version = map[pb.AccountTypeEnum]*pb.Version{}

	for _, version := range kc.Versions {
		versionInfo := strings.Split(version.Version, ".")
		if len(versionInfo) != 3 {
			return errors.Errorf("loadVersion error %v", kc.Versions)
		}

		v1, err := strconv.Atoi(versionInfo[0])
		if err != nil {
			return err
		}
		v2, err := strconv.Atoi(versionInfo[1])
		if err != nil {
			return err
		}
		v3, err := strconv.Atoi(versionInfo[2])
		if err != nil {
			return err
		}

		kc.Ver = &pb.Version{
			V1: int32(v1),
			V2: int32(v2),
			V3: int32(v3),
		}

		switch version.AccountType {
		case "ios":
			kc.accountType2Version[pb.AccountTypeEnum_Ios] = kc.Ver
		case "android":
			kc.accountType2Version[pb.AccountTypeEnum_Android] = kc.Ver
		case "wxgame":
			kc.accountType2Version[pb.AccountTypeEnum_Wxgame] = kc.Ver
			kc.accountType2Version[pb.AccountTypeEnum_WxgameIos] = kc.Ver
		}
	}

	return nil
}

func (kc *KingwarConfig) GetChannelTdkey(channel string) string {
	for _, _cfg := range kc.Channels {
		if _cfg.Channel == channel {
			return _cfg.Tdkey
		}
	}
	return ""
}

func (kc *KingwarConfig) GetChannelConfig(channel string) *ChannelConfig {
	for _, _cfg := range kc.Channels {
		if _cfg.Channel == channel {
			return _cfg
		}
	}
	return nil
}

func (kc *KingwarConfig) GetLoginChannelConfig(channel, loginChannel string) *LoginChannelConfig {
	for _, _cfg := range kc.Channels {
		if _cfg.Channel == channel {
			for _, lcfg := range _cfg.LoginChannels {
				if lcfg.Channel == loginChannel {
					return lcfg
				}
			}

			return nil
		}
	}
	return nil
}

func (kc *KingwarConfig) IsOldServer() bool {
	return kc.HostID < 1000
}

func (kc *KingwarConfig) IsXfServer() bool {
	return kc.HostID >= 1001
}

func (kc *KingwarConfig) IsOldXfServer() bool {
	return kc.HostID == 1001
}

func (kc *KingwarConfig) GetVersion(accountType pb.AccountTypeEnum) *pb.Version {
	if ver, ok := kc.accountType2Version[accountType]; ok {
		return ver
	} else if kc.Ver != nil {
		return kc.Ver
	} else {
		return &pb.Version{V3: 1}
	}
}

func LoadConfig() {
	attr := attribute.NewAttrMgr("config", "kingwar", true)
	err := attr.Load(true)
	if err != nil {
		panic(err)
	}

	cfg = &KingwarConfig{}
	_, err = toml.Decode(attr.GetStr("data"), cfg)
	if err != nil {
		panic(err)
	}

	err = cfg.loadVersion()
	if err != nil {
		panic(err)
	}
}

func ReloadConfig() error {
	attr := attribute.NewAttrMgr("config", "kingwar", true)
	err := attr.Load()
	if err != nil {
		return err
	}

	newCfg := &KingwarConfig{}
	_, err = toml.Decode(attr.GetStr("data"), newCfg)
	if err != nil {
		return err
	}

	err = newCfg.loadVersion()
	if err != nil {
		return err
	}

	cfg = newCfg
	return nil
}

func GetConfig() *KingwarConfig {
	return cfg
}
