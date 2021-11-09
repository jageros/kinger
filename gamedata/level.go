package gamedata

import (
	"encoding/json"

	//"kinger/gopuppy/common/glog"
	"fmt"
	"kinger/common/consts"
)

type Level struct {
	ID            int      `json:"__id__"`
	Offensive     int      `json:"offensive"`
	CampsLimit    []int    `json:"nation"`
	OwnHand       []uint32 `json:"own_hand"`
	EnemyHand     []uint32 `json:"enemy_hand"`
	OwnSide       [][]int  `json:"own_side"`
	EnemySide     [][]int  `json:"enemy_side"`
	Unlock        int      `json:"unlock"`
	RewardCer     []uint32 `json:"reward_cer"`
	RewardCard    []uint32 `json:"reward_card"`
	RewardGold    int      `json:"reward_gold"`
	SurrenderCon  int      `json:"surrenderCon"`
	EnergyCon     int      `json:"energyCon"`
	RewardExp     int      `json:"reward_exp"`
	RewardWeap    int      `json:"reward_weap"`
	RewardHor     int      `json:"reward_hor"`
	RewardMat     int      `json:"reward_mat"`
	RewardFor     int      `json:"reward_for"`
	RewardMed     int      `json:"reward_med"`
	RewardBan     int      `json:"reward_ban"`
	Chapter       []int    `json:"chapter"`
	GeneralUnlock uint32   `json:"generalUnlock"`
	BattleScale   int      `json:"battleScale"`
	RewardBox     string   `json:"reward_box"`
	RankCondition int      `json:"rankCondition"`
	BattleRes     int      `json:"battleRes"`
	//ChooseAmount int `json:"chooseAmount"`
	Name       int `json:"name"`
	IsRecharge int `json:"isRecharge"`

	IsRechargeUnlock bool
	LevelName        string
	ownGridCards     map[int]uint32
	enemyGridCards   map[int]uint32
}

func (l *Level) init() {
	l.IsRechargeUnlock = l.IsRecharge != 0

	l.ownGridCards = make(map[int]uint32)
	for _, info := range l.OwnSide {
		l.ownGridCards[info[1]] = uint32(info[0])
	}

	l.enemyGridCards = make(map[int]uint32)
	for _, info := range l.EnemySide {
		l.enemyGridCards[info[1]] = uint32(info[0])
	}

	l.LevelName = GetGameData(consts.Text).(*TextGameData).TEXT(l.Name)
}

func (l *Level) GetOwnGridCards() map[int]uint32 {
	return l.ownGridCards
}

func (l *Level) GetEnemyGridCards() map[int]uint32 {
	return l.enemyGridCards
}

func (l *Level) GetChapter() int {
	return l.Chapter[0]
}

func (l *Level) GetBattleScale() int {
	if l.BattleScale <= 0 {
		return consts.BtScale33
	}
	return l.BattleScale
}

func (l *Level) String() string {
	return fmt.Sprintf("[ID=%d, chapter=%d]", l.ID, l.GetChapter())
}

type LevelGameData struct {
	baseGameData
	levelMap        map[int]*Level
	LevelList       []*Level
	maxChapter      int
	maxLevelID      int
	unlockCardLevel map[int]uint32
	ChapterTreasure map[int]string
	Chapter2Levels  map[int][]*Level
}

func newLevelGameData() *LevelGameData {
	g := &LevelGameData{}
	g.i = g
	return g
}

func (ld *LevelGameData) name() string {
	return consts.Level
}

func (ld *LevelGameData) init(d []byte) error {
	ld.LevelList = make([]*Level, 0)

	err := json.Unmarshal(d, &ld.LevelList)
	if err != nil {
		return err
	}

	ld.levelMap = make(map[int]*Level)
	ld.unlockCardLevel = make(map[int]uint32)
	ld.ChapterTreasure = make(map[int]string)
	ld.Chapter2Levels = map[int][]*Level{}

	for _, lv := range ld.LevelList {
		lv.init()
		if lv.GetChapter() > ld.maxChapter {
			ld.maxChapter = lv.GetChapter()
		}
		ld.levelMap[lv.ID] = lv

		if lv.ID > ld.maxLevelID {
			ld.maxLevelID = lv.ID
		}

		if lv.GeneralUnlock > 0 {
			ld.unlockCardLevel[lv.ID] = lv.GeneralUnlock
		}

		chapter := lv.GetChapter()
		if lv.RewardBox != "" {
			ld.ChapterTreasure[chapter] = lv.RewardBox
		}

		lvs := ld.Chapter2Levels[chapter]
		ld.Chapter2Levels[chapter] = append(lvs, lv)
	}
	//glog.Infof("LevelList = %v", ld.levelList)

	return nil
}

func (ld *LevelGameData) GetFirstLevel() *Level {
	return ld.LevelList[0]
}

func (ld *LevelGameData) GetLevelData(levelID int) *Level {
	return ld.levelMap[levelID]
}

func (ld *LevelGameData) GetMaxChapter() int {
	return ld.maxChapter
}

func (ld *LevelGameData) GetMaxLevelID() int {
	return ld.maxLevelID
}

func (ld *LevelGameData) GetUnlockCards(curLevelID int) []uint32 {
	var cardIDs []uint32
	for levelID, unlockCardID := range ld.unlockCardLevel {
		if curLevelID > levelID {
			cardIDs = append(cardIDs, unlockCardID)
		}
	}

	return cardIDs
}

// return curLevel所属章节未通关的关卡
func (ld *LevelGameData) GetChapterUnClearLevels(curLevel int) []*Level {
	var levels []*Level
	data := ld.levelMap[curLevel]
	if data == nil {
		return levels
	}

	if lvs, ok := ld.Chapter2Levels[data.GetChapter()]; ok {
		for _, lv := range lvs {
			if lv.ID >= curLevel {
				levels = append(levels, lv)
			}
		}
		return levels
	} else {
		return levels
	}
}
