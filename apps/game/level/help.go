package level

import (
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"strconv"
	"kinger/proto/pb"
	"kinger/apps/game/module"
	"kinger/gopuppy/common/evq"
)

type helpRecord struct {
	cptAttr *attribute.MapAttr
	helpAttr *attribute.MapAttr
	beHelpAttr *attribute.MapAttr
}

func newHelpRecord(cptAttr *attribute.MapAttr) *helpRecord {
	return &helpRecord{cptAttr: cptAttr}
}

func (hr *helpRecord) getHelpAttr() *attribute.MapAttr {
	if hr.helpAttr != nil {
		return hr.helpAttr
	}
	hr.helpAttr = hr.cptAttr.GetMapAttr("help")
	if hr.helpAttr == nil {
		hr.helpAttr = attribute.NewMapAttr()
		hr.cptAttr.SetMapAttr("help", hr.helpAttr)
	}
	return hr.helpAttr
}

func (hr *helpRecord) getOrNewBeHelpAttr() *attribute.MapAttr {
	hr.getBeHelpAttr()
	if hr.beHelpAttr == nil {
		hr.beHelpAttr = attribute.NewMapAttr()
		hr.cptAttr.SetMapAttr("beHelp", hr.beHelpAttr)
	}
	return hr.beHelpAttr
}

func (hr *helpRecord) getBeHelpAttr() *attribute.MapAttr {
	if hr.beHelpAttr != nil {
		return hr.beHelpAttr
	}
	hr.beHelpAttr = hr.cptAttr.GetMapAttr("beHelp")
	return hr.beHelpAttr
}

func (hr *helpRecord) setHelpInfo(beHelpUid common.UUid, levelID int) {
	attr := hr.getHelpAttr()
	attr.SetUInt64("uid", uint64(beHelpUid))
	attr.SetInt("levelID", levelID)
}

func (hr *helpRecord) delHelp() {
	hr.helpAttr = nil
	hr.cptAttr.Del("help")
}

func (hr *helpRecord) getBeHelpUid() common.UUid {
	attr := hr.getHelpAttr()
	return common.UUid(attr.GetUInt64("uid"))
}

func (hr *helpRecord) getLevelID() int {
	attr := hr.getHelpAttr()
	return attr.GetInt("levelID")
}

func (hr *helpRecord) addNeedAskHelpLevel(levelID int) {
	attr := hr.getOrNewBeHelpAttr()
	levelKey := strconv.Itoa(levelID)
	levelBeHelpAttr := attr.GetListAttr(levelKey)
	if levelBeHelpAttr == nil {
		levelBeHelpAttr = attribute.NewListAttr()
		attr.SetListAttr(levelKey, levelBeHelpAttr)
	}
}

func (hr *helpRecord) delNeedAskHelpLevel(levelID int) {
	beHelpAttr := hr.getBeHelpAttr()
	if beHelpAttr != nil {
		beHelpAttr.Del(strconv.Itoa(levelID))
	}
}

func (hr *helpRecord) recordBeHelp(helperUid common.UUid, helperName string, levelID int, battleID common.UUid) {
	attr := hr.getOrNewBeHelpAttr()
	levelBeHelpAttr := attr.GetListAttr(strconv.Itoa(levelID))
	if levelBeHelpAttr == nil {
		levelBeHelpAttr = attribute.NewListAttr()
		attr.SetListAttr(strconv.Itoa(levelID), levelBeHelpAttr)
	}

	var recordAttr *attribute.MapAttr
	helperUid2 := uint64(helperUid)
	levelBeHelpAttr.ForEachIndex(func(index int) bool {
		recordAttr2 := levelBeHelpAttr.GetMapAttr(index)
		if recordAttr2.GetUInt64("uid") == helperUid2 {
			recordAttr = recordAttr2
			return false
		}
		return true
	})

	if recordAttr == nil {
		recordAttr = attribute.NewMapAttr()
		recordAttr.SetUInt64("uid", helperUid2)
		levelBeHelpAttr.AppendMapAttr(recordAttr)
	}
	if recordAttr.GetUInt64("battleID") <= 0 {
		recordAttr.SetUInt64("battleID", uint64(battleID))
		recordAttr.SetInt("cnt", recordAttr.GetInt("cnt")+1)
	}
}

func (hr *helpRecord) packBeHelpMsg(levelID int) *pb.LevelHelpRecord {
	records := &pb.LevelHelpRecord{}
	attr := hr.getBeHelpAttr()
	if attr == nil {
		return records
	}

	levelBeHelpAttr := attr.GetListAttr(strconv.Itoa(levelID))
	if levelBeHelpAttr == nil {
		return records
	}

	uid2Record := map[uint64]*pb.LevelHelpRecordItem{}
	var loadPlayerChans []chan *pb.SimplePlayerInfo
	levelBeHelpAttr.ForEachIndex(func(index int) bool {
		recordAttr := levelBeHelpAttr.GetMapAttr(index)
		item := &pb.LevelHelpRecordItem{
			HelpCnt: int32(recordAttr.GetInt("cnt")),
			VideoID: recordAttr.GetUInt64("battleID"),
		}
		item.IsWin = item.VideoID > 0
		uid2Record[recordAttr.GetUInt64("uid")] = item

		loadPlayerChans = append(loadPlayerChans, module.Player.LoadSimplePlayerInfoAsync(common.UUid(recordAttr.GetUInt64("uid"))))
		records.Records = append(records.Records, item)
		return true
	})

	if len(loadPlayerChans) > 0 {
		evq.Await(func() {
			for _, c := range loadPlayerChans {
				playerInfo := <- c
				if item, ok := uid2Record[playerInfo.Uid]; ok {
					item.HelperHeadImgUrl = playerInfo.HeadImgUrl
					item.HelperName = playerInfo.Name
					item.HelperHeadFrame = playerInfo.HeadFrame
				}
			}
		})
	}

	return records
}

func (hr *helpRecord) packNeedAskHelpLevels() []int32 {
	var levels []int32
	attr := hr.getBeHelpAttr()
	if attr != nil {
		attr.ForEachKey(func(key string) {
			levelID, _ := strconv.Atoi(key)
			levels = append(levels, int32(levelID))
		})
	}
	return levels
}
