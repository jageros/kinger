package pvp

import (
	"fmt"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/evq"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/proto/pb"
	"strconv"
	"time"
)

type leagueSeasonAttr struct {
	attr              map[int]*attribute.AttrMgr
	area2RewardAttr   map[int]map[int]*attribute.AttrMgr // map[area]map[serial]*attribute.AttrMgr
	loadingRewardAttr map[int]map[int]chan struct{}      // map[area]map[serial]chan struct{}
	cycleCfg          *gamedata.LeagueCycleTime
}

func (l *leagueSeasonAttr) init() {
	l.initLeagueCycleTime()
	l.attr = map[int]*attribute.AttrMgr{}
	l.area2RewardAttr = map[int]map[int]*attribute.AttrMgr{}
	l.loadingRewardAttr = map[int]map[int]chan struct{}{}
	areaData := gamedata.GetGameData(consts.AreaConfig)
	areaData.AddReloadCallback(l.loadAttr)
	l.loadAttr(areaData)
}

func (l *leagueSeasonAttr) loadAttr(data gamedata.IGameData) {
	data.(*gamedata.AreaConfigGameData).ForEachOpenedArea(func(config *gamedata.AreaConfig) {
		_, ok := l.attr[config.Area]
		if ok {
			return
		}

		attr := attribute.NewAttrMgr("leagueSeason", config.Area, true)
		attr.Load()
		l.attr[config.Area] = attr
	})
}

func (l *leagueSeasonAttr) reloadAttrForArea(area int, isCrossSeason bool) {
	attr := attribute.NewAttrMgr("leagueSeason", area, true)
	err := attr.Load()
	if err != nil {
		glog.Errorf("leagueSeasonAttr reloadAttrForArea error, area=%d isCrossSeason=%v err=%s", area,
			isCrossSeason, err)
		return
	}

	if isCrossSeason {
		l.loadRewardAttr(area, l.getCurSeasonSerial(area), false)
	}

	l.attr[area] = attr
	if isCrossSeason {
		l.onCrossSeason()
	}
}

func (l *leagueSeasonAttr) getCurSeasonSerial(area int) int {
	var s int
	if attr, ok := l.attr[area]; ok {
		s = attr.GetInt("league_serial")
	}
	if s <= 0 {
		s = 1
	}
	return s
}

func (l *leagueSeasonAttr) setSeasonSerial(area int, serial int, isCrossSeason bool) {
	if attr, ok := l.attr[area]; ok {
		attr.SetInt("league_serial", serial)
		attr.Save(true)
		arg := &pb.ReloadLeagueAttrArg{
			Area:          int32(area),
			AppID:         module.Service.GetAppID(),
			IsCrossSeason: isCrossSeason,
		}
		logic.BroadcastBackend(pb.MessageID_G2G_SAVE_LEAGUE_ATTR_RELOAD, arg)
	}
}

func (l *leagueSeasonAttr) getCurTimes(area int) int {
	if attr, ok := l.attr[area]; ok {
		return attr.GetInt("time_num")
	}
	return 0
}

func (l *leagueSeasonAttr) conformTimes(area int) bool {
	var num int
	if attr, ok := l.attr[area]; ok {
		defer attr.Save(false)
		num = attr.GetInt("time_num")
		num += 1
		if num >= l.cycleCfg.TimeNum {
			attr.SetInt("time_num", 0)
			return true
		} else {
			attr.SetInt("time_num", num)
			return false
		}
	}
	return false
}

func (l *leagueSeasonAttr) initLeagueCycleTime() {
	lt := gamedata.GetGameData(consts.FunctionPrice).(*gamedata.FunctionPriceGameData).LeagueCycle
	l.cycleCfg = lt
}

func (l *leagueSeasonAttr) crossSeasonUpdateAttr() {
	var isCrossSeason bool
	for area, _ := range l.attr {
		if l.conformTimes(area) {
			l.recordRankPlayerRewardInfo(area)
			serial := l.getCurSeasonSerial(area)
			serial += 1
			l.setSeasonSerial(area, serial, true)
			isCrossSeason = true
		}
	}

	if isCrossSeason {
		l.onCrossSeason()
	}
}

func (l *leagueSeasonAttr) onCrossSeason() {
	glog.Infof("on cross league season serial")
	module.Player.ForEachOnlinePlayer(func(player types.IPlayer) {
		player.GetComponent(consts.PvpCpt).(*pvpComponent).updateLeagueSeason()
	})
}

func (l *leagueSeasonAttr) recordRankPlayerRewardInfo(area int) {
	msg := &pb.GLeagueSeasonEndArg{Area: int32(area)}
	reply, err := logic.CallBackend("", 0, pb.MessageID_G2R_LEAGUE_SEASON_END, msg)
	if err == nil {
		area2UserRanking := reply.(*pb.Area2UserRanking)
		for k, area := range area2UserRanking.Areas {
			serial := l.getCurSeasonSerial(int(area))
			serialAttrID := fmt.Sprintf("%d_%d", area, serial)
			seasonAttr := attribute.NewAttrMgr("leagueSeasonReward", serialAttrID, true)
			leagueEndRewardAttr := seasonAttr.GetMapAttr("league_end_reward")
			if leagueEndRewardAttr == nil {
				leagueEndRewardAttr = attribute.NewMapAttr()
				seasonAttr.SetMapAttr("league_end_reward", leagueEndRewardAttr)
			}
			endRewards := gamedata.GetGameData(consts.LeagueReward).(*gamedata.LeagueRewardGameData).ID2LeagueReward
			for lv, rws := range endRewards {
				lvStr := strconv.Itoa(lv)
				lvAttr := leagueEndRewardAttr.GetListAttr(lvStr)
				if lvAttr == nil {
					lvAttr = attribute.NewListAttr()
					leagueEndRewardAttr.SetListAttr(lvStr, lvAttr)
				}
				for _, rw := range rws.Reward {
					lvAttr.AppendStr(rw)
				}

			}
			for rank, uid := range area2UserRanking.UserRanking[k].Uids {
				playerInfo := module.Player.GetSimplePlayerInfo(common.UUid(uid))
				rankScore := int(playerInfo.GetRankScore())
				llv, rankReward, kingFlag := gamedata.GetGameData(consts.LeagueRankReward).(*gamedata.LeagueRankRewardGameData).GetRewardByRank(int(rank))
				leagueLvl := gamedata.GetGameData(consts.League).(*gamedata.LeagueGameData).GetLeagueEndRewardLvlByScore(rankScore)
				if leagueLvl < llv {
					continue
				}

				uidStr := strconv.Itoa(int(uid))
				playerAttr := seasonAttr.GetMapAttr(uidStr)
				if playerAttr == nil {
					playerAttr = attribute.NewMapAttr()
					seasonAttr.SetMapAttr(uidStr, playerAttr)
				}

				playerAttr.SetInt("rank", rank)
				playerAttr.SetInt("king_flag", kingFlag)
				rewardAttr := playerAttr.GetListAttr("rewards")
				if rewardAttr == nil {
					rewardAttr = attribute.NewListAttr()
					playerAttr.SetListAttr("rewards", rewardAttr)
				}
				for _, rw := range rankReward {
					rewardAttr.AppendStr(rw)
				}
			}

			seasonAttr.Save(true)
			serial2RewardAttr, ok := l.area2RewardAttr[int(area)]
			if !ok {
				serial2RewardAttr = map[int]*attribute.AttrMgr{}
				l.area2RewardAttr[int(area)] = serial2RewardAttr
			}
			serial2RewardAttr[serial] = seasonAttr
		}
	}
}

func (l *leagueSeasonAttr) loadRewardAttr(area, serial int, justFromCache bool) *attribute.AttrMgr {
	serial2RewardAttr, ok := l.area2RewardAttr[area]
	if ok {
		if attr, ok := serial2RewardAttr[serial]; ok {
			return attr
		}
	} else {
		serial2RewardAttr = map[int]*attribute.AttrMgr{}
		l.area2RewardAttr[area] = serial2RewardAttr
	}

	if justFromCache {
		return nil
	}

	serial2loading, ok := l.loadingRewardAttr[area]
	if ok {
		if loading, ok := serial2loading[serial]; ok {
			evq.Await(func() {
				<-loading
			})
			return l.loadRewardAttr(area, serial, true)
		}
	} else {
		serial2loading = map[int]chan struct{}{}
		l.loadingRewardAttr[area] = serial2loading
	}

	c := make(chan struct{})
	serial2loading[serial] = c
	defer func() {
		delete(serial2loading, serial)
		close(c)
	}()

	serialAttrID := fmt.Sprintf("%d_%d", area, serial)
	rewardAttr := attribute.NewAttrMgr("leagueSeasonReward", serialAttrID, true)
	err := rewardAttr.Load()
	if err != nil {
		glog.Errorf("leagueSeasonAttr loadRewardAttr error, area=%d serial=%d err=%s", area, serial, err)
		return nil
	}

	serial2RewardAttr[serial] = rewardAttr
	return rewardAttr
}

func (l *leagueSeasonAttr) getPlayerLeagueRankReward(player types.IPlayer, serial int) (rank int, rewards []string,
	kingFlag int, Shortlisted bool) {

	area := player.GetArea()
	seasonAttr := l.loadRewardAttr(area, serial, false)
	if seasonAttr == nil {
		glog.Errorf("leagueSeasonAttr getPlayerLeagueRankReward no seasonAttr, uid=%d area=%d serial=%d",
			player.GetUid(), area, serial)
		return
	}

	uid := player.GetUid()
	uidStr := strconv.Itoa(int(uid))
	playerAttr := seasonAttr.GetMapAttr(uidStr)
	if playerAttr == nil {
		return
	}

	Shortlisted = true
	rank = playerAttr.GetInt("rank")
	kingFlag = playerAttr.GetInt("king_flag")
	rewardAttr := playerAttr.GetListAttr("rewards")
	if rewardAttr != nil {
		rewardAttr.ForEachIndex(func(index int) bool {
			rw := rewardAttr.GetStr(index)
			rewards = append(rewards, rw)
			return true
		})
	}
	return
}

func (l *leagueSeasonAttr) getLeagueEndReward(area, serial, lvl int) (reward []string) {
	seasonAttr := l.loadRewardAttr(area, serial, false)
	if seasonAttr == nil {
		glog.Errorf("leagueSeasonAttr getPlayerLeagueRankReward no seasonAttr, area=%d serial=%d", area, serial)
		return
	}
	leagueEndRewardAttr := seasonAttr.GetMapAttr("league_end_reward")
	if leagueEndRewardAttr != nil {
		lvStr := strconv.Itoa(lvl)
		rwListAttr := leagueEndRewardAttr.GetListAttr(lvStr)
		rwListAttr.ForEachIndex(func(index int) bool {
			reward = append(reward, rwListAttr.GetStr(index))
			return true
		})
	}
	return
}

func (l *leagueSeasonAttr) initCrossSeasonFunction() {
	if module.Service.GetAppID() != 1 {
		return
	}
	timer.RunEveryDay(0, 0, 0, func() {
		switch l.cycleCfg.TimeType {
		case "M":
			if time.Now().Day() == l.cycleCfg.TimeDay {
				l.crossSeasonUpdateAttr()
			}
		case "W":
			if int(time.Now().Weekday()) == l.cycleCfg.TimeDay {
				l.crossSeasonUpdateAttr()
			}
		case "D":
			l.crossSeasonUpdateAttr()
		default:

		}
	})

}

func (l *leagueSeasonAttr) getRemainTime(area int) int64 {
	var remainDay int
	var remainTime int64
	curNum := l.getCurTimes(area)
	switch l.cycleCfg.TimeType {
	case "M":
		y, m, _ := time.Now().Date()
		mm := int(m)
		mm += l.cycleCfg.TimeNum - curNum
		if mm > 12 {
			y += mm / 12
			mm = mm % 12
		}
		var mstr, dstr string
		if mm < 10 {
			mstr = fmt.Sprintf("0%d", mm)
		} else {
			mstr = fmt.Sprintf("%d", mm)
		}
		if l.cycleCfg.TimeDay < 10 {
			dstr = fmt.Sprintf("0%d", l.cycleCfg.TimeDay)
		} else {
			dstr = fmt.Sprintf("%d", l.cycleCfg.TimeDay)
		}
		tStr := fmt.Sprintf("%d-%s-%s 00:00:00", y, mstr, dstr)
		endTime, _ := utils.StringToTime(tStr, utils.TimeFormat2)
		remainDay = timer.GetDayNo(endTime.Unix()) - timer.GetDayNo() - 1
		remainTime = int64(timer.TimeDelta(0, 0, 0).Seconds() + float64(remainDay*86400))

	case "W":
		wd := (l.cycleCfg.TimeNum - curNum - 1) * 7
		td := time.Now().Weekday() + 1
		cfgDay := l.cycleCfg.TimeDay
		if cfgDay == 0 {
			cfgDay = 7
		}
		var dayNum int
		if int(td) > cfgDay {
			dayNum = 7 - int(td) + cfgDay + wd
		} else {
			dayNum = cfgDay - int(td) + wd
		}
		remainDay = dayNum
		remainTime = int64(timer.TimeDelta(0, 0, 0).Seconds() + float64(remainDay*86400))

	case "D":
		remainDay = l.cycleCfg.TimeNum - curNum - 1
		remainTime = int64(timer.TimeDelta(0, 0, 0).Seconds() + float64(remainDay*86400))

	default:

	}
	return remainTime
}
