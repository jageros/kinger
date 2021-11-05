package treasure

import (
	"github.com/gogo/protobuf/proto"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/timer"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"math"
	"time"
)

const (
	maxOpenTime = 2000000000
	ttUnknowTreasure = 1
	ttRewardTreasure = 2
	ttDailyTreasure = 3
)

type iTreasure interface {
	isOpen() bool
	isCanOpen() bool
	opened()
	packMsg(player types.IPlayer) proto.Marshaler
	isDobule(player types.IPlayer) bool
	getAttr() *attribute.MapAttr
	getModelID() string
}

type treasureSt struct {
	attr *attribute.MapAttr
}

func newTreasureByAttr(attr *attribute.MapAttr) *treasureSt {
	return &treasureSt{
		attr: attr,
	}
}

func newTreasure(id, pos int32, modelID string) *treasureSt {
	attr := attribute.NewMapAttr()
	attr.SetUInt32("id", uint32(id))
	attr.SetStr("modelID", modelID)
	attr.SetInt32("openTime", math.MaxInt32)
	attr.SetInt32("openStarCount", -1)
	attr.SetInt32("pos", pos)
	return newTreasureByAttr(attr)
}

func (t *treasureSt) getAttr() *attribute.MapAttr {
	return t.attr
}

// 唯一id
func (t *treasureSt) getID() uint32 {
	return t.attr.GetUInt32("id")
}

// 表里的id
func (t *treasureSt) getModelID() string {
	return t.attr.GetStr("modelID")
}

func (t *treasureSt) setModelID(modelID string) {
	t.attr.SetStr("modelID", modelID)
}

func (t *treasureSt) getOpenTime() int32 {
	return t.attr.GetInt32("openTime")
}

func (t *treasureSt) setOpenTime(time_ int32) {
	t.attr.SetInt32("openTime", time_)
}

func (t *treasureSt) setActivateTime(time_ int32)  {
	t.attr.SetInt32("activateTime", time_)
}

func (t *treasureSt) getPos() int32 {
	return t.attr.GetInt32("pos")
}

func (t *treasureSt) getOpenStarCount() int32 {
	return t.attr.GetInt32("openStarCount")
}

func (t *treasureSt) setOpenStartCount(count int32) {
	t.attr.SetInt32("openStarCount", count)
}

func (t *treasureSt) getGameData() *gamedata.Treasure {
	treasureGameData := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData)
	return treasureGameData.Treasures[t.getModelID()]
}

func (t *treasureSt) getRare() int {
	data := t.getGameData()
	if data != nil {
		return data.Rare
	}
	return 0
}


func (t *treasureSt) getOpenNeedTime() int {
	data := t.getGameData()
	if data == nil {
		return 0
	}
	return data.RewardUnlockTime
}

func (t *treasureSt) isActivated() bool {
	openTime := t.getOpenTime()
	return openTime >= 0 && openTime < maxOpenTime
}

func (t *treasureSt) isCanOpen() bool {
	openTime := t.getOpenTime()
	if openTime >= 0 {
		return int64(openTime) <= time.Now().Unix()
	} else {
		return t.getOpenStarCount() <= 0
	}
}

func (t *treasureSt) isOpen() bool {
	return t.attr.GetBool("isOpen")
}

func (t *treasureSt) opened() {
	t.attr.SetBool("isOpen", true)
}

func (t *treasureSt) packMsg(player types.IPlayer) proto.Marshaler {
	openTimeout := t.getOpenTime() - int32(time.Now().Unix())
	if openTimeout < 0 {
		openTimeout = 0
	}
	return &pb.Treasure{
		ID:            t.getID(),
		ModelID:       t.getModelID(),
		OpenTimeout:   openTimeout,
		OpenStarCount: -1,
		Pos:           t.getPos(),
	}
}

func (t *treasureSt) isDobule(player types.IPlayer) bool {
	return false
}


type dailyTreasureSt struct {
	treasureSt
}

func newDailyTreasureByAttr(attr *attribute.MapAttr) *dailyTreasureSt {
	t := &dailyTreasureSt{}
	t.attr = attr
	data := t.getGameData()
	if data != nil && t.getOpenStarCount() > int32(data.DailyUnlockStar) {
		t.setOpenStartCount(int32(data.DailyUnlockStar))
	}
	return t
}

func newDailyTreasure(id int32, data *gamedata.Treasure) *dailyTreasureSt {
	attr := attribute.NewMapAttr()
	attr.SetUInt32("id", uint32(id))
	attr.SetStr("modelID", data.ID)
	attr.SetInt32("openStarCount", int32(data.DailyUnlockStar))
	attr.SetBool("isOpen", false)
	attr.SetInt("dayno", timer.GetDayNo())
	return newDailyTreasureByAttr(attr)
}

func (t *dailyTreasureSt) setHelper(uid common.UUid, headImg, headFrame, name string) {
	t.attr.SetUInt64("helperUid", uint64(uid))
	t.attr.SetStr("helperHeadImg", headImg)
	t.attr.SetStr("helperHeadFrame", headImg)
	t.attr.SetStr("helperName", headImg)
}

func (t *dailyTreasureSt) isDobule(player types.IPlayer) bool {
	double := t.attr.GetBool("isDobule")
	if !double {
		num := module.OutStatus.BuffDoubleRewardOfVip(player, 1)
		if num == 2 {
			return true
		}
	}
	return double
}

func (t *dailyTreasureSt) beDouble() {
	t.attr.SetBool("isDobule", true)
}

func (t *dailyTreasureSt) getDayno() int {
	return t.attr.GetInt("dayno")
}

func (t *dailyTreasureSt) setDayno(dayno int) {
	t.attr.SetInt("dayno", dayno)
}

func (t *dailyTreasureSt) getDayIdx() int {
	return t.attr.GetInt("dayIdx")
}

func (t *dailyTreasureSt) setDayIdx(idx int) {
	t.attr.SetInt("dayIdx", idx)
}

func (t *dailyTreasureSt) packMsg(player types.IPlayer) proto.Marshaler {
	data := t.getGameData()
	ts := gamedata.GetGameData(consts.Treasure).(*gamedata.TreasureGameData).Team2DailyTreasure[data.Team]
	dayIdx := t.getDayIdx()
	remainAmount := len(ts) - dayIdx
	if remainAmount < 0 {
		remainAmount = 0
	}

	var completedPro int
	for i, t := range ts {
		if i < dayIdx {
			completedPro += t.DailyUnlockStar
		} else {
			break
		}
	}
	msg := &pb.DailyTreasure{
		ID:            t.getID(),
		ModelID:       t.getModelID(),
		OpenStarCount: t.getOpenStarCount(),
		IsOpen:        t.isOpen(),
		NextTime:      int32(timer.TimeDelta(0, 0, 0).Seconds()),
		IsDouble:      t.isDobule(player),
		RemainAmount: int32(remainAmount),
		CompletedPro: int32(completedPro),
	}

	helperUid := t.attr.GetUInt64("helperUid")
	if helperUid > 0 {
		msg.ShareInfo = &pb.DailyTreasureShareInfo{
			HelperUid: helperUid,
			HelperHeadImg: t.attr.GetStr("helperHeadImg"),
			HelperHeadFrame: t.attr.GetStr("helperHeadFrame"),
			HelperName: t.attr.GetStr("helperName"),
		}
	}
	return msg
}
