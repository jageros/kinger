package gamedata

import (
	"encoding/json"

	"kinger/common/consts"
)

type Text struct {
	ID int    `json:"__id__"`
	Cn string `json:"cn"`
	Tw string `json:"tw"`
	En string `json:"en"`
}

type TextGameData struct {
	baseGameData
	textMap map[int]*Text // map[id]*Text
}

func newTextGameData() *TextGameData {
	d := &TextGameData{}
	d.i = d
	return d
}

func (tg *TextGameData) name() string {
	return consts.Text
}

func (tg *TextGameData) init(d []byte) error {
	var _list []*Text
	err := json.Unmarshal(d, &_list)
	if err != nil {
		return err
	}

	tg.textMap = make(map[int]*Text)
	for _, d := range _list {
		tg.textMap[d.ID] = d
	}

	return nil
}

func (tg *TextGameData) TEXT(id int) string {
	if tx, ok := tg.textMap[id]; ok {
		return tx.Cn
	} else {
		return ""
	}
}
