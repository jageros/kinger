package gamedata

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"encoding/json"
	"kinger/gopuppy/common"
	"kinger/gopuppy/common/glog"
	"kinger/common/consts"
	"kinger/proto/pb"
	"sort"
)

type Card struct {
	GCardID        uint32  `json:"__id__"`
	CardID         uint32  `json:"cardId"`
	Camp           int     `json:"camp"`
	Type           int     `json:"type"`
	UpL            int     `json:"up_l"`
	UpU            int     `json:"up_u"`
	LeftL          int     `json:"left_l"`
	LeftU          int     `json:"left_u"`
	DownL          int     `json:"down_l"`
	DownU          int     `json:"down_u"`
	RightL         int     `json:"right_l"`
	RightU         int     `json:"right_u"`
	Skill          string  `json:"skill"`
	UpValueRate    float32 `json:"up_value_rate"`
	DownValueRate  float32 `json:"down_value_rate"`
	LeftValueRate  float32 `json:"left_value_rate"`
	RightValueRate float32 `json:"right_value_rate"`
	AdjFValue      float32 `json:"adj_f_value"`
	CardValue      float32 `json:"card_value"`
	Campaign       int     `json:"campaign"`
	Level          int     `json:"level"`
	LevelupNum     int     `json:"levelupNum"`
	LevelupGold    int     `json:"levelupGold"`
	LevelupWeap    int     `json:"levelupWeap"`
	LevelupHor     int     `json:"levelupHor"`
	LevelupMat     int     `json:"levelupMat"`
	Energy         float32 `json:"energy"`
	RobotWeight    int     `json:"robotWeight"`
	RobotRate      [][]int `json:"robotRate"`
	CardType       int     `json:"cardType"`
	Rare           int     `json:"rare"`
	Name int `json:"name"`
	CardOrder int `json:"cardOrder"`
	Version int `json:"version"`
	ResetLevel int `json:"resetLevel"`
	Strength int `json:"value"`
	Icon int `json:"icon"`
	Politics float64 `json:"politics"`
	Intelligence float64 `json:"intelligence"`
	Force float64 `json:"force"`
	Command float64 `json:"command"`
	ConsumeBook int `json:"consumeBook"`
	LevelLimit int `json:"levelLimit"`

	Head string
	skillIds []int32
}

func (c *Card) String() string {
	return fmt.Sprintf("[GCardID=%d, Camp=%s, Type=%d, Up=%s, "+
		"Left=%s, Down=%s, Right=%s, skill=%v]", c.GCardID, c.Camp, c.Type,
		c.UpL, c.LeftL, c.DownL, c.RightL, c.skillIds)
}

func (c *Card) IsSpCard() bool {
	return c.Rare >= 99
}

func (c *Card) GetName() string {
	return GetGameData(consts.Text).(*TextGameData).TEXT(c.Name)
}

func (c *Card) GetGCardID() uint32 {
	return c.GCardID
}

func (c *Card) GetCardID() uint32 {
	return c.CardID
}

func (c *Card) GetSkillIds() []int32 {
	return c.skillIds
}

func (c *Card) RandomUp() int {
	return rand.Intn(c.UpU-c.UpL+1) + c.UpL
}

func (c *Card) RandomDown() int {
	return rand.Intn(c.DownU-c.DownL+1) + c.DownL
}

func (c *Card) RandomLeft() int {
	return rand.Intn(c.LeftU-c.LeftL+1) + c.LeftL
}

func (c *Card) RandomRight() int {
	return rand.Intn(c.RightU-c.RightL+1) + c.RightL
}

func (c *Card) GetUpValueRate() float32 {
	return c.UpValueRate
}

func (c *Card) GetDownValueRate() float32 {
	return c.DownValueRate
}

func (c *Card) GetLeftValueRate() float32 {
	return c.LeftValueRate
}

func (c *Card) GetRightValueRate() float32 {
	return c.RightValueRate
}

func (c *Card) GetAdjFValue() float32 {
	return c.AdjFValue
}

func (c *Card) GetCardValue() float32 {
	return c.CardValue
}

func (c *Card) GetCamp() int {
	return c.Camp
}

func (c *Card) PackDiyFightCardInfo() *pb.DiyFightCardInfo {
	return nil
}

func (c *Card) GetLevel() int {
	return c.Level
}

func (c *Card) GetCardType() int {
	return c.CardType
}

func (c *Card) init() {
	if c.Skill != "" {
		sinfo := strings.Split(c.Skill, ";")
		for _, strs := range sinfo {
			s, err := strconv.Atoi(strs)
			if err != nil {
				glog.Errorf("Card init Skill, cardId=%d, skill=%s, err=%s", c.GCardID, c.Skill, err)
				continue
			}
			c.skillIds = append(c.skillIds, int32(s))
		}
	}

	c.Head = strconv.Itoa(int(c.CardID))
}

func (c *Card) IsSystemCard() bool {
	return c.CardID == 0 || c.Campaign != 0 || c.CardOrder == 0
}

func (c *Card) getRobotRateByPvpLevel(level int) int {
	for _, info := range c.RobotRate {
		if info[0] == level {
			return info[1]
		}
	}
	return 0
}

type pvpRobotCardList struct {
	weightTotal int
	pvpLevel    int
	l           []*Card
}

func (cl *pvpRobotCardList) Len() int {
	return len(cl.l)
}

func (cl *pvpRobotCardList) Swap(i, j int) {
	cl.l[i], cl.l[j] = cl.l[j], cl.l[i]
}

func (cl *pvpRobotCardList) Less(i, j int) bool {
	return cl.l[i].getRobotRateByPvpLevel(cl.pvpLevel) > cl.l[j].getRobotRateByPvpLevel(cl.pvpLevel)
}

func (cl *pvpRobotCardList) addCard(c *Card) {
	cl.l = append(cl.l, c)
}

func (cl *pvpRobotCardList) sort() {
	cardAmount := 0
	_set := common.UInt32Set{}
	for _, c := range cl.l {
		if !_set.Contains(c.CardID) {
			cardAmount++
			_set.Add(c.CardID)
		}
		cl.weightTotal += c.getRobotRateByPvpLevel(cl.pvpLevel)
	}

	if cardAmount < 5 {
		glog.Fatalf("pvpRobotCardList %d no enough", cl.pvpLevel)
	}

	sort.Sort(cl)
}

func (cl *pvpRobotCardList) randomHandCards(camp int) []*Card {
	var cards []*Card
	for len(cards) < 5 {
		rw := rand.Intn(cl.weightTotal + 1)
		tw := 0
	L1:
		for _, c := range cl.l {
			if c.Camp != camp && c.Camp != consts.Heroes {
				continue
			}

			for _, c2 := range cards {
				if c2.GetCardID() == c.CardID {
					continue L1
				}
			}

			tw += c.getRobotRateByPvpLevel(cl.pvpLevel)
			if rw < tw {
				cards = append(cards, c)
				break
			}
		}
	}

	glog.Debugf("randomHandCards %s", cards)
	return cards
}

// implement gamedata.IGameData
type PoolGameData struct {
	baseGameData
	cardPoolMap          map[int]map[int][]*Card  // map[camp]map[type][]*Card
	allCardLevelMap      map[uint32]map[int]*Card // map[cardId]map[level]*Card
	allCardMap           map[uint32]*Card         // map[gCardId]*Card
	typeCardMap          map[int]map[int][]*Card  // map[type]map[level][]*Card
	allCardList          []*Card
	pvpRobotCardMap      map[int]*pvpRobotCardList // map[pvpLevel]*pvpRobotCardList
	campaignCardLevelMap map[uint32]map[int]*Card  // map[cardId]map[level]*Card
	level2Cards map[int][]*Card
	maxRobotPvpLevel int
}

func newPoolGameData() *PoolGameData {
	p := &PoolGameData{}
	p.i = p
	return p
}

func (pd *PoolGameData) name() string {
	return consts.Pool
}

func (pd *PoolGameData) init(data []byte) error {
	pd.cardPoolMap = make(map[int]map[int][]*Card)
	pd.allCardLevelMap = make(map[uint32]map[int]*Card)
	pd.allCardList = make([]*Card, 0)
	pd.allCardMap = make(map[uint32]*Card)
	pd.typeCardMap = make(map[int]map[int][]*Card)
	pd.campaignCardLevelMap = make(map[uint32]map[int]*Card)
	pd.maxRobotPvpLevel = 10
	pd.pvpRobotCardMap = map[int]*pvpRobotCardList{
		2: &pvpRobotCardList{pvpLevel: 2},
		3: &pvpRobotCardList{pvpLevel: 3},
		4: &pvpRobotCardList{pvpLevel: 4},
		5: &pvpRobotCardList{pvpLevel: 5},
		6: &pvpRobotCardList{pvpLevel: 6},
		7: &pvpRobotCardList{pvpLevel: 7},
		8: &pvpRobotCardList{pvpLevel: 8},
		9: &pvpRobotCardList{pvpLevel: 9},
		10: &pvpRobotCardList{pvpLevel: 10},
	}

	err := json.Unmarshal(data, &pd.allCardList)
	if err != nil {
		return err
	}

	pd.level2Cards = map[int][]*Card{}
	heros := make(map[int][]*Card)
	for _, card := range pd.allCardList {
		card.init()
		camp := card.Camp

		pd.allCardMap[card.GCardID] = card

		for pvpLevel, _list := range pd.pvpRobotCardMap {
			if card.getRobotRateByPvpLevel(pvpLevel) > 0 {
				_list.addCard(card)
			}
		}

		if card.CardID != 0 && card.Level != 0 && card.Campaign == 1 {
			cardLevelMap, ok := pd.campaignCardLevelMap[card.CardID]
			if !ok {
				cardLevelMap = make(map[int]*Card)
				pd.campaignCardLevelMap[card.CardID] = cardLevelMap
			}
			cardLevelMap[card.Level] = card
		}

		if card.CardID != 0 && card.Campaign != 1 {
			// 玩家可得到的卡
			levelMap, ok := pd.typeCardMap[card.Type]
			if !ok {
				levelMap = make(map[int][]*Card)
				pd.typeCardMap[card.Type] = levelMap
			}
			levelCardList, ok := levelMap[card.Level]
			if !ok {
				levelCardList = make([]*Card, 0)
			}
			levelMap[card.Level] = append(levelCardList, card)

			cardLevelMap, ok := pd.allCardLevelMap[card.CardID]
			if !ok {
				cardLevelMap = make(map[int]*Card)
				pd.allCardLevelMap[card.CardID] = cardLevelMap
			}
			cardLevelMap[card.Level] = card

			var m map[int][]*Card
			if m, ok = pd.cardPoolMap[camp]; !ok {
				m = make(map[int][]*Card)
			}
			var p []*Card
			p, ok = m[card.Type]
			p = append(p, card)
			m[card.Type] = p
			pd.cardPoolMap[camp] = m
		}

		if !card.IsSpCard() && !card.IsSystemCard() {
			cs := pd.level2Cards[card.Level]
			pd.level2Cards[card.Level] = append(cs, card)
		}

		if camp == consts.Heroes {
			var p []*Card
			p, _ = heros[card.Type]
			p = append(p, card)
			heros[card.Type] = p
			continue
		}

	}

	for _, m := range pd.cardPoolMap {
		for t, p := range heros {
			_p, ok := m[t]
			if ok {
				m[t] = append(_p, p...)
			} else {
				m[t] = p
			}
		}
	}

	for _, _list := range pd.pvpRobotCardMap {
		_list.sort()
	}

	//glog.Infof("PoolList = %v", pd.cardPoolMap)

	return nil
}

func (pd *PoolGameData) GetAllCards() []*Card {
	return pd.allCardList
}

func (pd *PoolGameData) GetCardsByLevel(level int) []*Card {
	return pd.level2Cards[level]
}

func (pd *PoolGameData) GetCard(cardID uint32, level int) *Card {
	if cardLevelMap, ok := pd.allCardLevelMap[cardID]; ok {
		if c, ok := cardLevelMap[level]; ok {
			return c
		}
	}
	return nil
}

func (pd *PoolGameData) GetCampaignCard(cardID uint32, level int) *Card {
	if cardLevelMap, ok := pd.campaignCardLevelMap[cardID]; ok {
		if c, ok := cardLevelMap[level]; ok {
			return c
		}
	}
	return nil
}

func (pd *PoolGameData) GetCardByGid(gcardID uint32) *Card {
	return pd.allCardMap[gcardID]
}

func (pd *PoolGameData) RandomCardByType(_type, level int) *Card {
	var cardList []*Card
	if levelMap, ok := pd.typeCardMap[_type]; ok {
		cardList, ok = levelMap[level]
		if !ok {
			return nil
		}
	}

	if len(cardList) <= 0 {
		return nil
	}

	return cardList[rand.Intn(len(cardList))]
}

func (pd *PoolGameData) GetCardPoolMap() map[int]map[int][]*Card {
	return pd.cardPoolMap
}

func (pd *PoolGameData) RandomPvpRobotCards(pvpLevel, camp int, force bool) []*Card {
	if _list, ok := pd.pvpRobotCardMap[pvpLevel]; ok {
		return _list.randomHandCards(camp)
	} else {
		if force {
			return pd.RandomPvpRobotCards(pd.maxRobotPvpLevel, camp, false)
		}
		glog.Errorf("RandomPvpRobotCards no card pvpLevel=%d, camp=%d", pvpLevel, camp)
		return []*Card{}
	}
}
