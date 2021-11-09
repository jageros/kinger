package utils

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/common/forbidlist"
	"kinger/gamedata"
	"kinger/gopuppy/apps/center/mq"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	gconsts "kinger/gopuppy/common/consts"
	"kinger/gopuppy/common/wordfilter"
	"kinger/gopuppy/network/protoc"
	"kinger/proto/pb"
	"strings"
	"time"
)

func RegisterDirtyWords() {
	wordfilter.RegisterWords(func() map[int][]string {
		var wordsList = map[int][]string{}
		keys := map[int]string{gconsts.GeneralWords: "dirty_word", gconsts.FuzzyWords: "fuzzy_word", gconsts.AccurateWords: "accurate_word"}
		for k, attrId := range keys {
			var words []string
			attr := attribute.NewAttrMgr("gamedata", attrId, true)
			attr.Load()
			wordAttr := attr.GetListAttr("words")
			if wordAttr != nil {
				wordAttr.ForEachIndex(func(index int) bool {
					words = append(words, wordAttr.GetStr(index))
					return true
				})
			}
			wordsList[k] = words
		}
		return wordsList
	})
}

func PlayerMqPublish(uid common.UUid, type_ protoc.IMessageID, msg mq.IRmqMessge) {
	mq.Publish(fmt.Sprintf("player:%d", uid), type_, logic.GetAgentRegion(uid), msg)
}

func AddDirtyWords(addWordsStr, delWordsStr string, isAccurate bool) {
	addWords := strings.Split(addWordsStr, "///")
	delWords := strings.Split(delWordsStr, "///")

	var key string
	if isAccurate {
		key = "accurate_word"
	} else {
		key = "fuzzy_word"
	}
	flag := 0
	wordsList := map[bool][]string{true: addWords, false: delWords}
	for isAdd, words := range wordsList {
		if words == nil || len(words) <= 0 {
			continue
		}
		wordfilter.UpdateDirtyWords(words, isAccurate, isAdd)
		if module.Service.GetAppID() != 1 {
			continue
		}

		attr := attribute.NewAttrMgr("gamedata", key, true)
		attr.Load()
		wordAttr := attr.GetListAttr("words")
		if wordAttr == nil {
			wordAttr = attribute.NewListAttr()
			attr.SetListAttr("words", wordAttr)
		}
		if isAdd {
			for _, wrd := range words {
				if wrd != "" && wrd != " " {
					wordAttr.AppendStr(wrd)
				}
			}
		} else {
			for _, wrd := range words {
				if wrd != "" && wrd != " " {
					wordAttr.ForEachIndex(func(index int) bool {
						if wordAttr.GetStr(index) == wrd {
							wordAttr.DelByIndex(index)
							return false
						}
						return true
					})
				}
			}
		}

		attr.Save(false)
		flag = 1
	}

	if flag == 1 {
		arg := &pb.ImportWordArg{
			AddWordsStr: addWordsStr,
			DelWordsStr: delWordsStr,
			IsAccurate:  isAccurate,
		}
		logic.BroadcastBackend(pb.MessageID_G2G_ON_IMPORT_DIRTY_WORDS, arg)
	}
}

func ForbidAccount(uid common.UUid, forbidType int, isForbid bool, overTimes int64, msg string, isAuto bool) error {
	p := module.Player.GetPlayer(uid)
	if p == nil {
		return gamedata.InternalErr
	}
	var leftTime int
	if overTimes > 0 {
		leftTime = int(overTimes - time.Now().Unix())
		if leftTime <= 0 {
			return nil
		}
	} else {
		leftTime = int(overTimes)
	}

	if forbidType == consts.ForbidAccount {
		if isForbid {
			module.OutStatus.AddForbidStatus(p, consts.ForbidAccount, leftTime)
			if isAuto {
				forbidlist.AddForbidInfo(uid, msg, forbidType)
			}
			p.OnForbidLogin()
		} else {
			module.OutStatus.DelForbidStatus(p, consts.ForbidAccount)
		}
	}
	if forbidType == consts.ForbidChat {
		if isForbid {
			module.OutStatus.AddForbidStatus(p, consts.ForbidChat, leftTime)
		} else {
			module.OutStatus.DelForbidStatus(p, consts.ForbidChat)
		}
		logic.BroadcastBackend(pb.MessageID_L2CA_FORBID_CHAT, &pb.ForbidChatArg{
			Uid:      uint64(p.GetUid()),
			IsForbid: isForbid,
		})
	}
	if forbidType == consts.ForbidMonitor {
		if isForbid {
			module.OutStatus.AddForbidStatus(p, consts.ForbidMonitor, leftTime)
			if isAuto {
				forbidlist.AddForbidInfo(uid, msg, forbidType)
			}
		} else {
			module.OutStatus.DelForbidStatus(p, consts.ForbidMonitor)
		}
	}
	return nil
}

func UpdateForbidList(uid common.UUid, forbidType int, isForbid bool, opTime int64, isDelAll bool) {
	if isForbid {
		forbidlist.AddForbidInfo(uid, "", forbidType)
	} else {
		forbidlist.DelForbidInfo(uid, forbidType, opTime, isDelAll)
	}
}

func ForbidIpAddr(ipaddr string, isForbid bool) {
	if isForbid {
		forbidlist.AddForbidIP(ipaddr)
		module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
			if player.GetIP() == ipaddr {
				player.OnForbidLogin()
			}
		})
	} else {
		forbidlist.DelForbidIP(ipaddr)
	}
}

func IsForbidIPAddr(ipaddr string) bool {
	return forbidlist.IsForbidIP(ipaddr)
}

func InitForbidList() {
	forbidlist.InitForbidIP()
}

func CrossLeagueSeasonResetScore(oldMaxScore, oldRankScore int) (modifyMaxScore, modifyScore int) {
	baseScore := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetScoreById(1)
	if oldMaxScore < baseScore || oldRankScore < baseScore {
		return
	}

	funGameData := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData)
	pro := float64(funGameData.LeagueResetRewardProp) / 100
	compensatoryScore := int(float64(oldRankScore-baseScore) * pro)
	maxCompensatoryScore := funGameData.LeagueResetRewardMax
	if compensatoryScore > maxCompensatoryScore {
		compensatoryScore = maxCompensatoryScore
	}
	newScore := baseScore + compensatoryScore
	modifyScore = newScore - oldRankScore
	modifyMaxScore = newScore - oldMaxScore
	if newScore < baseScore {
		modifyMaxScore = baseScore - oldMaxScore
	}
	if modifyMaxScore >= 0 {
		modifyMaxScore = 0
	}
	if modifyScore >= 0 || newScore <= baseScore {
		modifyScore = 0
	}
	return
}
