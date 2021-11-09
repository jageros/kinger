package gamedata

import (
	"encoding/json"
	"kinger/common/consts"
)

type WinningRate struct {
	ID                         string  `json:"__id__"`
	RateDisparity              []int   `json:"rate_disparity"`
	ExpectedWinningRateAdv     float64 `json:"expected_winning_rate_adv"`
	ExpectedWinningRateInf     float64 `json:"expected_winning_rate_inf"`
	ExpectedWinningRateAdvKing float64 `json:"expected_winning_rate_adv_king"`
	ExpectedWinningRateInfKing float64 `json:"expected_winning_rate_inf_king"`

	maxDiff int
	minDiff int
}

func (w *WinningRate) init() {
	switch len(w.RateDisparity) {
	case 1:
		w.maxDiff = w.RateDisparity[0]
	case 2:
		w.minDiff = w.RateDisparity[0]
		w.maxDiff = w.RateDisparity[1]
	}
}

type WinningRateGameData struct {
	baseGameData
	WinningRates []*WinningRate
}

func newWinningRateGameData() *WinningRateGameData {
	r := &WinningRateGameData{}
	r.i = r
	return r
}

func (wg *WinningRateGameData) name() string {
	return consts.WinningRate
}

func (wg *WinningRateGameData) init(d []byte) error {
	var l []*WinningRate

	err := json.Unmarshal(d, &l)
	if err != nil {
		return err
	}

	wg.WinningRates = l
	for _, w := range l {
		w.init()
	}
	return nil
}

func (wg *WinningRateGameData) binarySearchWinRate(indexDiff, minIdx, maxIdx int) *WinningRate {
	idx := (minIdx + maxIdx) / 2
	w := wg.WinningRates[idx]
	if indexDiff >= w.minDiff && indexDiff <= w.maxDiff {
		return w
	} else if indexDiff < w.minDiff {
		if idx > minIdx {
			return wg.binarySearchWinRate(indexDiff, minIdx, idx-1)
		}
	} else {
		if idx < maxIdx {
			return wg.binarySearchWinRate(indexDiff, idx+1, maxIdx)
		}
	}

	return nil
}

func (wg *WinningRateGameData) GetExpectedWinningRate(indexDiff int, lvl int) float64 {
	indexDiff2 := indexDiff
	if indexDiff2 < 0 {
		indexDiff2 = -indexDiff2
	}

	w := wg.binarySearchWinRate(indexDiff2, 0, len(wg.WinningRates)-1)
	if w == nil {
		return 0.5
	} else if indexDiff < 0 {
		if lvl >= 9 {
			return w.ExpectedWinningRateInfKing
		}
		return w.ExpectedWinningRateInf
	} else {
		if lvl >= 9 {
			return w.ExpectedWinningRateAdvKing
		}
		return w.ExpectedWinningRateAdv
	}
}
