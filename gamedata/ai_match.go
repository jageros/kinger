package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type AiMatch struct {
	ID         int   `json:"__id__"`
	Winning    []int `json:"winning"`
	PositiveIQ int   `json:"PositiveIQ"`
	NegativeIQ int   `json:"NegativeIQ"`
	FoolSign   int   `json:"foolSign"`

	maxRate int
	minRate int
}

func (a *AiMatch) init() {
	switch len(a.Winning) {
	case 1:
		a.maxRate = a.Winning[0]
	case 2:
		a.minRate = a.Winning[0]
		a.maxRate = a.Winning[1]
	}
}

type AiMatchGameData struct {
	baseGameData
	ais         []*AiMatch
	FoolWinRate int
}

func newAiMatchGameData() *AiMatchGameData {
	b := &AiMatchGameData{}
	b.i = b
	return b
}

func (bg *AiMatchGameData) name() string {
	return consts.AiMatch
}

func (bg *AiMatchGameData) init(d []byte) error {
	var l []*AiMatch
	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	bg.ais = l
	for _, a := range l {
		a.init()
		if a.FoolSign > 0 {
			bg.FoolWinRate = a.maxRate
		}
	}

	return nil
}

func (bg *AiMatchGameData) binarySearchIQ(winRate, minIdx, maxIdx int) *AiMatch {
	idx := (minIdx + maxIdx) / 2
	ai := bg.ais[idx]
	if winRate >= ai.minRate && winRate <= ai.maxRate {
		return ai
	} else if winRate < ai.minRate {
		if idx < maxIdx {
			return bg.binarySearchIQ(winRate, idx+1, maxIdx)
		}
	} else {
		if idx > minIdx {
			return bg.binarySearchIQ(winRate, minIdx, idx-1)
		}
	}

	return nil
}

func (bg *AiMatchGameData) GetAiIQ(winRate int) *AiMatch {
	return bg.binarySearchIQ(winRate, 0, len(bg.ais)-1)
}
