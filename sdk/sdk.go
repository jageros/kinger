package sdk

import (
	"kinger/common/config"
	"kinger/gopuppy/common"
	"net/http"
	"strings"
)

var allSdks = map[string]map[string]ISdk{}

type ISdk interface {
	LoginAuth(channelUid, token string) error
	RechargeAuthSign(request *http.Request) (uid uint64, channelUid string, cpOrderID, channelOrderID string,
		paymentAmount int, reply []byte, needCheckMoney, ok bool)
}

func GetSdk(channel, loginChannel string) ISdk {
	if sdks, ok := allSdks[channel]; ok {
		if loginChannel == "" {
			loginChannel = channel
		}

		if s, ok := sdks[loginChannel]; ok {
			return s
		} else {
			return nil
		}

	} else {
		return nil
	}
}

func parseExt(ext string) (uint64, string) {
	extInfo := strings.Split(ext, "_")
	cpOrderID := extInfo[0]
	uid := uint64(common.ParseUUidFromString(extInfo[1]))
	if uid == 0 {
		// 客户端bug ?
		strUid := cpOrderID[:len(cpOrderID)-16]
		uid = uint64(common.ParseUUidFromString(strUid))
	}
	return uid, cpOrderID
}

func addSdk(channel, loginChannel string, s ISdk) {
	if loginChannel == "" {
		loginChannel = channel
	}

	sdks, ok := allSdks[channel]
	if !ok {
		sdks = map[string]ISdk{}
		allSdks[channel] = sdks
	}
	sdks[loginChannel] = s
}

func addSdkV2(channel, loginChannel string, newSdk func(cfg *config.LoginChannelConfig) ISdk) {
	cfg := config.GetConfig().GetLoginChannelConfig(channel, loginChannel)
	if cfg != nil {
		addSdk(channel, loginChannel, newSdk(cfg))
	}
}

func Initialize() {
	addSdk("lzd_pkgsdk", "", newFire233())
	addSdk("ios_fire233", "", newIosFire233())
	addSdk("facebook", "", newHandJoy())
	addSdk("google", "", newGoogleSdk())

	addSdk("xianfeng", "",
		newXianFeng(config.GetConfig().GetLoginChannelConfig("lzd_xianfeng_taptap", "xianfeng")))
	addSdk("lzd_xianfeng_taptap", "xianfeng",
		newXianFeng(config.GetConfig().GetLoginChannelConfig("lzd_xianfeng_taptap", "xianfeng")))
	addSdk("lzd_xianfeng_recharge", "xianfeng",
		newXianFeng(config.GetConfig().GetLoginChannelConfig("lzd_xianfeng_taptap", "xianfeng")))
	addSdkV2("lzd_xianfeng_i4", "i4", newAiSiSdk)

	addSdkV2("iclock_morefun", "ios", newIClockMoreFun)
	addSdkV2("iclock_morefun", "googleplay", newIClockMoreFun)
	addSdkV2("iclock_morefun", "android", newIClockMoreFun)

	addSdkV2("lzd_xianfeng_tx", "qq", newTencentQQYSDK)
	addSdkV2("lzd_xianfeng_tx", "wechat", newTencentWechatYSDK)

	addSdkV2("lzd_xianfeng_multilan", "xianfeng", newMultilanXianFeng)
	addSdkV2("lzd_xianfeng_channel", "xianfeng", newXianFeng)
}
