package gamedata

import (
	"encoding/json"
	//"kinger/gopuppy/common/glog"
	"kinger/common/consts"
)

type Treasure struct {
	ID               string `json:"__id__"`
	Team             int    `json:"team"`
	Rare             int    `json:"rare"`
	RewardProb       int    `json:"reward_prob"`
	DailyProb        int    `json:"daily_prob"`
	GoldMin          int    `json:"goldMin"`
	GoldMax          int    `json:"goldMax"`
	RandomCard       int    `json:"randomCard"`
	FakeRandomCard   int    `json:"fakeRandomCard"`
	CardCnt          int    `json:"cardCnt"`
	RewardUnlockTime int    `json:"reward_unlockTime"`
	DailyUnlockStar  int    `json:"daily_unlockstar"`
	Reward           []uint32  `json:"reward"`
	JadeMax int `json:"jadeMax"`
	JadeMin int `json:"jadeMin"`
	CardStar1 int `json:"cardStar1"`
	CardStar2 int `json:"cardStar2"`
	CardStar3 int `json:"cardStar3"`
	CardStar4 int `json:"cardStar4"`
	CardStar5 int `json:"cardStar5"`
	QuestUnlockCnt int `json:"quest_unlockCnt"`
	CardSkins []string `json:"skin"`
	Title int `json:"title"`
	Emojis []int `json:"emojis"`
	HeadFrames []string `json:"headFrame"`
	EventItemCnt int `json:"eventItemCnt"`
	EventProp int `json:"eventProp"`
	BowlderMax int `json:"bowlderMax"`
	BowlderMin int `json:"bowlderMin"`
	RandomReward string `json:"randomReward"`
	Star int `json:"star"`
	RewardTbl string `json:"reward_tbl"`
	Camp int `json:"camp"`
}

func (t *Treasure) GetNewCardNum() int {
	if t.Team <= 1 {
		return 0
	}
	if t.Rare == 3 {
		return 1
	} else if t.Rare == 4 {
		return 2
	} else {
		return 0
	}
}

func (t *Treasure) GetName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(t.Title)
}

type TreasureGameData struct {
	baseGameData
	AllTreasures       []*Treasure
	Treasures          map[string]*Treasure
	TreasuresOfTeam    map[int][]*Treasure
	Team2DailyTreasure map[int][]*Treasure
}

func newTreasureGameData() *TreasureGameData {
	t := &TreasureGameData{}
	t.i = t
	return t
}

func (t *TreasureGameData) name() string {
	return consts.Treasure
}

func (t *TreasureGameData) init(d []byte) error {
	var l []*Treasure

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	//glog.Infof("treasures = %s", l)
	t.Treasures = map[string]*Treasure{}
	t.TreasuresOfTeam = map[int][]*Treasure{}
	t.Team2DailyTreasure = map[int][]*Treasure{}
	t.AllTreasures = l

	for _, t2 := range l {
		t.Treasures[t2.ID] = t2
		l := t.TreasuresOfTeam[t2.Team]
		t.TreasuresOfTeam[t2.Team] = append(l, t2)

		if t2.DailyUnlockStar > 0 {
			dailyTreasure := t.Team2DailyTreasure[t2.Team]
			t.Team2DailyTreasure[t2.Team] = append(dailyTreasure, t2)
		}
	}

	return nil
}

func (t *TreasureGameData) GetTreasureByBXID(tid string) *Treasure {
	if trea, ok := t.Treasures[tid]; ok {
		return trea
	}
	return nil
}