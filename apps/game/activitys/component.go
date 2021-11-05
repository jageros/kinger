package activitys

import (
	"fmt"
	"kinger/gopuppy/attribute"
	"kinger/gopuppy/common/glog"
	"kinger/gopuppy/common/timer"
	"kinger/gopuppy/common/utils"
	"kinger/apps/game/activitys/consume"
	"kinger/apps/game/activitys/dailyrecharge"
	"kinger/apps/game/activitys/dailyshare"
	"kinger/apps/game/activitys/fight"
	"kinger/apps/game/activitys/firstrecharge"
	"kinger/apps/game/activitys/growplan"
	"kinger/apps/game/activitys/login"
	"kinger/apps/game/activitys/loginrecharge"
	"kinger/apps/game/activitys/online"
	"kinger/apps/game/activitys/rank"
	"kinger/apps/game/activitys/recharge"
	aTypes "kinger/apps/game/activitys/types"
	"kinger/apps/game/activitys/win"
	"kinger/apps/game/module"
	"kinger/apps/game/module/types"
	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/proto/pb"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"kinger/apps/game/activitys/spring"
)

var _ types.IPlayerComponent = &activityComponent{}
var _ aTypes.IPlayerCom = &activityComponent{}

const (
	//奖励类型
	ty_unknow_ = iota
	ty_gold_
	ty_jade_
	ty_headFrame_
	ty_chatPop_
	ty_card_
	ty_randomCard_
	ty_ticket_
	ty_treasure_
	ty_skin_
	ty_eventItem_
)

type activityComponent struct {
	attr               *attribute.MapAttr
	player             types.IPlayer
	activityTagList    []int32
	oldActivityTagList []int32
}

func (ac *activityComponent) ComponentID() string {
	return consts.ActivityCpt
}

func (ac *activityComponent) GetPlayer() types.IPlayer {
	return ac.player
}

func (ac *activityComponent) OnInit(player types.IPlayer) {
	ac.player = player
}

func (ac *activityComponent) OnLogin(isRelogin, isRestore bool) {
	ac.initTagList(true)
	login.OnLogin(ac.player)
	recharge.OnLogin(ac.player)
	online.OnLogin(ac.player)
	fight.OnLogin(ac.player)
	win.OnLogin(ac.player)
	rank.OnLogin(ac.player)
	consume.OnLogin(ac.player)
	loginrecharge.OnLogin(ac.player)
	firstrecharge.OnLogin(ac.player)
	growplan.OnLogin(ac.player)
	spring.OnLogin(ac.player)
	dailyrecharge.OnLogin(ac.player)
	dailyshare.OnLogin(ac.player)
	ac.oldActivityTagList = ac.activityTagList
}

func (ac *activityComponent) OnLogout() {
}

func (ac *activityComponent) OnCrossDay (dayno int) {
	if ac.player.GetDataDayNo() == dayno {
		return
	}
	login.OnCrossDay(ac.player, dayno)
	recharge.OnCrossDay(ac.player, dayno)
	fight.OnCrossDay(ac.player, dayno)
	win.OnCrossDay(ac.player, dayno)
	rank.OnCrossDay(ac.player, dayno)
	consume.OnCrossDay(ac.player, dayno)
	loginrecharge.OnCrossDay(ac.player, dayno)
	growplan.OnCrossDay(ac.player, dayno)
	spring.OnCrossDay(ac.player, dayno)
	dailyrecharge.OnCrossDay(ac.player, dayno)
	dailyshare.OnCrossDay(ac.player, dayno)
}

func (ac *activityComponent) InitAttr(key string) *attribute.MapAttr {
	attr := ac.attr.GetMapAttr(key)
	if attr == nil {
		attr = attribute.NewMapAttr()
		ac.attr.SetMapAttr(key, attr)
	}
	return attr
}

func (ac *activityComponent) UpdateActivityTagList(idList []int32) {
	ac.activityTagList = idList
}

func (ac *activityComponent) GetActivityTagList() []int32 {
	return ac.activityTagList
}

func (ac *activityComponent) initTagList(plog bool) {
	lable := []int{}
	ac.activityTagList = []int32{}
	for aid, a := range mod.id2Activity {
		if ac.ConformTime(aid) && ac.ConformOpen(aid) {
			if plog {
				ac.LogActivity(aid, a.GetActivityVersion(), a.GetActivityType(), 0, aTypes.ActivityOnStart)
			}
			lable = append(lable, aid)
		}
	}
	sort.Ints(lable)
	for _, v := range lable {
		ac.activityTagList = append(ac.activityTagList, int32(v))
	}
	login.UpdateTagList(ac.player)
	recharge.UpdateTabList(ac.player)
	fight.UpdateTabList(ac.player)
	win.UpdateTabList(ac.player)
	rank.UpdateTabList(ac.player)
	consume.UpdateTabList(ac.player)
	loginrecharge.UpdateTabList(ac.player)
	firstrecharge.UpdateTabList(ac.player)
	growplan.UpdateTabList(ac.player)
	glog.Infof("activityComponent initTagList %v", ac.activityTagList)
}

func (ac *activityComponent) ConformOpen(activityID int) bool {
	activityData := mod.GetActivityByID(activityID)
	if activityData == nil {
		return false
	}
	comData := activityData.GetOpenCondition()
	if comData == nil || comData.ID == 0 {
		return true
	}
	if ac.player.GetPvpLevel() < comData.RankLevel {
		return false
	}
	if comData.IsVip {
		if !ac.player.IsVip() {
			return false
		}
	}
	if comData.InitialCamp != 0 && ac.player.GetInitCamp() != comData.InitialCamp {
		return false
	}
	if comData.AllCardCnt > 0 && comData.AllCardCnt > len(module.Card.GetAllCollectCards(ac.player)) {
		return false
	}
	for _, v := range comData.CardLevel {
		str := strings.Split(v, ":")
		cardID, err := strconv.Atoi(str[0])
		if err != nil {
			glog.Errorf("ConformOpen string0 to int return err=%s, uid=%d, activityID=%d", err, ac.player.GetUid(), activityID)
			return false
		}
		cardLevel, err := strconv.Atoi(str[1])
		if err != nil {
			glog.Errorf("ConformOpen string1 to int return err=%s, uid=%d, activityID=%d", err, ac.player.GetUid(), activityID)
			return false
		}
		card := module.Card.GetCollectCard(ac.player, uint32(cardID))
		if card == nil {
			glog.Errorf("ConformOpen GetCollectCard return err=%s, uid=%d, activityID=%d, cardID=%d", err, ac.player.GetUid(), activityID, cardID)
			return false
		} else {
			if card.GetLevel() < cardLevel {
				return false
			}
		}
	}
	if ac.GetCreateDayNum() < comData.PlayDay {
		return false
	}

	if comData.CreatTimeBefore != "" {
		creatBeforeTim, err := utils.StringToTime(comData.CreatTimeBefore, utils.TimeFormat2)
		if err != nil || timer.GetDayNo(creatBeforeTim.Unix()) < timer.GetDayNo(ac.player.GetCreateTime()) {
			return false
		}
	}

	if ac.player.GetBackHandAmount()+ac.player.GetFirstHandAmount() < comData.FightCnt {
		return false
	}

	if module.Level.GetCurLevel(ac.player)-1 < comData.PassLevel {
		return false
	}

	if comData.OffensiveRate != 0 {
		if ac.player.GetFirstHandAmount() != 0 {
			fhr := int(ac.player.GetFirstHandWinAmount() / ac.player.GetFirstHandAmount() * 100)
			if fhr < comData.OffensiveRate {
				return false
			}
		} else {
			return false
		}
	}

	if comData.DefensiveRate != 0 {
		if ac.player.GetBackHandAmount() != 0 {
			fhr := int(ac.player.GetBackHandWinAmount() / ac.player.GetBackHandAmount() * 100)
			if fhr < comData.DefensiveRate {
				return false
			}
		} else {
			return false
		}
	}

	if !comData.AreaLimit.IsEffective(ac.player.GetArea()) {
		return false
	}

	if len(comData.Platform) > 0 {
		flag := 0
		playerChannel := ac.player.GetChannel()
		for _, platform := range comData.Platform {
			if platform == playerChannel {
				flag = 1
				break
			}
		}
		if flag == 0 {
			return false
		}
	}


	return true
}

func (ac *activityComponent) ConformTime(activityID int) bool {
	actData := mod.GetActivityByID(activityID)
	if actData == nil {
		return false
	}
	tim := actData.GetTimeCondition()
	if tim == nil || tim.ID == 0 {
		return false
	}
	switch tim.TimeType {
	case consts.CreateDurationDay:
		cDayNum := ac.GetCreateDayNum()
		if cDayNum >= tim.RegisterFirstFewDay && (cDayNum < (tim.RegisterFirstFewDay+tim.DurationDay) || tim.DurationDay == 0) {
			return true
		}
	case consts.TimeToTime:
		if time.Now().After(tim.StartTime) {
			t := time.Time{}
			if tim.EndTime == t {
				return true
			} else {
				if time.Now().Before(tim.EndTime) {
					return true
				}
			}

		}
	case consts.DayOfWeek:
		for _, v := range tim.OpenDayOfWeek {
			if time.Weekday(v) == time.Now().Weekday() {
				return true
			}
		}
	}
	return false
}

func (ac *activityComponent) GetCreateDayNum() int {
	createTime := ac.player.GetCreateTime()
	dayNum := timer.GetDayNo() - timer.GetDayNo(createTime) +1
	if dayNum < 1 {
		dayNum = 1
	}
	return dayNum
}

func (ac *activityComponent) GetLastOnlineDay() int {
	lastOnlineTime, err := utils.UnixToTime(int64(ac.player.GetLastOnlineTime()))
	if err != nil {
		return 0
	}
	return lastOnlineTime.Day()
}

func (ac *activityComponent) genModifyResourceReason(activityID, activityType, rewardID int) string {
	return fmt.Sprintf("%s%d_%d_%d", consts.RmrActivityRewardPrefix, activityType, activityID, rewardID)
}

func (ac *activityComponent) GiveReward(stuff string, num int, rd *pb.Reward, activityID, rewardID int) error {
	activityData := mod.GetActivityByID(activityID)
	if activityData == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("GiveReward get activity data err=%s, uid=%d, activityID=%d, rewardID=%d", err, ac.player.GetUid(), activityID, rewardID)
		return err
	}
	actType := activityData.GetActivityType()
	rewardTy := getRewardType(stuff)
	switch rewardTy {
	case ty_gold_:
		module.Player.ModifyResource(ac.player, consts.Gold, num, ac.genModifyResourceReason(activityID, actType, rewardID))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, goldNum=%d", activityID, actType, ac.player.GetUid(), num)
	case ty_jade_:
		module.Player.ModifyResource(ac.player, consts.Jade, num, ac.genModifyResourceReason(activityID, actType, rewardID))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, jadeNum=%d", activityID, actType, ac.player.GetUid(), num)
	case ty_headFrame_:
		module.Bag.AddHeadFrame(ac.player, strconv.Itoa(num))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, headFrameID=%d", activityID, actType, ac.player.GetUid(), num)
	case ty_chatPop_:
		module.Bag.AddChatPop(ac.player, strconv.Itoa(num))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, chatpopID=%d", activityID, actType, ac.player.GetUid(), num)
	case ty_randomCard_:
		cardNum := num/8 + 1
		card := utils.RandUInt32Sample(module.Card.GetUnlockCards(ac.player, 0), cardNum, false)
		cardSum := utils.RandFewNumberWithSum(num, cardNum)
		ac.giveCard(card, cardSum, rd, activityID, rewardID)
	case ty_ticket_:
		module.Player.ModifyResource(ac.player, consts.AccTreasureCnt, num, ac.genModifyResourceReason(activityID, actType, rewardID))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, ticketNum=%d", activityID, actType, ac.player.GetUid(), num)
	case ty_card_:
		cardID, err := strconv.Atoi(stuff)
		if err != nil {
			return err
		}
		var card []uint32
		var nums []int
		card = append(card, uint32(cardID))
		nums = append(nums, num)
		ac.giveCard(card, nums, rd, activityID, rewardID)
	case ty_treasure_:
		for i := 0; i < num; i++ {
			treasureReward := module.Treasure.OpenTreasureByModelID(ac.player, stuff, false)
			rd.TreasureReward = append(rd.TreasureReward, treasureReward)
		}
		rd.RewardList[stuff] = int32(num)
	case ty_skin_:
		module.Bag.AddCardSkin(ac.player, stuff)
		rd.RewardList[stuff] = int32(num)
	case ty_eventItem_:
		module.Player.ModifyResource(ac.player, consts.EventItem1, num, ac.genModifyResourceReason(activityID, actType, rewardID))
		rd.RewardList[stuff] = int32(num)
		glog.Infof("GiveReward modify player resource, activityID=%d, activityType=%d, uid=%d, EventItemNum=%d", activityID, actType, ac.player.GetUid(), num)
	default:
	}
	return nil
}

func (ac *activityComponent) giveCard(cardID []uint32, num []int, rd *pb.Reward, activityID, rewardID int) {
	activityData := mod.GetActivityByID(activityID)
	if activityData == nil {
		err := gamedata.GameError(aTypes.GetActivityError)
		glog.Errorf("giveCard get activity data err=%s, uid=%d, activityID=%d, rewardID=%d", err, ac.player.GetUid(), activityID, rewardID)
		return
	}
	actType := activityData.GetActivityType()
	cardCpt := ac.player.GetComponent(consts.CardCpt).(types.ICardComponent)
	cardMap := map[uint32]*pb.CardInfo{}
	for key, val := range cardID {
		rd.RewardList[strconv.Itoa(int(val))] = int32(num[key])
		cardMap[val] = &pb.CardInfo{
			Amount: int32(num[key]),
		}
	}
	cardCpt.ModifyCollectCards(cardMap)
	for k, cid := range cardID {
		glog.Infof("giveCard modify player collect cards, activityID=%d, activityType=%d, rewardID=%d, uid=%d, cardID=%d, cardNum=%d",
			activityID, actType, rewardID, ac.player.GetUid(), cid, num[k])
	}
}

func (ac *activityComponent) LogActivity(aid, vid, aty, rid, eid int) {
	glog.JsonInfo("activity", glog.Uint64("uid", uint64(ac.player.GetUid())), glog.String("channel",
		ac.player.GetChannel()), glog.String("accountType", ac.player.GetAccountType().String()),
		glog.Int("activityID", aid), glog.Int("version", vid), glog.Int("activityType", aty),
		glog.Int("rewardID", rid), glog.Int("event", eid), glog.Int("area", ac.player.GetArea()),
		glog.String("subChannel", ac.player.GetSubChannel()))
}

func (ac *activityComponent) OnHeartBeat() {
	ac.initTagList(false)
	flag := 0
	for _, aid := range ac.oldActivityTagList {
		if !aTypes.IsInArry(aid, ac.activityTagList) {
			msg := &pb.ActivityStatusChange{
				ActivityID: aid,
				Status:     false,
			}
			ac.player.GetAgent().PushClient(pb.MessageID_S2C_REFRESH_ACTIVITY_STATUS, msg)
			flag = 1
		}
	}
	for _, aid := range ac.activityTagList {
		if !aTypes.IsInArry(aid, ac.oldActivityTagList) {
			msg := &pb.ActivityStatusChange{
				ActivityID: aid,
				Status:     true,
			}
			ac.player.GetAgent().PushClient(pb.MessageID_S2C_REFRESH_ACTIVITY_STATUS, msg)
			flag = 1
			act := mod.GetActivityByID(int(aid))
			if act.GetActivityType() == consts.ActivityOfOnline {
				online.UpdataPlayerInfo(ac.player)
			}
		}
	}
	if flag == 1 {
		ac.oldActivityTagList = ac.activityTagList
	}
}

func (p *activityComponent) PushFinshNum(aid, rid, num  int) {
	msg := &pb.ActivityFinshChange{
		ActivityID: int32(aid),
		FinshNum:   int32(num),
		RewardID:   int32(rid),
	}
	p.player.GetAgent().PushClient(pb.MessageID_S2C_REFRESH_ACTIVITY_FINSHNUM, msg)
}

func (p *activityComponent) PushReceiveStatus(aid, rid int, rst pb.ActivityReceiveStatus) {
	msg := &pb.ActivityReceiveChange{
		ActivityID:    int32(aid),
		RewardID:      int32(rid),
		ReceiveStatus: rst,
	}
	p.player.GetAgent().PushClient(pb.MessageID_S2C_REFRESH_ACTIVITY_RECEIVE_STATUS, msg)
}

func (p *activityComponent) Conform(aid int) bool {
	return p.ConformOpen(aid) && p.ConformTime(aid) && aTypes.IsInArry(int32(aid), p.GetActivityTagList())
}


func getRewardType(stuff string) int {
	isNumbel := func() bool {
		for _, v := range stuff {
			if !unicode.IsNumber(v) {
				return false
			}
		}
		return true
	}

	switch {
	case stuff == "gold":
		return ty_gold_
	case stuff == "jade":
		return ty_jade_
	case stuff == "headFrame":
		return ty_headFrame_
	case stuff == "chatPop":
		return ty_chatPop_
	case stuff == "card":
		return ty_randomCard_
	case stuff == "ticket":
		return ty_ticket_
	case stuff == "eventItem":
		return ty_eventItem_
	case isNumbel():
		return ty_card_
	case strings.HasPrefix(stuff, "BX"):
		return ty_treasure_
	case strings.HasPrefix(stuff, "SK"):
		return ty_skin_
	default:
		return ty_unknow_
	}
}

