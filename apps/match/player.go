package main

import (
	"time"

	"kinger/common/consts"
	"kinger/gamedata"
	"kinger/gopuppy/apps/logic"
	"kinger/gopuppy/common"
	"kinger/proto/pb"
)

type iMatchPlayerData interface {
	GetName() string
	GetCamp() int32
	GetPvpScore() int32
	GetMmr() int32
	GetHandCards() []*pb.SkinGCard
	GetDrawCards() []*pb.SkinGCard
	GetHeadImgUrl() string
	GetHeadFrame() string
	GetWinRate() int32
	GetArea() int32
	GetCountryFlag() string
}

type iMatchPlayer interface {
	packFighterData() *pb.FighterData
	getAgent() *logic.PlayerAgent
	getSeasonDataID() int
	getMmr() int
	getArea() int
	getPvpScore() int
	getPvpLevel() int
	getCamp() int
	getHandCards() []*pb.SkinGCard
}

type matchPlayer struct {
	agent              *logic.PlayerAgent
	camp               int
	pvpLevel           int
	pvpTeam            int
	pvpScore           int
	matchScore         int
	beginMatchTime     time.Time
	mmr                int
	cardStrength       int
	robotID            common.UUid
	name               string
	handCards          []*pb.SkinGCard
	drawCards          []*pb.SkinGCard
	gridCards          []*pb.InGridCard
	headImgUrl         string
	headFrame          string
	countryFlag        string
	seasonDataID       int
	winRate            int
	rebornCnt          int
	area               int
	streakWinCnt       int
	streakLoseCnt      int
	lastOppUid         common.UUid
	rechargeMatchIndex int
}

func (mp *matchPlayer) getCamp() int {
	return mp.camp
}

func (mp *matchPlayer) getHandCards() []*pb.SkinGCard {
	return mp.handCards
}

func (mp *matchPlayer) getRechargeMatchIndex() int {
	return mp.rechargeMatchIndex
}

func (mp *matchPlayer) getStreakWinCnt() int {
	return mp.streakWinCnt
}

func (mp *matchPlayer) getStreakLoseCnt() int {
	return mp.streakLoseCnt
}

func (mp *matchPlayer) getLastOppUid() common.UUid {
	return mp.lastOppUid
}

func (mp *matchPlayer) getEquipAmount() int {
	var amount int
	for _, card := range mp.handCards {
		if card.Equip != "" {
			amount++
		}
	}
	return amount
}

func (mp *matchPlayer) getSeasonDataID() int {
	return mp.seasonDataID
}

func (mp *matchPlayer) getAgent() *logic.PlayerAgent {
	return mp.agent
}

func (mp *matchPlayer) getArea() int {
	return mp.area
}

func (mp *matchPlayer) getWinRate() int {
	return mp.winRate
}

func (mp *matchPlayer) getBeginMatchTime() time.Time {
	return mp.beginMatchTime
}

func (mp *matchPlayer) getRebornCnt() int {
	return mp.rebornCnt
}

func (mp *matchPlayer) getUid() common.UUid {
	return mp.agent.GetUid()
}

func (mp *matchPlayer) getGridCard() []*pb.InGridCard {
	return mp.gridCards
}

func (mp *matchPlayer) setGridCard(gridCards []*pb.InGridCard) {
	mp.gridCards = gridCards
}

func (mp *matchPlayer) packFighterData() *pb.FighterData {
	ft := &pb.FighterData{
		Uid:          uint64(mp.agent.GetUid()),
		ClientID:     uint64(mp.agent.GetClientID()),
		GateID:       mp.agent.GetGateID(),
		HandCards:    mp.handCards,
		DrawCardPool: mp.drawCards,
		Name:         mp.name,
		Camp:         int32(mp.camp),
		PvpScore:     int32(mp.pvpScore),
		IsRobot:      mp.agent.IsRobot(),
		RobotID:      uint64(mp.robotID),
		Mmr:          int32(mp.mmr),
		GridCards:    mp.getGridCard(),
		HeadImgUrl:   mp.headImgUrl,
		HeadFrame:    mp.headFrame,
		WinRate:      int32(mp.winRate),
		Area:         int32(mp.area),
		Region:       mp.agent.GetRegion(),
		CountryFlag:  mp.countryFlag,
	}

	return ft
}

func (mp *matchPlayer) packMsg() *pb.MatchPlayer {
	return &pb.MatchPlayer{
		Uid:  uint64(mp.agent.GetUid()),
		Name: mp.name,
	}
}

func (mp *matchPlayer) getPvpLevel() int {
	return mp.pvpLevel
}

func (mp *matchPlayer) getPvpTeam() int {
	return mp.pvpTeam
}

func (mp *matchPlayer) getPvpScore() int {
	return mp.pvpScore
}

func (mp *matchPlayer) getMatchScore() int {
	return mp.matchScore
}

func (mp *matchPlayer) getMmr() int {
	return mp.mmr
}

func (mp *matchPlayer) setPvpLevel(lv int) {
	mp.pvpLevel = lv
}

func (mp *matchPlayer) isMatchTimout(now time.Time) bool {
	return now.Sub(mp.beginMatchTime) >= matchTimeout
}

func (mp *matchPlayer) getCardStrength() int {
	return mp.cardStrength
}

func newMatchPlayer(agent *logic.PlayerAgent, data iMatchPlayerData, pvpData *pb.BeginMatchArg) *matchPlayer {

	p := &matchPlayer{
		agent:          agent,
		camp:           int(data.GetCamp()),
		beginMatchTime: time.Now(),
		name:           data.GetName(),
		handCards:      data.GetHandCards(),
		drawCards:      data.GetDrawCards(),
		pvpScore:       int(data.GetPvpScore()),
		mmr:            int(data.GetMmr()),
		headImgUrl:     data.GetHeadImgUrl(),
		headFrame:      data.GetHeadFrame(),
		countryFlag:    data.GetCountryFlag(),
		winRate:        int(data.GetWinRate()),
		area:           int(data.GetArea()),
	}

	if pvpData != nil {
		p.cardStrength = int(pvpData.CardStrength)
		p.seasonDataID = int(pvpData.SeasonDataID)
		p.rebornCnt = int(pvpData.RebornCnt)
		p.streakWinCnt = int(pvpData.StreakWinCnt)
		p.streakLoseCnt = int(pvpData.StreakLoseCnt)
		p.lastOppUid = common.UUid(pvpData.LastOppUid)
		p.rechargeMatchIndex = int(pvpData.RechargeMatchIndex)
		p.matchScore = int(pvpData.MatchScore)
	}

	if !agent.IsRobot() {
		rankGameData := gamedata.GetGameData(consts.Rank).(*gamedata.RankGameData)
		p.pvpLevel = rankGameData.GetPvpLevelByStar(p.pvpScore)
		gdata := rankGameData.Ranks[p.pvpLevel]
		p.pvpTeam = gdata.Team
	}
	return p
}

func newMatchPlayerByRobot(robot iMatchRobot, oopPlayer iMatchPlayer) *matchPlayer {

	p := &matchPlayer{
		agent:          logic.NewRobotAgent(),
		camp:           robot.getCamp(),
		beginMatchTime: time.Now(),
		name:           gamedata.RandomRobotName(oopPlayer.getPvpLevel()),
		handCards:      robot.getHandCards(),
		gridCards:      robot.getGridCards(),
		pvpScore:       robot.getPvpScore(),
		pvpLevel:       robot.getPvpLevel(),
		robotID:        robot.getID(),
		mmr:            oopPlayer.getMmr(),
		headImgUrl:     robot.getHeadImgUrl(),
		headFrame:      robot.getHeadFrame(),
	}

	return p
}
